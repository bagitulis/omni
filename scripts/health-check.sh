#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# OMNI Service Health Check
# Usage: health-check.sh [--target local|vps] [--service NAME] [--json] [--dry-run]
# =============================================================================

TARGET="${TARGET:-local}"
SERVICE="all"
TIMEOUT="${TIMEOUT:-30}"
JSON_MODE=false
DRY_RUN=false
DISK_THRESHOLD="${DISK_USAGE_THRESHOLD:-90}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
TOTAL=0; PASSED=0; FAILED=0; WARNED=0; RESULTS=()

# =============================================================================
# Argument parsing
# =============================================================================
while [[ $# -gt 0 ]]; do
    case "$1" in
        --target)   TARGET="$2";   shift 2 ;;
        --service)  SERVICE="$2";  shift 2 ;;
        --timeout)  TIMEOUT="$2";  shift 2 ;;
        --json)     JSON_MODE=true; shift ;;
        --dry-run)  DRY_RUN=true;  shift ;;
        -h|--help)
            cat << 'HELP'
Usage: health-check.sh [OPTIONS]
  --target local|vps   Target environment (default: local)
  --service NAME       all|backend|frontend|db|nginx|redis|disk
  --timeout SECONDS    Per-service timeout (default: 30)
  --json               Output as JSON for CI/automation
  --dry-run            Print what would be checked and exit
  -h, --help           Show this help message
HELP
            exit 0 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

[[ "$JSON_MODE" == "true" ]] && RED='' && GREEN='' && YELLOW='' && BOLD='' && NC=''

# =============================================================================
# URL / Path helpers — no hardcoded domains
# =============================================================================
backend_url() {
    [[ "$TARGET" == "local" ]] && { echo "http://localhost:8080"; return; }
    local d="${DOMAIN_NAME:-${VPS_HOST:-}}"
    [[ -z "$d" ]] && echo "http://localhost:8080" || echo "https://${d}"
}

frontend_url() {
    [[ "$TARGET" == "local" ]] && { echo "http://localhost:3000"; return; }
    local d="${DOMAIN_NAME:-${VPS_HOST:-}}"
    [[ -z "$d" ]] && echo "http://localhost:3000" || echo "https://${d}"
}

nginx_url() {
    if [[ "$TARGET" == "local" ]]; then
        echo "http://localhost:80/nginx-health"
    else
        local d="${DOMAIN_NAME:-${VPS_HOST:-}}"
        [[ -z "$d" ]] && echo "http://localhost:80/nginx-health" || echo "https://${d}/nginx-health"
    fi
}

disk_path() { [[ "$TARGET" == "vps" ]] && echo "/opt/omni" || echo "/"; }

# =============================================================================
# Result recording
# =============================================================================
record() {
    local name="$1" status="$2" msg="$3"
    TOTAL=$((TOTAL + 1))
    case "$status" in
        healthy)   PASSED=$((PASSED + 1)) ;;
        unhealthy) FAILED=$((FAILED + 1)) ;;
        warning)   WARNED=$((WARNED + 1)) ;;
    esac

    if [[ "$JSON_MODE" == "true" ]]; then
        RESULTS+=("{\"service\":\"${name}\",\"status\":\"${status}\",\"message\":\"${msg}\"}")
        return
    fi

    local icon
    case "$status" in
        healthy)   icon="${GREEN}✓${NC}" ;;
        unhealthy) icon="${RED}✗${NC}" ;;
        warning)   icon="${YELLOW}⚠${NC}" ;;
    esac
    echo -e "  ${icon} ${BOLD}${name}${NC}: ${msg}"
}

# =============================================================================
# Shared HTTP check — backend, frontend, nginx
# =============================================================================
http_check() {
    local name="$1" url="$2" verify_type="$3" verify_pattern="$4"
    local body http_code err_line

    body=$(curl -sS --max-time "$TIMEOUT" -o /dev/stdout -w '\n%{http_code}' "$url" 2>&1) && {
        http_code=$(echo "$body" | tail -1)
        body=$(echo "$body" | sed '$d')

        [[ "$http_code" != "200" ]] && { record "$name" "unhealthy" "HTTP ${http_code}"; return; }

        case "$verify_type" in
            json_field)
                echo "$body" | grep -q "$verify_pattern" \
                    && record "$name" "healthy" "HTTP 200, ${verify_pattern} present" \
                    || record "$name" "warning" "HTTP 200 but unexpected response format" ;;
            html)
                echo "$body" | grep -qi "$verify_pattern" \
                    && record "$name" "healthy" "HTTP 200, HTML detected" \
                    || record "$name" "warning" "HTTP 200 but no HTML detected" ;;
            any) record "$name" "healthy" "HTTP 200" ;;
        esac
        return
    }

    # Connection failed — classify error without leaking secrets
    err_line=$(echo "$body" | head -1)
    if echo "$err_line" | grep -qi "connection refused"; then
        record "$name" "unhealthy" "not running (connection refused)"
    elif echo "$err_line" | grep -qi "could not resolve"; then
        record "$name" "unhealthy" "DNS resolution failure"
    elif echo "$err_line" | grep -qi "timed out"; then
        record "$name" "unhealthy" "timeout after ${TIMEOUT}s"
    else
        record "$name" "unhealthy" "connection failed"
    fi
}

# =============================================================================
# Individual service checks
# =============================================================================
check_backend()  { http_check "backend"  "$(backend_url)/api/health" json_field '"status"'; }
check_frontend() { http_check "frontend" "$(frontend_url)/"         html       '<!doctype html\|<html'; }
check_nginx()    { http_check "nginx"    "$(nginx_url)"             any        ""; }

check_db() {
    local output
    output=$(timeout "$TIMEOUT" pg_isready -h localhost 2>&1) && {
        echo "$output" | grep -q "accepting connections" \
            && record "db" "healthy" "accepting connections" \
            || record "db" "warning" "unexpected pg_isready output"
        return
    }
    command -v pg_isready &>/dev/null \
        && record "db" "unhealthy" "not accepting connections" \
        || record "db" "unhealthy" "pg_isready not found"
}

check_redis() {
    local output
    output=$(timeout "$TIMEOUT" redis-cli ping 2>&1) && {
        [[ "$output" == "PONG" ]] \
            && record "redis" "healthy" "PONG received" \
            || record "redis" "unhealthy" "unexpected response: ${output}"
        return
    }
    command -v redis-cli &>/dev/null \
        && record "redis" "unhealthy" "not running or connection refused" \
        || record "redis" "unhealthy" "redis-cli not found"
}

check_disk() {
    local dp usage
    dp="$(disk_path)"
    [[ ! -e "$dp" ]] && dp="/"

    usage=$(df "$dp" 2>/dev/null | tail -1 | awk '{print $(NF-1)}' | tr -d '%') && {
        if   [[ "$usage" -lt "$DISK_THRESHOLD" ]]; then
            record "disk" "healthy"   "${usage}% used (threshold: ${DISK_THRESHOLD}%)"
        elif [[ "$usage" -lt 95 ]]; then
            record "disk" "warning"   "${usage}% used (threshold: ${DISK_THRESHOLD}%)"
        else
            record "disk" "unhealthy" "${usage}% used — critical (threshold: ${DISK_THRESHOLD}%)"
        fi
        return
    }
    record "disk" "unhealthy" "failed to read disk usage for ${dp}"
}

# =============================================================================
# Dry run — prints what would be checked and exits
# =============================================================================
if [[ "$DRY_RUN" == "true" ]]; then
    echo "Health Check — Dry Run"
    echo "  Target:   ${TARGET}"
    echo "  Service:  ${SERVICE}"
    echo "  Timeout:  ${TIMEOUT}s"
    echo "  Disk threshold: ${DISK_THRESHOLD}%"
    echo ""

    svcs="${SERVICE}"
    [[ "$svcs" == "all" ]] && svcs="backend frontend db nginx redis disk"

    for svc in $svcs; do
        case "$svc" in
            backend)  echo "  backend:  GET $(backend_url)/api/health" ;;
            frontend) echo "  frontend: GET $(frontend_url)/" ;;
            db)       echo "  db:       pg_isready -h localhost" ;;
            nginx)    echo "  nginx:    GET $(nginx_url)" ;;
            redis)    echo "  redis:    redis-cli ping" ;;
            disk)     echo "  disk:     df $(disk_path) (threshold: ${DISK_THRESHOLD}%)" ;;
            *)        echo "  ${svc}:    unknown service" ;;
        esac
    done
    exit 0
fi

# =============================================================================
# Main — run health checks
# =============================================================================
if [[ "$JSON_MODE" != "true" ]]; then
    echo ""
    echo -e "${BOLD}OMNI Health Check${NC}"
    echo -e "  Target: ${TARGET} | Service: ${SERVICE} | Timeout: ${TIMEOUT}s"
    echo ""
fi

svcs="${SERVICE}"
[[ "$svcs" == "all" ]] && svcs="backend frontend db nginx redis disk"

for svc in $svcs; do
    case "$svc" in
        backend)  check_backend ;;
        frontend) check_frontend ;;
        db)       check_db ;;
        nginx)    check_nginx ;;
        redis)    check_redis ;;
        disk)     check_disk ;;
        *) echo "Unknown service: $svc" >&2; exit 1 ;;
    esac
done

# =============================================================================
# Summary output — JSON or human-readable
# =============================================================================
if [[ "$JSON_MODE" == "true" ]]; then
    # Build valid JSON array
    json="["
    for i in "${!RESULTS[@]}"; do
        [[ $i -gt 0 ]] && json+=","
        json+="${RESULTS[$i]}"
    done
    json+="]"
    echo "$json"
else
    echo ""
    echo -e "${BOLD}Summary:${NC} ${PASSED} healthy, ${FAILED} unhealthy, ${WARNED} warning(s) — ${TOTAL} total"
    if   [[ $FAILED -gt 0 ]]; then echo -e "${RED}Some services are unhealthy.${NC}"
    elif [[ $WARNED -gt 0 ]]; then echo -e "${YELLOW}Some services have warnings.${NC}"
    else echo -e "${GREEN}All services healthy.${NC}"
    fi
    echo ""
fi

# Exit 0 if all healthy, exit 1 if any unhealthy
[[ $FAILED -gt 0 ]] && exit 1
exit 0
#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# OMNI Post-Migration Monitoring Script (48h after cutover)
# Usage: monitor-post-migration.sh [--target local|vps] [--hours N] [--json] [--alert]
# =============================================================================

TARGET="${TARGET:-local}"
HOURS="${MONITOR_HOURS:-48}"
JSON_MODE=false
ALERT_MODE=false
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"
CONTAINER="${BACKEND_CONTAINER:-omni-backend}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
TOTAL_ALERTS=0

# =============================================================================
# Argument parsing
# =============================================================================
while [[ $# -gt 0 ]]; do
    case "$1" in
        --target)   TARGET="$2";   shift 2 ;;
        --hours)    HOURS="$2";    shift 2 ;;
        --json)     JSON_MODE=true; shift ;;
        --alert)    ALERT_MODE=true; shift ;;
        --webhook)  ALERT_WEBHOOK="$2"; shift 2 ;;
        -h|--help)
            cat << 'HELP'
Usage: monitor-post-migration.sh [OPTIONS]
  --target local|vps   Target environment (default: local)
  --hours N            Monitor window in hours (default: 48)
  --json               Output as JSON for automation
  --alert              Send alerts on threshold breach
  --webhook URL        Webhook URL for alerts (or set ALERT_WEBHOOK env)
  -h, --help           Show this help message

Monitors:
  1. Token refresh failures per platform (shopee/lazada/tiktok)
  2. Credential load failures
  3. OAuth callback failures
  4. Platform API auth errors (401/403)
  5. Alert threshold: any non-zero count for token refresh failures
HELP
            exit 0 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

[[ "$JSON_MODE" == "true" ]] && RED='' && GREEN='' && YELLOW='' && BOLD='' && NC=''

# =============================================================================
# Helper: Get Docker logs for time window
# =============================================================================
get_logs() {
    local since="${HOURS}h"
    if [[ "$TARGET" == "local" ]]; then
        docker logs "$CONTAINER" --since "$since" 2>&1
    else
        # VPS: assumes SSH access
        ssh "${VPS_USER:-root}@${VPS_HOST:-localhost}" \
            "docker logs $CONTAINER --since $since" 2>&1
    fi
}

# =============================================================================
# Helper: Send alert via webhook
# =============================================================================
send_alert() {
    local severity="$1" message="$2"
    TOTAL_ALERTS=$((TOTAL_ALERTS + 1))
    
    if [[ "$ALERT_MODE" == "true" ]] && [[ -n "$ALERT_WEBHOOK" ]]; then
        local payload
        payload=$(cat <<EOF
{
  "text": "🚨 OMNI Post-Migration Alert [${severity}]: ${message}",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
        )
        curl -sS -X POST -H "Content-Type: application/json" -d "$payload" "$ALERT_WEBHOOK" >/dev/null 2>&1 || true
    fi
}

# =============================================================================
# Counter functions
# =============================================================================
count_token_refresh_failures() {
    local platform="$1"
    local logs="$2"
    
    # Count refresh failures: log.Error messages containing platform refresh
    local count
    count=$(echo "$logs" | grep -c "\[${platform^^} REFRESH\].*\(error\|failed\|Error\)" 2>/dev/null || echo "0")
    
    # Also count generic token refresh failures for this platform
    local generic_count
    generic_count=$(echo "$logs" | grep -c "token.*refresh.*failed.*${platform}" 2>/dev/null || echo "0")
    
    echo $((count + generic_count))
}

count_credential_load_failures() {
    local logs="$1"
    
    # Count credential load failures: failed to get tenant DB, failed to get connection, etc.
    local count
    count=$(echo "$logs" | grep -c "failed to get.*credential\|credential.*not found\|failed to load.*credential" 2>/dev/null || echo "0")
    
    # Count connection lookup failures
    local conn_count
    conn_count=$(echo "$logs" | grep -c "no credential connection found\|failed to get connection" 2>/dev/null || echo "0")
    
    echo $((count + conn_count))
}

count_oauth_callback_failures() {
    local logs="$1"
    
    # Count OAuth callback failures
    local count
    count=$(echo "$logs" | grep -c "callback.*failed\|token_exchange_failed\|invalid_attempt\|app_not_configured" 2>/dev/null || echo "0")
    
    # Count platform-specific callback errors
    local platform_count
    platform_count=$(echo "$logs" | grep -c "Shopee callback failed\|Lazada.*token exchange failed\|TikTok.*token exchange failed" 2>/dev/null || echo "0")
    
    echo $((count + platform_count))
}

count_platform_auth_errors() {
    local logs="$1"
    
    # Count 401/403 errors from platform APIs
    local count_401
    count_401=$(echo "$logs" | grep -c "\"code\":401\|status.*401\|401 Unauthorized" 2>/dev/null || echo "0")
    
    local count_403
    count_403=$(echo "$logs" | grep -c "\"code\":403\|status.*403\|403 Forbidden" 2>/dev/null || echo "0")
    
    # Count platform-specific auth errors
    local auth_errors
    auth_errors=$(echo "$logs" | grep -c "invalid_access_token\|token.*expired\|auth.*failed\|unauthorized" 2>/dev/null || echo "0")
    
    echo $((count_401 + count_403 + auth_errors))
}

# =============================================================================
# Main monitoring logic
# =============================================================================
echo ""
echo -e "${BOLD}OMNI Post-Migration Monitor${NC}"
echo -e "  Target: ${TARGET} | Window: ${HOURS}h | Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo ""

# Get logs
echo "Fetching logs..."
LOGS=$(get_logs)
LOG_LINES=$(echo "$LOGS" | wc -l)
echo "  Retrieved ${LOG_LINES} log lines"
echo ""

# Initialize results
declare -A RESULTS
RESULTS[token_refresh_shopee]=0
RESULTS[token_refresh_lazada]=0
RESULTS[token_refresh_tiktok]=0
RESULTS[credential_load]=0
RESULTS[oauth_callback]=0
RESULTS[platform_auth]=0

# Count failures
echo -e "${BOLD}Scanning for failures...${NC}"
echo ""

# 1. Token refresh failures per platform
echo "  Token Refresh Failures:"
for platform in shopee lazada tiktok; do
    count=$(count_token_refresh_failures "$platform" "$LOGS")
    RESULTS[token_refresh_${platform}]=$count
    
    if [[ "$JSON_MODE" != "true" ]]; then
        if [[ $count -gt 0 ]]; then
            echo -e "    ${RED}✗${NC} ${platform}: ${count} failures"
        else
            echo -e "    ${GREEN}✓${NC} ${platform}: 0 failures"
        fi
    fi
done

# 2. Credential load failures
count=$(count_credential_load_failures "$LOGS")
RESULTS[credential_load]=$count
if [[ "$JSON_MODE" != "true" ]]; then
    echo ""
    echo "  Credential Load Failures:"
    if [[ $count -gt 0 ]]; then
        echo -e "    ${RED}✗${NC} Total: ${count} failures"
    else
        echo -e "    ${GREEN}✓${NC} Total: 0 failures"
    fi
fi

# 3. OAuth callback failures
count=$(count_oauth_callback_failures "$LOGS")
RESULTS[oauth_callback]=$count
if [[ "$JSON_MODE" != "true" ]]; then
    echo ""
    echo "  OAuth Callback Failures:"
    if [[ $count -gt 0 ]]; then
        echo -e "    ${RED}✗${NC} Total: ${count} failures"
    else
        echo -e "    ${GREEN}✓${NC} Total: 0 failures"
    fi
fi

# 4. Platform API auth errors (401/403)
count=$(count_platform_auth_errors "$LOGS")
RESULTS[platform_auth]=$count
if [[ "$JSON_MODE" != "true" ]]; then
    echo ""
    echo "  Platform API Auth Errors (401/403):"
    if [[ $count -gt 0 ]]; then
        echo -e "    ${RED}✗${NC} Total: ${count} errors"
    else
        echo -e "    ${GREEN}✓${NC} Total: 0 errors"
    fi
fi

# =============================================================================
# Alert threshold check
# =============================================================================
echo ""
echo -e "${BOLD}Alert Threshold Check:${NC}"
echo "  Threshold: Any non-zero count for token refresh failures"

ALERT_TRIGGERED=false
for platform in shopee lazada tiktok; do
    count=${RESULTS[token_refresh_${platform}]}
    if [[ $count -gt 0 ]]; then
        ALERT_TRIGGERED=true
        message="Token refresh failure detected for ${platform}: ${count} failures in last ${HOURS}h"
        send_alert "HIGH" "$message"
        echo -e "  ${RED}🚨 ALERT:${NC} ${message}"
    fi
done

if [[ "$ALERT_TRIGGERED" == "false" ]]; then
    echo -e "  ${GREEN}✓${NC} No alerts triggered"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BOLD}Summary:${NC}"

TOTAL_FAILURES=0
for key in "${!RESULTS[@]}"; do
    TOTAL_FAILURES=$((TOTAL_FAILURES + RESULTS[$key]))
done

if [[ $TOTAL_FAILURES -eq 0 ]]; then
    echo -e "  ${GREEN}✓ All systems nominal${NC}"
    echo "  No failures detected in the last ${HOURS} hours"
else
    echo -e "  ${RED}✗ ${TOTAL_FAILURES} total failures detected${NC}"
    echo ""
    echo "  Breakdown:"
    echo "    Token Refresh (Shopee): ${RESULTS[token_refresh_shopee]}"
    echo "    Token Refresh (Lazada): ${RESULTS[token_refresh_lazada]}"
    echo "    Token Refresh (TikTok): ${RESULTS[token_refresh_tiktok]}"
    echo "    Credential Load: ${RESULTS[credential_load]}"
    echo "    OAuth Callback: ${RESULTS[oauth_callback]}"
    echo "    Platform Auth (401/403): ${RESULTS[platform_auth]}"
fi

# =============================================================================
# JSON output
# =============================================================================
if [[ "$JSON_MODE" == "true" ]]; then
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "target": "${TARGET}",
  "window_hours": ${HOURS},
  "metrics": {
    "token_refresh_failures": {
      "shopee": ${RESULTS[token_refresh_shopee]},
      "lazada": ${RESULTS[token_refresh_lazada]},
      "tiktok": ${RESULTS[token_refresh_tiktok]}
    },
    "credential_load_failures": ${RESULTS[credential_load]},
    "oauth_callback_failures": ${RESULTS[oauth_callback]},
    "platform_auth_errors": ${RESULTS[platform_auth]},
    "total_failures": ${TOTAL_FAILURES}
  },
  "alerts_triggered": $([ "$ALERT_TRIGGERED" == "true" ] && echo "true" || echo "false"),
  "alert_count": ${TOTAL_ALERTS}
}
EOF
fi

# =============================================================================
# Exit code
# =============================================================================
if [[ "$ALERT_TRIGGERED" == "true" ]]; then
    exit 1
fi
exit 0

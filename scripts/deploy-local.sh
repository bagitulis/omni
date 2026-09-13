#!/usr/bin/env bash
set -euo pipefail

# OMNI local deployment for development.
# Uses .env.local for local development, never touches production .env.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env.local"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DRY_RUN=false
FORCE=false
BUILD=false
TUNNEL=false
SPEC="standard"

info() { printf "%b\n" "${BLUE}INFO:${NC} $*"; }
success() { printf "%b\n" "${GREEN}SUCCESS:${NC} $*"; }
warn() { printf "%b\n" "${YELLOW}WARNING:${NC} $*"; }
error() { printf "%b\n" "${RED}ERROR:${NC} $*" >&2; }
die() { error "$*"; exit 1; }

usage() {
    cat <<'USAGE'
Usage: scripts/deploy-local.sh [options]

Deploy OMNI locally for development using Docker Compose.

Options:
  --dry-run                 Print compose config without starting, exit 0
  --spec highspec|standard|lowspec
                            Select compose overlay (default: standard)
  --tunnel                  Enable Cloudflare tunnel mode
  --build                   Force rebuild Docker images
  --force                   Stop existing containers before starting
  -h, --help                Show this help

Examples:
  scripts/deploy-local.sh --dry-run
  scripts/deploy-local.sh --spec lowspec --dry-run
  scripts/deploy-local.sh --build
  scripts/deploy-local.sh --tunnel --spec highspec
USAGE
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --spec)
            [[ $# -ge 2 ]] || die "--spec requires highspec, standard, or lowspec"
            SPEC="$2"
            shift 2
            ;;
        --tunnel) TUNNEL=true; shift ;;
        --build) BUILD=true; shift ;;
        --force) FORCE=true; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "Unknown option: $1" ;;
    esac
done

case "$SPEC" in
    highspec|standard|lowspec) ;;
    *) die "Invalid --spec '${SPEC}'. Use highspec, standard, or lowspec." ;;
esac

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-}"

detect_runtime() {
    if [[ -n "$CONTAINER_RUNTIME" ]]; then
        command -v "$CONTAINER_RUNTIME" &>/dev/null \
            || die "CONTAINER_RUNTIME='${CONTAINER_RUNTIME}' is set but not found in PATH."
        return
    fi

    if command -v docker &>/dev/null; then
        CONTAINER_RUNTIME=docker
    elif command -v podman &>/dev/null; then
        CONTAINER_RUNTIME=podman
    else
        die "No container runtime found. Install Docker or Podman, or set CONTAINER_RUNTIME=docker|podman."
    fi
}

run_compose() {
    "$CONTAINER_RUNTIME" compose "$@"
}

check_docker() {
    detect_runtime
    info "Using container runtime: ${CONTAINER_RUNTIME}"

    if ! "$CONTAINER_RUNTIME" info &>/dev/null 2>&1; then
        if [[ "$CONTAINER_RUNTIME" == "podman" ]]; then
            die "Podman is not running. Start it with: podman machine start"
        fi
        die "Docker daemon is not running. Start Docker and try again."
    fi
}

check_ports() {
    local ports=(80 443 5432 6379 8080)
    local conflict=false

    for port in "${ports[@]}"; do
        if command -v ss &>/dev/null; then
            if ss -tlnp 2>/dev/null | grep -qE "[:.]${port}\\s"; then
                error "Port ${port} is already in use"
                conflict=true
            fi
        elif command -v netstat &>/dev/null; then
            if netstat -tlnp 2>/dev/null | grep -qE "[:.]${port}\\s"; then
                error "Port ${port} is already in use"
                conflict=true
            fi
        fi
    done

    if [[ "$conflict" == true ]]; then
        die "Port conflict(s) detected. Stop conflicting services and retry."
    fi
    info "Port check passed (80, 443, 5432, 6379, 8080)"
}

check_existing_containers() {
    local compose_args=(-f docker-compose.tunnel.yml -f "docker-compose.tunnel.${SPEC}.yml")

    if run_compose "${compose_args[@]}" ps -q 2>/dev/null | grep -q .; then
        if [[ "$FORCE" == true ]]; then
            warn "Existing containers found; stopping (--force)"
            run_compose "${compose_args[@]}" down
        else
            die "Existing containers found. Use --force to stop them first."
        fi
    fi
}

# ---------------------------------------------------------------------------
# Environment
# ---------------------------------------------------------------------------
load_env() {
    if [[ ! -f "$ENV_FILE" ]]; then
        if [[ "$DRY_RUN" == true ]]; then
            warn "Missing ${ENV_FILE}; using placeholder values for dry-run"
            export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-dryrun_placeholder}"
            export REDIS_PASSWORD="${REDIS_PASSWORD:-dryrun_placeholder}"
            export ENCRYPTION_KEY="${ENCRYPTION_KEY:-dryrun_placeholder_32chars_long!!}"
            export JWT_SECRET="${JWT_SECRET:-dryrun_placeholder}"
            export JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-dryrun_placeholder}"
            export DOMAIN_NAME="localhost"
            return 0
        fi
        die ".env.local not found. Copy from .env.example: cp .env.example .env.local"
    fi

    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
}

validate_env() {
    info "Validating environment variables"
    local deploy_target=""
    [[ "$TUNNEL" == true ]] && deploy_target="local"
    DEPLOY_TARGET="$deploy_target" DEPLOY_PROFILE=development \
        bash "${SCRIPT_DIR}/validate-env.sh"
}

ensure_localhost() {
    local domain="${DOMAIN_NAME:-localhost}"

    if [[ "$domain" != "localhost" && "$domain" != "127.0.0.1" ]]; then
        warn "DOMAIN_NAME='${domain}' overridden to 'localhost' for local development"
    fi

    export DOMAIN_NAME="localhost"
    export FRONTEND_URL="${FRONTEND_URL:-http://localhost}"
    export CORS_ORIGINS="${CORS_ORIGINS:-http://localhost,http://localhost:3000,http://localhost:5174}"
}

# ---------------------------------------------------------------------------
# Compose helpers
# ---------------------------------------------------------------------------
compose_args() {
    printf '%s\n' "-f" "docker-compose.tunnel.yml" "-f" "docker-compose.tunnel.${SPEC}.yml"
}

print_plan() {
    local tunnel_mode="disabled"
    [[ "$TUNNEL" == true ]] && tunnel_mode="enabled"

    cat <<PLAN
${BLUE}Local Deployment Plan${NC}
  Env file:      ${ENV_FILE}
  Spec:          ${SPEC}
  Compose files: docker-compose.tunnel.yml + docker-compose.tunnel.${SPEC}.yml
  Tunnel:        ${tunnel_mode}
  Force build:   ${BUILD}
  Force restart: ${FORCE}
  Dry run:       ${DRY_RUN}

Steps:
  1. Validate .env.local exists
  2. Check Docker daemon is running
  3. Check port conflicts (80, 443, 5432, 6379, 8080)
  4. Check for existing containers
  5. Validate environment variables
  6. Ensure localhost domain (override production domains)
  7. Start services with the detected container runtime (docker or podman)
  8. Verify health endpoint
PLAN
}

print_compose_config() {
    local args=()
    while IFS= read -r line; do
        args+=("$line")
    done < <(compose_args)

    info "Resolved compose configuration:"
    echo "---"
    run_compose "${args[@]}" config 2>/dev/null || {
        warn "${CONTAINER_RUNTIME} compose config failed; showing file references instead"
        echo "Compose files:"
        echo "  - docker-compose.tunnel.yml"
        echo "  - docker-compose.tunnel.${SPEC}.yml"
    }
}

# ---------------------------------------------------------------------------
# Services
# ---------------------------------------------------------------------------
start_services() {
    local args=()
    while IFS= read -r line; do
        args+=("$line")
    done < <(compose_args)

    local up_args=(-d)
    [[ "$BUILD" == true ]] && up_args+=(--build)

    if [[ "$TUNNEL" == true ]]; then
        info "Starting all services including Cloudflare tunnel"
        run_compose "${args[@]}" up "${up_args[@]}"
    else
        info "Starting local services (without Cloudflare tunnel)"
        run_compose --profile dev-tools "${args[@]}" up "${up_args[@]}" \
            postgres pgbouncer redis backend frontend nginx pgweb
    fi
}

health_check() {
    info "Waiting for services to become healthy..."
    sleep 10

    if curl -fsS http://localhost/api/health &>/dev/null; then
        success "Health check passed"
        info "Frontend: http://localhost/"
        info "API:      http://localhost/api/"
        info "PgWeb:    http://localhost:8081/"
        return 0
    fi

    error "Health check failed. Printing recent logs:"
    local args=()
    while IFS= read -r line; do
        args+=("$line")
    done < <(compose_args)
    run_compose "${args[@]}" logs --tail=60 2>/dev/null || true
    exit 1
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
deploy() {
    print_plan

    if [[ "$DRY_RUN" == true ]]; then
        print_compose_config
        success "Dry run complete; no containers started"
        return 0
    fi

    check_docker
    check_ports
    check_existing_containers
    validate_env
    ensure_localhost
    start_services
    health_check
    success "Local deployment complete"
}

main() {
    load_env
    deploy
}

main "$@"

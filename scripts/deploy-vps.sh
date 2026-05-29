#!/usr/bin/env bash
set -euo pipefail

# OMNI VPS deployment via rsync over SSH.
# Secrets are sourced from .env but never printed.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DRY_RUN=false
SKIP_BUILD=false
RUN_BACKUP=false
ROLLBACK_TS=""
SPEC="${DEPLOY_SPEC:-${DEPLOY_PROFILE:-highspec}}"

info() { printf "%b\n" "${BLUE}INFO:${NC} $*"; }
success() { printf "%b\n" "${GREEN}SUCCESS:${NC} $*"; }
warn() { printf "%b\n" "${YELLOW}WARNING:${NC} $*"; }
error() { printf "%b\n" "${RED}ERROR:${NC} $*" >&2; }
die() { error "$*"; exit 1; }

usage() {
    cat <<'USAGE'
Usage: scripts/deploy-vps.sh [options]

Options:
  --dry-run                 Print planned actions without remote mutation
  --spec highspec|standard|lowspec
                            Select compose overlay (default: DEPLOY_SPEC,
                            then DEPLOY_PROFILE, then highspec)
  --skip-build              Restart using existing images; skip Docker build
  --backup                  Run backup before deploy if backup tooling exists
  --rollback <timestamp>    Restore /opt/omni-backup-<timestamp> to deploy path
  -h, --help                Show this help
USAGE
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --spec)
            [[ $# -ge 2 ]] || die "--spec requires highspec, standard, or lowspec"
            SPEC="$2"
            shift 2
            ;;
        --skip-build) SKIP_BUILD=true; shift ;;
        --backup) RUN_BACKUP=true; shift ;;
        --rollback)
            [[ $# -ge 2 ]] || die "--rollback requires a backup timestamp"
            ROLLBACK_TS="$2"
            shift 2
            ;;
        -h|--help) usage; exit 0 ;;
        *) die "Unknown option: $1" ;;
    esac
done

case "$SPEC" in
    highspec|standard|lowspec) ;;
    production) SPEC="highspec" ;;
    *) die "Invalid --spec '${SPEC}'. Use highspec, standard, or lowspec." ;;
esac

load_env() {
    if [[ ! -f "$ENV_FILE" ]]; then
        if [[ "$DRY_RUN" == true ]]; then
            warn "Missing ${ENV_FILE}; using placeholder dry-run target values"
            VPS_HOST="${VPS_HOST:-vps.example.com}"
            VPS_USER="${VPS_USER:-root}"
            VPS_DEPLOY_PATH="${VPS_DEPLOY_PATH:-/opt/omni}"
            DOMAIN_NAME="${DOMAIN_NAME:-example.com}"
            return 0
        fi
        die "Missing ${ENV_FILE}. Create it from .env.example before deploy."
    fi
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
}

required_env() {
    local name="$1"
    [[ -n "${!name:-}" ]] || die "${name} must be set in environment"
}

ssh_key_option() {
    local key_path="${SSH_KEY_PATH:-${VPS_SSH_KEY:-}}"
    if [[ -n "$key_path" && "$key_path" != __SET_IN_GH_SECRETS__ ]]; then
        printf -- '-i %q' "$key_path"
    fi
}

remote_target() {
    printf '%s@%s' "${VPS_USER}" "${VPS_HOST}"
}

ssh_base_args() {
    local key_path="${SSH_KEY_PATH:-${VPS_SSH_KEY:-}}"
    printf '%s\n' "-o" "StrictHostKeyChecking=yes" "-o" "ConnectTimeout=10" "-p" "${VPS_PORT:-22}"
    if [[ -n "$key_path" && "$key_path" != __SET_IN_GH_SECRETS__ ]]; then
        printf '%s\n' "-i" "$key_path"
    fi
}

run_ssh() {
    local command="$1"
    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN ssh $(remote_target) '${command}'"
        return 0
    fi

    local attempt
    for attempt in 1 2 3; do
        if ssh $(ssh_base_args) "$(remote_target)" "$command"; then
            return 0
        fi
        if [[ "$attempt" -lt 3 ]]; then
            warn "SSH command failed; retrying in 10s (${attempt}/3)"
            sleep 10
        fi
    done
    return 1
}

validate_local_env() {
    info "Validating local environment with scripts/validate-env.sh"
    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN would run: DEPLOY_TARGET=vps DEPLOY_PROFILE=production scripts/validate-env.sh"
        return 0
    fi
    DEPLOY_TARGET=vps DEPLOY_PROFILE=production bash "${SCRIPT_DIR}/validate-env.sh"
}

print_plan() {
    cat <<PLAN
${BLUE}Deployment plan${NC}
  Target:        $(remote_target)
  Deploy path:   ${VPS_DEPLOY_PATH}
  Spec:          ${SPEC}
  Compose files: docker-compose.tunnel.yml + docker-compose.tunnel.${SPEC}.yml
  Skip build:    ${SKIP_BUILD}
  Backup:        ${RUN_BACKUP}
  Dry run:       ${DRY_RUN}

Steps:
  1. Validate .env locally with scripts/validate-env.sh
  2. Test SSH connectivity with 3 retries
  3. Check remote disk has at least 5GB available
  4. Confirm remote .env exists; do not create it automatically
  5. Warn on deploy lock or dirty VPS git working tree when present
  6. Create pre-deploy copy: ${VPS_DEPLOY_PATH}-backup-\$(date +%s)
  7. Rsync repository with progress/compression; exclude .git/, node_modules/, .env, *.log, backups/
  8. docker compose down when containers exist
  9. docker compose up -d $([[ "$SKIP_BUILD" == true ]] || printf '%s' '--build')
  10. Verify http://localhost/api/health on VPS and print URLs
PLAN
}

test_ssh_connectivity() {
    info "Testing SSH connectivity"
    run_ssh "echo connected >/dev/null"
}

check_disk_space() {
    info "Checking remote disk space (>= 5GB required)"
    run_ssh "available=\$(df -Pk '${VPS_DEPLOY_PATH}' 2>/dev/null | awk 'NR==2 {print \$4}' || df -Pk / | awk 'NR==2 {print \$4}'); if [ \"\${available:-0}\" -lt 5242880 ]; then echo 'Remote disk space below 5GB free' >&2; exit 1; fi"
}

check_remote_env() {
    info "Checking remote .env exists"
    run_ssh "test -f '${VPS_DEPLOY_PATH}/.env' || { echo 'create .env on VPS first: ${VPS_DEPLOY_PATH}/.env' >&2; exit 1; }"
}

check_lock_and_dirty_tree() {
    info "Checking remote deploy lock and working tree"
    run_ssh "if [ -f '${VPS_DEPLOY_PATH}/.deploy.lock' ]; then echo 'ERROR: deploy lock exists at ${VPS_DEPLOY_PATH}/.deploy.lock — remove it to proceed' >&2; exit 1; fi; if [ -d '${VPS_DEPLOY_PATH}/.git' ] && command -v git >/dev/null 2>&1; then cd '${VPS_DEPLOY_PATH}' && if [ -n \"\$(git status --porcelain 2>/dev/null)\" ]; then echo 'WARNING: remote git working tree has local changes; rsync deploy will overwrite files'; fi; fi"
}

list_remote_backups() {
    info "Available remote backups:"
    run_ssh "ls -d '${VPS_DEPLOY_PATH}-backup-'* 2>/dev/null || true"
}

create_predeploy_backup() {
    info "Creating pre-deploy backup copy"
    run_ssh "if [ -d '${VPS_DEPLOY_PATH}' ]; then cp -r '${VPS_DEPLOY_PATH}' '${VPS_DEPLOY_PATH}-backup-'\$(date +%s); else echo 'First deploy: ${VPS_DEPLOY_PATH} does not exist yet, skipping backup copy'; fi"
}

run_backup_tooling() {
    [[ "$RUN_BACKUP" == true ]] || return 0
    info "Running application backup before deploy"
    run_ssh "cd '${VPS_DEPLOY_PATH}' && if [ -x scripts/backup.sh ]; then scripts/backup.sh; elif [ -f build.py ]; then python build.py backup; else echo 'No backup tooling found on VPS' >&2; exit 1; fi"
}

rsync_repo() {
    info "Transferring repository to VPS with rsync"
    local dry_args=()
    [[ "$DRY_RUN" == true ]] && dry_args+=(--dry-run)
    rsync -az --progress \
        "${dry_args[@]}" \
        --exclude='.git/' \
        --exclude='node_modules/' \
        --exclude='.env' \
        --exclude='*.log' \
        --exclude='backups/' \
        -e "ssh -o StrictHostKeyChecking=yes -o ConnectTimeout=10 -p ${VPS_PORT:-22} $(ssh_key_option)" \
        "${ROOT_DIR}/" "$(remote_target):${VPS_DEPLOY_PATH}/"
}

remote_deploy() {
    local compose="docker compose -f docker-compose.tunnel.yml -f docker-compose.tunnel.${SPEC}.yml"
    local up_command="${compose} up -d"
    [[ "$SKIP_BUILD" == true ]] || up_command="${up_command} --build"

    info "Restarting remote Docker Compose stack"
    run_ssh "cd '${VPS_DEPLOY_PATH}' && if ${compose} ps -q 2>/dev/null | grep -q .; then ${compose} down; else echo 'First deploy: no existing containers to stop'; fi && ${up_command}"
}

health_check() {
    info "Waiting for services before health check"
    run_ssh "sleep 10"
    info "Checking remote health endpoint"
    if run_ssh "curl -fsS http://localhost/api/health >/dev/null"; then
        success "Remote health check passed"
        return 0
    fi

    error "Remote health check failed. Printing Docker logs."
    run_ssh "cd '${VPS_DEPLOY_PATH}' && docker compose -f docker-compose.tunnel.yml -f docker-compose.tunnel.${SPEC}.yml logs --tail=120" || true
    error "Suggested rollback: scripts/deploy-vps.sh --rollback <timestamp>"
    list_remote_backups || true
    exit 1
}

rollback() {
    [[ "$ROLLBACK_TS" =~ ^[A-Za-z0-9._-]+$ ]] || die "Invalid rollback timestamp '${ROLLBACK_TS}'"
    print_plan
    info "Rollback requested for ${VPS_DEPLOY_PATH}-backup-${ROLLBACK_TS}"
    list_remote_backups
    run_ssh "test -d '${VPS_DEPLOY_PATH}-backup-${ROLLBACK_TS}' || { echo 'Backup not found: ${VPS_DEPLOY_PATH}-backup-${ROLLBACK_TS}' >&2; exit 1; }"
    run_ssh "rsync -av '${VPS_DEPLOY_PATH}-backup-${ROLLBACK_TS}/' '${VPS_DEPLOY_PATH}/' && cd '${VPS_DEPLOY_PATH}' && docker compose -f docker-compose.tunnel.yml -f docker-compose.tunnel.${SPEC}.yml up -d"
    success "Rollback command completed"
}

deploy() {
    print_plan
    if [[ "$DRY_RUN" == true ]]; then
        success "Dry run complete; no remote mutation performed"
        return 0
    fi
    validate_local_env
    test_ssh_connectivity
    check_disk_space
    check_remote_env
    check_lock_and_dirty_tree
    create_predeploy_backup
    run_backup_tooling
    rsync_repo
    remote_deploy
    health_check
    success "Deployment complete"
    info "Frontend URL: https://${DOMAIN_NAME:-<domain-name-not-set>}"
    info "Health URL:   https://${DOMAIN_NAME:-<domain-name-not-set>}/api/health"
}

main() {
    load_env
    if [[ "$DRY_RUN" == true ]]; then
        VPS_HOST="${VPS_HOST:-vps.example.com}"
        DOMAIN_NAME="${DOMAIN_NAME:-example.com}"
    fi
    VPS_USER="${VPS_USER:-root}"
    VPS_DEPLOY_PATH="${VPS_DEPLOY_PATH:-/opt/omni}"
    required_env VPS_HOST
    required_env VPS_USER
    required_env VPS_DEPLOY_PATH

    if [[ -n "$ROLLBACK_TS" ]]; then
        rollback
    else
        deploy
    fi
}

main "$@"

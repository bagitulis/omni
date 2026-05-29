#!/usr/bin/env bash
set -euo pipefail

# OMNI Backup — pg_dump (PostgreSQL) + Redis + Uploads + Config with retention.


SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TARGET="${TARGET:-local}"
OUTPUT_DIR="${BACKUP_PATH:-${ROOT_DIR}/backups}"
DRY_RUN=false
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"
BACKUP_TYPE="full"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

info()    { printf "%b\n" "${BLUE}INFO:${NC} $*"; }
success() { printf "%b\n" "${GREEN}SUCCESS:${NC} $*"; }
warn()    { printf "%b\n" "${YELLOW}WARNING:${NC} $*"; }
error()   { printf "%b\n" "${RED}ERROR:${NC} $*" >&2; }
die()     { error "$*"; exit 1; }
usage() {
    cat <<'USAGE'
Usage: scripts/backup.sh [options]


Options:
  --target local|vps         Deployment target (default: local)
  --output /path/to/backups  Output directory (default: BACKUP_PATH or ./backups)
  --dry-run                  Print steps without executing
  --retention 7              Days to keep (default: BACKUP_RETENTION_DAYS or 7)
  --type full|db-only|volumes-only
  --type full|db-only|volumes-only  Backup scope (default: full)
Examples:
  scripts/backup.sh --dry-run
  scripts/backup.sh --type db-only
  scripts/backup.sh --target vps --retention 14
USAGE

while [[ $# -gt 0 ]]; do
    case "$1" in
        --target)
            [[ $# -ge 2 ]] || die "--target requires local or vps"
            TARGET="$2"; shift 2 ;;
        --output)
            [[ $# -ge 2 ]] || die "--output requires a path"
            OUTPUT_DIR="$2"; shift 2 ;;
        --dry-run) DRY_RUN=true; shift ;;
        --retention)
            [[ $# -ge 2 ]] || die "--retention requires a number of days"
            RETENTION_DAYS="$2"; shift 2 ;;
        --type)
            [[ $# -ge 2 ]] || die "--type requires full, db-only, or volumes-only"
            BACKUP_TYPE="$2"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) die "Unknown option: $1" ;;
    esac
done

[[ "$BACKUP_TYPE" =~ ^(full|db-only|volumes-only)$ ]] || \
    die "Invalid --type: $BACKUP_TYPE (must be full, db-only, or volumes-only)"

# Pre-flight: disk space check
check_disk_space() {
    local dir="$1"
    local avail_kb free_pct
    avail_kb=$(df -Pk "$dir" 2>/dev/null | awk 'NR==2{print $4}')
    if [[ -z "$avail_kb" ]]; then
        warn "Cannot determine disk space for $dir"
        return 0
    fi
    free_pct=$(awk "BEGIN{printf \"%d\", ($avail_kb/($(df -Pk "$dir" | awk 'NR==2{print $3}') + avail_kb)) * 100}" 2>/dev/null || echo 100)
    if [[ "$free_pct" -lt 20 ]]; then
        die "Insufficient disk space: ${free_pct}% free on $dir (requires >= 20%)"
    fi
    info "Disk space: ${free_pct}% free on $dir"
}

# Lock file: /tmp/omni-backup.lock (VPS) or OUTPUT_DIR/.lock (local)
LOCK_FILE="$([[ "$TARGET" == "vps" ]] && echo "/tmp/omni-backup.lock" || echo "${OUTPUT_DIR}/.lock")"

acquire_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local pid
        pid=$(< "$LOCK_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            die "Backup already running (PID $pid). Lock: $LOCK_FILE"
        fi
        warn "Stale lock found (PID $pid not running), removing"
        rm -f "$LOCK_FILE"
    fi
    echo $$ > "$LOCK_FILE"
}

release_lock() { rm -f "${LOCK_FILE:-}"; }
trap release_lock EXIT

# Docker helpers
PG_CONTAINER="omni-postgres"
REDIS_CONTAINER="omni-redis"

docker_exec() {
    local container="$1"; shift
    docker exec "$container" "$@"
}

container_running() {
    docker ps --filter "name=$1" --filter "status=running" --format '{{.Names}}' 2>/dev/null | grep -q "^${1}$"
}

# Backup functions
backup_postgres() {
    local dest="$1/omni_${TIMESTAMP}.dump"
    if $DRY_RUN; then
        info "[DRY-RUN] pg_dump -U <user> -d <db> -Fc → $dest"
        return 0
    fi
    if ! container_running "$PG_CONTAINER"; then
        warn "PostgreSQL container ($PG_CONTAINER) not running, skipping DB backup"
        return 1
    fi
    info "Dumping PostgreSQL (custom format)..."
    docker_exec "$PG_CONTAINER" pg_dump \
        -U "${POSTGRES_USER:-omni}" \
        -d "${POSTGRES_DB:-omni_main}" \
        -Fc --no-owner --no-privileges \
        -f "/tmp/omni_backup.dump"
    docker cp "${PG_CONTAINER}:/tmp/omni_backup.dump" "$dest"
    docker_exec "$PG_CONTAINER" rm -f /tmp/omni_backup.dump
    local checksum
    checksum=$(sha256sum "$dest" | awk '{print $1}')
    success "PostgreSQL: $dest ($(du -h "$dest" | awk '{print $1}'), sha256:${checksum:0:12})"
}

backup_redis() {
    local dest="$1/redis_${TIMESTAMP}.rdb"
    if $DRY_RUN; then
        info "[DRY-RUN] redis-cli BGSAVE + copy dump.rdb → $dest"
        return 0
    fi
    if ! container_running "$REDIS_CONTAINER"; then
        warn "Redis container ($REDIS_CONTAINER) not running, skipping Redis backup"
        return 1
    fi
    info "Saving Redis RDB (BGSAVE)..."
    local redis_args=()
    [[ -n "${REDIS_PASSWORD:-}" ]] && redis_args+=("-a" "$REDIS_PASSWORD" "--no-auth-warning")
    docker_exec "$REDIS_CONTAINER" redis-cli "${redis_args[@]}" BGSAVE
    sleep 2
    docker cp "${REDIS_CONTAINER}:/data/dump.rdb" "$dest"
    local checksum
    checksum=$(sha256sum "$dest" | awk '{print $1}')
    success "Redis: $dest ($(du -h "$dest" | awk '{print $1}'), sha256:${checksum:0:12})"
}

backup_uploads() {
    local uploads_dir="${ROOT_DIR}/backend/uploads"
    local dest="$1/uploads_${TIMESTAMP}.tar.gz"
    if $DRY_RUN; then
        info "[DRY-RUN] tar.gz uploads/ → $dest"
        return 0
    fi
    if [[ ! -d "$uploads_dir" ]]; then
        warn "Uploads directory not found: $uploads_dir, skipping"
        return 1
    fi
    info "Archiving uploads directory..."
    tar -czf "$dest" \
        --exclude='.git' \
        -C "${ROOT_DIR}/backend" uploads
    local checksum
    checksum=$(sha256sum "$dest" | awk '{print $1}')
    success "Uploads: $dest ($(du -h "$dest" | awk '{print $1}'), sha256:${checksum:0:12})"
}

backup_config() {
    local dest="$1/config_${TIMESTAMP}.tar.gz"
    if $DRY_RUN; then
        info "[DRY-RUN] tar.gz .env + nginx configs → $dest"
        return 0
    fi
    info "Archiving config files..."
    local tmp_dir
    tmp_dir=$(mktemp -d)
    local files_added=0
    if [[ -f "${ROOT_DIR}/.env" ]]; then
        cp "${ROOT_DIR}/.env" "$tmp_dir/env"
        files_added=$((files_added + 1))
    fi
    if [[ -d "${ROOT_DIR}/nginx" ]]; then
        tar -cf - -C "${ROOT_DIR}" --exclude='.git' nginx | tar -xf - -C "$tmp_dir"
        files_added=$((files_added + 1))
    fi
    if [[ "$files_added" -eq 0 ]]; then
        warn "No config files found (.env, nginx/), skipping"
        rm -rf "$tmp_dir"
        return 1
    fi
    tar -czf "$dest" -C "$tmp_dir" .
    rm -rf "$tmp_dir"
    local checksum
    checksum=$(sha256sum "$dest" | awk '{print $1}')
    success "Config: $dest ($(du -h "$dest" | awk '{print $1}'), sha256:${checksum:0:12})"
}

# Manifest — JSON with timestamps, sizes, checksums (no secret values)
write_manifest() {
    local backup_dir="$1"; shift
    local manifest="$backup_dir/manifest.json"
    if $DRY_RUN; then
        info "[DRY-RUN] Write manifest JSON with timestamps, sizes, checksums"
        return 0
    fi
    info "Writing backup manifest..."
    local files_json=""
    for f in "$@"; do
        [[ -f "$f" ]] || continue
        local fname fsize fchecksum
        fname=$(basename "$f")
        fsize=$(stat -c%s "$f" 2>/dev/null || stat -f%z "$f" 2>/dev/null || echo 0)
        fchecksum=$(sha256sum "$f" 2>/dev/null | awk '{print $1}' || echo "unavailable")
        [[ -n "$files_json" ]] && files_json="${files_json},"
        files_json="${files_json}{\"file\":\"${fname}\",\"size_bytes\":${fsize},\"sha256\":\"${fchecksum}\"}"
    done
    cat > "$manifest" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "target": "${TARGET}",
  "type": "${BACKUP_TYPE}",
  "retention_days": ${RETENTION_DAYS},
  "files": [${files_json}]
}
EOF
    success "Manifest: $manifest"
}

# Retention cleanup — always keep at least 1 backup
cleanup_old_backups() {
    if $DRY_RUN; then
        info "[DRY-RUN] Delete backups older than ${RETENTION_DAYS} days (keep >= 1)"
        return 0
    fi
    info "Applying retention policy: keep ${RETENTION_DAYS} days, minimum 1 backup"
    local candidates
    candidates=$(find "$OUTPUT_DIR" -maxdepth 1 -type d -name "backup-*" -mtime "+${RETENTION_DAYS}" 2>/dev/null || true)
    if [[ -z "$candidates" ]]; then
        info "No backups older than ${RETENTION_DAYS} days"
        return 0
    fi
    local total
    total=$(find "$OUTPUT_DIR" -maxdepth 1 -type d -name "backup-*" | wc -l)
    local removed=0
    while IFS= read -r dir; do
        [[ -z "$dir" ]] && continue
        if [[ $((total - removed)) -le 1 ]]; then
            warn "Keeping $dir (minimum 1 backup retention)"
            break
        fi
        info "Removing old backup: $(basename "$dir")"
        rm -rf "$dir"
        removed=$((removed + 1))
    done <<< "$candidates"
    [[ $removed -gt 0 ]] && success "Retention: removed $removed old backup(s)"
}

# Main
main() {
    info "OMNI Backup — target=$TARGET type=$BACKUP_TYPE retention=${RETENTION_DAYS}d"
    mkdir -p "$OUTPUT_DIR"
    check_disk_space "$OUTPUT_DIR"
    acquire_lock

    local backup_dir="${OUTPUT_DIR}/backup-${TIMESTAMP}"
    $DRY_RUN || mkdir -p "$backup_dir"

    local created_files=()

    case "$BACKUP_TYPE" in
        full)
            backup_postgres "$backup_dir" && created_files+=("${backup_dir}/omni_${TIMESTAMP}.dump")
            backup_redis "$backup_dir"    && created_files+=("${backup_dir}/redis_${TIMESTAMP}.rdb")
            backup_uploads "$backup_dir"  && created_files+=("${backup_dir}/uploads_${TIMESTAMP}.tar.gz")
            backup_config "$backup_dir"   && created_files+=("${backup_dir}/config_${TIMESTAMP}.tar.gz")
            ;;
        db-only)
            backup_postgres "$backup_dir" && created_files+=("${backup_dir}/omni_${TIMESTAMP}.dump")
            backup_redis "$backup_dir"    && created_files+=("${backup_dir}/redis_${TIMESTAMP}.rdb")
            ;;
        volumes-only)
            backup_uploads "$backup_dir"  && created_files+=("${backup_dir}/uploads_${TIMESTAMP}.tar.gz")
            backup_config "$backup_dir"   && created_files+=("${backup_dir}/config_${TIMESTAMP}.tar.gz")
            ;;
    esac

    write_manifest "$backup_dir" "${created_files[@]}"
    cleanup_old_backups

    $DRY_RUN && { success "Dry-run complete — no files created"; exit 0; }

    if [[ ${#created_files[@]} -eq 0 ]]; then
        warn "Backup completed with warnings — no files created (services may be down)"
    fi
    success "Backup complete: $backup_dir ($(du -sh "$backup_dir" | awk '{print $1}') total)"
}

main

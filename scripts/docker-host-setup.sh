#!/usr/bin/env bash
set -euo pipefail

# OMNI Docker host setup — one-time VPS preparation for Ubuntu 24.04.
# Installs Docker CE from official apt repo, configures daemon logging,
# logrotate, UFW firewall, and creates the /opt/omni deploy directory.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_PATH="/opt/omni"
DAEMON_JSON="/etc/docker/daemon.json"
LOGROTATE_CONF="/etc/logrotate.d/docker-containers"
MIN_DISK_WARN_KB=5242880   # 5 GB in KB
MIN_DISK_ABORT_KB=1048576  # 1 GB in KB

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DRY_RUN=false

info() { printf "%b\n" "${BLUE}INFO:${NC} $*"; }
success() { printf "%b\n" "${GREEN}SUCCESS:${NC} $*"; }
warn() { printf "%b\n" "${YELLOW}WARNING:${NC} $*"; }
error() { printf "%b\n" "${RED}ERROR:${NC} $*" >&2; }
die() { error "$*"; exit 1; }

usage() {
    cat <<'USAGE'
Usage: scripts/docker-host-setup.sh [options]

One-time Docker host setup for Ubuntu 24.04 VPS.

Options:
  --dry-run     Print planned actions without executing
  -h, --help    Show this help

Actions performed:
  1. Detect OS (Ubuntu/Debian) and version
  2. Install Docker CE from official apt repository
  3. Install Docker Compose v2 plugin (bundled with docker-ce)
  4. Configure Docker daemon log rotation (json-file, 10m, 3 files)
  5. Add current user to docker group
  6. Create /opt/omni deploy directory
  7. Configure logrotate for Docker container logs
  8. Configure UFW firewall (allow SSH 22, HTTP 80)
USAGE
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "Unknown option: $1" ;;
    esac
done

run_cmd() {
    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN: $*"
        return 0
    fi
    "$@"
}

check_root() {
    info "Checking root privileges"
    if [[ "$EUID" -ne 0 ]]; then
        warn "This script should be run as root (current EUID=$EUID)"
        warn "Docker installation and daemon configuration require root privileges"
        if [[ "$DRY_RUN" == true ]]; then
            warn "DRY RUN continuing without root — some steps would fail in real execution"
            return 0
        fi
        die "Re-run with sudo: sudo $0"
    fi
    success "Running as root"
}

detect_os() {
    info "Detecting operating system"

    local os_id=""
    local os_version=""
    local os_description=""

    if command -v lsb_release >/dev/null 2>&1; then
        os_id="$(lsb_release -is)"
        os_version="$(lsb_release -rs)"
        os_description="$(lsb_release -d | cut -f2-)"
    elif [[ -f /etc/os-release ]]; then
        os_id="$(. /etc/os-release && echo "${ID}")"
        os_version="$(. /etc/os-release && echo "${VERSION_ID}")"
        os_description="$(. /etc/os-release && echo "${PRETTY_NAME}")"
    else
        die "Cannot detect OS. No lsb_release or /etc/os-release found"
    fi

    info "Detected: ${os_description} (${os_id} ${os_version})"

    case "${os_id,,}" in
        ubuntu)
            if [[ "$os_version" != "24.04" ]]; then
                warn "Ubuntu ${os_version} detected — script tested on 24.04"
                warn "Docker apt repo URL may differ for your version"
            fi
            success "Ubuntu ${os_version} — supported"
            ;;
        debian)
            success "Debian ${os_version} — supported"
            ;;
        *)
            error "Unsupported OS: ${os_description}"
            error ""
            error "This script supports Ubuntu and Debian only."
            error "For other distributions, install Docker manually:"
            error "  https://docs.docker.com/engine/install/"
            error ""
            error "After Docker is installed, re-run with --skip-docker-install"
            error "to configure daemon, logrotate, and firewall only."
            exit 1
            ;;
    esac

    OS_ID="${os_id,,}"
    OS_VERSION="${os_version}"
}

check_disk_space() {
    info "Checking disk space for ${DEPLOY_PATH}"

    local available_kb
    available_kb="$(df -Pk / 2>/dev/null | awk 'NR==2 {print $4}')"

    if [[ -z "${available_kb:-}" ]]; then
        warn "Could not determine available disk space"
        return 0
    fi

    local available_gb
    available_gb=$(( available_kb / 1048576 ))

    if [[ "$available_kb" -lt "$MIN_DISK_ABORT_KB" ]]; then
        die "Insufficient disk space: ${available_gb}GB free (minimum 1GB required)"
    fi

    if [[ "$available_kb" -lt "$MIN_DISK_WARN_KB" ]]; then
        warn "Low disk space: ${available_gb}GB free (recommended 5GB+)"
        return 0
    fi

    success "Disk space: ${available_gb}GB free"
}

install_docker() {
    if command -v docker >/dev/null 2>&1; then
        local version
        version="$(docker --version 2>/dev/null || echo 'unknown')"
        info "Docker already installed: ${version}"
        success "Skipping Docker installation"
        return 0
    fi

    info "Installing Docker CE from official apt repository"

    run_cmd apt-get update
    run_cmd apt-get install -y ca-certificates curl gnupg

    info "Adding Docker GPG key"
    run_cmd bash -c "install -m 0755 -d /etc/apt/keyrings"
    run_cmd bash -c "curl -fsSL https://download.docker.com/linux/${OS_ID}/gpg -o /etc/apt/keyrings/docker.asc"
    run_cmd chmod a+r /etc/apt/keyrings/docker.asc

    info "Adding Docker apt repository"
    local codename
    if [[ "$OS_ID" == "ubuntu" ]]; then
        codename="$(lsb_release -cs 2>/dev/null || echo 'noble')"
    else
        codename="$(. /etc/os-release && echo "${VERSION_CODENAME}")"
    fi

    run_cmd bash -c "echo 'deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/${OS_ID} ${codename} stable' > /etc/apt/sources.list.d/docker.list"

    run_cmd apt-get update
    run_cmd apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    success "Docker CE installed with Compose v2 plugin"
}

configure_docker_daemon() {
    info "Configuring Docker daemon (${DAEMON_JSON})"

    local daemon_content
    daemon_content=$(cat <<'JSON'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
JSON
)

    if [[ -f "$DAEMON_JSON" ]]; then
        local existing
        existing="$(cat "$DAEMON_JSON" 2>/dev/null || echo '{}')"
        if echo "$existing" | grep -q '"log-driver".*"json-file"' && \
           echo "$existing" | grep -q '"max-size".*"10m"' && \
           echo "$existing" | grep -q '"max-file".*"3"'; then
            info "Docker daemon already configured with json-file log rotation"
            return 0
        fi
        warn "Existing ${DAEMON_JSON} found — will be overwritten"
        if [[ "$DRY_RUN" == false ]]; then
            cp "$DAEMON_JSON" "${DAEMON_JSON}.bak.$(date +%s)"
            info "Backup created: ${DAEMON_JSON}.bak.*"
        fi
    fi

    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN: Would write ${DAEMON_JSON}:"
        echo "$daemon_content"
    else
        mkdir -p /etc/docker
        echo "$daemon_content" > "$DAEMON_JSON"
    fi

    success "Docker daemon configured: json-file log rotation (10m, 3 files)"

    info "Restarting Docker daemon to apply configuration"
    if [[ "$DRY_RUN" == false ]]; then
        if systemctl is-active --quiet docker; then
            run_cmd systemctl restart docker
        else
            run_cmd systemctl start docker
            run_cmd systemctl enable docker
        fi
    else
        info "DRY RUN: Would restart Docker daemon"
    fi
}

configure_docker_group() {
    info "Configuring docker group for current user"

    local current_user="${SUDO_USER:-$USER}"

    if id -nG "$current_user" 2>/dev/null | grep -qw docker; then
        info "User ${current_user} already in docker group"
        return 0
    fi

    run_cmd usermod -aG docker "$current_user"
    success "Added ${current_user} to docker group"

    if [[ "$DRY_RUN" == false ]]; then
        warn "Log out and back in for docker group to take effect"
        warn "Or run: newgrp docker"
    fi
}

create_deploy_directory() {
    info "Creating deploy directory: ${DEPLOY_PATH}"

    if [[ -d "$DEPLOY_PATH" ]]; then
        info "${DEPLOY_PATH} already exists"
        local owner
        owner="$(stat -c '%U:%G' "$DEPLOY_PATH" 2>/dev/null || echo 'unknown')"
        if [[ "$owner" == "root:root" ]]; then
            success "Ownership correct: root:root"
        else
            warn "Ownership is ${owner} — setting to root:root"
            run_cmd chown root:root "$DEPLOY_PATH"
        fi
        return 0
    fi

    run_cmd mkdir -p "$DEPLOY_PATH"
    run_cmd chown root:root "$DEPLOY_PATH"
    run_cmd chmod 755 "$DEPLOY_PATH"
    success "Created ${DEPLOY_PATH} (root:root, 755)"
}

configure_logrotate() {
    info "Configuring logrotate for Docker container logs"

    local logrotate_content
    logrotate_content=$(cat <<'LOGROTATE'
/var/lib/docker/containers/*/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
    maxsize 50M
}
LOGROTATE
)

    if [[ -f "$LOGROTATE_CONF" ]]; then
        local existing_md5
        local new_md5
        existing_md5="$(md5sum "$LOGROTATE_CONF" 2>/dev/null | cut -d' ' -f1 || echo 'none')"
        new_md5="$(echo "$logrotate_content" | md5sum | cut -d' ' -f1)"
        if [[ "$existing_md5" == "$new_md5" ]]; then
            info "Logrotate configuration already up to date"
            return 0
        fi
        warn "Existing ${LOGROTATE_CONF} found — updating"
    fi

    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN: Would write ${LOGROTATE_CONF}:"
        echo "$logrotate_content"
    else
        echo "$logrotate_content" > "$LOGROTATE_CONF"
    fi

    success "Logrotate configured for Docker container logs (daily, 7 rotations, 50M max)"
}

configure_ufw() {
    info "Configuring UFW firewall"

    if ! command -v ufw >/dev/null 2>&1; then
        info "Installing UFW"
        run_cmd apt-get update
        run_cmd apt-get install -y ufw
    fi

    local ufw_status
    if [[ "$DRY_RUN" == true ]]; then
        info "DRY RUN: Would check and configure UFW rules"
        info "DRY RUN: Would allow SSH (port 22)"
        info "DRY RUN: Would allow HTTP (port 80) — Cloudflare handles SSL"
        info "DRY RUN: Would NOT open port 443 (Cloudflare handles SSL termination)"
        info "DRY RUN: Would enable UFW"
        return 0
    fi

    ufw_status="$(ufw status numbered 2>/dev/null || echo '')"

    # SSH port 22
    if echo "$ufw_status" | grep -qE '22(/tcp| ).*ALLOW'; then
        info "SSH (port 22) already allowed — skipping"
    else
        run_cmd ufw allow 22/tcp comment 'SSH'
        success "Allowed SSH (port 22)"
    fi

    # HTTP port 80
    if echo "$ufw_status" | grep -qE '80(/tcp| ).*ALLOW'; then
        info "HTTP (port 80) already allowed — skipping"
    else
        run_cmd ufw allow 80/tcp comment 'HTTP (Cloudflare)'
        success "Allowed HTTP (port 80)"
    fi

    info "Port 443 intentionally NOT opened — Cloudflare handles SSL termination"

    # Enable UFW if not already active
    if echo "$ufw_status" | grep -q 'Status: active'; then
        info "UFW already active"
    else
        run_cmd bash -c "echo 'y' | ufw enable"
        success "UFW enabled"
    fi
}

print_summary() {
    cat <<SUMMARY
${BLUE}Docker Host Setup Summary${NC}
  OS:            ${OS_ID:-detecting...} ${OS_VERSION:-}
  Deploy path:   ${DEPLOY_PATH}
  Daemon config: ${DAEMON_JSON}
  Log rotation:  json-file (max-size 10m, max-file 3)
  Logrotate:     ${LOGROTATE_CONF}
  UFW rules:     allow 22/tcp (SSH), allow 80/tcp (HTTP)
  Port 443:      NOT opened (Cloudflare SSL termination)
  Dry run:       ${DRY_RUN}
SUMMARY
}

main() {
    info "OMNI Docker Host Setup"
    [[ "$DRY_RUN" == true ]] && warn "Dry run mode — no changes will be made"

    check_root
    detect_os
    check_disk_space

    print_summary

    install_docker
    configure_docker_daemon
    configure_docker_group
    create_deploy_directory
    configure_logrotate
    configure_ufw

    if [[ "$DRY_RUN" == true ]]; then
        success "Dry run complete — no changes were made"
    else
        success "Docker host setup complete"
        info "Next steps:"
        info "  1. Log out and back in (or run: newgrp docker)"
        info "  2. Verify: docker run --rm hello-world"
        info "  3. Deploy: scripts/deploy-vps.sh"
    fi
}

main "$@"

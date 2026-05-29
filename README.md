# OMNI

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19-blue.svg)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

A multi-platform e-commerce order management system (OMS) integrating Shopee, Lazada, and TikTok Shop into a unified dashboard.

## Features

- **Multi-Platform Integration**: Connect and manage orders from Shopee, Lazada, and TikTok Shop
- **Unified Dashboard**: Single interface for all marketplace operations
- **Smart Database Backup**: PostgreSQL backup with change detection, SHA-256 checksums, and automatic chunking
- **Multi-Tenant Architecture**: Schema-based tenant isolation for secure data separation
- **JWT Authentication**: Secure access with refresh tokens and rate limiting
- **Automated Build System**: Python-based Docker orchestration with 22+ error recovery patterns

## Quick Start

```bash
# 1. Clone and navigate to project
cd omni

# 2. Configure environment
cp .env.example .env
# Edit .env with your API keys and secrets

# 3. Build and run
python build.py smart

# 4. Access the application
# Frontend: http://localhost/
# API: http://localhost/api/
```

## Project Structure

```
omni/
├── backend/              # Go backend (Gin + GORM)
│   ├── cmd/server/       # Application entry point
│   ├── internal/         # Internal packages
│   │   ├── handler/      # HTTP handlers
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Database access
│   │   └── model/        # Data models
│   ├── pkg/              # Shared utilities
│   └── shopee-sdk/       # Platform SDKs
├── frontend/             # React 19 frontend
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── pages/        # Route pages
│   │   ├── api/          # API clients
│   │   └── stores/       # Zustand state
│   └── public/
├── scripts/              # Build and deployment scripts
├── docker-compose.yml    # Docker orchestration
└── build.py              # Main build entry point
```

## Database Operations

```bash
# Create backup (smart - only if changes detected)
python build.py backup

# Restore from latest backup
python build.py restore

# Full rebuild with restore
python build.py full --restore
```

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 20+
- Python 3.10+
- Docker and Docker Compose

### Backend Development

```bash
cd backend
go mod download
go run ./cmd/server/
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

## Architecture

OMNI follows a layered architecture pattern:

```
Request → Handler → Service → Repository → Database
```

| Layer | Responsibility |
|-------|---------------|
| Handler | HTTP request parsing, validation, response formatting |
| Service | Business logic, orchestration, transformations |
| Repository | Database access and query execution |

## Configuration

Key environment variables (see `.env.example`):

- `JWT_SECRET` - JWT signing key (min 32 chars)
- `JWT_REFRESH_SECRET` - Refresh token signing key
- `ENCRYPTION_KEY` - Data encryption key (32 chars)
- `SHOPEE_*` - Shopee API credentials
- `LAZADA_*` - Lazada API credentials
- `TIKTOK_*` - TikTok Shop API credentials

## Documentation

- [Backend Documentation](backend/README.md)
- [Frontend Documentation](frontend/README.md)
- [Build System Guide](scripts/python-build/README.md)
- [API Documentation](backend/docs/api.md)

## Deployment

OMNI supports local development and VPS production deployments via Docker Compose. Choose the target that fits your environment.

### Prerequisites

- Docker Engine 24+
- Docker Compose v2 (plugin)
- SSH key pair for VPS deployments
- GitHub account for CI/CD (optional)

### Quick Start

Local development:

```bash
# Copy and configure local environment
cp .env.example .env.local
# Edit .env.local with local values

# Deploy locally
make deploy-local
```

VPS production:

```bash
# Configure environment
cp .env.example .env
# Edit .env with production values

# Set up the VPS (one-time)
make setup-vps

# Deploy to VPS
make deploy-vps
```

### Environment Setup

Copy `.env.example` to the file matching your target:

| Target | File | Usage |
|---|---|---|
| Local | `.env.local` | Development on your machine |
| VPS | `.env` | Production server deployment |

Set `DEPLOY_TARGET` to `local` or `vps` and `DEPLOY_SPEC` to one of the resource profiles below. Variables marked `__SET_IN_GH_SECRETS__` are injected by CI/CD and must not be stored in env files.

### Deployment Targets

| Target | Spec | RAM | CPU | Postgres Buffers | Use Case |
|---|---|---|---|---|---|
| VPS | `highspec` | 8 GB | 2 cores | 1 GB | Production with full workload |
| VPS | `standard` | 4 GB | 2 cores | 256 MB | Typical production |
| VPS | `lowspec` | 2 GB | 1 core | 128 MB | Staging or low-traffic |
| Local | `standard` | 4 GB | 2 cores | 256 MB | Development default |
| Local | `highspec` | 8 GB | 2 cores | 1 GB | Performance testing |
| Local | `lowspec` | 2 GB | 1 core | 128 MB | Resource-constrained dev |

Local deployments can optionally enable Cloudflare Tunnel mode with `--tunnel` to test tunnel connectivity before pushing to VPS.

### Compose Layering

Docker Compose files are layered to keep configuration modular:

```
docker-compose.tunnel.yml          (base: services, networks, volumes)
  + docker-compose.tunnel.<spec>.yml  (overlay: resource limits, tuning)
```

| File | Purpose |
|---|---|
| `docker-compose.tunnel.yml` | Base services, ports, volumes, networks |
| `docker-compose.tunnel.standard.yml` | Resource limits for 4 GB RAM |
| `docker-compose.tunnel.highspec.yml` | Resource limits for 8 GB RAM, tuned Postgres |
| `docker-compose.tunnel.lowspec.yml` | Resource limits for 2 GB RAM |

The deploy scripts select the overlay automatically based on `DEPLOY_SPEC`.

### CI/CD

Pushing to the `main` branch triggers an automated deployment pipeline:

1. Backend tests (`go test ./...`)
2. Frontend tests (`npm test`)
3. Docker image build on VPS via SSH
4. Compose restart with health verification

Manual deployments can be triggered from GitHub Actions using `workflow_dispatch`.

Configure the following GitHub Secrets under Settings > Secrets > Actions:

| Secret | Description |
|---|---|
| `VPS_HOST` | Server public IP address |
| `VPS_USER` | SSH username (e.g. `root`) |
| `VPS_SSH_KEY` | Private SSH key contents (include BEGIN/END markers) |
| `VPS_DEPLOY_PATH` | Deploy directory on VPS (e.g. `/opt/omni`) |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token for alerts |
| `TELEGRAM_CHAT_ID` | Telegram chat ID for alerts |

Set secrets via GitHub CLI:

```bash
gh secret set VPS_HOST --body '<YOUR_VPS_IP>'
gh secret set VPS_USER --body '<YOUR_SSH_USER>'
gh secret set VPS_SSH_KEY < <YOUR_SSH_KEY_PATH>
gh secret set VPS_DEPLOY_PATH --body '<YOUR_DEPLOY_PATH>'
gh secret set TELEGRAM_BOT_TOKEN --body '<YOUR_TELEGRAM_BOT_TOKEN>'
gh secret set TELEGRAM_CHAT_ID --body '<YOUR_TELEGRAM_CHAT_ID>'
```

### Health Check

Run a health check against all services:

```bash
# Local services
make health

# VPS services
make health TARGET=vps
```

The health check verifies Docker containers, HTTP endpoints, database connectivity, Redis, and disk usage.

### Backup

Create a database backup:

```bash
# Local backup
make backup

# VPS backup
make backup TARGET=vps
```

Or use the Python build system directly:

```bash
python build.py backup
```

Backups are stored in `backups/smart/` with SHA-256 checksums and change detection.

### Rollback

Roll back a VPS deployment to a specific commit:

```bash
make deploy-vps ROLLBACK=<commit-sha>
```

This checks out the specified commit, rebuilds, and redeploys. Use `git log --oneline` to find the target commit.

### Monitoring

OMNI sends deployment alerts via Telegram when `ALERT_ENABLED=true`. Alerts fire on:

- Deployment failures
- Health check failures
- Disk usage exceeding the configured threshold

Configure `ALERT_TELEGRAM_BOT_TOKEN` and `ALERT_TELEGRAM_CHAT_ID` in your environment or GitHub Secrets.

### Make Targets Reference

| Target | Description |
|---|---|
| `make deploy-vps` | Deploy to VPS via rsync over SSH |
| `make deploy-local` | Deploy locally with Docker Compose |
| `make deploy-dry-run` | Preview VPS deploy without making changes |
| `make setup-vps` | One-time VPS preparation (Docker, UFW, dirs) |
| `make health` | Check all service health endpoints |
| `make health TARGET=vps` | Check VPS service health |
| `make backup` | Create database backup |
| `make backup TARGET=vps` | Create VPS database backup |
| `make test` | Run backend and frontend tests |
| `make build` | Build backend and frontend locally |

## License

Proprietary - All rights reserved.

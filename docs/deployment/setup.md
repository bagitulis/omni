---
last_updated: 2026-05-03
updated_by: agent
relates_to: build.py, docker-compose*.yml, .env.example
stale_if_changed:
  - .env.example
  - build.py
  - docker-compose.tunnel.yml
  - backend/Dockerfile
  - frontend/Dockerfile
---

# Development & Deployment Setup

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Backend development |
| Node.js | 20+ | Frontend development |
| Python | 3.10+ | Build system (build.py) |
| Docker | Latest | Container orchestration |
| Docker Compose | v2+ | Multi-service management |

---

## Quick Start (New Machine)

```bash
# 1. Clone repository
git clone <repo-url> omni && cd omni

# 2. Configure environment
cp .env.example .env
# Edit .env — set JWT_SECRET, ENCRYPTION_KEY, platform API keys

# 3. Build and run everything
python build.py smart

# 4. Access
# Frontend: http://localhost/
# API:      http://localhost/api/
# pgweb:    http://localhost:8081/
```

---

## Environment Variables

### Security (REQUIRED in production)

| Variable | Description | Default | Notes |
|----------|-------------|---------|-------|
| `JWT_SECRET` | Access token signing key | (none) | Min 32 chars, REQUIRED |
| `JWT_REFRESH_SECRET` | Refresh token signing key | (none) | REQUIRED |
| `ENCRYPTION_KEY` | Data encryption key | (none) | Must be exactly 32 chars |
| `REDIS_PASSWORD` | Redis authentication | (none) | Set in production |
| `CLOUDFLARE_TUNNEL_TOKEN` | Tunnel auth for home hosting | (none) | Only for tunnel deployment |

### Application

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment mode | `development` |
| `PORT` | Backend port | `3000` |
| `LOG_LEVEL` | Logging verbosity | `info` |
| `DATABASE_PATH` | SQLite database path | `/app/config/databases` |

### Token & Rate Limiting

| Variable | Description | Default |
|----------|-------------|---------|
| `ACCESS_TOKEN_TTL` | Access token lifetime | `30m` |
| `REFRESH_TOKEN_TTL` | Refresh token lifetime | `168h` (7 days) |
| `AUTH_RATE_LIMIT` | Auth endpoint rate | `30/min/IP` |
| `LOGIN_RATE_LIMIT` | Login attempts rate | `10/min/IP` |
| `API_RATE_LIMIT` | General API rate | `1000/min` |
| `SYNC_RATE_LIMIT` | Platform sync rate | `5000/min/tenant` |
| `WEBHOOK_RATE_LIMIT` | Webhook rate | `1000/min` |

### Account Security

| Variable | Description | Default |
|----------|-------------|---------|
| `MAX_LOGIN_ATTEMPTS` | Failed attempts before lockout | `10` |
| `LOCKOUT_DURATION` | Lockout period | `15m` |

### Platform API Keys

| Variable | Platform | Purpose |
|----------|----------|---------|
| `SHOPEE_PARTNER_ID` | Shopee | Partner ID |
| `SHOPEE_PARTNER_KEY` | Shopee | Partner secret key |
| `SHOPEE_REDIRECT_URL` | Shopee | OAuth callback URL |
| `LAZADA_APP_KEY` | Lazada | App key |
| `LAZADA_APP_SECRET` | Lazada | App secret |
| `LAZADA_REDIRECT_URL` | Lazada | OAuth callback URL |
| `TIKTOK_APP_KEY` | TikTok | App key |
| `TIKTOK_APP_SECRET` | TikTok | App secret |
| `TIKTOK_REDIRECT_URL` | TikTok | OAuth callback URL |

### Email (Optional)

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server | (none) |
| `SMTP_PORT` | SMTP port | (none) |
| `SMTP_USER` | SMTP username | (none) |
| `SMTP_PASS` | SMTP password | (none) |

---

## Build System (build.py)

### Build Modes

| Mode | Command | When to Use |
|------|---------|-------------|
| **smart** | `python build.py smart` | Code changes (rebuilds only affected images) |
| **quickfix** | `python build.py quickfix` | Service issues (restart without rebuild) |
| **full** | `python build.py full` | Clean rebuild from scratch |
| **full --restore** | `python build.py full --restore` | New PC setup (rebuild + restore DB) |
| **backup** | `python build.py backup` | After database schema changes |
| **restore** | `python build.py restore` | Restore database from backup |
| **clean** | `python build.py clean` | Clean Docker resources |
| **status** | `python build.py status` | Show container status |
| **validate** | `python build.py validate` | Validation only (no build) |

### Flags

| Flag | Purpose |
|------|---------|
| `--spec=lowspec` | Use low-spec resource overrides |
| `--skip-frontend` | Skip frontend build |
| `--dry-run` | Validation mode (no actual build) |

### Typical Workflows

```bash
# Daily development
python build.py smart

# After database migration
python build.py smart && python build.py backup

# Something broke, quick restart
python build.py quickfix

# New machine setup
python build.py full --restore

# Low-spec machine
python build.py smart --spec=lowspec
```

---

## Docker Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| omni-postgres | postgres:16-alpine | 5432 | Database |
| omni-pgbouncer | edoburu/pgbouncer | - | Connection pooling |
| omni-redis | redis:7-alpine | 6379 | Caching |
| omni-backend | Custom (Go) | 3000 | API server |
| omni-frontend | Custom (React+Nginx) | 5174 | SPA |
| omni-nginx | nginx:alpine | 80 | Reverse proxy |
| omni-cloudflared | cloudflare/cloudflared | - | Tunnel (optional) |
| omni-pgweb | sosedoff/pgweb | 8081 | DB admin UI |

### Resource Allocation

| Service | CPU | RAM | Notes |
|---------|-----|-----|-------|
| Backend | 1.0 | 512MB | Standard spec |
| Frontend | 0.5 | 256MB | Standard spec |
| PostgreSQL | 1.0 | 512MB | Standard spec |
| Redis | 0.25 | 128MB | Standard spec |
| Nginx | 0.5 | 128MB | Standard spec |
| PgBouncer | 0.25 | 64MB | Connection pooling |

---

## Local Development (Without Docker)

### Backend

```bash
cd backend
go mod download
go run ./cmd/server/
# Runs on http://localhost:3000
```

### Frontend

```bash
cd frontend
npm install
npm run dev
# Runs on http://localhost:5173 (Vite dev server)
```

### Testing

```bash
# Backend
cd backend
go build ./...          # Compile check
go test ./...           # Unit tests
go test -race ./...     # Race condition detection

# Frontend
cd frontend
npm run build           # TypeScript + build check
npm run lint            # ESLint
npm test                # Unit tests
```

---

## CI/CD Pipeline

GitHub Actions runs on every push/PR:

1. **Test Backend**: Build, vet, test (45% coverage gate on security middleware)
2. **Test Frontend**: Install, lint, test
3. **Runtime Strict** (main branch only): Full Docker stack + integration + security + E2E + Lighthouse tests

See `.github/workflows/ci.yml` for details.

---

## Makefile Targets

```bash
# Root
make test              # Backend + frontend tests
make test-backend      # Backend only
make test-frontend     # Frontend only
make build             # Build all

# Backend
cd backend
make build             # Build binary
make test              # Unit tests
make test-integration  # Integration tests
make test-security     # Security tests
make test-coverage     # Coverage report (HTML)
make lint              # golangci-lint
make swagger           # Generate swagger docs
make mocks             # Generate mocks
```

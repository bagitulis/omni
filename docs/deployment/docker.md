---
last_updated: 2026-05-03
updated_by: agent
relates_to: docker-compose*.yml, nginx/
stale_if_changed:
  - docker-compose.tunnel.yml
  - docker-compose.tunnel.standard.yml
  - docker-compose.tunnel.lowspec.yml
  - nginx/nginx.conf
  - nginx/conf.d/tunnel.conf
  - backend/Dockerfile
  - frontend/Dockerfile
---

# Docker & Nginx Configuration

## Docker Compose Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    omni-nginx (:80)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ / → frontend │  │ /api → backend│  │ /uploads     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘  │
└─────────┼─────────────────┼─────────────────────────────┘
          │                 │
          ▼                 ▼
┌─────────────────┐  ┌─────────────────┐
│ omni-frontend   │  │ omni-backend    │
│ (:5174)         │  │ (:3000)         │
│ React + Nginx   │  │ Go + Gin        │
└─────────────────┘  └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │ omni-pgbouncer  │
                     │ (conn pooling)  │
                     └────────┬────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼                               ▼
     ┌─────────────────┐            ┌─────────────────┐
     │ omni-postgres   │            │ omni-redis      │
     │ (:5432)         │            │ (:6379)         │
     └─────────────────┘            └─────────────────┘
```

---

## Nginx Routing Rules

### Route Priority (top to bottom)

| Route | Target | Rate Limit | Timeout | Notes |
|-------|--------|-----------|---------|-------|
| `/health` | 200 OK | None | - | Docker healthcheck |
| `/api/webhooks` | Backend | None | 30s | Platform callbacks |
| `/api/platform-auth` | Backend | None | 30s | OAuth callbacks |
| `/api/*/sync` | Backend | 10/s burst 20 | 300s | Long-running syncs |
| `/api/master-products/import` | Backend | 10/s burst 20 | 300s | Bulk import |
| `/api/*` | Backend | 10/s burst 20 | 60s | General API |
| `/uploads/*` | Static files | None | - | 30-day cache |
| `/*` | Frontend | 100/s burst 50 | - | SPA routing |

### Security Rules

- **Blocked paths**: `.*` (dotfiles), `.env`, `.git`, `.sql`, `.conf`, `.bak`
- **Headers**: HSTS, X-Frame-Options, X-Content-Type-Options, X-XSS-Protection
- **Server tokens**: Hidden (no nginx version exposure)

---

## Nginx Performance Settings

| Setting | Value | Purpose |
|---------|-------|---------|
| Worker processes | auto | Match CPU cores |
| Worker connections | 2048 | Max concurrent connections |
| Client max body | 50MB | Upload limit |
| Gzip level | 6 | Compression (min 1000 bytes) |
| Keepalive timeout | 65s | Connection reuse |
| Sendfile | on | Kernel-level file transfer |

---

## Dockerfile Details

### Backend (Multi-stage)

```
Stage 1 (Builder): golang:1.24-alpine
  - Dependencies: gcc, musl-dev, git, libwebp-dev
  - Build: CGO_ENABLED=1 go build -o server ./cmd/server/

Stage 2 (Runtime): alpine:3.19
  - Dependencies: ca-certificates, sqlite, tzdata, libwebp
  - Binary: /app/server
  - Port: 3000
  - Healthcheck: wget /api/health every 30s
```

### Frontend (Multi-stage)

```
Stage 1 (Builder): node:20-alpine
  - Install: npm ci
  - Build: npm run build

Stage 2 (Runtime): nginx:1.27-alpine
  - Dist: /usr/share/nginx/html
  - Port: 80
  - Healthcheck: curl / every 30s
```

---

## Resource Profiles

### Standard Spec (default)

| Service | CPU Limit | RAM Limit |
|---------|-----------|-----------|
| Backend | 1.5 | 1536MB |
| Frontend | 0.75 | 512MB |
| PostgreSQL | 1.0 | 512MB |
| Redis | 0.5 | 256MB |
| Nginx | 0.5 | 128MB |
| PgBouncer | 0.25 | 64MB |

### Low Spec

Use `python build.py smart --spec=lowspec` for reduced resource allocation.

---

## Healthchecks

| Service | Endpoint | Interval | Timeout | Retries |
|---------|----------|----------|---------|---------|
| Backend | `GET /api/health` | 30s | 10s | 3 |
| Frontend | `GET /` | 30s | 10s | 3 |
| Nginx | `GET /health` | 10s | 5s | 3 |
| PostgreSQL | `pg_isready` | 5s | 5s | 5 |
| Redis | `redis-cli ping` | 10s | 5s | 5 |

---

## Volumes

| Volume | Host Path | Container Path | Purpose |
|--------|-----------|---------------|---------|
| postgres-data | `./data/postgres` | `/var/lib/postgresql/data` | Database persistence |
| backend-logs | `./logs` | `/app/logs` | Application logs |
| uploads | `./backend/uploads` | `/app/uploads` | User uploads |
| redis-data | Docker volume | `/data` | Redis persistence |

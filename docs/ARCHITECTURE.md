# 🚀 OMNI Production Architecture

**Last Updated:** 20 Januari 2026

---

## 📊 System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         PRODUCTION STACK                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────────────┐  │
│   │   nginx     │────▶│  Go Backend │────▶│   PostgreSQL 16     │  │
│   │  (Port 80)  │     │ (Port 8080) │     │    (Port 5432)      │  │
│   │             │     │             │     │                     │  │
│   │  Frontend   │     │    GORM     │     │   omni_main DB      │  │
│   │  + Reverse  │     │  + Gin      │     │   Schema per tenant │  │
│   │    Proxy    │     │             │     │                     │  │
│   └─────────────┘     └─────────────┘     └─────────────────────┘  │
│         │                   │                       │               │
│         │                   │                       │               │
│         ▼                   ▼                       ▼               │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────────────┐  │
│   │   Redis     │◀────│   Cache     │     │   Tenant Schemas    │  │
│   │  (6379)     │     │   Layer     │     │                     │  │
│   └─────────────┘     └─────────────┘     │  • system           │  │
│                                           │  • tenant_yumna_*   │  │
│                                           │  • tenant_tika_*    │  │
│                                           └─────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🐳 Docker Containers

| Container     | Image              | Port | Status     | Purpose            |
| ------------- | ------------------ | ---- | ---------- | ------------------ |
| omni-frontend | omni-frontend      | 80   | ✅ Healthy | Vue.js SPA + nginx |
| omni-backend  | omni-backend       | 8080 | ✅ Healthy | Go API Server      |
| omni-postgres | postgres:16-alpine | 5432 | ✅ Healthy | Main Database      |
| omni-redis    | redis:7-alpine     | 6379 | ✅ Running | Cache/Session      |

### Optional Services (Profile-based)

| Container        | Profile  | Port | Purpose           |
| ---------------- | -------- | ---- | ----------------- |
| omni-pgadmin     | `admin`  | 5050 | DB Management UI  |
| omni-cloudflared | `tunnel` | -    | Cloudflare Tunnel |

---

## 🚀 Quick Start

```powershell
# Start production stack
.\start-production.ps1

# With pgAdmin
.\start-production.ps1 -WithAdmin

# Stop all
.\start-production.ps1 -Stop

# View status
.\start-production.ps1 -Status

# View logs
.\start-production.ps1 -Logs
```

Or directly with docker-compose:

```bash
# Start
docker compose -f docker-compose.production.yml up -d

# With pgAdmin
docker compose -f docker-compose.production.yml --profile admin up -d

# Stop
docker compose -f docker-compose.production.yml down
```

---

## 🔗 Service URLs

| Service    | URL                                |
| ---------- | ---------------------------------- |
| Frontend   | http://localhost                   |
| API Health | http://localhost/api/health        |
| Direct API | http://localhost:8080/api/health   |
| pgAdmin    | http://localhost:5050 (if enabled) |
| PostgreSQL | localhost:5432                     |

---

## 📁 Key Files

| File                            | Purpose                    |
| ------------------------------- | -------------------------- |
| `docker-compose.production.yml` | Main production stack      |
| `start-production.ps1`          | Startup script             |
| `frontend/nginx-default.conf`   | nginx reverse proxy config |
| `backend-go/Dockerfile`         | Go backend container       |
| `backend-go/internal/config/`   | DB driver & config         |

---

## 🔧 Environment Variables

```env
# Required
JWT_SECRET=your-jwt-secret-min-32-chars
ENCRYPTION_KEY=your-encryption-key-base64

# PostgreSQL (defaults provided)
POSTGRES_USER=omni
POSTGRES_PASSWORD=omni_secure_2026
POSTGRES_DB=omni_main

# Optional
CORS_ORIGINS=http://localhost,https://yndigital.my.id
CLOUDFLARE_TUNNEL_TOKEN=your-token  # For tunnel profile
```

---

## 📊 Performance Metrics

Based on load testing (100 concurrent connections, 1000 requests):

| Metric       | Value       |
| ------------ | ----------- |
| Throughput   | 2,150 req/s |
| P50 Latency  | 26.36 ms    |
| P95 Latency  | 127.27 ms   |
| P99 Latency  | 403.13 ms   |
| Success Rate | 100%        |

---

## 🗄️ Database Schema

PostgreSQL menggunakan schema-per-tenant:

```
omni_main (database)
├── system (schema)
│   └── global_config
├── tenant_yumna_bertigamart (schema)
│   ├── users
│   ├── shopee_orders
│   ├── shopee_products
│   └── ...
└── tenant_tika_nusseyba (schema)
    ├── users
    ├── shopee_orders
    └── ...
```

---

## 🔄 Migration from Node.js

**Status:** ✅ Complete (20 Januari 2026)

- Node.js backend: **DEPRECATED**
- Go backend: **PRODUCTION**
- SQLite: **MIGRATED** to PostgreSQL
- Data migrated: 4,661 rows

---

## 🛟 Troubleshooting

### Container won't start

```powershell
# Check logs
docker logs omni-backend

# Restart
docker compose -f docker-compose.production.yml restart backend
```

### Database connection issues

```powershell
# Check PostgreSQL
docker exec omni-postgres psql -U omni -d omni_main -c "SELECT 1;"

# Check from Go backend
docker exec omni-backend wget -qO- http://localhost:8080/api/health
```

### Network issues

```powershell
# Recreate network
docker compose -f docker-compose.production.yml down
docker network rm omni-network
docker compose -f docker-compose.production.yml up -d
```

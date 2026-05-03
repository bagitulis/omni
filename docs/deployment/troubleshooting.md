---
last_updated: 2026-05-03
updated_by: agent
relates_to: build.py, docker-compose*.yml
stale_if_changed:
  - build.py
  - docker-compose.tunnel.yml
---

# Troubleshooting Guide

## Common Issues

### Build Issues

#### `python build.py smart` fails

```bash
# 1. Check Docker is running
docker info

# 2. Try quickfix first (restart without rebuild)
python build.py quickfix

# 3. If still failing, full rebuild
python build.py full

# 4. Nuclear option: clean everything and rebuild
python build.py clean && python build.py full
```

#### Frontend build fails with TS1261 (casing error)

**Cause**: Windows case-insensitive filesystem conflicts with import paths.

```bash
# Fix: Rename directory via temp name
cd frontend/src/components
mv modals _temp_modals
mv _temp_modals modals
```

**Prevention**: Always use lowercase directory names in React frontend imports.

#### Backend build fails with CGO errors

**Cause**: Missing C compiler for SQLite CGO bindings.

```bash
# Windows: Install TDM-GCC or MSYS2
# Docker: Already handled in Dockerfile (gcc, musl-dev)
```

---

### Runtime Issues

#### Backend container keeps restarting

```bash
# Check logs
docker logs omni-backend --tail 50

# Common causes:
# 1. Missing environment variables → check .env
# 2. Database connection failed → check postgres is healthy
# 3. Port conflict → check nothing else on port 3000
```

#### Frontend shows blank page

```bash
# Check frontend container
docker logs omni-frontend --tail 20

# Check nginx routing
docker logs omni-nginx --tail 20

# Verify frontend build succeeded
docker exec omni-frontend ls /usr/share/nginx/html/
```

#### API returns 502 Bad Gateway

```bash
# Backend is down or not ready
docker ps | grep omni-backend

# Check backend health
curl http://localhost:3000/api/health

# Restart backend
docker restart omni-backend
```

#### Rate limiting (429 Too Many Requests)

```
# Default limits:
# API: 10 req/s per IP
# Auth: 30 req/min per IP
# Login: 10 attempts/min per IP

# For development, these are usually not an issue
# For production, adjust in .env or nginx config
```

---

### Database Issues

#### PostgreSQL won't start

```bash
# Check logs
docker logs omni-postgres --tail 30

# Common: data directory permissions
# Fix: Remove data and reinitialize
rm -rf data/postgres
python build.py full
```

#### Database backup fails

```bash
# Check if changes exist
python build.py backup
# If "no changes detected", backup is skipped (expected)

# Force backup
python build.py backup --force
```

#### Need to restore database

```bash
# From latest backup
python build.py restore

# Full rebuild with restore
python build.py full --restore
```

---

### Platform Integration Issues

#### Shopee/Lazada/TikTok OAuth fails

1. Check API keys in `.env` are correct
2. Check redirect URLs match platform developer console
3. Check platform API status (may be down)
4. Check token expiration — refresh tokens may need renewal

#### Platform sync timeout

```
# Sync operations have 300s (5 min) timeout
# For very large catalogs, this may not be enough
# Check nginx/conf.d/tunnel.conf for timeout settings
```

---

### Development Issues

#### Hot reload not working (frontend)

```bash
# Vite dev server
cd frontend
npm run dev
# Access via http://localhost:5173 (not through nginx)
```

#### Go tests fail with "database locked"

**Cause**: SQLite concurrent access in tests.

```bash
# Run tests sequentially
go test -p 1 ./...
```

#### MCP server connection issues

```bash
# Check MCP server is running
# MCP servers communicate via stdio, not HTTP
# Verify the command in your MCP client config
```

---

## Diagnostic Commands

```bash
# Container status
docker ps -a | grep omni

# All logs
docker compose -f docker-compose.tunnel.yml logs --tail 50

# Specific service logs
docker logs omni-backend --tail 50 -f
docker logs omni-frontend --tail 50 -f
docker logs omni-nginx --tail 50 -f

# Resource usage
docker stats --no-stream | grep omni

# Network inspection
docker network inspect omni_default

# Enter container shell
docker exec -it omni-backend sh
docker exec -it omni-postgres psql -U postgres

# Health check
curl http://localhost/health
curl http://localhost/api/health
curl http://localhost:8081  # pgweb
```

---

## Recovery Procedures

### Full Reset (Last Resort)

```bash
# WARNING: This destroys all local data
python build.py clean
rm -rf data/postgres
python build.py full --restore  # Restore from backup
```

### Partial Reset (Keep Database)

```bash
python build.py clean
python build.py smart  # Rebuild without touching DB
```

### Restore from Backup

```bash
# Backups are in backups/smart/
# manifest.json tracks backup metadata
python build.py restore
```

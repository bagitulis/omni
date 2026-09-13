---
last_updated: 2026-05-03
updated_by: agent
relates_to: /
stale_if_changed:
  - backend/cmd/server/main.go
  - backend/internal/
  - frontend/src/
  - docker-compose*.yml
  - build.py
---

# OMNI - Project Context

> **Purpose**: High-level context for AI agents. Read this FIRST before any task.
> **Audience**: AI agents (Sisyphus, Atlas, Hephaestus, Sisyphus-Junior, Oracle)

---

## What Is OMNI?

OMNI is a **multi-platform e-commerce Order Management System (OMS)** that integrates Shopee, Lazada, and TikTok Shop into a unified dashboard. It's a proprietary SaaS product with multi-tenant architecture.

**Business Domain**: E-commerce operations management — order processing, product catalog sync, inventory management, and marketplace analytics.

**Target Users**: E-commerce sellers managing multiple marketplace accounts simultaneously.

---

## Tech Stack

| Layer | Technology | Version | Notes |
|-------|-----------|---------|-------|
| Backend | Go (Gin + GORM) | 1.24 | Layered architecture |
| Frontend | React + TypeScript | 19 | Vite 6, Ant Design 5, Zustand |
| Database | **PostgreSQL only** | 16 | Multi-tenant (schema-based). SQLite is used **only** for in-memory unit tests, never at runtime. |
| Build | Python (build.py) | 3.10+ | Docker orchestration |
| Proxy | Nginx | latest | Reverse proxy + static |
| MCP Servers | Go | 1.21+ | GitHub, Antigravity, Shopee Ads, TikTok Ads |

---

## Architecture Pattern

```
Request → Nginx → Backend API → Handler → Service → Repository → Database
                → Frontend (SPA) → Vite dev / static build
```

| Layer | Responsibility | Location |
|-------|---------------|----------|
| Handler | Parse request, validate, format response | `backend/internal/handler/` |
| Service | Business logic, orchestration, transformations | `backend/internal/service/` |
| Repository | Database access ONLY | `backend/internal/repository/` |
| Model | Data structures, GORM models | `backend/internal/model/` |

**CRITICAL RULE**: NO business logic in Handler layer. NO database access in Service layer.

---

## Multi-Tenant Architecture

- **Isolation**: Schema-based tenant separation in PostgreSQL (`tenant_{tenant_id}` + `search_path`)
- **Tenant ID**: Required on ALL protected endpoints (no default tenant)
- **Connection model**: One pooled Postgres connection **per tenant** (`maxOpen=25` each), fronted by PgBouncer in production. Do NOT remove PgBouncer — without it, ~8 tenants exhaust Postgres `max_connections`.
- **Auth**: JWT with refresh tokens, rate limiting

---

## Platform SDKs (LOCAL - Check Here First)

| Platform | Local Path | Key Files |
|----------|-----------|-----------|
| Shopee | `backend/shopee-sdk/` | orders.go, products.go, client.go |
| Lazada | `backend/lazada-sdk/` | order.go, product.go, auth.go |
| Lazada IOP | `backend/lazada_sdk/iop-sdk-go/` | Official IOP SDK |
| TikTok | `backend/tiktok_sdk/` | Comprehensive SDK (100+ files) |

**ALWAYS check local SDK before external docs.**

---

## Key Business Concepts

| Concept | Description |
|---------|-------------|
| Order | Marketplace order with items, buyer info, shipping |
| Product | Catalog item synced from marketplace platforms |
| Inventory | Stock levels across platforms |
| Platform | Connected marketplace account (Shopee/Lazada/TikTok) |
| Tenant | Isolated business entity (multi-tenant) |
| Shop | Individual marketplace shop under a tenant |

---

## Current State & Known Issues

### Active Development Areas
- React frontend migration (from Vue legacy)
- Product management consolidation
- Platform integration improvements

### Known Constraints
- Windows development environment (case-sensitivity issues)
- Directory casing: React uses lowercase (`modals/`), Vue legacy uses uppercase (`Modals/`)
- Frontend build: TS1261 errors from casing mismatches on Windows
- **Database is PostgreSQL-only at runtime.** `config.SetDatabaseDriver()` in `internal/app/app.go` hardcodes `DriverPostgres`; `DB_DRIVER` env var is not read. There is no SQLite runtime code path.
- **Unit tests use SQLite (`glebarez/sqlite`) in-memory, but production SQL uses Postgres-only syntax.** Queries in `services/inventory/inventory_filters.go` (`data->>'stock'`, `::INTEGER`, `~` regex), `inventory_service.go` (`data::text ILIKE`), `platform/config_manager_base.go` (`gen_random_uuid()`, `ON CONFLICT ... DO UPDATE`), and all `ILIKE` usages **cannot execute on SQLite**. Green unit tests do not validate these code paths.
- **`basePath` is a tenant-config concept, not a DB path.** It is threaded through handlers/routes/services to locate `tenants.json`. `config.GetTenantDB(tenantID, basePath)` ignores `basePath` internally (Postgres connection is keyed by tenant only). Do not remove `basePath` from signatures without tracing all call sites.

### What NOT To Do
- Never use `git rm` on `backups/smart/` or `nginx/logs/error.log` (intentionally tracked)
- Never add default tenant fallback (`if tenantID == "" { tenantID = "default" }`)
- Never use Google Fonts (system fonts only)
- Never use camelCase in JSON API responses (always snake_case)
- Never put business logic in handlers
- Never suppress TypeScript errors with `as any` or `@ts-ignore`
- Never document the database as "PostgreSQL / SQLite" — SQLite is test-only

---

## Build & Deploy

```bash
python build.py smart      # Smart build (recommended for code changes)
python build.py quickfix   # Fix service issues without rebuild
python build.py full       # Full rebuild from scratch
python build.py backup     # Database backup (MANDATORY after schema changes)
python build.py restore    # Restore from latest backup
```

**Access Points**:
- Frontend: `http://localhost/`
- API: `http://localhost/api/`
- Backend direct: `http://localhost:3000/`

---

## Documentation Map

| Need | Location |
|------|----------|
| AI Agent Rules | `AGENTS.md` (root) |
| Backend Rules | `backend/AGENTS.md` |
| Frontend Rules | `frontend/AGENTS.md` |
| MCP Rules | `mcp-servers/AGENTS.md` |
| API Endpoints | `docs/api/endpoints.md` |
| Database Schema | `docs/database/schema.md` |
| Architecture | `docs/architecture/overview.md` |
| Deployment | `docs/deployment/setup.md` |
| UI Design System | `docs/UI/` |
| Platform Integrations | `docs/integrations/` |
| Build System | `scripts/python-build/README.md` |

---

## Rules Hierarchy (for AI Agents)

```
AGENTS.md (ROOT)              ← Critical invariants (ALWAYS read)
├── backend/AGENTS.md         ← When working in backend/
├── frontend/AGENTS.md        ← When working in frontend/
├── mcp-servers/AGENTS.md     ← When working in mcp-servers/
├── SISYPHUS_RULES.md         ← Orchestrator rules
├── EXECUTOR_RULES.md         ← Executor rules
├── PROMETHEUS_RULES.md       ← Planner rules
└── TESTING_RULES.md          ← Testing constitution
```

**Precedence**: Root invariants > Directory context > Agent-role rules > Skills

---

## Quick Reference for Common Tasks

| Task | Start Here |
|------|-----------|
| Fix backend bug | Trace: handler → service → repository → model |
| Add API endpoint | Create handler + service + repository + route registration |
| Fix frontend bug | Check component → hook → API client → types |
| Add frontend page | Create page + route + API integration + types |
| Database change | Migration → build.py smart → build.py backup |
| Platform integration | Check local SDK first → then external docs |

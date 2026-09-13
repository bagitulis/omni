---
alwaysApply: true
description: Project architecture overview (synced from project root)
synced_from: ARCHITECTURE.md
---

<!-- AUTO-SYNCED from ARCHITECTURE.md by .claude/sync_project_rules.py -->
<!-- Source: ARCHITECTURE.md | DO NOT EDIT — edit the source file and re-run sync -->

# Architecture: Omni

> Auto-generated reference for AI agents. Read this before searching the codebase.

## Stack

### Backend
- **Language:** Go 1.24 (primary runtime)
- **Module:** `github.com/omni/backend`
- **Framework:** Gin (HTTP router)
- **ORM:** GORM
- **Database:** PostgreSQL 16 only (multi-tenant schema-based). SQLite is in-memory for unit tests, never runtime.
- **Logging:** zerolog
- **Architecture:** Handler → Service → Repository → Database

### Frontend
- **Framework:** React 19 + TypeScript 5.7+ + Vite 6
- **UI Library:** Ant Design 5
- **State:** Zustand (client) + TanStack Query (server)
- **Router:** React Router 7

### MCP Servers
- **Language:** Go 1.21+
- **Protocol:** JSON-RPC 2.0 (MCP)
- **Integrations:** Google Sheets API, Python (ads analysis)

## Folder Structure

```
omni/
├── backend/                    # Go backend service
│   ├── cmd/
│   │   └── server/             # Main server entrypoint (main.go)
│   ├── internal/
│   │   ├── app/                # App wiring, startup sync, auto-function handlers
│   │   ├── config/             # Config loading, DB init, migrations, GORM logger
│   │   ├── dto/
│   │   │   ├── request/        # Request DTOs (shopee, lazada, tiktok)
│   │   │   └── response/       # Response DTOs, API response types
│   │   ├── errors/             # Shared error types
│   │   ├── handlers/           # HTTP handlers (Gin)
│   │   │   ├── google/         # Google auth, sheets, quota, service account
│   │   │   ├── inventory/      # Inventory handlers
│   │   │   ├── lazada/         # Lazada-specific handlers
│   │   │   ├── master_product/ # Master product handlers
│   │   │   ├── shopee/         # Shopee-specific handlers
│   │   │   └── tiktok/         # TikTok-specific handlers
│   │   ├── middleware/         # Auth, tenant, CORS middleware
│   │   ├── mocks/              # Test mocks
│   │   ├── models/             # GORM models (DB schema)
│   │   ├── repositories/       # Database access layer (GORM)
│   │   ├── routes/             # Route registration
│   │   ├── services/           # Business logic layer
│   │   │   ├── analytics/      # Multi-platform analytics + intelligence
│   │   │   ├── autofunction/   # Auto-function orchestration
│   │   │   ├── cache/          # Caching service
│   │   │   ├── google/         # Google Sheets/OAuth services
│   │   │   ├── image/          # Image processing
│   │   │   ├── inventory/      # Cross-platform inventory management
│   │   │   ├── jobs/           # Background job services
│   │   │   ├── label/          # Label/shipping services
│   │   │   ├── lazada/         # Lazada platform integration
│   │   │   ├── master_product/ # Master product sync
│   │   │   ├── monitoring/     # Monitoring services
│   │   │   ├── oauth/          # OAuth token management
│   │   │   ├── operations/     # Cross-platform operations
│   │   │   ├── orders/         # Order management
│   │   │   ├── platform/       # Platform abstraction
│   │   │   ├── products/       # Multi-platform product cloning
│   │   │   ├── quota/          # API quota management
│   │   │   ├── route/          # Route services
│   │   │   ├── security/       # Security services
│   │   │   ├── sheets/         # Google Sheets integration
│   │   │   ├── shop/           # Shop management
│   │   │   ├── shopee/         # Shopee platform integration
│   │   │   ├── sku/            # SKU management
│   │   │   ├── spreadsheet/    # Spreadsheet export
│   │   │   ├── sync/           # Order/product sync orchestration
│   │   │   ├── tiktok/         # TikTok platform integration
│   │   │   ├── webhooks/       # Webhook handling
│   │   │   └── wholesale/      # Wholesale pricing
│   │   ├── testutils/          # Shared test utilities
│   │   └── utils/
│   │       ├── http/           # HTTP utilities
│   │       └── logger/         # Logger utilities
│   ├── migrations/             # SQL migration files
│   ├── pkg/
│   │   ├── lazada/             # Lazada SDK wrapper
│   │   ├── shopee/             # Shopee SDK wrapper
│   │   └── tiktok/             # TikTok SDK wrapper
│   ├── lazada-sdk/             # Lazada SDK (local)
│   ├── lazada_sdk/             # Lazada IOP SDK (official)
│   ├── shopee-sdk/             # Shopee SDK (local)
│   ├── tiktok_sdk/             # TikTok SDK (comprehensive, 100+ files)
│   ├── config/static/          # Static config files (Google credentials)
│   ├── data/                   # SQLite databases (dev, per-tenant)
│   ├── logs/                   # Application logs
│   ├── tests/                  # Integration tests
│   └── go.mod / go.sum
├── frontend/                   # React 19 frontend
│   ├── src/
│   │   ├── api/                # API client functions
│   │   ├── components/
│   │   │   ├── layout/         # AppLayout, Sidebar, Header, MobileNav
│   │   │   ├── ui/             # Extended Ant Design components
│   │   │   ├── tables/         # Data tables
│   │   │   ├── forms/          # Form components
│   │   │   ├── modals/         # Modal components (LOWERCASE — Windows bug risk)
│   │   │   ├── orders/         # Order-specific components
│   │   │   └── shared/         # Shared components (gallery, etc.)
│   │   ├── pages/              # Route pages
│   │   ├── hooks/              # Custom hooks
│   │   ├── stores/             # Zustand stores
│   │   ├── types/              # TypeScript interfaces (snake_case)
│   │   ├── lib/                # Utilities
│   │   └── styles/             # Theme and global CSS
│   ├── e2e/                    # Playwright E2E tests
│   ├── tests/                  # Unit/integration tests
│   └── package.json
├── mcp-servers/                # MCP server binaries (Go)
│   ├── bin/                    # Compiled MCP binaries
│   ├── cmd/
│   │   ├── github/             # GitHub accounts MCP
│   │   ├── antigravity/        # Antigravity accounts MCP
│   │   ├── shopee-ads/         # Shopee ads analyzer MCP
│   │   └── tiktok-ads/         # TikTok ads analyzer MCP
│   └── pkg/
│       ├── mcp/                # MCP protocol implementation
│       ├── sheets/             # Google Sheets integration
│       ├── totp/               # TOTP/2FA generator
│       └── python/             # Python runner for ads
├── rules-master/               # Rules control plane (rules.json, sync_rules.py)
├── docs/                       # Architecture, API, database, deployment docs
├── nginx/                      # Nginx config and SSL
├── scripts/                    # Build, deploy, DB scripts
├── data/postgres/              # PostgreSQL data (Docker volume)
└── AGENTS.MD                   # Project rules (mandatory)
```

## Key Search Patterns

```
# Backend handlers
backend/internal/handlers/*.go                    # Root-level handlers
backend/internal/handlers/shopee/*.go             # Shopee handlers
backend/internal/handlers/lazada/*.go             # Lazada handlers
backend/internal/handlers/tiktok/*.go             # TikTok handlers
backend/internal/handlers/inventory/*.go          # Inventory handlers
backend/internal/handlers/master_product/*.go     # Master product handlers
backend/internal/handlers/google/*.go             # Google handlers

# Backend services
backend/internal/services/*.go                    # Root-level services
backend/internal/services/shopee/*.go             # Shopee services
backend/internal/services/lazada/*.go             # Lazada services
backend/internal/services/tiktok/*.go             # TikTok services
backend/internal/services/sync/*.go               # Sync orchestration
backend/internal/services/analytics/*.go          # Analytics services
backend/internal/services/products/*.go           # Product cloning
backend/internal/services/inventory/*.go          # Inventory management

# Backend repositories
backend/internal/repositories/*.go               # All repositories

# Backend models
backend/internal/models/*.go                     # GORM models

# Backend routes
backend/internal/routes/*.go                     # Route registration

# Platform SDKs (search here FIRST for platform APIs)
backend/shopee-sdk/                              # Shopee SDK
backend/lazada-sdk/                              # Lazada SDK
backend/tiktok_sdk/                              # TikTok SDK
backend/pkg/shopee/                              # Shopee SDK wrapper
backend/pkg/lazada/                              # Lazada SDK wrapper

# Frontend
frontend/src/api/*.ts                            # API client functions
frontend/src/pages/*.tsx                         # Route pages
frontend/src/components/**/*.tsx                 # React components
frontend/src/hooks/*.ts                          # Custom hooks
frontend/src/stores/*.ts                         # Zustand stores
frontend/src/types/*.ts                          # TypeScript types

# Config
backend/internal/config/config.go               # App config
backend/internal/config/migration.go            # DB migrations
backend/migrations/*.sql                         # SQL migration files
```

## Build Commands

```bash
# Backend
cd backend
go build ./...                    # Compile-check all packages
go build -o bin/omni-backend.exe ./cmd/server  # Build server binary
go test ./...                     # Run all tests (MUST pass before commit)

# Docker (recommended for full stack)
python build.py smart             # Smart build (RECOMMENDED for code changes)
python build.py quickfix          # Fix service issues without rebuild
python build.py full              # Full rebuild from scratch (no cache)
python build.py backup            # Backup DB (MANDATORY after schema changes)

# Frontend
cd frontend
npm run build                     # Production build (tsc + vite)
npm run dev                       # Dev server (Vite)
npm run lint                      # ESLint
npm run test                      # Vitest (unit tests)

# MCP Servers
cd mcp-servers
go build -o bin/mcp-github.exe ./cmd/github
go build -o bin/mcp-antigravity.exe ./cmd/antigravity
go build -o bin/mcp-shopee-ads.exe ./cmd/shopee-ads
go build -o bin/mcp-tiktok-ads.exe ./cmd/tiktok-ads
```

## Architecture Flow

```
Request → Gin Router (internal/routes)
  → Middleware (auth, tenant isolation)
  → Handler (internal/handlers)
      → Validate request
      → Call Service
  → Service (internal/services)
      → Business logic
      → Call Repository
  → Repository (internal/repositories)
      → GORM DB access
      → PostgreSQL (tenant_{tenantID} schema)
  → Response (snake_case JSON)
```

## Multi-Tenant Architecture

- PostgreSQL schema-based isolation: `tenant_{tenantID}`
- Every protected endpoint MUST extract and validate `tenant_id`
- **NO DEFAULT TENANT**: `if tenantID == "" { return error "Missing tenant_id" }`
- SQLite used for development (per-tenant `.db` files in `backend/data/`)

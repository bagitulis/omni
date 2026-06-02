---
last_updated: 2026-06-02
updated_by: agent
relates_to: backend/internal/, frontend/src/
stale_if_changed:
  - backend/cmd/server/main.go
  - backend/internal/handler/
  - backend/internal/service/
  - backend/internal/repository/
  - frontend/src/api/
  - frontend/src/pages/
---

# Architecture Overview

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         NGINX (Reverse Proxy)                    │
│  ┌─────────────────────────┐  ┌─────────────────────────────┐  │
│  │  / → Frontend (static)  │  │  /api/* → Backend (:3000)   │  │
│  └─────────────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                                    │
         ▼                                    ▼
┌─────────────────────┐          ┌─────────────────────────────┐
│   React Frontend    │          │        Go Backend            │
│   (Vite SPA)        │          │                              │
│                     │          │  ┌───────────────────────┐   │
│  • React 19         │  HTTP    │  │      Handlers         │   │
│  • Ant Design 5     │ ──────▶  │  │  (Parse + Validate)   │   │
│  • Zustand          │          │  └───────────┬───────────┘   │
│  • TanStack Query   │          │              │               │
│  • TypeScript       │          │  ┌───────────▼───────────┐   │
│                     │          │  │      Services         │   │
└─────────────────────┘          │  │  (Business Logic)     │   │
                                 │  └───────────┬───────────┘   │
                                 │              │               │
                                 │  ┌───────────▼───────────┐   │
                                 │  │    Repositories       │   │
                                 │  │  (Database Access)    │   │
                                 │  └───────────┬───────────┘   │
                                 │              │               │
                                 └──────────────┼───────────────┘
                                                │
                                 ┌──────────────▼───────────────┐
                                 │         Database              │
                                 │  PostgreSQL (prod) / SQLite   │
                                 │  Multi-tenant (schema-based)  │
                                 └──────────────────────────────┘
```

---

## Backend Architecture (Go)

### Layered Architecture

```
backend/
├── cmd/server/           # Entry point (main.go)
│   └── main.go           # Server bootstrap, DI, router setup
├── internal/             # Core application code
│   ├── config/           # Configuration loading
│   ├── handler/          # HTTP handlers (request/response)
│   ├── service/          # Business logic layer
│   ├── repository/       # Database access layer
│   ├── model/            # GORM models + domain types
│   ├── middleware/       # Auth, CORS, rate limiting
│   └── dto/              # Data Transfer Objects
├── pkg/                  # Shared utilities
├── shopee-sdk/           # Shopee platform SDK
├── lazada-sdk/           # Lazada platform SDK
├── lazada_sdk/           # Lazada IOP SDK (official)
└── tiktok_sdk/           # TikTok Shop SDK
```

### Layer Responsibilities

| Layer | Can Call | Cannot Call | Responsibility |
|-------|---------|-------------|---------------|
| Handler | Service | Repository, DB | Parse HTTP, validate input, format response |
| Service | Repository, other Services | Handler, DB directly | Business logic, orchestration |
| Repository | Database (GORM) | Handler, Service | CRUD operations, queries |
| Model | Nothing | Everything | Data structures, validation tags |

### Dependency Injection

```go
// Typical initialization flow in main.go
db := database.Connect(config)
repo := repository.NewOrderRepository(db)
svc := service.NewOrderService(repo)
handler := handler.NewOrderHandler(svc)
router.RegisterOrderRoutes(handler)
```

---

## Frontend Architecture (React)

### Structure

```
frontend/src/
├── api/                  # Axios API clients
│   ├── client.ts         # Base axios instance + interceptors
│   ├── orders.ts         # Order API functions
│   ├── products.ts       # Product API functions
│   └── auth.ts           # Authentication API
├── components/           # Reusable UI components
│   ├── layout/           # AppLayout, Sidebar, Header
│   ├── ui/               # Extended Ant Design components
│   ├── tables/           # Data table components
│   ├── forms/            # Form components
│   └── modals/           # Modal components
├── pages/                # Route-level pages
│   ├── auth/             # Login, Register
│   ├── dashboard/        # Dashboard
│   ├── orders/           # Order management
│   ├── products/         # Product catalog
│   ├── inventory/        # Inventory management
│   └── settings/         # App settings
├── hooks/                # Custom React hooks
├── stores/               # Zustand state stores
├── types/                # TypeScript interfaces (snake_case!)
├── lib/                  # Utility functions
└── styles/               # Theme + global CSS
```

### Data Flow

```
User Action → Component → Hook (TanStack Query) → API Client → Backend
                                    ↓
                              Zustand Store (if global state needed)
```

### Key Patterns

- **API Types**: Always `snake_case` to match backend JSON
- **State Management**: Zustand for global state, TanStack Query for server state
- **Styling**: Ant Design theme tokens (no hardcoded colors)
- **Routing**: React Router v6
- **Font**: System fonts only (12px base, data-dense)

---

## Multi-Tenant Architecture

### Tenant Isolation Strategy

```
PostgreSQL (Production):
  ┌─────────────────────────────────────┐
  │  Database: omni                      │
  │  ├── Schema: tenant_abc123          │
  │  │   ├── orders                     │
  │  │   ├── products                   │
  │  │   └── ...                        │
  │  ├── Schema: tenant_def456          │
  │  │   ├── orders                     │
  │  │   ├── products                   │
  │  │   └── ...                        │
  │  └── Schema: public (shared)        │
  │      ├── tenants                    │
  │      └── users                      │
  └─────────────────────────────────────┘

SQLite (Development):
  backend/data/
  ├── tenant_abc123.db
  ├── tenant_def456.db
  └── shared.db
```

### Authentication Flow

```
Login → JWT Access Token (short-lived) + Refresh Token (long-lived)
     → Token includes tenant_id claim
     → Middleware extracts tenant_id from token
     → All queries scoped to tenant
```

---

## Credential System

All platform credentials (Shopee, Lazada, TikTok) are stored in PostgreSQL. No `.env` fallback exists. PostgreSQL is the single source of truth.

### Storage Tables

```
PostgreSQL (public schema):
  credential_app_configs     — App-level credentials (partner ID, API keys)
  credential_connections      — Store-level connections (access/refresh tokens)
  credential_audit_events    — Lifecycle audit log (never contains secrets)
  platform_configs (being migrated) — Legacy bridge table, transitioning to canonical credential tables
```

### Resolution Chain

```
CredentialService.GetPlatformCredentials(tenantID, platform)
  │
  ├── CredentialRepository.GetAppConfig(tenantID, platform)
  │     → credential_app_configs table
  │     → PartnerID, PartnerKey (Shopee) or AppKey, AppSecret (Lazada/TikTok)
  │
  └── CredentialRepository.ListConnections(tenantID, platform)
        → credential_connections table
        → AccessToken, RefreshToken, TokenExpiry, RefreshExpiry
```

### Encryption

All secrets are encrypted at rest using Fernet (AES-128-CBC). The encryption key comes from the `ENCRYPTION_KEY` environment variable. Tokens and API keys are encrypted before storage and decrypted on read by the repository layer.

### Token Lifecycle

```
OAuth Callback → CredentialService stores tokens → credential_connections
Auto-Refresh   → TokenManager refreshes → writes back to credential_connections
Manual Token   → POST /credentials/platforms/:platform/connections/manual-token
Audit Trail    → Every mutation logs a credential_audit_events row (metadata only, no secrets)
```

### Optimistic Locking

Credential connections use a `version` column for optimistic concurrency control. Manual token application and auto-refresh both check the version before writing, preventing silent token overwrites.

---

## MCP Servers Architecture

```
mcp-servers/
├── cmd/
│   ├── github/           # GitHub accounts management
│   ├── antigravity/      # Antigravity accounts management
│   ├── shopee-ads/       # Shopee ads analyzer
│   └── tiktok-ads/       # TikTok ads analyzer
└── internal/             # Shared MCP utilities
```

Each MCP server:
- Implements JSON-RPC 2.0 protocol
- Runs as standalone process
- Communicates via stdio
- Provides tools for AI agent consumption

---

## Build & Deployment Architecture

```
┌──────────────────────────────────────────────────┐
│                  build.py                          │
│  (Python orchestrator - 22+ error recovery)       │
│                                                   │
│  Modes:                                           │
│  • smart    → Detect changes, rebuild affected    │
│  • full     → Clean rebuild from scratch          │
│  • quickfix → Restart services without rebuild    │
│  • backup   → Smart database backup               │
│  • restore  → Restore from backup                 │
└──────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────┐
│              Docker Compose                        │
│                                                   │
│  Services:                                        │
│  • backend  (Go binary)                           │
│  • frontend (Nginx + static build)                │
│  • postgres (Database)                            │
│  • nginx    (Reverse proxy)                       │
└──────────────────────────────────────────────────┘
```

---

## Design Decisions

See `docs/architecture/decisions/` for Architecture Decision Records (ADRs).

Key decisions:
1. **Multi-tenant via schema isolation** — Security + performance balance
2. **Layered architecture** — Separation of concerns, testability
3. **Local SDKs** — Control over platform API integration
4. **React migration** — Modern stack replacing Vue legacy
5. **build.py orchestrator** — Complex Docker management with error recovery

# Omni — Full-Stack OMS (Order Management System)

Multi-tenant e-commerce order management: Go backend (Gin + GORM) + React 19 frontend (Ant Design 5) + MCP servers.

## Quick Start

```bash
# Backend (Go)
cd backend && go mod download && go build ./cmd/server/ && ./server

# Frontend (React + Vite)
cd frontend && pnpm install && pnpm dev

# Database migrations
cd backend && go run cmd/migrate/main.go

# MCP Servers
cd mcp-servers && go build ./...
```

## Project Structure

- `backend/` — Go backend (Gin HTTP, GORM ORM, zerolog)
  - `cmd/server/` — Main entrypoint
  - `internal/handlers/` — HTTP handlers (shopee, lazada, tiktok, google, inventory, master_product)
  - `internal/services/` — Business logic layer
  - `internal/repositories/` — Data access layer (GORM)
  - `internal/models/` — DB schema models
  - `internal/middleware/` — Auth, tenant, CORS
  - `internal/config/` — Config loading, DB init, migrations
- `frontend/` — React 19 + TypeScript + Vite 6
  - `src/pages/` — Route-based pages (Dashboard, Orders, Products, Inventory, etc.)
  - `src/components/` — Reusable UI (tables, forms, modals, layout)
  - `src/stores/` — Zustand state management
  - `src/api/` — API client + hooks (TanStack Query)
- `mcp-servers/` — MCP protocol servers (Google Sheets, Python ads analysis)
- `scripts/` — Setup, migration, and utility scripts
- `docs/` — Documentation + screenshots
- `.claude/` — Claude Code config and rules
- `notebooks/` — Jupyter notebooks for data analysis

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.24, Gin, GORM, zerolog |
| **Database** | PostgreSQL 16 only (prod + dev); SQLite in-memory for unit tests only |
| **Frontend** | React 19, TypeScript 5.7+, Vite 6 |
| **UI** | Ant Design 5, Zustand, TanStack Query |
| **MCP** | Go 1.21+, JSON-RPC 2.0 |
| **Python** | Ads analysis, data processing |

## Platform Integrations

- **Shopee** — Orders, products, inventory sync
- **Lazada** — Orders, products sync
- **TikTok** — Orders sync
- **Google** — Auth, Sheets API, service accounts
- **Multi-tenant** — Schema-based tenant isolation (`tenant_{tenant_id}` + `search_path`). Postgres-only at runtime; a pooled connection is created per tenant and PgBouncer fronts them in production.

## Available Rules

Located in `.claude/rules/`. All rules auto-apply (`alwaysApply: true`).

| Rule | Purpose |
|------|---------|
| `project-context.md` | Project-specific coding standards and structure |
| `git-safety.md` | Git operation safety and auto-commit policy |
| `code-quality.md` | ~300 line guideline, research protocol, commit evidence |
| `delegation-rules.md` | Agent delegation, escalation, session continuity |
| `react-rules.md` | React 19 + Ant Design 5 standards (omni-specific) |
| `proxy-no-vision.md` | Proxy vision support detection and fallbacks |
| `playwright-screenshots.md` | Screenshot naming conventions (omni-specific) |

## Coding Standards

See `.claude/rules/project-context.md` for detailed coding standards.
See `.claude/rules/react-rules.md` for React/frontend-specific standards.

## Git Safety

See `.claude/rules/git-safety.md` for git operation rules and auto-commit policy.

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
- `rules-master/` — Master rules for auto-sync to all projects
- `docs/` — Documentation + screenshots
- `.claude/` — Claude Code config, rules, BMAD skills
- `.agents/` — Agent skills (BMAD method, 43 skills)
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

## AI Platform Support

- **Claude Code** — Primary (this project)
- **OpenCode** — Via AI.py hub (`python AI.py` from extensions/)
- Provider abstraction: `extensions/opencode-configs/ai_providers.py`
- Cross-platform: each platform's config is independent — switching doesn't break the other

## Available BMAD Skills (43 skills)

Located in `.agents/skills/`. Invoke via slash command or natural language.

### Core Agents
- `bmad-agent-analyst` — Business analyst (Mary)
- `bmad-agent-architect` — System architect (Winston)
- `bmad-agent-dev` — Senior developer (Amelia)
- `bmad-agent-pm` — Product manager (John)
- `bmad-agent-tech-writer` — Documentation specialist (Paige)
- `bmad-agent-ux-designer` — UX/UI designer (Sally)

### Product Workflow
- `bmad-brainstorming` — Ideation sessions with diverse techniques
- `bmad-create-prd` / `bmad-edit-prd` / `bmad-validate-prd` — PRD lifecycle
- `bmad-product-brief` — Product brief creation
- `bmad-create-architecture` — Architecture design
- `bmad-create-epics-and-stories` / `bmad-create-story` — Story management
- `bmad-sprint-planning` / `bmad-sprint-status` — Sprint tracking
- `bmad-check-implementation-readiness` — Validate specs before coding
- `bmad-dev-story` / `bmad-quick-dev` — Story implementation
- `bmad-evaluate-project` — Project evaluation (omni-specific)

### Research & Discovery
- `bmad-domain-research` — Industry/domain research
- `bmad-market-research` — Market & competitor analysis
- `bmad-technical-research` — Technology evaluation

### Quality & Review
- `bmad-code-review` — Adversarial code review
- `bmad-review-adversarial-general` — Critical analysis
- `bmad-review-edge-case-hunter` — Edge case detection
- `bmad-editorial-review-prose` / `bmad-editorial-review-structure` — Writing review
- `bmad-qa-generate-e2e-tests` — Automated test generation

### Documentation
- `bmad-document-project` — Generate project documentation
- `bmad-index-docs` / `bmad-shard-doc` — Doc organization
- `bmad-distillator` — Document compression

### Meta & Utilities
- `bmad-customize` — Skill customization
- `bmad-help` — BMad guidance
- `bmad-generate-project-context` — Generate project-context.md
- `bmad-correct-course` — Mid-sprint changes
- `bmad-checkpoint-preview` — Human review
- `bmad-party-mode` — Multi-agent discussion
- `bmad-retrospective` — Post-epic review
- `bmad-prfaq` — Working backwards
- `bmad-create-ux-design` — UX specifications
- `bmad-advanced-elicitation` — Deep critique methods

**Auto-routing**: Rules in `.claude/rules/bmad-auto-router.md` automatically detect task type and invoke appropriate skills in YOLO mode.

## Available Rules (7 rules)

Located in `.claude/rules/`. All rules auto-apply (`alwaysApply: true`).

| Rule | Purpose |
|------|---------|
| `project-context.md` | Project-specific coding standards and structure |
| `git-safety.md` | Git operation safety and auto-commit policy |
| `code-quality.md` | ~300 line guideline, research protocol, commit evidence |
| `delegation-rules.md` | Agent delegation, escalation, session continuity |
| `bmad-auto-router.md` | Auto-detect task type → invoke BMAD skill |
| `react-rules.md` | React 19 + Ant Design 5 standards (omni-specific) |
| `proxy-no-vision.md` | Proxy vision support detection and fallbacks |
| `playwright-screenshots.md` | Screenshot naming conventions (omni-specific) |

## Coding Standards

See `.claude/rules/project-context.md` for detailed coding standards.
See `.claude/rules/react-rules.md` for React/frontend-specific standards.

## Git Safety

See `.claude/rules/git-safety.md` for git operation rules and auto-commit policy.

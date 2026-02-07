# OMNI - Core Development Rules

> **STATUS: MANDATORY** | **Version: 5.2** | **Updated: 2026-02-05**
>
> This file is the **IMMUTABLE CONSTITUTION** for ALL agents.
> These rules apply GLOBALLY - cannot be overridden by child rules.

---

## ⚠️ CRITICAL RULES TABLE (MANDATORY - READ FIRST)

**BEFORE doing ANY work, verify you understand these critical rules:**

| Rule              | Requirement                     | Verification                   |
| ----------------- | ------------------------------- | ------------------------------ |
| **SRP**           | One function = one purpose      | Check before writing code      |
| **DRY**           | No duplicated logic             | Extract to shared utilities    |
| **OOP**           | Proper encapsulation            | Use interfaces and composition |
| **300 Lines**     | Max per file (500 for models)   | Split if exceeding             |
| **Commit ALL**    | Include ALL changed files       | No cherry-picking allowed      |
| **Push Always**   | Push after every commit         | User expects remote sync       |
| **Parallel Work** | Don't wait idle for delegations | Fire agents, continue working  |

**REMINDER TRIGGERS:**

- Starting a new task → Re-read this checklist
- About to commit → Verify ALL files included
- Delegating task → Fire in background, continue immediately
- Editing code → Check SRP/DRY/OOP compliance

---

## Rules Hierarchy

```
AGENTS.md (ROOT)              ← YOU ARE HERE - Critical invariants
    ├── backend/AGENTS.md     ← Go/Backend context when working in backend/
    ├── frontend/AGENTS.md    ← React Frontend context (primary app)
    ├── frontend-vue/AGENTS.md← Vue Frontend context (legacy app)
    ├── mcp-servers/AGENTS.md ← MCP context when working in mcp-servers/
    │
    ├── SISYPHUS_RULES.md     ← Orchestrator-specific rules
    ├── EXECUTOR_RULES.md     ← Executor-specific rules
    └── PROMETHEUS_RULES.md   ← Planner-specific rules

Skills (auto-loaded via oh-my-opencode.json):
    ├── react-frontend-rules  ← React 19 + Ant Design specs
    ├── playwright-layout-eval← UI/Layout verification with Playwright
    ├── stuck-recovery        ← Subagent timeout/failure recovery
    └── [agent]-rules         ← Agent-specific rules (oracle, librarian, etc.)
```

**Precedence:** Root invariants > Directory context > Agent-role rules > Skills

---

## Agent Rules Distribution

| Agent               | Additional Rules                                  |
| ------------------- | ------------------------------------------------- |
| **Sisyphus**        | `SISYPHUS_RULES.md` + skill: stuck-recovery       |
| **Sisyphus-Junior** | `EXECUTOR_RULES.md` + skill: stuck-recovery       |
| **Atlas**           | `EXECUTOR_RULES.md` + skill: react-frontend-rules |
| **Hephaestus**      | `EXECUTOR_RULES.md` + skill: react-frontend-rules |
| **Prometheus**      | `PROMETHEUS_RULES.md`                             |
| **Oracle**          | skill: oracle-rules                               |
| **Librarian**       | skill: librarian-rules                            |
| **Explore**         | skill: explore-rules                              |
| **Metis**           | skill: metis-rules                                |
| **Momus**           | skill: momus-rules                                |

### React Migration Skills

| Skill                    | Purpose                                      |
| ------------------------ | -------------------------------------------- |
| `react-frontend-rules`   | React 19 + Ant Design + Ginee-style specs    |
| `playwright-layout-eval` | UI/Layout verification against Design System |

**Note:** For visual-engineering category, both skills auto-loaded.

---

## 1. Critical Invariants (MUST NOT BE VIOLATED)

### ENGLISH ONLY (Code & UI)

All code, comments, variable names, UI labels, and menu items MUST be in English.

```
❌ WRONG: const pesananBaru = []; // Ambil data pesanan
❌ WRONG: <button>Simpan Data</button>
✅ CORRECT: const newOrders = []; // Fetch order data
✅ CORRECT: <button>Save Data</button>
```

### NO FALSE POSITIVES

```
❌ WRONG: success: true + error message
✅ CORRECT: success: false + proper HTTP status code
```

### NO DEFAULT TENANT

```
❌ WRONG: if tenantID == "" { tenantID = "default" }
✅ CORRECT: if tenantID == "" { return error "Missing tenant_id" }
```

### JSON = snake_case

All JSON response/request MUST use snake_case. Backend and Frontend must match.

### Max Lines Per File (CODE ONLY)

| Type       | Limit     |
| ---------- | --------- |
| General    | 300 lines |
| Models     | 500 lines |
| Migrations | Unlimited |
| Docs/Rules | Unlimited |

### Code Quality Principles (MANDATORY)

All code MUST follow these principles:

| Principle      | Rule                                                                   |
| -------------- | ---------------------------------------------------------------------- |
| **DRY**        | Don't Repeat Yourself - extract duplicated code into functions/modules |
| **SRP**        | Single Responsibility Principle - one function/class = one purpose     |
| **OOP**        | Use proper encapsulation, inheritance where appropriate                |
| **Clean Code** | Remove dead code, unused imports, commented-out code                   |

```
❌ WRONG: Copy-paste same logic in 3 places
✅ CORRECT: Extract to shared utility function

❌ WRONG: Function that does 5 different things
✅ CORRECT: Split into 5 focused functions

❌ WRONG: Leave unused variables and imports
✅ CORRECT: Clean up before commit
```

---

## 2. Architecture Pattern

```
Request → Handler → Service → Repository → Database
```

| Layer          | Responsibility                                 |
| -------------- | ---------------------------------------------- |
| **Handler**    | Parse, validate, call service, format response |
| **Service**    | Business logic, orchestration, transformations |
| **Repository** | Database access ONLY                           |

**CRITICAL:** NO business logic in Handler layer.

---

## 3. Testing Requirements

```bash
# MUST pass before task complete
go build ./...
go test ./...
```

---

## 4. SDK Priority (External APIs)

```
1. LOCAL SDK FIRST    → backend/*sdk*/
2. Existing Pattern   → internal/
3. External Docs      → ONLY if stuck
```

| Platform | Local Path            |
| -------- | --------------------- |
| Shopee   | `backend/shopee-sdk/` |
| Lazada   | `backend/lazada-sdk/` |
| TikTok   | `backend/tiktok_sdk/` |

---

## 5. Git Operations

| Allowed      | Forbidden          |
| ------------ | ------------------ |
| `git add`    | `git reset --hard` |
| `git commit` | `git rebase`       |
| `git push`   | `git push --force` |
|              | `git branch -D`    |

### Commit Policy (MANDATORY - NO PERMISSION NEEDED)

- **AUTO-COMMIT**: Commit ALL changed files after EVERY completed task
- **AUTO-PUSH**: Push immediately after commit - NO confirmation required
- Sensitive files already handled by user via .gitignore
- NEVER use destructive git commands that could damage remote repository

**Workflow:**

```
Task Complete → git add -A → git commit → git push → Done
```

---

## 6. Database Configuration

```
Default: SQLite (development)
  Path: backend/data/ (per-tenant databases)

PostgreSQL (production - via env vars):
  DB_DRIVER=postgres
  PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE
```

---

## 7. Delegation Quick Reference

| Situation             | Delegate To                   |
| --------------------- | ----------------------------- |
| Large task/feature    | Prometheus → Sisyphus         |
| Search in codebase    | @explore                      |
| Search external docs  | @librarian                    |
| Architecture question | @oracle                       |
| UI/Frontend           | category="visual-engineering" |
| Quick fix             | category="quick"              |

---

## 8. Response Format

```
SUCCESS: { "success": true, "data": ... }
ERROR:   { "success": false, "error": "message" }
```

---

## 9. Subagent Stuck Recovery

| Signal                   | Action                           |
| ------------------------ | -------------------------------- |
| `Tool execution aborted` | Retry with session_id            |
| Empty response           | Wait 5s, check background_output |
| 3x same error            | Fallback to other agent          |

Fallback chain:

- oracle → librarian → manual
- librarian → explore → manual
- explore → grep/glob → manual

---

## 10. Anti-Patterns (GLOBAL - ALL AGENTS)

| Forbidden               | Do Instead              |
| ----------------------- | ----------------------- |
| `as any`, `@ts-ignore`  | Fix types properly      |
| Empty catch blocks      | Handle or log error     |
| Delete failing tests    | Fix the code            |
| Shotgun debugging       | Trace flow first        |
| `success: true` + error | Use proper status codes |

---

**For detailed context:** See directory-specific `AGENTS.md` files.

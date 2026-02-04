# OMNI - Core Development Rules

> **STATUS: MANDATORY** | **Version: 5.1** | **Updated: 2026-02-04**
>
> This file is the **IMMUTABLE CONSTITUTION** for ALL agents.
> These rules apply GLOBALLY - cannot be overridden by child rules.

---

## Rules Hierarchy

```
AGENTS.md (ROOT)              ← YOU ARE HERE - Critical invariants
    ├── backend/AGENTS.md     ← Go/Backend context when working in backend/
    ├── frontend/AGENTS.md    ← Vue/Frontend context when working in frontend/
    ├── mcp-servers/AGENTS.md ← MCP context when working in mcp-servers/
    │
    ├── SISYPHUS_RULES.md     ← Orchestrator-specific rules
    ├── EXECUTOR_RULES.md     ← Executor-specific rules
    └── PROMETHEUS_RULES.md   ← Planner-specific rules
```

**Precedence:** Root invariants > Directory context > Agent-role rules

---

## Agent Rules Distribution

| Agent               | Additional Rules                            |
| ------------------- | ------------------------------------------- |
| **Sisyphus**        | `SISYPHUS_RULES.md` + skill: stuck-recovery |
| **Sisyphus-Junior** | `EXECUTOR_RULES.md` + skill: stuck-recovery |
| **Prometheus**      | `PROMETHEUS_RULES.md`                       |
| **Oracle**          | skill: oracle-rules                         |
| **Librarian**       | skill: librarian-rules                      |
| **Explore**         | skill: explore-rules                        |
| **Metis**           | skill: metis-rules                          |
| **Momus**           | skill: momus-rules                          |

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

### Commit Policy

- Commit ALL files when requested (sensitive files already handled by user via .gitignore)
- Always push after commit when requested
- NEVER use destructive git commands that could damage remote repository

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

## 10. Anti-Patterns (GLOBAL - SEMUA AGENT)

| Forbidden               | Do Instead              |
| ----------------------- | ----------------------- |
| `as any`, `@ts-ignore`  | Fix types properly      |
| Empty catch blocks      | Handle or log error     |
| Delete failing tests    | Fix the code            |
| Shotgun debugging       | Trace flow first        |
| `success: true` + error | Use proper status codes |

---

**For detailed context:** See directory-specific `AGENTS.md` files.

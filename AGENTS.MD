# OMNI - Core Development Rules

> **STATUS: MANDATORY** | **Version: 6.0** | **Updated: 2026-02-15**
>
> This file is the **IMMUTABLE CONSTITUTION** for ALL agents.
> These rules apply GLOBALLY - cannot be overridden by child rules.

---

## ⚠️ CRITICAL RULES TABLE (MANDATORY - READ FIRST)

**BEFORE doing ANY work, verify you understand these critical rules:**

| Rule              | Requirement                            | Verification                    |
| ----------------- | -------------------------------------- | ------------------------------- |
| **SRP**           | One function = one purpose             | Check before writing code       |
| **DRY**           | No duplicated logic                    | Extract to shared utilities     |
| **OOP**           | Proper encapsulation                   | Use interfaces and composition  |
| **~300 Lines**    | Code file quality signal (SRP/DRY/OOP) | If exceeded → review & refactor |
| **Commit ALL**    | Include ALL changed files              | No cherry-picking allowed       |
| **Push Always**   | Push after every commit                | User expects remote sync        |
| **Parallel Work** | Don't wait idle for delegations        | Fire agents, continue working   |

**REMINDER TRIGGERS:**

- Starting a new task → Re-read this checklist
- About to commit → Verify ALL files included
- Delegating task → Fire in background, continue immediately
- Editing code → Check SRP/DRY/OOP compliance
- File > ~300 lines → Review for SRP/DRY/OOP violation

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

| Agent               | Additional Rules                                                   |
| ------------------- | ------------------------------------------------------------------ |
| **Sisyphus**        | `SISYPHUS_RULES.md` + skills: stuck-recovery, ulw-loop             |
| **Sisyphus-Junior** | `EXECUTOR_RULES.md` + skills: stuck-recovery, react-frontend-rules |
| **Atlas**           | `EXECUTOR_RULES.md` + skills: stuck-recovery, react-frontend-rules |
| **Hephaestus**      | `EXECUTOR_RULES.md` + skills: stuck-recovery, react-frontend-rules |
| **Prometheus**      | `PROMETHEUS_RULES.md`                                              |
| **Oracle**          | skill: oracle-rules                                                |
| **Librarian**       | skill: librarian-rules                                             |
| **Explore**         | skill: explore-rules                                               |
| **Metis**           | skill: metis-rules                                                 |
| **Momus**           | skill: momus-rules                                                 |

### React Migration Skills

| Skill                    | Purpose                                      |
| ------------------------ | -------------------------------------------- |
| `react-frontend-rules`   | React 19 + Ant Design + Ginee-style specs    |
| `playwright-layout-eval` | UI/Layout verification against Design System |

**Note:** For visual-engineering category, both skills auto-loaded.

### Config Variants

Not all configs include every agent. The agent table above is the **full set** (base + copilot + antigravity configs).

| Config          | Agents Available        |
| --------------- | ----------------------- |
| Mix Copilot     | All agents listed above |
| Mix Antigravity | All agents listed above |

> **Note**: All profiles include advisory agents (oracle, librarian, explore, metis, momus).
> The fallback matrix below applies when an advisory agent **fails or is unavailable** (e.g., rate limit, timeout).

#### Advisory Agent Fallback Matrix

When an advisory agent is unavailable or fails, use these alternative paths:

| Normal Path         | Mix Mode Fallback                                                  |
| ------------------- | ------------------------------------------------------------------ |
| `@oracle` consult   | `category="deep"` with detailed analysis prompt                    |
| `@librarian` lookup | `category="deep"` with explicit search instructions + web tools    |
| `@explore` search   | Direct `grep`/`glob`/`ast_grep_search` tools                       |
| `@metis` analysis   | Orchestrator performs pre-planning inline (follow Metis framework) |
| `@momus` review     | Orchestrator performs plan review inline (follow Momus criteria)   |
| Deep also fails     | Ask user for guidance — do NOT loop endlessly                      |

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

> **Protected endpoint** = any endpoint that reads/writes tenant-specific data.
> Only truly public endpoints (health checks, login, static assets) are exempt.

```
❌ WRONG: if tenantID == "" { tenantID = "default" }
✅ CORRECT: if tenantID == "" { return error "Missing tenant_id" }
```

### JSON = snake_case

All JSON response/request MUST use snake_case. Backend and Frontend must match.

### ~300 Lines Per Code File (Quality Signal)

> **~300 lines is NOT a hard limit.** It's a quality signal that MUST trigger a refactor attempt.
> If a code file exceeds ~300 lines, you MUST attempt to refactor it (extract helpers, split by responsibility, remove dead code).
> After a genuine refactor effort, if the minimum achievable is slightly above 300 (e.g. 310-330) and the code satisfies SRP/DRY/OOP with no dead code — that's acceptable.
> This is NOT a license for 400+ line files. If your file is 400+ lines, you haven't refactored hard enough.

| Applies To                             | Rule                                           |
| -------------------------------------- | ---------------------------------------------- |
| Code files (.go, .ts, .tsx, .js, .vue) | ~300 lines — MUST refactor if exceeded         |
| Code files after refactor              | Slightly above 300 OK if SRP/DRY/OOP satisfied |
| Code files 400+ lines                  | NOT acceptable — refactor harder or split      |
| Migrations                             | No limit                                       |
| Docs / Rules / Config / JSON           | No limit                                       |

**AFTER EVERY file edit, ask yourself:**

1. Is this file doing more than one thing? → Split by responsibility
2. Is there duplicated logic? → Extract to shared utility
3. Is there dead code / unused imports? → Remove them
4. Is there commented-out code? → Remove it

**How to split when file is too long:**

```
1. Extract helper functions    → {module}_helpers.go / utils.ts
2. Extract types/interfaces    → {module}_types.go / types.ts
3. Extract constants           → {module}_constants.go / constants.ts
4. Split by responsibility     → order_create.go + order_list.go + order_export.go
```

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

### Understand Flow BEFORE Acting

> Before fixing any bug or implementing features, research the RELEVANT flow for the topic.
> Never guess — trace with evidence. If you fix the same thing twice, STOP: you don't understand the flow.

### No Premature Done

> Before claiming completion, self-check: (1) ALL changes verified with evidence? (2) Any remaining TODOs?
> (3) Did I TEST or just ASSUME? (4) Re-read original task — fully addressed?
> If ANY answer is "no" → NOT done, keep working.

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

## 3. Build, Test & Deploy

### Backend Testing

```bash
# MUST pass before task complete
go build ./...
go test ./...
```

### Test Failure Policy

- **Your changes caused a test failure** → MUST fix before task is complete
- **Pre-existing test failures** → Document them explicitly, prove your changes did NOT introduce or worsen them
- **No test file exists for changed code** → MUST create test file with meaningful tests covering your changes
- **Never delete or skip failing tests** to make the build "pass"

### Frontend Testing

```bash
# React frontend (primary app)
npm run build    # in frontend/
npm run lint     # in frontend/

# Vue frontend (legacy app)
npm run build    # in frontend-vue/
```

### Docker Build & Deploy (build.py)

Use `build.py` for all Docker operations:

```bash
python build.py smart      # Smart build (RECOMMENDED for code changes)
python build.py quickfix    # Fix service issues without rebuild
python build.py full        # Full rebuild from scratch (no cache)
```

### Database Backup on Schema Changes (MANDATORY)

When ANY database schema is modified (migrations, DDL changes, new tables, altered columns):

```bash
# MANDATORY: Run backup AFTER schema changes are applied via build.py
python build.py backup
```

> **Why AFTER?** The backup captures the latest schema so restore always has the most recent structure.
> This ensures the team always has a backup file reflecting the current database state.

| Trigger                   | Action                                             |
| ------------------------- | -------------------------------------------------- |
| New migration file added  | `python build.py smart` → `python build.py backup` |
| Table/column altered      | `python build.py smart` → `python build.py backup` |
| Code-only change (no DDL) | `python build.py smart` (no backup needed)         |

---

## 4. SDK Priority (External APIs)

```
1. LOCAL SDK FIRST    → backend/*sdk*/
2. Existing Pattern   → internal/
3. External Docs      → WHEN LOCAL IS INSUFFICIENT
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
- **NEVER commit secrets**: `.env` files, API keys, passwords, credentials, tokens. If unsure, ask the user.

> **Exception — Hephaestus**: Hephaestus (testing agent) MUST NOT auto-commit or auto-push.
> Hephaestus requires **explicit user confirmation** before any git commit/push.
> This is intentional — testing results must be reviewed before committing.
> All other agents follow the auto-commit/auto-push policy above.

> **Task vs Batch**: A "task" = one logical unit of work (e.g., fix a bug, add a feature, update docs).
> In loop mode (ULW), the orchestrator may batch multiple small tasks into one commit — this is acceptable.
> The rule is: **never leave completed work uncommitted**.

**Workflow:**

```
Task Complete → git add -A → git commit → git push → Done
```

### Screenshot Evidence Policy (GLOBAL)

> **All screenshots (any agent, any purpose) MUST be saved under `.sisyphus/evidence/`.**
> Never save screenshots to root folder or random locations.

```
.sisyphus/evidence/
├── task-{N}-{component}-desktop.png
├── task-{N}-{component}-mobile.png
├── bug-{page}-{issue}.png
└── verification-{feature}.png
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

> **Launch delegations aggressively for maximum throughput.** Don't hold back — fire multiple agents in parallel.
> **MAXIMUM 3 background delegations RUNNING simultaneously.** System queues excess tasks automatically (`defaultConcurrency: 3`).
> This limit is **orchestrator-controlled** (Sisyphus manages the pipeline).
> Executors do NOT self-delegate unless instructed by the orchestrator.

<!-- MASTER:delegation-routing-compact -->
| Situation             | Delegate To                   |
| --------------------- | ----------------------------- |
| Large task/feature    | Prometheus → Sisyphus         |
| Search in codebase    | `@explore`                    |
| Search external docs  | `@librarian`                  |
| Architecture question | `@oracle`                     |
| UI/Frontend           | category="visual-engineering" |
| Quick fix             | category="quick"              |
<!-- /MASTER:delegation-routing-compact -->

### Delegation Failure Escalation (MANDATORY)

<!-- MASTER:failure-escalation -->
When a delegated task **fails or produces incorrect results** (NOT due to connection loss or timeout):

| Failure Type             | Action                                                     |
| ------------------------ | ---------------------------------------------------------- |
| **Timeout / connection** | Retry with `session_id` in same category                   |
| **Wrong output / error** | **MUST retry using `category="deep"`** on the same task    |
| **Deep also fails**      | Escalate to `@oracle` for analysis, then retry or ask user |

> **Why `deep`?** The `deep` category uses a stronger reasoning model with autonomous problem-solving.
> It performs thorough research before acting — ideal for tasks that lighter categories failed on.
> This prevents wasting retries on the same weak model that already failed.

> **Fallback note:** If `@oracle` is unavailable (rate limit, timeout), use `category="deep"` with detailed analysis prompt instead. See AGENTS.md § Advisory Agent Fallback Matrix.
<!-- /MASTER:failure-escalation -->

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

<!-- MASTER:fallback-chain -->
- oracle → librarian → manual
- librarian → explore + websearch → manual
- explore → grep/glob → manual
<!-- /MASTER:fallback-chain -->

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

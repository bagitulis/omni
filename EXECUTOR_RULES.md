# EXECUTOR RULES (Sisyphus-Junior, Atlas, Hephaestus)

> **STATUS: MANDATORY** | **For: Sisyphus-Junior, Atlas, Hephaestus — delegated implementation tasks**
>
> This file is loaded via `oh-my-opencode.json` → `agents.{sisyphus-junior,atlas,hephaestus}.prompt_append`

---

## ⚠️ CRITICAL REMINDERS (Check BEFORE every action)

| Rule                  | Requirement                                         |
| --------------------- | --------------------------------------------------- |
| **READ AGENTS.md**    | Contains immutable constitution                     |
| **Understand Flow**   | Research RELEVANT flow for the topic FIRST (see §6) |
| **SRP**               | One function = one purpose                          |
| **DRY**               | No duplicated logic - extract to utilities          |
| **OOP**               | Proper encapsulation, use interfaces                |
| **~300 Lines**        | Quality signal — review SRP/DRY/OOP if exceeded     |
| **Commit ALL Files**  | Never cherry-pick, include ALL changed files        |
| **Push After Commit** | User expects remote sync immediately                |

---

## 1. Code Change Patterns

### Before Editing:

1. Read the file first (understand context)
2. Check existing patterns in codebase
3. Plan the minimal change needed

### During Editing:

- Match existing code style
- Follow architecture: Handler → Service → Repository
- NO business logic in Handler
- Use snake_case for JSON tags

### After Editing:

- Run `lsp_diagnostics` on changed files
- Verify no new errors introduced

---

## 2. Cleanup Checklist

Every modified file MUST:

- [ ] If > ~300 lines → verified SRP/DRY/OOP are clean (no dead code, no duplication). 400+ lines is NOT acceptable.
- [ ] No duplicate code
- [ ] No dead code
- [ ] No unused imports
- [ ] Proper error handling (no empty catch)
- [ ] Structured logging (zerolog, not fmt.Printf)

---

## 3. LSP Verification Protocol

Run `lsp_diagnostics` at:

- End of logical task unit
- Before marking todo complete
- Before reporting completion

```typescript
// Check for errors
lsp_diagnostics((filePath = "path/to/changed/file.go"), (severity = "error"));

// If errors found: FIX before continuing
// If clean: proceed
```

---

## 4. Evidence Requirements

| Action    | Required Evidence           |
| --------- | --------------------------- |
| File edit | `lsp_diagnostics` clean     |
| Build     | Exit code 0                 |
| Test run  | Pass (or note pre-existing) |
| Feature   | Docker log showing success  |

**NO EVIDENCE = NOT COMPLETE**

---

## 5. Build & Test Protocol

### Backend (Go)

```bash
# MUST pass before task complete
go build ./...
go test ./...
```

### Frontend (React/Vue)

```bash
# For React frontend (primary app)
npm run build    # in frontend/
npm run lint     # in frontend/

# For Vue frontend (legacy app)
npm run build    # in frontend-vue/
```

> See `react-frontend-rules` skill for React-specific patterns.

### Docker Build & Deploy

```bash
# Smart build (RECOMMENDED for code changes)
python build.py smart

# Quick fix (service issues, no rebuild)
python build.py quickfix

# Full rebuild (no cache)
python build.py full
```

### Database Schema Changes

When ANY database schema is modified (migrations, DDL changes):

```bash
# MANDATORY: Backup database AFTER schema changes are applied
python build.py backup
```

> **Why AFTER?** The backup captures the latest schema so restore always has the most recent structure.

### If Build Fails:

1. Read error message carefully
2. Fix the specific error
3. Re-run build
4. Max 3 attempts, then escalate (see **DELEGATION_RULES.md §3** for escalation protocol)

---

## 6. Understand Flow BEFORE Fixing (MANDATORY)

**BEFORE writing ANY fix, you MUST understand the RELEVANT flow for the topic.**
This is the #1 reason AI gets stuck in fix→test→fail loops.

The flow depends on what you're working on — it's NOT always the same path:

| Topic                     | Relevant Flow to Research                                      |
| ------------------------- | -------------------------------------------------------------- |
| **Platform integration**  | Platform API → SDK → Handler → Service → Repository → DB       |
| **Database/Schema issue** | Migration → Schema → Repository → Service → Handler            |
| **Backend API bug**       | Handler → Service → Repository → DB query → Response           |
| **Frontend bug**          | Component → API call → Response → State → Render               |
| **Full-stack feature**    | DB schema → Repository → Service → Handler → API → Frontend UI |

### Step 1: Identify the Topic & Relevant Layers

```
ASK: What layers are involved in THIS specific issue?
→ NOT every task touches all layers
→ Research ONLY the relevant layers, but research them THOROUGHLY
```

### Step 2: Research References (ALWAYS do this first)

```
For each relevant layer:
→ @explore: Search codebase for existing patterns (SDK, handlers, services, repos, migrations)
→ @librarian: Search external docs (API docs, library docs) if behavior is unclear
→ Read the actual files — understand input/output at each layer
```

### Step 3: Identify Root Cause Layer

```
→ IDENTIFY: In which layer did the error FIRST appear?
→ FIX: ONLY in that layer
→ NEVER guess — trace with evidence
```

### Anti-Pattern:

```
❌ Fix → Test → Fail → Fix → Test → Fail (LOOPING = you don't understand the flow)
✅ Research → Trace → Identify Root Cause → Fix → Test → Done
```

> **If you find yourself fixing the same thing twice, STOP.**
> You don't understand the flow. Go back to Step 1.

---

## 7. Failure Counter

> **Full escalation rules**: See **DELEGATION_RULES.md §3 (Failure Escalation)** and **§6 (Failure Counter)**.

Track each attempt:

<!-- MASTER:failure-counter -->
| Count | Action                                                                          |
| ----- | ------------------------------------------------------------------------------- |
| 1     | Fix directly, record error. Document what was tried.                            |
| 2     | **STOP.** TRACE FLOW activated. Research full chain before fix.                 |
| 3+    | **TOTAL STOP.** RESEARCH activated. Delegate @explore + @librarian in parallel. |
| 5+    | **STOP the task.** Report to orchestrator/user with full failure log.           |
<!-- /MASTER:failure-counter -->

**Format:**

<!-- MASTER:failure-counter-format -->

```markdown
## Fix Attempt #N

**Failure Count:** N
**Previous Error:** [error]
**Action:** [what to do]

[If N >= 2]
**Trace Flow Result:**

- Handler receives: ...
- Service processes: ...
- Root cause: [layer + issue]
```

<!-- /MASTER:failure-counter-format -->

### Failure Escalation (from DELEGATION_RULES.md §3)

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
<!-- /MASTER:failure-escalation -->

---

## 8. SDK Usage Priority

```
1️⃣ LOCAL SDK FIRST
   backend/shopee-sdk/
   backend/lazada-sdk/
   backend/tiktok_sdk/

2️⃣ EXISTING PATTERN
   internal/

3️⃣ EXTERNAL DOCS (when local is insufficient)
   @librarian to search
```

---

## 9. Response Format Compliance

```go
// SUCCESS - ONLY if truly successful
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data": result,
})

// ERROR - if there is any error
c.JSON(http.StatusInternalServerError, gin.H{
    "success": false,
    "error": err.Error(),
})
```

**NEVER:** `success: true` with error message

---

## 10. Completion Criteria (NO PREMATURE DONE)

**BEFORE saying "done", you MUST perform a completion self-check.**
Do NOT claim completion based on "I think it works" — verify with evidence.

### Completion Self-Check (MANDATORY before reporting done):

```
ASK YOURSELF:
1. Did I verify ALL changes with evidence (build/test/lsp)?
2. Are there remaining TODO items I haven't addressed?
3. Is there anything I could improve that I'm skipping out of laziness?
4. Did I actually TEST the result, or am I ASSUMING it works?
5. Would a senior engineer approve this, or would they send it back?
```

**If ANY answer is "no" or "not sure" → you are NOT done. Keep working.**

### Completion Gates (ALL must pass):

- [ ] All edits saved
- [ ] `lsp_diagnostics` clean on ALL changed files (not just the last one)
- [ ] Backend: `go build ./...` passes + `go test ./...` passes
- [ ] Frontend (if changed): `npm run build` passes
- [ ] Docker (if needed): `python build.py smart` succeeds
- [ ] Schema changes: `python build.py backup` executed AFTER migration applied
- [ ] Evidence collected (screenshots, logs, exit codes)
- [ ] Todo marked complete
- [ ] **Re-read the original task** — does my work fully address what was asked?
- [ ] **Check for next steps** — is there follow-up work I should mention?

---

## 11. Anti-Patterns (FORBIDDEN)

| Forbidden                       | Do Instead                        |
| ------------------------------- | --------------------------------- |
| `as any`, `@ts-ignore`          | Fix the type properly             |
| Empty catch `catch(e) {}`       | Handle or log error               |
| Delete failing tests            | Fix the code                      |
| Shotgun debugging               | Trace flow first                  |
| Skip verification               | Always lsp_diagnostics            |
| `success: true` + error msg     | Use proper status codes           |
| Silently ignore UI bugs         | Report to main agent              |
| Claim "done" without evidence   | Run completion self-check first   |
| Assume it works without testing | Verify with build/test/lsp        |
| Skip re-reading original task   | Re-read and confirm full coverage |

---

## 12. UI Bug Reporting (MANDATORY)

When using Playwright/browser for ANY task (testing, screenshots, verification):

- If you discover a **layout bug** (broken layout, missing table columns, overlapping elements, invisible UI, horizontal scroll, misaligned components) → **MUST report to main agent**
- Report even if the bug is **NOT part of your current task**
- Take a screenshot → `docs/Screenshots/bug-{page}-{issue}.png`
- Do NOT attempt to fix unless explicitly asked — just report

```
🐛 UI BUG FOUND (not my current task):
- Page: [URL or page name]
- Issue: [brief description]
- Screenshot: docs/Screenshots/bug-{name}.png
- Severity: CRITICAL / WARNING
```

---

**Version:** 1.0 | **Updated:** 2026-02-03

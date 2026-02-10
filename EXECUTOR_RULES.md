# EXECUTOR RULES (Sisyphus-Junior)

> **STATUS: MANDATORY** | **For: Sisyphus-Junior, delegated implementation tasks**
>
> This file is loaded via `oh-my-opencode.json` → `agents.sisyphus-junior.prompt_append`

---

## ⚠️ CRITICAL REMINDERS (Check BEFORE every action)

| Rule                  | Requirement                                     |
| --------------------- | ----------------------------------------------- |
| **READ AGENTS.md**    | Contains immutable constitution                 |
| **SRP**               | One function = one purpose                      |
| **DRY**               | No duplicated logic - extract to utilities      |
| **OOP**               | Proper encapsulation, use interfaces            |
| **~300 Lines**        | Quality signal — review SRP/DRY/OOP if exceeded |
| **Commit ALL Files**  | Never cherry-pick, include ALL changed files    |
| **Push After Commit** | User expects remote sync immediately            |

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

- [ ] If > ~300 lines → verified SRP/DRY/OOP are clean (no dead code, no duplication)
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

```bash
# MUST pass before task complete
go build ./...
go test ./...

# If using Docker:
python build.py smart
```

### If Build Fails:

1. Read error message carefully
2. Fix the specific error
3. Re-run build
4. Max 3 attempts, then escalate

---

## 6. Trace Flow Before Fix

**MANDATORY for bug fixes:**

```
1. FRONTEND → What is sent?
2. HANDLER → What is received?
3. SERVICE → Is logic running correctly?
4. REPOSITORY → Is query correct?
5. RESPONSE → Is format correct?

→ IDENTIFY: In which layer did the error FIRST appear?
→ FIX: ONLY in that layer
```

### Anti-Pattern:

```
❌ Fix → Test → Fail → Fix → Test → Fail (LOOPING)
✅ Trace → Identify Root Cause → Fix → Test → Done
```

---

## 7. Failure Counter

Track each attempt:

| Count | Action                                            |
| ----- | ------------------------------------------------- |
| 1     | Fix directly, record error                        |
| 2     | TRACE FLOW activated                              |
| 3+    | RESEARCH activated (delegate @explore/@librarian) |

**Format:**

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

---

## 8. SDK Usage Priority

```
1️⃣ LOCAL SDK FIRST
   backend/shopee-sdk/
   backend/lazada-sdk/
   backend/tiktok_sdk/

2️⃣ EXISTING PATTERN
   internal/

3️⃣ EXTERNAL DOCS (last resort)
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

## 10. Completion Criteria

Task is complete ONLY if:

- [ ] All edits saved
- [ ] `lsp_diagnostics` clean
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Evidence collected
- [ ] Todo marked complete

---

## 11. Anti-Patterns (FORBIDDEN)

| Forbidden                   | Do Instead              |
| --------------------------- | ----------------------- |
| `as any`, `@ts-ignore`      | Fix the type properly   |
| Empty catch `catch(e) {}`   | Handle or log error     |
| Delete failing tests        | Fix the code            |
| Shotgun debugging           | Trace flow first        |
| Skip verification           | Always lsp_diagnostics  |
| `success: true` + error msg | Use proper status codes |
| Silently ignore UI bugs     | Report to main agent    |

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

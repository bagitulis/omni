# EXECUTOR RULES (Sisyphus-Junior)

> **STATUS: MANDATORY** | **For: Sisyphus-Junior, delegated implementation tasks**
>
> Rules ini di-load via `opencode.json` → `agents.sisyphus-junior.prompt_append`

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

Setiap file yang diubah WAJIB:

- [ ] < 300 lines (models: 500)
- [ ] No duplicate code
- [ ] No dead code
- [ ] No unused imports
- [ ] Proper error handling (no empty catch)
- [ ] Structured logging (zerolog, bukan fmt.Printf)

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
# WAJIB sebelum task complete
go build ./...
go test ./...

# Jika ada Docker:
python build.py smart
```

### Jika Build Gagal:

1. Read error message carefully
2. Fix the specific error
3. Re-run build
4. Max 3 attempts, lalu escalate

---

## 6. Trace Flow Before Fix

**WAJIB untuk bug fixes:**

```
1. FRONTEND → Apa yang dikirim?
2. HANDLER → Apa yang diterima?
3. SERVICE → Logic berjalan benar?
4. REPOSITORY → Query correct?
5. RESPONSE → Format benar?

→ IDENTIFY: Di layer mana error PERTAMA KALI muncul?
→ FIX: HANYA di layer tersebut
```

### Anti-Pattern:

```
❌ Fix → Test → Gagal → Fix → Test → Gagal (LOOPING)
✅ Trace → Identify Root Cause → Fix → Test → Done
```

---

## 7. Failure Counter

Track setiap attempt:

| Count | Action                                            |
| ----- | ------------------------------------------------- |
| 1     | Fix langsung, catat error                         |
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
1️⃣ LOCAL SDK DULU
   backend/shopee-sdk/
   backend/lazada-sdk/
   backend/tiktok_sdk/

2️⃣ EXISTING PATTERN
   internal/

3️⃣ EXTERNAL DOCS (last resort)
   @librarian untuk cari
```

---

## 9. Response Format Compliance

```go
// SUCCESS - HANYA jika benar-benar sukses
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data": result,
})

// ERROR - jika ada error apapun
c.JSON(http.StatusInternalServerError, gin.H{
    "success": false,
    "error": err.Error(),
})
```

**NEVER:** `success: true` dengan error message

---

## 10. Completion Criteria

Task selesai HANYA jika:

- [ ] Semua edits saved
- [ ] `lsp_diagnostics` clean
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Evidence collected
- [ ] Todo marked complete

---

## 11. Anti-Patterns (DILARANG)

| Forbidden                   | Do Instead              |
| --------------------------- | ----------------------- |
| `as any`, `@ts-ignore`      | Fix the type properly   |
| Empty catch `catch(e) {}`   | Handle or log error     |
| Delete failing tests        | Fix the code            |
| Shotgun debugging           | Trace flow first        |
| Skip verification           | Always lsp_diagnostics  |
| `success: true` + error msg | Use proper status codes |

---

**Version:** 1.0 | **Updated:** 2026-02-03

# OMNI - Core Development Rules

> **STATUS: MANDATORY** | **Version: 4.0** | **Updated: 2026-02-03**
>
> File ini adalah CORE RULES untuk SEMUA agent.
> Untuk rules spesifik, lihat file terpisah atau skill yang di-load.

---

## Rules Distribution

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

## 1. Critical Rules (WAJIB - TIDAK BOLEH DILANGGAR)

### NO FALSE POSITIVES

```go
// ❌ WRONG
c.JSON(http.StatusOK, gin.H{"success": true, "message": "Failed: " + err.Error()})

// ✅ CORRECT
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
```

### NO DEFAULT TENANT

```go
// ❌ WRONG
if tenantID == "" { tenantID = "default" }

// ✅ CORRECT
if tenantID == "" {
    c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
    return
}
```

### JSON = snake_case

```go
type Order struct {
    OrderSN string `json:"order_sn"`  // ✅ CORRECT
    // OrderSN string `json:"orderSn"` // ❌ WRONG
}
```

### Max Lines Per File

| Type       | Limit     |
| ---------- | --------- |
| General    | 300 lines |
| Models     | 500 lines |
| Migrations | Unlimited |

---

## 2. Architecture Pattern

```
Request → Handler → Service → Repository → Database
            ↓          ↓           ↓
         Parse      Logic       Query
         Validate   Process     CRUD
         Response   Transform
```

| Layer          | Responsibility                                 |
| -------------- | ---------------------------------------------- |
| **Handler**    | Parse, validate, call service, format response |
| **Service**    | Business logic, orchestration, transformations |
| **Repository** | Database access ONLY                           |

---

## 3. Testing Requirements

```bash
# WAJIB pass sebelum task complete
go build ./...
go test ./...
```

---

## 4. SDK Priority (External APIs)

```
1. LOCAL SDK DULU    → backend/*sdk*/
2. Existing Pattern  → internal/
3. External Docs     → HANYA jika stuck
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

---

## 6. Database Configuration

```
Default: SQLite (development)
  Path: backend/data/ (per-tenant databases)

PostgreSQL (production - via env vars):
  DB_DRIVER=postgres
  PG_HOST=localhost
  PG_PORT=5432
  PG_USER=omni
  PG_PASSWORD=(from PG_PASSWORD env)
  PG_DATABASE=omni_main
  PG_SSLMODE=disable
```

---

## 7. Delegation Quick Reference

| Situasi               | Delegate Ke                   |
| --------------------- | ----------------------------- |
| Task besar/feature    | Prometheus → Sisyphus         |
| Cari di codebase      | @explore                      |
| Cari docs external    | @librarian                    |
| Architecture question | @oracle                       |
| UI/Frontend           | category="visual-engineering" |
| Quick fix             | category="quick"              |

---

## 8. Response Format

```go
// SUCCESS
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data": result,
})

// ERROR
c.JSON(http.StatusInternalServerError, gin.H{
    "success": false,
    "error": "message",
})
```

---

## 9. Subagent Stuck Recovery

Jika subagent timeout/stuck:

| Signal                   | Action                           |
| ------------------------ | -------------------------------- |
| `Tool execution aborted` | Retry dengan session_id          |
| Empty response           | Wait 5s, check background_output |
| 3x same error            | Fallback ke agent lain           |

Fallback chain:

- oracle → librarian → manual
- librarian → explore → manual
- explore → grep/glob → manual

---

**Full details: Lihat file rules spesifik per agent**

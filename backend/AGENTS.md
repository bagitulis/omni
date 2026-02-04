# Backend Development Context

> Auto-injected when working in `backend/` directory.
> **Parent rules:** See root `AGENTS.md` for critical rules.

---

## Stack

- Go 1.21+
- Gin framework (NOT Fiber)
- GORM
- PostgreSQL (multi-tenant, schema-based)
- zerolog for logging

---

## Architecture Pattern

```
Request → Handler → Service → Repository → Database
```

| Layer      | Location                 | Responsibility                           |
| ---------- | ------------------------ | ---------------------------------------- |
| Handler    | `internal/handlers/`     | Parse request, validate, format response |
| Service    | `internal/services/`     | Business logic, orchestration            |
| Repository | `internal/repositories/` | Database access ONLY                     |

**CRITICAL:** NO business logic in Handler layer.

---

## SDK Locations (CARI DI SINI DULU!)

| Platform   | Path                             | Key Files                         |
| ---------- | -------------------------------- | --------------------------------- |
| Shopee     | `backend/shopee-sdk/`            | orders.go, products.go, client.go |
| Lazada     | `backend/lazada-sdk/`            | order.go, product.go, auth.go     |
| Lazada IOP | `backend/lazada_sdk/iop-sdk-go/` | Official IOP SDK                  |
| TikTok     | `backend/tiktok_sdk/`            | Comprehensive SDK (100+ files)    |

---

## Naming Conventions

| Type          | Convention     | Example           |
| ------------- | -------------- | ----------------- |
| Struct fields | PascalCase     | `OrderSN`         |
| JSON tags     | **snake_case** | `json:"order_sn"` |
| Local vars    | camelCase      | `orderList`       |
| Constants     | PascalCase     | `MaxRetries`      |
| Packages      | lowercase      | `handlers`        |

---

## Critical Rules

### NO FALSE POSITIVES

```go
// ❌ WRONG - success true tapi ada error
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

---

## Logging (zerolog ONLY)

```go
// ✅ CORRECT
log.Info().Str("order_sn", order.SN).Msg("Order created")
log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get order")

// ❌ WRONG
fmt.Printf("Order created: %s\n", order.SN)
```

---

## Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to get order: %w", err)
}

// ❌ NEVER empty catch
// if err != nil { } ← DILARANG
```

---

## Database (GORM)

- Use `context.Context` for all DB operations
- Tenant isolation via schema: `tenant_{tenantID}`
- Connection pooling configured in config

```go
// ✅ CORRECT - with context
db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&orders)

// ❌ WRONG - tanpa context
db.Where("tenant_id = ?", tenantID).Find(&orders)
```

---

## Response Format

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

## Testing

```bash
# WAJIB pass sebelum task complete
go build ./...
go test ./...
```

---

## Anti-Patterns

| Forbidden                 | Do Instead                    |
| ------------------------- | ----------------------------- |
| Business logic in Handler | Move to Service layer         |
| `fmt.Printf` for logging  | Use zerolog                   |
| Empty error handling      | Always handle or wrap errors  |
| DB calls tanpa context    | Always use `WithContext(ctx)` |
| camelCase JSON tags       | Use snake_case                |

---
globs: ["**/*.go", "backend/**"]
description: Go/Backend development rules for OMNI project
---

# Go Backend Development Rules

## Architecture Pattern

```
Handler → Service → Repository → Database
```

| Layer      | Responsibility            | Location                 |
| ---------- | ------------------------- | ------------------------ |
| Handler    | Parse, validate, response | `internal/handlers/`     |
| Service    | Business logic            | `internal/services/`     |
| Repository | Database access           | `internal/repositories/` |

---

## Naming Conventions

| Type          | Convention | Example           |
| ------------- | ---------- | ----------------- |
| Struct fields | PascalCase | `OrderSN`         |
| JSON tags     | snake_case | `json:"order_sn"` |
| Local vars    | camelCase  | `orderList`       |
| Constants     | PascalCase | `MaxRetries`      |
| Packages      | lowercase  | `handlers`        |

---

## Critical Rules

### NO FALSE POSITIVES

```go
// ❌ WRONG
c.JSON(http.StatusOK, gin.H{"success": true, "message": "Failed"})

// ✅ CORRECT
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
```

### NO DEFAULT TENANT

```go
// ❌ WRONG
if tenantID == "" { tenantID = "default" }

// ✅ CORRECT
if tenantID == "" {
    return ErrMissingTenantID
}
```

---

## Logging

Use zerolog ONLY:

```go
// ✅ CORRECT
log.Info().Str("order_sn", order.SN).Msg("Order created")

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

// Never empty catch
// ❌ WRONG: if err != nil { }
```

---

## Database (GORM)

- Use context for all operations
- Tenant isolation via schema: `tenant_{tenantID}`
- Connection pooling configured in config

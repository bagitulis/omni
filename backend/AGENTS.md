# Backend Development Context

> Auto-injected when reading files in `backend/` directory.

## Stack

- Go 1.21+
- Gin framework (NOT Fiber)
- GORM
- PostgreSQL (multi-tenant, schema-based)
- zerolog for logging

## Architecture

```
Handler → Service → Repository → Database
```

| Layer      | Location                 | Responsibility            |
| ---------- | ------------------------ | ------------------------- |
| Handler    | `internal/handlers/`     | Parse, validate, response |
| Service    | `internal/services/`     | Business logic            |
| Repository | `internal/repositories/` | Database access ONLY      |

## SDK Locations

| Platform | Path                  |
| -------- | --------------------- |
| Shopee   | `backend/shopee-sdk/` |
| Lazada   | `backend/lazada-sdk/` |
| TikTok   | `backend/tiktok_sdk/` |

## Critical Rules

1. **NO FALSE POSITIVES** - `success: true` hanya untuk sukses
2. **NO DEFAULT TENANT** - Error jika tenant_id kosong
3. **JSON = snake_case** - Semua JSON tags
4. **Max 300 lines** per file (models: 500)

## Database

- Driver: PostgreSQL
- Multi-tenancy: Schema-based (`tenant_{tenantID}`)
- Connection: Via GORM with pooling

## Testing

```bash
go build ./...
go test ./...
```

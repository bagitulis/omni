# Multi-Tenancy Implementation Analysis Report

## Executive Summary

The system implements **SCHEMA-BASED MULTI-TENANCY** with **HYBRID APPLICATION-LEVEL TENANT FILTERING**. This is a robust and secure approach with properly isolated databases per tenant at the PostgreSQL schema level, supplemented by application-level tenant ID validation in most critical areas.

### Architecture Type: **PostgreSQL Schema Isolation + Application Filtering**
- **Database Level**: PostgreSQL schema-per-tenant (e.g., `tenant_yumna_bertigamart`, `tenant_tika_nusseyba`)
- **Application Level**: Tenant ID extraction from JWT and header validation
- **Security Model**: Defense-in-depth with multiple validation layers

---

## 1. ARCHITECTURE OVERVIEW

### 1.1 Database Isolation Strategy

**Location**: `backend/internal/config/database.go`, `backend/internal/config/database_postgres.go`

The system uses PostgreSQL schema isolation as the **primary security boundary**:

```go
// Schema naming convention
schemaName := fmt.Sprintf("tenant_%s", tenantID)

// Connection DSN includes search_path to restrict to specific schema
options='-csearch_path=tenant_{tenantID},public'
```

**Key Mechanism**:
- Each tenant gets a **separate database connection** with `search_path` set to their schema
- Connections are **cached per tenant** in memory
- PostgreSQL enforces schema boundary at database driver level
- All queries from a tenant connection only see tables in their schema

### 1.2 Tenant Context Propagation

**Location**: `backend/internal/middleware/auth.go`, `backend/internal/middleware/tenant.go`

**Flow**:
1. JWT token extracted from `Authorization: Bearer <token>` header
2. JWT decoded to get `tenantID` claim (from `internal/utils/jwt.go`)
3. TenantID set in Gin context via `c.Set("tenantID", claims.TenantID)`
4. Tenant middleware validates tenantID against `tenants.json` config
5. All handlers retrieve tenantID from context and pass to repository layer

### 1.3 Tenant Configuration

**Location**: `backend/internal/config/config.go`

Tenants loaded from `config/static/tenants.json`:
```json
{
  "yumna_bertigamart": {
    "db_path": "config/databases/yumna_bertigamart.db",
    "shop_name": "Bertiga Mart",
    "is_global": false
  }
}
```

---

## 2. KEY COMPONENTS & FILE LOCATIONS

### 2.1 Authentication & Tenant Extraction

| Component | Location | Purpose |
|-----------|----------|---------|
| JWT Service | `backend/internal/utils/jwt.go` | Creates/validates tokens with tenant_id claim |
| Auth Middleware | `backend/internal/middleware/auth.go` | Extracts tenantID from JWT claims |
| Tenant Middleware | `backend/internal/middleware/tenant.go` | Validates tenantID against config |
| Auth Handler | `backend/internal/handlers/auth_handler.go` | Login logic using multi-tenant service |

### 2.2 Database Layer

| Component | Location | Purpose |
|-----------|----------|---------|
| Database Config | `backend/internal/config/database.go` | Connection management per tenant |
| PostgreSQL Config | `backend/internal/config/database_postgres.go` | Schema-level isolation setup |
| Tenant Service | `backend/internal/services/tenant_service.go` | Tenant discovery & DB routing |
| Models | `backend/internal/models/*.go` | All have TenantID field with index |

### 2.3 Repository Layer (Data Access)

**Good Examples** (with tenant_id filters):
- `backend/internal/repositories/master_product_repository.go` - Uses Where("tenant_id = ?", tenantID)
- `backend/internal/repositories/audit_repository.go` - Tenant-specific audit logs

**At-Risk Repositories** (queries without explicit tenant filter):
- `backend/internal/repositories/shopee_order_repository.go` - FindAll() doesn't filter by tenant
- `backend/internal/repositories/lazada_order_repository.go` - Same issue
- `backend/internal/repositories/tiktok_order_repository.go` - Same issue
- `backend/internal/repositories/oauth_repository.go` - FindStateByState() has no tenant filter

### 2.4 Handler Layer (HTTP Endpoints)

All protected routes use dual middleware:
```go
protected := router.Group("/api/products")
protected.Use(middleware.Auth())      // Validates JWT
protected.Use(middleware.Tenant())    // Validates & sets tenantID
```

---

## 3. TENANT FILTERING PATTERNS

### 3.1 Correct Pattern - Explicit Tenant Filtering

**Master Product Repository**:
```go
func (r *MasterProductRepository) FindByTenantAndID(
    ctx context.Context, 
    tenantID string, 
    id uint) (*models.MasterProduct, error) {
    
    return r.db.WithContext(ctx).
        Where("id = ? AND tenant_id = ?", id, tenantID).
        First(&product)
}
```

### 3.2 At-Risk Pattern

**Shopee Order Repository**:
```go
// ISSUE: FindAll() doesn't filter by tenant
func (r *ShopeeOrderRepository) FindAll(
    ctx context.Context, 
    page, pageSize int) ([]models.ShopeeOrder, int64, error) {
    
    r.db.Model(&models.ShopeeOrder{}).Count(&total)  // Counts ALL orders!
    
    err := r.db.WithContext(ctx).
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&orders).Error  // Gets ALL orders!
}
```

**Mitigation**: Database-level schema isolation prevents cross-tenant leakage, but application-level validation is missing.

---

## 4. MULTI-TENANCY ISOLATION MECHANISMS

### 4.1 Three-Layer Defense

```
Layer 1: JWT Token Validation
    ↓ (extracts tenantID)
Layer 2: Tenant Configuration Validation
    ↓ (validates against tenants.json)
Layer 3: PostgreSQL Schema Isolation
    ↓ (search_path restricts to schema)
Database Level: Enforced by PostgreSQL
```

### 4.2 Key Security Features

**Hard Guard against Missing Tenant ID**:
```go
var ErrMissingTenantID = errors.New(
    "tenant_id is required - no default tenant allowed")

func GetTenantDB(tenantID string, basePath string) (*gorm.DB, error) {
    if tenantID == "" {
        return nil, ErrMissingTenantID  // FAIL FAST
    }
}
```

**Tenant ID in ALL Model Tables**:
Every data model includes indexed TenantID field.

**No Default Tenant Fallback**: System enforces explicit tenant context.

---

## 5. POTENTIAL SECURITY CONCERNS

### 5.1 CRITICAL: Repository Methods Lacking Tenant Filters

**Issue**: Methods like FindAll(), FindByOrderSN() don't include tenant_id in WHERE clause.

**Risk Level**: LOW-MEDIUM (mitigated by PostgreSQL schema isolation)

**Affected Files**:
- `backend/internal/repositories/shopee_order_repository.go`
- `backend/internal/repositories/lazada_order_repository.go`
- `backend/internal/repositories/tiktok_order_repository.go`

**Remediation**: Add tenant_id filters to all repository queries for defense-in-depth.

### 5.2 MEDIUM: OAuth State Query Without Tenant Context

**File**: `backend/internal/repositories/oauth_repository.go`

**Issue**: FindStateByState() queries by state value only, no tenant_id filter.

**Risk**: OAuth state tokens could theoretically be reused across tenants.

**Current Mitigation**: States expire after 10 minutes.

**Recommendation**: Add tenantID validation to OAuth state queries.

### 5.3 LOW: Webhook Tenant Parameter Validation

**File**: `backend/internal/handlers/webhook_extended_handler.go`

Webhooks accept tenantId from URL parameter but validate it against config.

**Status**: ACCEPTABLE - webhooks are intentionally unauthenticated for external platforms.

---

## 6. MISSING TENANT CHECKS

### 6.1 Most Areas Have Proper Protection

**Routes with protection**: 
- ✅ Product routes: Auth() + Tenant()
- ✅ Order routes: Auth() + Tenant()
- ✅ Platform auth: Auth() + Tenant()
- ✅ Webhook routes: Auth() + Tenant()

**System-level separation**:
- Users stored in system schema (shared across tenants)
- Tenant association through JWT token
- No cross-tenant user access possible

---

## 7. ROW-LEVEL SECURITY ANALYSIS

### 7.1 Is RLS Implemented?

**Answer**: NO explicit PostgreSQL Row-Level Security policies.

**Why It's OK**: PostgreSQL schema isolation is STRONGER than RLS because:
- Tenant A's connection cannot see tenant B's schema at all
- No risk of WHERE clause bypass
- Simpler to implement and maintain

**Alternative**: Application-level filtering via tenan

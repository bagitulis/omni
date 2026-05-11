# Backend Services - Development Rules

> **Parent:** `backend/AGENTS.md` | **Files:** 198 Go files

---

## Structure

```
services/
├── analytics/        # 26 files - Multi-platform analytics + intelligence
├── products/         # 21 files - Multi-platform product cloning
├── sync/             # 15 files - Order/product sync orchestration
├── inventory/        # 15 files - Cross-platform inventory management
├── ads/              # 12 files - Shopee/TikTok ads analytics + scoring
├── shopee/           # 12 files - Shopee platform integration
├── master_product/   # 12 files - Master product sync
├── tiktok/           # Platform-specific TikTok operations
├── lazada/           # Platform-specific Lazada operations
└── auth_service.go   # Authentication service
```

---

## Service Pattern (MANDATORY)

```go
// ✅ CORRECT - Service struct with dependencies
type ProductService struct {
    repo   repositories.ProductRepository
    cache  *CacheService
    logger zerolog.Logger
}

// ✅ CORRECT - Constructor with dependency injection
func NewProductService(repo repositories.ProductRepository, cache *CacheService) *ProductService {
    return &ProductService{
        repo:   repo,
        cache:  cache,
        logger: log.With().Str("service", "product").Logger(),
    }
}

// ❌ WRONG - Global variables or singletons
var productService *ProductService
```

---

## Platform Integration Rules

| Platform | SDK Location  | Wrapper       | Pattern                |
| -------- | ------------- | ------------- | ---------------------- |
| Shopee   | `shopee-sdk/` | `pkg/shopee/` | SDK → pkg → service    |
| Lazada   | `lazada-sdk/` | `pkg/lazada/` | SDK → pkg → service    |
| TikTok   | `tiktok_sdk/` | Direct        | SDK → service (no pkg) |

```go
// ✅ CORRECT - Use pkg wrapper for Shopee/Lazada
import "omni/backend/pkg/shopee"
client := shopee.NewClient(...)

// ✅ CORRECT - Direct SDK for TikTok (no wrapper yet)
import "omni/backend/tiktok_sdk/apis"

// ❌ WRONG - Mixing platform APIs
import shopee "omni/backend/pkg/shopee"
import tiktok "omni/backend/pkg/tiktok"  // Don't mix in same file
```

---

## Cross-Platform Service Pattern

For services that span multiple platforms (analytics, products, sync):

```go
// ✅ CORRECT - Base service with platform-specific implementations
type BaseEscrowService struct {
    tenantID string
    db       *gorm.DB
}

type ShopeeEscrowService struct {
    BaseEscrowService
    shopeeClient *shopee.Client
}

type TikTokEscrowService struct {
    BaseEscrowService
    tiktokClient *tiktok.Client
}
```

---

## File Naming Convention

| Type              | Pattern                          | Example                   |
| ----------------- | -------------------------------- | ------------------------- |
| Core service      | `{domain}_service.go`            | `auth_service.go`         |
| Platform-specific | `{domain}_{platform}.go`         | `clone_shopee.go`         |
| Helpers           | `{domain}_{platform}_helpers.go` | `clone_shopee_helpers.go` |
| Types/DTOs        | `{domain}_dto.go`                | `clone_dto.go`            |

---

## Analytics Intelligence Submodule

`services/analytics/intelligence/` contains ML-adjacent algorithms:

| File             | Purpose                     |
| ---------------- | --------------------------- |
| `scorer.go`      | ROAS/CTR scoring algorithms |
| `probability.go` | Conversion probability      |
| `volatility.go`  | Price/stock volatility      |
| `simulator.go`   | What-if scenarios           |
| `trend.go`       | Trend detection             |

**Rule:** Keep algorithms pure (no DB access). Pass data in, get results out.

---

## Anti-Patterns (FORBIDDEN)

```go
// ❌ WRONG - Business logic in repository
func (r *OrderRepo) GetOrdersWithDiscount() {}  // Discount is business logic

// ❌ WRONG - Direct DB access in service (bypass repository)
func (s *OrderService) GetOrder() {
    s.db.Where("id = ?", id).First(&order)  // Use repository!
}

// ❌ WRONG - Hardcoded tenant
if tenantID == "" { tenantID = "default" }

// ❌ WRONG - Empty error handling
if err != nil { }

// ✅ CORRECT - Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to sync order %s: %w", orderSN, err)
}
```

---

<!-- MASTER:file-size-quality -->
## ~300 Lines Per File (Quality Signal)

> **~300 lines is NOT a hard limit.** It's a quality signal that MUST trigger a refactor attempt.
> If a code file exceeds ~300 lines, you MUST attempt to refactor it (extract helpers, split by responsibility, remove dead code).
> After a genuine refactor effort, if the minimum achievable is slightly above 300 (e.g. 310-330) and the code satisfies SRP/DRY/OOP with no dead code — that's acceptable.
> This is NOT a license for 400+ line files. If your file is 400+ lines, you haven't refactored hard enough.

| Type           | Guideline                                      |
| -------------- | ---------------------------------------------- |
| Code files     | ~300 lines — MUST refactor if exceeded         |
| After refactor | Slightly above 300 OK if SRP/DRY/OOP satisfied |
| 400+ lines     | NOT acceptable — refactor harder or split      |

**EXEMPT from line limit (NO refactoring required):**
- Test files (`*_test.go`, `*.test.ts`, `*.spec.ts`) — tests are naturally long
- Generated/config files (`.json`, `.yaml`, `go.sum`, `package-lock.json`)
- Migration files (`migration*.go`, `*.sql`)
- Documentation files (`.md`)
- Type definition files that are purely interfaces/types

**ONLY applies to:** Business logic, handlers, services, components, utilities
<!-- /MASTER:file-size-quality -->

**If exceeding:** Split by platform or responsibility.

# Backend Handlers - Development Rules

> **Parent:** `backend/AGENTS.md` | **Files:** 141+ Go files

---

## Structure

```
handlers/
├── shopee/           # 11 files - Shopee platform handlers
├── tiktok/           # 9 files - TikTok platform handlers
├── lazada/           # 7 files - Lazada platform handlers
├── analytics/        # 10 files - Analytics handlers
├── inventory/        # 7 files - Inventory handlers
├── master_product/   # 6 files - Master product handlers
├── google/           # 7 files - Google Sheets integration
├── ml/               # ML-related handlers
└── *.go              # 70+ flat handlers (orders, products, auth, oauth)
```

---

## Handler Pattern (MANDATORY)

```go
// ✅ CORRECT - Handler calls service only
func (h *OrderHandler) GetOrders(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    if tenantID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
        return
    }

    orders, err := h.orderService.GetOrders(c.Request.Context(), tenantID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

// ❌ WRONG - Business logic in handler
func (h *OrderHandler) GetOrders(c *gin.Context) {
    orders := h.db.Where("status = ?", "pending").Find(&orders)  // NO DB access!
    for _, o := range orders {
        o.CalculateDiscount()  // NO business logic!
    }
}
```

---

## Response Format (MANDATORY)

```go
// ✅ SUCCESS
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "data": result,
})

// ✅ ERROR
c.JSON(http.StatusBadRequest, gin.H{
    "success": false,
    "error": "Validation failed: missing order_sn",
})

// ❌ WRONG - success: true with error message
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "message": "Failed to process",  // NEVER!
})
```

---

## Tenant Extraction (MANDATORY)

```go
// ✅ CORRECT - Get from middleware context
tenantID := c.GetString("tenant_id")
if tenantID == "" {
    c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
    return
}

// ❌ WRONG - Default tenant
if tenantID == "" {
    tenantID = "default"  // NEVER!
}

// ❌ WRONG - Get from query param without validation
tenantID := c.Query("tenant_id")  // Use middleware context!
```

---

## Request Validation

```go
// ✅ CORRECT - Validate and bind
type CreateOrderRequest struct {
    OrderSN   string `json:"order_sn" binding:"required"`
    ProductID int64  `json:"product_id" binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }
    // Call service...
}
```

---

## Platform-Specific Handlers

Each platform subdirectory follows same patterns but uses platform-specific services:

| Directory | Service         | SDK Used      |
| --------- | --------------- | ------------- |
| `shopee/` | `ShopeeService` | `pkg/shopee/` |
| `tiktok/` | `TikTokService` | `tiktok_sdk/` |
| `lazada/` | `LazadaService` | `pkg/lazada/` |

```go
// ✅ CORRECT - Platform handler uses platform service
func (h *ShopeeHandler) SyncOrders(c *gin.Context) {
    err := h.shopeeService.SyncOrders(ctx, tenantID)  // Platform-specific
}

// ❌ WRONG - Mixing platforms
func (h *ShopeeHandler) SyncOrders(c *gin.Context) {
    h.tiktokService.SyncOrders(...)  // Wrong service!
}
```

---

## File Naming Convention

| Type             | Pattern                   | Example             |
| ---------------- | ------------------------- | ------------------- |
| Domain handler   | `{domain}_handler.go`     | `order_handler.go`  |
| CRUD operations  | `{domain}_{operation}.go` | `product_create.go` |
| Platform subdirs | `{subdomain}.go`          | `shopee/orders.go`  |

---

## Anti-Patterns (FORBIDDEN)

```go
// ❌ WRONG - Database access in handler
h.db.Where(...).Find(&orders)

// ❌ WRONG - Business logic in handler
if order.Total > 100 { order.Discount = 10 }

// ❌ WRONG - Calling repository directly
h.orderRepo.FindByID(...)  // Use service layer!

// ❌ WRONG - Inconsistent response format
c.JSON(200, orders)  // Missing success/data wrapper

// ❌ WRONG - Panic instead of error response
panic("something went wrong")
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
<!-- /MASTER:file-size-quality -->

**If exceeding:** Split by operation (create, read, update, delete).

# Order Management Quick Reference

## Available Order Actions by Platform

| Action | Shopee | Lazada | TikTok |
|--------|--------|--------|--------|
| **List Orders** | GET /api/shopee/orders | GET /api/lazada/orders | GET /api/tiktok/orders |
| **Ship** | POST /ship (1-step) | POST /ship (2-step) | POST /ship |
| **Cancel** | POST /cancel | POST /cancel (item-based) | POST /cancel |
| **Batch Limit** | 50 orders | Item-based | Token paginated |
| **Status Levels** | Order-level | Item-level | Order-level |

## Order Data Fields Returned

### All Platforms
- `order_no` - Frontend-friendly order number
- `platform` - "shopee" / "lazada" / "tiktok"
- `status` - Platform-specific status
- `total_amount` - Order total
- `currency` - Currency code
- `buyer_username` - Buyer identifier
- `sku` - Product SKU
- `product_name` - Product name
- `variation_name` - Variant/model name
- `qty` - Quantity (note: "qty" not "quantity")
- `created_at`, `updated_at` - Timestamps

### For Processed Orders
- `tracking_number` - Shipping tracking
- `shipping_carrier` - Logistics provider

## Sync Endpoints

```
POST /api/orders/sync/{category}
  - Categories: unpaid, unprocess, processed
  - Query: ?days=7&platforms=shopee,lazada

POST /api/orders/sync-all
  - Syncs all categories from all platforms

GET /api/orders/{category}
  - Retrieves synced orders from database
```

## Frontend Composable Usage

```typescript
import { useOrderManager } from '@/components/OrderManager/composables/useOrderManager'

const {
  orders,           // Current orders
  activeTab,        // Current tab
  loading,          // Loading state
  filteredOrders,   // Filtered by search/platform
  changeTab,        // (tabValue) => void
  refreshData,      // () => void
  initializeFromUrl // () => void
} = useOrderManager()
```

## Key Implementation Files

**Backend:**
- `backend/internal/handlers/order_sync.go` - Sync handlers
- `backend/internal/handlers/order_manager.go` - Retrieval handlers  
- `backend/internal/services/sync/order_operations.go` - Sync logic
- `backend/pkg/shopee|tiktok/order.go` - SDK wrappers
- `backend/lazada-sdk/order.go` - Lazada SDK

**Frontend:**
- `frontend/src/components/OrderManager/composables/useOrderManager.ts`
- `frontend/src/services/api.ts` - API client

## Status Mapping

```
Frontend    Shopee      Lazada           TikTok
──────────────────────────────────────────────────
unpaid      UNPAID      PENDING          UNPAID
unprocess   UNSHIPPED   READY_TO_SHIP    UNSHIPPED
processed   SHIPPED     SHIPPED          SHIPPED
```

## Important Constraints

1. **Shopee**: Max 50 orders per batch request
2. **Lazada**: 2-step ship (pack then ready-to-ship)
3. **TikTok**: Package-based shipping
4. **All**: No product images in order responses
5. **All**: 7-day default sync window

## Product Images

⚠️ **ORDER RESPONSES DO NOT INCLUDE PRODUCT IMAGES**

Images available from:
- `ShopeeProduct.Image` - Product catalog
- `LazadaProduct` - Product table
- `TiktokProduct` - Product table

To show images with orders:
1. Fetch order
2. Get product by SKU
3. Extract image_url from product

## Response Format Example

```json
{
  "success": true,
  "items": [
    {
      "order_no": "220123ABC123",
      "platform": "shopee",
      "sku": "PROD-001",
      "product_name": "Product Name",
      "variation_name": "Color: Red",
      "qty": 2,
      "status": "UNSHIPPED",
      "tracking_number": null,
      "total_amount": 250.00,
      "currency": "IDR",
      "buyer_username": "buyer123"
    }
  ],
  "count": 1
}
```

## Common API Patterns

### Sync-then-Fetch Pattern
```typescript
// Step 1: Sync from APIs
await fetch('/api/orders/sync/unpaid', { method: 'POST' })

// Step 2: Fetch from local DB
const orders = await fetch('/api/orders/unpaid')
```

### Platform-Specific Operations
```typescript
// Ship order
POST /api/{platform}/orders/ship
{
  "order_sn": "...",
  "tracking_number": "..." // Shopee
  // OR
  "order_id": "...",
  "package_id": "...", // TikTok
  "shipping_provider": "..."
  // OR
  "order_item_ids": [...], // Lazada
  "shipping_provider": "..."
}

// Cancel
POST /api/{platform}/orders/cancel
{
  "order_sn": "...",
  "cancel_reason": "..."
}
```

## Database Tables

Per platform (Shopee, Lazada, TikTok):
- `{Platform}Order` - Order header
- `{Platform}OrderItem` - Order line items

All tables include:
- `tenant_id` - Multi-tenancy support
- `created_at`, `updated_at` - Audit timestamps

## Debugging Tips

1. Check order status mapping in `order_status.go`
2. Verify API credentials in GlobalConfig
3. Check batch size in `order_operations.go` (line ~60)
4. Review sync results for per-platform errors
5. Orders cleared before sync - check database timestamps

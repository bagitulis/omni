# Vue Frontend API Endpoints Inventory

## Last Updated
- Scan Date: 2024
- Scope: frontend/src directory
- Method: Comprehensive grep + manual inspection

## Routes Summary
```
/ (Dashboard - default)
/login (Authentication)
/dev-preview
/dev-preview/:component
/product-manager (with child routes: /shopee, /lazada, /tiktok)
/order-manager
/script-monitor (with child routes: /current, /queue, /history, /auto-functions)
/report (with child routes: /shopee, /tiktok)
/analytics (with child routes: /hub, /simulator, /classification, /ml, /tiktok-ads, /shopee-ads, /ai-reports)
/settings (with child routes: /google-sheets, /webhook)
/inventory
/route-mapping
/master-products
/master-products/add
/master-products/import
/master-products/:id
```

---

## API Endpoints Used

### Authentication
- POST /auth/login - Login with credentials + reCAPTCHA
- POST /auth/register - Register new user
- GET /auth/me - Get current user profile
- POST /auth/logout - Logout user
- POST /auth/refresh - Refresh JWT token

### Master Products
- GET /master-products - List all master products with pagination
- GET /master-products/:id - Get single master product
- POST /master-products - Create new master product
- PUT /master-products/:id - Update master product
- DELETE /master-products/:id - Delete master product
- PUT /master-products/skus/:id - Update single SKU price/stock
- PUT /master-products/skus/batch - Batch update SKUs
- GET /master-products/import/preview?platform=shopee&item_id=123 - Preview platform import
- POST /master-products/import - Import product from platform
- GET /master-products/:id/mapping - Get SKU mapping status
- POST /master-products/mapping/auto - Auto-map SKU to platform products
- POST /master-products/mapping/link - Manual link SKU to platform
- DELETE /master-products/mapping/link - Unlink SKU from platform
- POST /master-products/:id/sync - Sync product to platform
- GET /master-products/:id/sync-status - Get product sync status

### Orders
- GET /orders/unpaid - Get unpaid orders
- GET /orders/unprocess - Get orders to ship
- GET /orders/processed - Get shipped orders
- POST /orders/sync/:category - Sync orders by category (unpaid, unprocess, processed)
- POST /orders/sync-all - Sync all orders
- POST /orders/locked-today - Get locked orders for today
- POST /orders/today - Get today's orders

### Inventory
- GET /inventory/columns/available - Get available column definitions
- GET /inventory/columns/selected - Get selected column configuration
- PUT /inventory/columns/selected - Save selected column configuration
- GET /inventory/:itemKey - Get inventory item by key
- PUT /inventory/:itemKey - Update inventory item
- POST /inventory/sync/from-sheets - Sync inventory from Google Sheets
- POST /inventory/sync/to-sheets - Sync inventory to Google Sheets
- GET /inventory/sync-status - Get sync status

### Routes Configuration
- GET /api/routes-config/all - Get all route configurations
- GET /api/routes-config - Get route configs with caching
- POST /api/routes-config - Create new route config
- PATCH /api/routes-config/:routeId - Update route config
- DELETE /api/routes-config/:routeId - Delete route config
- POST /api/routes-config/bulk-update - Bulk update routes
- DELETE (multiple) - Bulk delete routes
- POST /api/routes-config/apply-preset/:preset - Apply preset to routes

### Route Execution Configuration
- GET /route-execution-config - Get all execution configs
- GET /route-execution-config/:routeKey/mode - Get execution mode
- POST /route-execution-config - Create execution config
- PUT /route-execution-config/:routeKey - Update execution config
- DELETE /route-execution-config/:routeKey - Delete execution config

### Route Mapping
- GET /route-mapping/detailed-mapping - Get detailed route mapping
- GET /route-mapping/statistics - Get mapping statistics

### Shopee Database/Products
- GET /shopee/db/products/list - List Shopee products
- GET /shopee/db/products/base - Get base product info
- GET /shopee/db/products/model - Get product models
- GET /shopee/db/products/variants/:id - Get product variants
- GET /shopee/db/products/stock/:id - Get product stock
- GET /shopee/db/products/pricing/:id - Get product pricing
- GET /shopee/db/products/categories/:id - Get product categories
- GET /shopee/db/products/search?keyword=:keyword - Search products
- GET /shopee/db/products/detail/:productId - Get product detail
- GET /shopee/db/stats - Get database statistics
- GET /shopee/db/sync/unprocessed - Get unprocessed products
- GET /shopee/db/sync/no-models - Get products without models
- GET /shopee/db/sync/logs?limit=:limit - Get sync logs
- POST /shopee/db/products/batch-update - Batch update products
- POST /shopee/db/products/sync - Sync products
- POST /shopee/db/products/publish - Publish products

### TikTok Products
- POST /tiktok/products/search - Search TikTok products
- POST /tiktok/products/get-all - Get all TikTok products
- GET /tiktok/products/:productId - Get single product
- POST /tiktok/products/refresh-all - Refresh all products
- GET /tiktok/db/products/list - List TikTok database products
- GET /tiktok/db/products/master - Get master products
- GET /tiktok/db/products/models/:productId - Get product models
- GET /tiktok/db/products/variants/:productId - Get product variants
- GET /tiktok/db/products/:productId - Get product detail
- GET /tiktok/db/statistics - Get database statistics
- DELETE /tiktok/db/products/:productId - Delete product

### Product Creation (Platform-Specific)
- GET /api/:platform/categories - Get product categories
- GET /api/:platform/categories?parent_id=:parentId - Get child categories
- GET /api/:platform/categories/search?keyword=:keyword - Search categories
- GET /api/:platform/categories/:categoryId/attributes - Get category attributes
- POST /api/:platform/categories/recommend - Get recommended category
- GET /api/:platform/categories/popular - Get popular categories
- GET /api/:platform/brands?category_id=:categoryId - Get brands for category
- GET /api/:platform/delivery-options - Get delivery options
- POST /api/:platform/images/upload - Upload product image
- POST /api/:platform/products/create - Create new product

### Shopee-Specific
- GET /shopee/db/products - Get products list
- (Used by: ShopeeProductManager component)

### Lazada-Specific
- GET /lazada/db/products - Get products list
- (Used by: LazadaProductManager component)

### Images
- GET /images/gallery - Get image gallery
- POST /images/upload - Upload image

### Shipping/Labels
- GET /shipping/files - List shipping files
- POST /shipping/process-file - Process shipping file
- (Various label endpoints - check shippingLabelService.ts)

### Analytics/Ads
- GET /api/ads/:platform/summary - Get ads summary
- GET /api/ads/:platform/trends - Get ads trends
- GET /api/ads/:platform/performance - Get ads performance

### Google Sheets Integration
- GET /api/google/registry - Get registry
- POST /api/google/registry/:registryId - Create/update registry
- DELETE /api/google/registry/:registryId - Delete registry
- Various spreadsheet operations

### Webhooks
- GET /api/webhooks/:tenantId/:platform - Webhook endpoint (conditional)
- GET /api/webhooks/:platform - Webhook endpoint (default)
- Callback URLs:
  - /api/platform-auth/shopee/callback
  - /api/platform-auth/tiktok/callback
  - /api/platform-auth/lazada/callback

### System/Health
- GET /health - Health check
- GET /status - Server status
- GET /token-status - Token status
- POST /tokens/refresh-all?force=:force - Refresh all tokens
- POST /execute - Generic operation execution
- GET /debug/check-functions - Check function status

### Order Export
- POST /execute-order-export - Execute order export

---

## HTTP Methods Summary
| Method | Count | Common Endpoints |
|--------|-------|------------------|
| GET | 40+ | Fetching data, listing items |
| POST | 30+ | Creating, syncing, executing |
| PUT | 10+ | Updating resources |
| PATCH | 5+ | Partial updates |
| DELETE | 8+ | Removing resources |

---

## Service Files Map

### Main Services
- `api.ts` - Core axios servic

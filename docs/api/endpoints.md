---
last_updated: 2026-05-03
updated_by: agent
relates_to: backend/internal/routes/
stale_if_changed:
  - backend/internal/routes/*.go
  - backend/internal/handler/
  - backend/cmd/server/main.go
---

# API Endpoint Catalog

> **Base URL**: `http://localhost/api/` (via Nginx) or `http://localhost:3000/api/` (direct)
> **Response Format**: `{ "success": true, "data": {...} }` or `{ "success": false, "error": "message" }`
> **Auth**: JWT Bearer token in `Authorization` header (protected routes)

---

## Public Routes (No Auth Required)

### Health & Status

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/status` | Status with token data |
| GET | `/api/csrf-token` | Get CSRF token |
| GET | `/api/docs` | API documentation |
| GET | `/api/docs/swagger.json` | Swagger JSON spec |

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | User login |
| POST | `/api/auth/register` | User registration |
| POST | `/api/auth/refresh` | Refresh JWT token |
| GET | `/api/auth/verify` | Verify token validity |
| GET | `/api/auth/dev-info` | Dev login info (dev only) |
| POST | `/api/auth/dev-login` | Dev login (dev only) |

### Captcha

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/captcha/status` | Get captcha status |
| GET | `/api/captcha/site-key` | Get captcha site key |
| POST | `/api/captcha/verify` | Verify captcha |

### Webhooks (Platform Callbacks)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/webhooks/shopee` | Shopee webhook receiver |
| POST | `/api/webhooks/lazada` | Lazada webhook receiver |
| POST | `/api/webhooks/tiktok` | TikTok webhook receiver |
| POST | `/api/webhooks/:tenantId/shopee` | Tenant-specific Shopee webhook |
| POST | `/api/webhooks/:tenantId/lazada` | Tenant-specific Lazada webhook |
| POST | `/api/webhooks/:tenantId/tiktok` | Tenant-specific TikTok webhook |

### OAuth Callback

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/platform-auth/callback/:platform` | OAuth callback handler |

---

## Protected Routes (Auth + Tenant Required)

### User Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/logout` | Logout |
| POST | `/api/auth/change-password` | Change password |
| POST | `/api/auth/profile` | Update profile |
| GET | `/api/auth/me` | Get current user |
| GET | `/api/auth/tenants` | Get user's tenants |
| POST | `/api/auth/switch-tenant` | Switch active tenant |
| GET | `/api/users` | List users |
| POST | `/api/users` | Create user |
| GET | `/api/users/:id` | Get user by ID |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user |
| POST | `/api/users/:id/unlock` | Unlock locked user |

### Platform Authentication

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/platform-auth/urls` | Get OAuth URLs for all platforms |
| GET | `/api/platform-auth/status` | Get connection status |
| GET | `/api/platform-auth/initiate/:platform` | Initiate OAuth flow |
| GET | `/api/platform-auth/logs` | Get OAuth logs |
| POST | `/api/platform-auth/check-all` | Check all connections |
| GET | `/api/platform-auth/shopee/disconnect` | Disconnect Shopee |
| GET | `/api/platform-auth/lazada/disconnect` | Disconnect Lazada |
| GET | `/api/platform-auth/tiktok/shops` | Get TikTok shops |
| GET | `/api/platform-auth/tiktok/active-shop` | Get active TikTok shop |

### Token Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/tokens/status` | Get all token status |
| GET | `/api/tokens/status/:platform` | Get platform token status |
| POST | `/api/tokens/refresh/:platform` | Refresh platform token |
| POST | `/api/tokens/refresh-all` | Refresh all tokens |

### Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/orders` | Get unprocessed orders |
| GET | `/api/orders/unpaid` | Get unpaid orders |
| GET | `/api/orders/processed` | Get processed orders |
| GET | `/api/orders/today` | Get orders today |
| POST | `/api/orders/today` | Sync orders today |
| POST | `/api/orders/sync-all` | Sync all orders |
| GET | `/api/orders/:orderSn` | Get order by order SN |
| POST | `/api/orders/sync/:category` | Sync by category |
| POST | `/api/orders/sync/platform/:platform` | Sync platform orders |
| GET | `/api/orders/category/:category` | Get orders by category |
| GET | `/api/orders/details/:platform` | Get order details |
| POST | `/api/orders/bulk-print-labels` | Bulk print shipping labels |
| POST | `/api/orders/bulk-ship` | Bulk ship orders |
| GET | `/api/orders/locked-today` | Get saved locked orders |
| POST | `/api/orders/locked-today` | Get locked today orders |

### Locked Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/locked-orders` | Get locked orders |
| POST | `/api/locked-orders` | Save locked orders |
| DELETE | `/api/locked-orders` | Clear locked orders |

### Master Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/products` | Get master product list |
| GET | `/api/products/master` | Get master product list |
| GET | `/api/products/master/stats` | Get master product stats |
| GET | `/api/products/master/:id` | Get product by ID |
| POST | `/api/products/master` | Create product |
| PUT | `/api/products/master/:id` | Update product |
| DELETE | `/api/products/master/:id` | Delete product |
| PUT | `/api/products/master/skus/batch` | Batch update SKUs |
| PUT | `/api/products/master/skus/:id` | Update single SKU |
| POST | `/api/products/master/sync-selected` | Sync selected products |

### Master Product Import

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/master-products/import/preview` | Preview import (file upload) |
| GET | `/api/master-products/import/template` | Download import template |
| POST | `/api/master-products/import` | Execute import |
| POST | `/api/master-products/import/from-staging/shopee` | Import from Shopee staging |
| POST | `/api/master-products/import/from-staging/tiktok` | Import from TikTok staging |
| POST | `/api/master-products/import/from-staging/lazada` | Import from Lazada staging |

### Master Product Mapping

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/master-products/mapping/auto` | Auto-map SKU |
| POST | `/api/master-products/mapping/auto-link` | Auto-map and link batch |
| POST | `/api/master-products/mapping/link` | Manual link |
| DELETE | `/api/master-products/mapping/link` | Unlink |
| GET | `/api/master-products/:id/mapping` | Get mapping status |
| POST | `/api/master-products/images/backfill` | Backfill images |
| POST | `/api/master-products/:id/images/refresh` | Refresh product images |
| POST | `/api/master-products/:id/sync` | Sync to platform |
| GET | `/api/master-products/:id/sync-status` | Get sync status |

### Platform Products

#### Shopee Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/shopee/products` | Get products from API |
| GET | `/api/shopee/products/:itemId` | Get product by ID |
| POST | `/api/shopee/products` | Create product |
| PUT | `/api/shopee/products/:itemId` | Update product |
| DELETE | `/api/shopee/products/:itemId` | Delete product |
| POST | `/api/shopee/sync/products` | Sync products |
| GET | `/api/shopee/db/products` | Get DB products |
| GET | `/api/shopee/db/products/master` | Get master products |
| GET | `/api/shopee/db/products/full/:itemId` | Get full product |
| GET | `/api/shopee/db/products/search` | Search products |

#### Lazada Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/lazada/products` | Get products from API |
| GET | `/api/lazada/products/:itemId` | Get product by ID |
| POST | `/api/lazada/products` | Create product |
| PUT | `/api/lazada/products/:itemId` | Update product |
| DELETE | `/api/lazada/products/:itemId` | Delete product |
| POST | `/api/lazada/sync/products` | Sync products |
| GET | `/api/lazada/db/products` | Get DB products |
| GET | `/api/lazada/products/categories` | Get categories |

#### TikTok Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/tiktok/products` | Get products from API |
| GET | `/api/tiktok/products/:productId` | Get product by ID |
| POST | `/api/tiktok/products` | Create product |
| PUT | `/api/tiktok/products/:productId` | Update product |
| DELETE | `/api/tiktok/products/:productId` | Delete product |
| POST | `/api/tiktok/sync/products` | Sync products |
| GET | `/api/tiktok/db/products` | Get DB products |
| GET | `/api/tiktok/products/categories` | Get categories |
| POST | `/api/tiktok/products/upload-image` | Upload image |

### Product Creation & Cloning

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/products/create/shopee` | Create on Shopee |
| POST | `/api/products/create/lazada` | Create on Lazada |
| POST | `/api/products/create/tiktok` | Create on TikTok |
| GET | `/api/products/categories/:platform` | Get categories |
| POST | `/api/products/clone` | Clone product |
| POST | `/api/products/clone/batch` | Batch clone |
| GET | `/api/products/clone/status/:id` | Get clone status |

### Stock & Price

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/stock` | List stock |
| GET | `/api/stock/alerts` | Get stock alerts |
| GET | `/api/stock/:sku` | Get stock by SKU |
| PUT | `/api/stock/:sku` | Update stock |
| POST | `/api/stock/bulk` | Bulk update stock |
| GET | `/api/price` | List prices |
| GET | `/api/price/:sku` | Get price by SKU |
| PUT | `/api/price/:sku` | Update price |
| POST | `/api/price/bulk` | Bulk update prices |

### Inventory Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/inventory` | Get inventory list |
| GET | `/api/inventory/config` | Get inventory config |
| PUT | `/api/inventory/config` | Update inventory config |
| GET | `/api/inventory/stats` | Get inventory stats |
| POST | `/api/inventory` | Create record |
| GET | `/api/inventory/:keyValue` | Get record by key |
| PUT | `/api/inventory/:keyValue` | Update record |
| DELETE | `/api/inventory/:keyValue` | Delete record |
| POST | `/api/inventory/sync/from-sheets` | Sync from Google Sheets |
| POST | `/api/inventory/sync/to-sheets` | Sync to Google Sheets |
| POST | `/api/inventory/update-stock` | Update stock |
| POST | `/api/inventory/update-stock-batch` | Batch update stock |
| POST | `/api/inventory/update-price` | Update price |
| POST | `/api/inventory/update-price-batch` | Batch update price |
| POST | `/api/inventory/batch-check-sku` | Batch check SKU |

### Wholesale Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/wholesale/settings` | Get wholesale settings |
| PUT | `/api/wholesale/settings` | Update wholesale settings |
| POST | `/api/wholesale/calculate` | Calculate wholesale |
| POST | `/api/wholesale/apply` | Apply wholesale |
| POST | `/api/wholesale/shopee/batch-add` | Batch add wholesale |
| POST | `/api/wholesale/shopee/batch-delete` | Batch delete |
| POST | `/api/wholesale/shopee/batch-mpq` | Batch set MPQ |
| POST | `/api/wholesale/shopee/batch-update-skus` | Batch update by SKUs |
| POST | `/api/wholesale/tiktok/batch-mpq` | TikTok batch MPQ |

### Shipping

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/shopee/shipping/options` | Get Shopee shipping options |
| POST | `/api/shopee/shipping/arrange` | Arrange Shopee shipment |
| GET | `/api/shopee/shipping/tracking/:orderSn` | Get tracking |
| GET | `/api/shopee/shipping/label/:orderSn` | Get shipping label |
| POST | `/api/tiktok/shipping/arrange` | Arrange TikTok shipment |
| GET | `/api/tiktok/shipping/document/:packageId` | Get TikTok shipping doc |
| POST | `/api/tiktok/shipping/download/batch` | Batch download docs |

### Shopee Wallet & Escrow

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/shopee/wallet/balance` | Get wallet balance |
| GET | `/api/shopee/wallet/transactions` | Get transactions |
| GET | `/api/shopee/wallet/income` | Get net income |
| POST | `/api/shopee/wallet/report` | Get wallet report |
| POST | `/api/shopee/wallet/export` | Export wallet data |
| POST | `/api/shopee/wallet/escrow-detail` | Get escrow detail |

### Google Integration

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/google/auth/status` | Get Google auth status |
| GET | `/api/google/sheets/list` | List spreadsheets |
| GET | `/api/google/sheets/data` | Get spreadsheet data |
| POST | `/api/google/sheets/create` | Create spreadsheet |
| GET | `/api/google/service-accounts` | List service accounts |
| GET | `/api/google/quota/status` | Get quota status |

### Analytics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/analytics/dashboard` | Dashboard summary |
| GET | `/api/analytics/orders` | Order analytics |
| GET | `/api/analytics/revenue` | Revenue analytics |
| GET | `/api/analytics/settings` | Analytics settings |
| PUT | `/api/analytics/settings` | Update analytics settings |

### Job Queue & Automation

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/jobs` | Add job |
| GET | `/api/jobs` | List jobs |
| GET | `/api/jobs/:id` | Get job |
| DELETE | `/api/jobs/:id` | Cancel job |
| GET | `/api/jobs/status` | Get queue status |
| GET | `/api/jobs/monitor` | Get monitor |
| GET | `/api/jobs/history` | Get history |
| GET | `/api/jobs/auto-functions` | List auto-functions |
| POST | `/api/jobs/auto-functions` | Create auto-function |
| POST | `/api/jobs/auto-functions/:name/run` | Run auto-function |

### Reports

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/reports/shopee/ads` | Get Shopee ads report |
| GET | `/api/reports/shopee/ads/latest` | Get latest Shopee ads |
| GET | `/api/reports/tiktok/ads` | Get TikTok ads report |
| GET | `/api/reports/tiktok/ads/latest` | Get latest TikTok ads |

### Notifications

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notifications` | List notifications |
| POST | `/api/notifications` | Create notification |
| GET | `/api/notifications/stream` | SSE notification stream |
| GET | `/api/notifications/unread-count` | Get unread count |
| PATCH | `/api/notifications/:id/read` | Mark as read |
| PATCH | `/api/notifications/read-all` | Mark all as read |
| DELETE | `/api/notifications/:id` | Delete notification |
| GET | `/api/notifications/settings` | Get settings |
| PUT | `/api/notifications/settings` | Update settings |

### Images

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/images/upload` | Upload image |
| GET | `/api/images/gallery` | Get image gallery |
| GET | `/api/images/:id` | Get image by ID |
| DELETE | `/api/images/:id` | Delete image |

### Settings

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/settings/inventory` | Get inventory settings |
| PUT | `/api/settings/inventory` | Update inventory settings |
| GET | `/api/settings/google-sheets` | Get Google Sheets settings |
| PUT | `/api/settings/google-sheets` | Update Google Sheets settings |
| GET | `/api/settings/general` | Get general settings |
| POST | `/api/settings/general` | Update general settings |

### Audit & Monitoring

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/audit` | Get audit logs |
| GET | `/api/audit/logs/tenant` | Get tenant audit logs |
| GET | `/api/audit/user/:userId` | Get user audit logs |
| GET | `/api/monitoring/metrics` | Get metrics |
| GET | `/api/monitoring/health/detailed` | Detailed health check |

---

## Rate Limiting

| Endpoint Group | Limit | Notes |
|---------------|-------|-------|
| `/api/webhooks/*` | None | Platform callbacks |
| `/api/platform-auth/*` | None | OAuth flows |
| `/api/*/sync` | 10/s burst 20 | 300s timeout |
| `/api/master-products/import` | 10/s burst 20 | 300s timeout |
| `/api/*` (general) | 10/s burst 20 | 60s timeout |
| `/` (frontend) | 100/s burst 50 | Static assets |

---

## Authentication Flow

```
1. POST /api/auth/login → { access_token, refresh_token }
2. Use access_token in Authorization: Bearer <token>
3. When expired → POST /api/auth/refresh with refresh_token
4. Token includes tenant_id claim → all queries scoped to tenant
```

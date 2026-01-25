# Comprehensive Frontend-Backend Gap Analysis & Migration Plan

**Date:** January 25, 2026  
**Project:** Omni Integration (Migrating Node.js to Go)  
**Status:** FIXED - Frontend Updated, Some Backend Endpoints Need Verification

---

## 1. Executive Summary

Analisis ini menyoroti ketidaksesuaian antara endpoint yang dipanggil oleh Frontend (berbasis arsitektur legacy Node.js) dengan endpoint yang tersedia di Backend Go (arsitektur baru).

**Status Perbaikan:**

- [x] Frontend `apiOperationMappers.ts` sudah diperbaiki
- [x] Mapping endpoint stock/price sudah benar
- [x] Mapping endpoint auth/token sudah benar
- [x] Mapping endpoint sheets export sudah benar
- [ ] Endpoint `export_orders` per platform perlu verifikasi
- [ ] Endpoint `update_code` belum diimplementasi di Go

---

## 2. Route Mapping Summary

### 2.1 Inventory Operations (FIXED)

| Operation           | Node.js Backend                                           | Go Backend                           | Frontend (Updated)               | Status |
| :------------------ | :-------------------------------------------------------- | :----------------------------------- | :------------------------------- | :----: |
| Update Stock        | `/api/stock/update-stock` + `/api/inventory/update-stock` | `/api/inventory/update-stock`        | `/inventory/update-stock`        |   ✅   |
| Update Stock Batch  | `/api/stock/update-stock-batch`                           | `/api/inventory/update-stock-batch`  | `/inventory/update-stock-batch`  |   ✅   |
| Update Price        | `/api/price/update-price` + `/api/inventory/update-price` | `/api/inventory/update-price`        | `/inventory/update-price`        |   ✅   |
| Update Price Batch  | `/api/price/update-price-batch`                           | `/api/inventory/update-price-batch`  | `/inventory/update-price-batch`  |   ✅   |
| Lookup Platform IDs | `/api/stock/lookup-platform-ids`                          | `/api/inventory/lookup-platform-ids` | `/inventory/lookup-platform-ids` |   ✅   |

### 2.2 Authentication & Token (FIXED)

| Operation             | Node.js Backend                          | Go Backend                              | Frontend (Updated)                   | Status |
| :-------------------- | :--------------------------------------- | :-------------------------------------- | :----------------------------------- | :----: |
| OAuth Initiate        | `/api/platform-auth/:platform/authorize` | `/api/platform-auth/initiate/:platform` | `/platform-auth/initiate/{platform}` |   ✅   |
| OAuth Callback        | `/api/platform-auth/:platform/callback`  | `/api/platform-auth/callback/:platform` | N/A (redirect)                       |   ✅   |
| Refresh Token         | `/api/tokens/refresh-all`                | `/api/tokens/refresh/:platform`         | `/tokens/refresh/{platform}`         |   ✅   |
| Token Status          | `/api/token-status`                      | `/api/token-status` (alias)             | `/token-status`                      |   ✅   |
| Platform Token Status | `/api/:platform/token-status`            | `/api/:platform/token-status` (alias)   | `/{platform}/token-status`           |   ✅   |

### 2.3 Sheets Export Operations (FIXED)

| Operation              | Node.js Backend                                    | Go Backend                              | Frontend (Updated)                  | Status |
| :--------------------- | :------------------------------------------------- | :-------------------------------------- | :---------------------------------- | :----: |
| Wallet to Sheets       | `/api/sheets/operations` (walletSheetsRouter)      | `/api/shopee/wallet/export-to-sheets`   | `/shopee/wallet/export-to-sheets`   |   ✅   |
| Shipping Fee to Sheets | `/api/sheets/operations` (shippingFeeSheetsRouter) | `/api/shopee/shipping/export-to-sheets` | `/shopee/shipping/export-to-sheets` |   ✅   |

### 2.4 Order Export (NEEDS VERIFICATION)

| Operation            | Node.js Backend          | Go Backend                  | Frontend (Updated)      |  Status   |
| :------------------- | :----------------------- | :-------------------------- | :---------------------- | :-------: |
| Export Shopee Orders | `/api/n8n/export-orders` | `/api/shopee/orders/export` | `/shopee/orders/export` | ⚠️ Verify |
| Export Lazada Orders | `/api/n8n/export-orders` | `/api/lazada/orders/export` | `/lazada/orders/export` | ⚠️ Verify |
| Export TikTok Orders | `/api/n8n/export-orders` | `/api/tiktok/orders/export` | `/tiktok/orders/export` | ⚠️ Verify |

---

## 3. Node.js Backend Route Reference

Berdasarkan `backend-node/src/config/routeConfig.ts`:

```typescript
// Inventory Operations - Dual mounting for compatibility
app.use("/api/stock", stockRoutes);
app.use("/api/inventory", stockRoutes); // Also mounted here
app.use("/api/price", priceRoutes);
app.use("/api/inventory", priceRoutes); // Also mounted here

// Shopee Specific
app.use("/api/shopee/wallet", shopeeWalletRouter);
app.use("/api/shopee/shipping", shopeeShippingRouter);
app.use("/api/shopee/operations", shopeeOperationsRouter);

// Sheets Integration
app.use("/api/sheets/operations", walletSheetsRouter);
app.use("/api/sheets/operations", shippingFeeSheetsRouter);

// Token Management
app.use("/api", tokenStatusRoutes); // /token-status, /:platform/token-status
app.use("/api/platform-auth", platformAuthRoutes);
```

---

## 4. Go Backend Route Reference

Berdasarkan `backend/internal/routes/`:

```go
// routes.go - Token & Auth
tokenAlias.GET("/token-status", tokenHandler.GetAllTokenStatus)
tokenAlias.GET("/:platform/token-status", tokenHandler.GetPlatformTokenStatus)
tokens.POST("/refresh/:platform", tokenHandler.RefreshToken)
platformAuth.GET("/initiate/:platform", oauthHandler.InitiateAuth)

// extended_routes.go - Inventory
inv.POST("/update-stock", stockHandler.UpdateStock)
inv.POST("/update-stock-batch", stockHandler.UpdateStockBatch)
inv.POST("/update-price", priceHandler.UpdatePrice)
inv.POST("/update-price-batch", priceHandler.UpdatePriceBatch)
inv.POST("/lookup-platform-ids", stockHandler.LookupPlatformIds)

// extended_routes.go - Shopee Wallet & Shipping
wallet.POST("/export-to-sheets", handler.ExportToSheets)
shipping.POST("/export-to-sheets", handler.ExportToSheets)
```

---

## 5. Files Modified

| File                                           | Change Type  | Description                                       |
| :--------------------------------------------- | :----------- | :------------------------------------------------ |
| `frontend/src/services/apiOperationMappers.ts` | **MODIFIED** | Updated all endpoint mappings to match Go backend |
| `docs/frontend_backend_gap_analysis.md`        | **CREATED**  | This documentation                                |

---

## 6. Frontend Changes Applied

### `apiOperationMappers.ts` - Final State

```typescript
// Shopee Operations
export function mapShopeeOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock", // ✅ Fixed
    update_price: "/inventory/update-price", // ✅ Fixed
    export_orders: "/shopee/orders/export", // ⚠️ Verify
    get_token: "/platform-auth/initiate/shopee", // ✅ Fixed
    refresh_token: "/tokens/refresh/shopee", // ✅ Fixed
    update_code: "/shopee/operations/update-code", // ⚠️ Not in Go
  };
  // ...
}

// Lazada Operations
export function mapLazadaOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock", // ✅ Fixed
    update_price: "/inventory/update-price", // ✅ Fixed
    export_orders: "/lazada/orders/export", // ⚠️ Verify
    get_token: "/platform-auth/initiate/lazada", // ✅ Fixed
    refresh_token: "/tokens/refresh/lazada", // ✅ Fixed
    update_code: "/lazada/operations/update-code", // ⚠️ Not in Go
  };
  // ...
}

// TikTok Operations
export function mapTikTokOperation(operationName: string): EndpointMapping {
  const operationMap: Record<string, string> = {
    update_stock: "/inventory/update-stock", // ✅ Fixed
    update_price: "/inventory/update-price", // ✅ Fixed
    export_orders: "/tiktok/orders/export", // ⚠️ Verify
    get_token: "/platform-auth/initiate/tiktok", // ✅ Fixed
    refresh_token: "/tokens/refresh/tiktok", // ✅ Fixed
    update_code: "/tiktok/operations/update-code", // ⚠️ Not in Go
  };
  // ...
}

// Sheets Operations
export function mapSheetsOperation(operation: string): string {
  const sheetOperationMap: Record<string, string> = {
    wallet_to_sheets: "/shopee/wallet/export-to-sheets", // ✅ Fixed
    shipping_fee_to_sheets: "/shopee/shipping/export-to-sheets", // ✅ Fixed
  };
  return sheetOperationMap[operation] || "/execute-sheets";
}
```

---

## 7. Remaining Work (Backend Go)

### 7.1 Endpoints yang Perlu Diverifikasi

1. **Order Export per Platform**
   - `/api/shopee/orders/export`
   - `/api/lazada/orders/export`
   - `/api/tiktok/orders/export`

   Node.js menggunakan `/api/n8n/export-orders` yang generik. Go backend mungkin perlu endpoint spesifik per platform.

### 7.2 Endpoints yang Belum Ada di Go

1. **Update Code** (`/*/operations/update-code`)
   - Fungsi ini tidak ditemukan di Go backend
   - Perlu cek apakah fitur ini masih digunakan

2. **Fallback Execute** (`/execute`, `/execute-sheets`)
   - Node.js memiliki generic execute endpoint
   - Go menggunakan endpoint spesifik per operasi (lebih baik)

---

## 8. Testing Commands

```bash
# Build Frontend
cd frontend && npm run build

# Test Go Backend Health
curl http://localhost:3000/api/health

# Test Inventory Update Stock (Go)
curl -X POST http://localhost:3000/api/inventory/update-stock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -H "x-tenant-id: yumna_bertigamart" \
  -d '{"sku": "TEST-SKU", "platform": "shopee"}'

# Test Token Status (Go)
curl http://localhost:3000/api/token-status \
  -H "Authorization: Bearer <token>" \
  -H "x-tenant-id: yumna_bertigamart"

# Test Wallet Export to Sheets (Go)
curl -X POST http://localhost:3000/api/shopee/wallet/export-to-sheets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -H "x-tenant-id: yumna_bertigamart" \
  -d '{"month": 1, "year": 2026}'
```

---

## 9. Payload Format Reference

### Update Stock (Go Backend)

```json
{
  "sku": "SKU-001",
  "platform": "shopee", // Optional: single platform
  "platforms": ["shopee", "lazada", "tiktok"] // Optional: multiple
}
```

### Wallet Export to Sheets (Go Backend)

```json
{
  "month": 1,
  "year": 2026,
  "spreadsheet_id": "optional-spreadsheet-id",
  "sheet_name": "optional-sheet-name"
}
```

---

## 10. Conclusion

**Status: MOSTLY COMPLETE**

Frontend telah diperbarui untuk menggunakan endpoint Go backend yang benar. Sisa pekerjaan:

1. Verifikasi endpoint order export di Go backend
2. Pertimbangkan apakah `update_code` perlu dimigrasi ke Go

**Priority:** MEDIUM - Fitur utama (stock, price, auth) sudah berfungsi.

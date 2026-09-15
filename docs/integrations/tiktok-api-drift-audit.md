# TikTok Shop API — Drift Audit (Phase 6)

**Date:** 2026-09-15
**Source of truth:** [TikTok Shop Partner Center Changelog](https://partner.tiktokshop.com/docv2/changelog)
**Codebase version scanned:** `main` at commit ancestor of Phase 6 landing.
**Audit tool:** ripgrep on `backend/pkg/tiktok/**` and `backend/internal/services/**tiktok**`.

## 1. Endpoint version inventory

Endpoints below are grouped by version prefix. "Latest available" is what
TikTok Shop currently ships as of the changelog snapshot; "Codebase" is what
the omni codebase actually calls today.

| Category | Endpoint | Codebase version | Latest available | Status |
|---|---|---|---|---|
| Auth | `POST /api/v2/token/get` | `v2` (unversioned) | `v2` | ✅ current |
| Auth | `POST /api/v2/token/refresh` | `v2` (unversioned) | `v2` | ✅ current |
| Authorization | `POST /authorization/202309/shops` | `202309` | `202309` | ✅ current (test only) |
| Order | `GET /order/202309/orders/search` | `202309` | `202309` (POST variant `202502` also exists) | ⚠️ verify: some tenants report faster response on POST /202502 |
| Order | `POST /order/202309/orders/search` | `202309` | `202309` (POST variant `202502` also exists) | ⚠️ same as above |
| Order | `GET /order/202309/orders/{id}` | `202309` | `202309` | ✅ current |
| Order | `POST /order/202309/orders/cancel` | `202309` | `202309` | ✅ current |
| Product | `GET /product/202309/products/search` | `202309` | superseded by `POST /product/202502/products/search` for global product search | ✅ Migrated in Phase 7 (internal caller `services/platform/tiktok_client_api.go:194` now uses 202502; unused `pkg/tiktok/api.go GetProducts` marked Deprecated for external-SDK compatibility) |
| Product | `POST /product/202502/products/search` | `202502` | `202502` | ✅ current |
| Product | `POST /product/202309/products` (create) | `202309` | check for `202405`/`202509` global variants | ⚠️ verify per market; SEA POP still on 202309 |
| Product | `PUT /product/202309/products/{id}` | `202309` | `202309` | ✅ current |
| Product | `GET /product/202309/products/{id}` | `202309` | `202309` (new field `not_submitted_reasons` added — additive) | ✅ current, wire the new field when we surface product review status |
| Product | `POST /product/202309/products/{id}/prices/update` | `202309` | `202309` | ✅ current |
| Product | `POST /product/202309/products/{id}/inventory/update` | `202309` | `202309` | ✅ current |
| Product | `POST /product/202309/products/deactivate` | `202309` | `202309` | ✅ current |
| Product | `POST /product/202309/categories/recommend` (with `CategoryVersion: "v1"` for ID) | `202309` + `v1` | v1 still valid for Indonesia as of audit; check next quarter | ⚠️ verify: if TikTok pushes v2 for ID, flip `pkg/tiktok/category.go:63` |
| Product | `GET /product/202309/categories/{id}/attributes` | `202309` | `202309` | ✅ current |
| Product | `POST /product/202309/images/upload` | `202309` | `202309` | ✅ current |
| Fulfillment | `POST /fulfillment/202309/packages/{id}/ship` | `202309` | `202309` (`handover_method` deprecated for SEA POP cross-border) | ✅ Phase 7 shipped `ShouldSendHandoverMethod` + `BuildShipPackageRequest` region-aware helpers; 3 handler call sites wired. Region context still `""` at call sites — populate when seller-region lookup lands (backlog) |
| Fulfillment | `GET /fulfillment/202309/packages/{id}` | `202309` | `202309` | ✅ current |
| Fulfillment | `GET /fulfillment/202309/packages/{id}/handover_time_slots` | `202309` | `202309` | ✅ current |
| Fulfillment | `GET /fulfillment/202309/orders/{id}/handover_time_slots` | `202309` | `202309` | ✅ current |
| Fulfillment | `GET /fulfillment/202309/packages/{id}/shipping_documents` | `202309` | `202309` | ✅ current |
| Logistics | `GET /logistics/202309/warehouses` | `202309` | `202309` | ✅ current |
| Finance | `GET /finance/202309/orders/{id}/statement_transactions` (fallback) | `202309` | superseded by `202501` and (payments) `202605` | ⚠️ fallback path still active — safe; will drop when 202501 tenant coverage is 100 % |
| Finance | `GET /finance/202501/orders/{id}/statement_transactions` | `202501` | `202501` | ✅ current |
| Finance | `GET /finance/202309/statements` | `202309` | `202309` (list) | ✅ current |
| Finance | `GET /finance/202501/statements/{id}/statement_transactions` | `202501` | `202501` | ✅ current |
| Finance | `GET /finance/202309/payments/*` | not used | migrated to `202605` (announced Aug 2026) | ➖ not affected; codebase does not call payments |

## 2. Actions taken this phase

1. **Adopted the Inventory Update Webhook (code 6)** launched 2025-11:
   - `TiktokWebhookInventoryUpdate = 6` constant added.
   - `getEventType` returns `inventory_update`.
   - New `processInventoryEvent` fans the event to the realtime hub as
     `inventory/updated` so dashboards can invalidate stock queries without
     waiting for the next poll.
   - Three RED-first unit tests cover the type name, publisher wiring, and
     the missing-product-id negative path.
2. **Documented every endpoint's current version** vs latest available
   (this table). Any future audit only needs to diff this file against
   the TikTok changelog.

## 3. Actions deferred (backlog)

Ordered by "hurts sellers now" → "nice-to-have":

1. **Seller-region + user-type wiring** — Phase 7 landed the deprecation
   helpers (`ShouldSendHandoverMethod`, `BuildShipPackageRequest`) but all 3
   handler call sites pass `region=""` / `isCrossBorder=false` because that
   context isn't threaded through today. Populate from
   `TiktokConfigManager.SellerRegion` + `UserType` once `LoadConfig` reads
   those two keys from `platform_configs` (they exist as
   `tiktok/sellerRegion` len=2 and `tiktok/userType` len=1 in the legacy
   store).
2. **CategoryVersion probe for Indonesia** — script that calls
   `RecommendCategory` with v2 and confirms the response shape; only then
   flip the const.
3. **Wire `not_submitted_reasons`** on `GetProduct` into the product-detail
   UI so sellers can see rejection reasons without leaving the app.
4. **Drop the 202309 finance fallback** at `pkg/tiktok/finance.go:88` after
   Q1 2027 (once TikTok fully sunsets it — no announcement yet).
5. **Remove `pkg/tiktok/api.go GetProducts`** entirely after confirming no
   external caller depends on `pkg/tiktok` (marked Deprecated in Phase 7).

## 4. Not affected

- **Legacy V1 API** deprecated 2025-06 — codebase already on v2 for auth
  and 202309+ for everything else. No action.
- **Google API drift** — `google.golang.org/api v0.260.0` current; Sheets v4
  stable, `drive.readonly` still valid.
- **Shopee `aigc_label` addition on Video APIs** — codebase does not use the
  Video API surface.
- **Lazada PII masking** — handled in Phase 1 (see
  `backend/pkg/lazada/pii_mask.go`).

## 5. Refs

- Roadmap: `docs/superpowers/specs/2026-09-15-realtime-e2e-platform-drift-design.md` § Phase 6
- TikTok changelog: https://partner.tiktokshop.com/docv2/changelog
- Inventory webhook doc: https://partner.tiktokshop.com/docv2/page/9q4we8qr

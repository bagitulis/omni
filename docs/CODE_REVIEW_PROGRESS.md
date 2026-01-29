# CODE REVIEW PROGRESS - OMNI PROJECT

**Tanggal Review**: 29 Januari 2026  
**Status**: DALAM PROSES - Session 3 (README Consolidation)
**Reviewer**: AI Assistant

---

## 📋 SESSION SUMMARY

### Session 1: Initial Review & Documentation

- ✅ Comprehensive code review documented
- ✅ Updated README.md with detailed rules
- ✅ Fixed agentscopy.md contradictions
- ✅ Fixed 8 false positive patterns in order_manager.go

### Session 2: Naming Convention Fixes (snake_case)

- ✅ Fixed 9 backend files (JSON tags camelCase → snake_case)
- ✅ Fixed 6 frontend files (wholesale components)
- ✅ Backend build: SUCCESS
- ✅ Frontend build: SUCCESS

### Session 3: README Consolidation (CURRENT)

- ✅ Merged README.md + agentscopy.md into single comprehensive guide
- ✅ Added AI Workflow Decision Tree with Mermaid diagram
- ✅ Added detailed Validation Checklist
- ✅ Updated Naming Convention to **Hybrid Approach** (API snake_case, internal camelCase)
- ✅ Added Testing & Docker Build policies
- ✅ Updated this tracking document

---

## 🎯 ATURAN TEKNIS YANG WAJIB DIPATUHI (Updated from README.md)

### 1. Critical Rules (Top 10) - CANNOT BE VIOLATED

1. ❌ **NO FALSE POSITIVES** - Never `success: true` on errors
2. ❌ **NO ALIASES** - Fix names directly, no workarounds
3. ❌ **NO DEFAULT TENANT** - Always validate tenant_id
4. 🎯 **ALL JSON TAGS = snake_case** - API responses MUST use snake_case
5. 📏 **MAX 300 LINES PER FILE** - Exceptions: models 500, migrations unlimited
6. 🔐 **ALWAYS USE context.Context** - All DB/network operations
7. 🏗️ **CLEAN ARCHITECTURE** - Handler→Service→Repository
8. 📝 **STRUCTURED LOGGING ONLY** - Use zerolog, no fmt.Printf
9. 🗄️ **NO DB CHANGES WITHOUT MIGRATION** - scripts/postgres/\*.sql required
10. 🧪 **100% TEST SUCCESS REQUIRED** - Before marking complete

### 2. Naming Convention - HYBRID APPROACH (Updated!)

**New Policy**: Different layers use their natural conventions

| Layer        | Context              | Convention | Example                                 |
| ------------ | -------------------- | ---------- | --------------------------------------- |
| **Database** | PostgreSQL Columns   | snake_case | `item_id`, `product_name`, `created_at` |
| **Database** | PostgreSQL Tables    | snake_case | `shopee_orders`, `platform_configs`     |
| **Backend**  | Go Struct Fields     | PascalCase | `ItemID`, `ProductName`, `CreatedAt`    |
| **Backend**  | Go unexported vars   | camelCase  | `itemService`, `getTenant()`            |
| **Backend**  | GORM column tag      | snake_case | `gorm:"column:item_id"`                 |
| **Backend**  | JSON tag (API)       | snake_case | `json:"item_id"`                        |
| **Frontend** | API Type Definitions | snake_case | `item_id: number` (match API)           |
| **Frontend** | Local Variables      | camelCase  | `const itemId = data.item_id`           |
| **Frontend** | Component Props      | camelCase  | `defineProps<{ itemId: number }>`       |
| **Frontend** | Vue Template         | kebab-case | `:item-id="itemId"`                     |

**Why Hybrid?**

- API boundary uses snake_case (REST standard)
- Internal code uses natural language convention
- Better IDE support and developer experience

### 3. Testing Policy (Updated!)

**New Workflow:**

```
1. Complete ALL implementation tasks
2. Complete ALL cleanup tasks
3. THEN run tests:
   • go build ./...
   • go test ./...
4. If tests FAIL:
   • Fix the issues
   • Re-run tests
   • Repeat until 100% pass
5. DO NOT mark complete until:
   ✅ All builds pass
   ✅ All tests pass
   ✅ No false positives
```

**Rationale**: Avoids Docker issues during development, tests verify complete solution

### 4. Docker Build Policy (Updated!)

**Decision Tree:**

```
Changed only service logic (Go files)?
  → build.py smart (restarts service, ~30s)

Service error but code is correct?
  → build.py quick-fix (restart only, ~10s)

Changed dependencies/Dockerfile?
  → ASK USER for approval
  → If denied: use smart build
  → If approved: full rebuild (~5-10min)
```

**Default**: `build.py smart` → `build.py quick-fix` → ask for `full build`

---

## 📊 RINGKASAN TEMUAN

| Kategori                              | Backend | Frontend | Total | Severity |
| ------------------------------------- | ------- | -------- | ----- | -------- |
| Files > 300 lines                     | 9       | 53       | 62    | HIGH     |
| camelCase JSON (should be snake_case) | 100+    | -        | 100+  | HIGH     |
| Missing WithContext DB                | 30+     | -        | 30+   | MEDIUM   |
| False positive patterns               | 8       | -        | 8     | HIGH     |
| Silent error catching                 | -       | 9        | 9     | HIGH     |
| Duplicate code patterns               | 3 major | 3 major  | 6     | MEDIUM   |
| Hardcoded values                      | -       | 20+      | 20+   | LOW      |

---

## ✅ COMPLETED FIXES

### Backend Fixes (Session 1-2)

#### 1. False Positive Patterns [HIGH] - ✅ COMPLETED

**Fixed 8 instances in `internal/handlers/order_manager.go`:**

- Lines 74-82: Order sync service unavailable
- Lines 92-99: Failed to get orders
- Lines 148-155: Sync orders endpoint
- Lines 264-273: Failed to get locked orders
- Lines 321-329: Service not available
- Lines 468-475: Failed to get order today items
- Lines 517-524: Sync order today endpoint
- Lines 526-533: Service not initialized

**Status**: ✅ 100% Complete (8/8 fixed)

#### 2. camelCase JSON Tags [HIGH] - ✅ COMPLETED (Partial)

**Fixed 9 backend files:**

- ✅ `internal/handlers/lazada/orders.go` - 6 fields
- ✅ `internal/handlers/wholesale_dto.go` - 8 fields
- ✅ `internal/services/inventory/inventory_service.go` - 8 fields
- ✅ `internal/utils/jwt.go` - 2 fields
- ✅ `internal/utils/logger/logger.go` - 3 fields
- ✅ `internal/services/sync/delta_sync_service.go` - 10 fields
- ✅ `internal/services/shopee_escrow_service.go` - 20+ fields
- ✅ `internal/services/shopee/wallet_service.go` - 25+ fields
- ✅ `internal/services/shopee/shipping_service.go` - 15+ fields

**Status**: 🟡 In Progress (~10% complete, 90+ files remaining)

### Frontend Fixes (Session 2)

#### 1. Wholesale Components snake_case [HIGH] - ✅ COMPLETED

**Fixed 6 frontend files:**

- ✅ `src/services/wholesaleService.ts` - Removed duplicate methods
- ✅ `src/components/content/InventoryContent/WholesaleUpdateModal.vue`
- ✅ `src/components/content/InventoryContent/WholesaleBatchDeleteModal.vue`
- ✅ `src/components/content/InventoryContent/WholesaleSettingsTab.vue`
- ✅ `src/components/content/InventoryContent/composables/useWholesaleUpdate.ts`
- ✅ `src/components/content/InventoryContent/composables/useInventoryWholesale.ts`

**Status**: ✅ 100% Complete (wholesale module)

### Documentation Updates (Session 3)

#### 1. README.md Consolidation - ✅ COMPLETED

- ✅ Merged README.md + agentscopy.md into single source of truth
- ✅ Added Mermaid decision tree diagram
- ✅ Added comprehensive validation checklist
- ✅ Updated naming convention to Hybrid approach
- ✅ Added detailed testing & docker policies
- ✅ Removed emoji from production code examples (kept in docs only)
- ✅ Total length: ~1800 lines (comprehensive)

**Status**: ✅ 100% Complete

---

## 🔄 PENDING FIXES

### HIGH Priority (Must Fix Soon)

#### 1. Files > 300 Lines [HIGH]

**Backend (9 files):**

- [ ] `internal/handlers/order_manager.go` (632 lines) - Split into platform-specific handlers
- [ ] `internal/services/analytics/shopee_escrow_sync.go` (426 lines)
- [ ] `internal/services/analytics/intelligence/simulator.go` (372 lines)
- [ ] `internal/services/wholesale/shopee_wholesale_service.go` (333 lines)
- [ ] `internal/services/platform/shopee_client.go` (309 lines)
- [ ] `internal/handlers/wholesale_batch_handler.go` (309 lines)
- [ ] `internal/services/analytics/ml_service.go` (300 lines)
- [ ] `internal/handlers/filter_preference.go` (300 lines)
- [ ] `internal/handlers/auth_handler.go` (300 lines)

**Frontend (Top 20):**

- [ ] 484 lines: `src/views/analytics/ShopeeAnalytics.vue`
- [ ] 462 lines: `src/components/analytics/ml/ProductScoreTable.vue`
- [ ] 459 lines: `src/views/analytics/TiktokAnalytics.vue`
- [ ] 459 lines: `src/composables/useAnalytics.ts`
- [ ] 447 lines: `src/composables/useTiktokAnalytics.ts`
- [ ] 416 lines: `src/components/analytics/ml/PortfolioHealthCard.vue`
- [ ] 406 lines: `src/components/content/InventoryContent/InventoryContent.vue`
- [ ] 401 lines: `src/components/ProductManager/LazadaProductManager/LazadaProductManager.vue`
- [ ] 393 lines: `src/components/analytics/ShopeeAdsUpload.vue`
- [ ] 386 lines: `src/components/AddProduct/CategorySelector.vue`
- [ ] (+ 43 more files, see Appendix)

#### 2. Remaining camelCase JSON Tags [HIGH]

**~90+ files remaining** with camelCase JSON tags that need snake_case

**Recommended approach:**

- Use grep/search to find all remaining instances
- Batch fix by file type (handlers, services, DTOs)
- Update corresponding frontend TypeScript interfaces

#### 3. Missing WithContext in DB Operations [MEDIUM]

**30+ instances** missing `WithContext(ctx)`:

- [ ] `internal/config/shopee_client.go` (1 instance)
- [ ] `internal/handlers/wholesale_batch.go` (1 instance)
- [ ] `internal/services/spreadsheet/registry_service.go` (5 instances)
- [ ] `internal/services/route/config_service.go` (8 instances)
- [ ] `internal/services/autofunction/config_manager.go` (11 instances)
- [ ] `internal/services/autofunction/scheduler.go` (1 instance)
- [ ] `internal/services/autofunction/executor.go` (3 instances)
- [ ] `internal/services/jobs/queue_manager.go` (5 instances)

### MEDIUM Priority

#### 4. Duplicate Code Patterns [MEDIUM]

**Backend (3 major patterns):**

- [ ] Order Repository pattern (shopee/lazada/tiktok) - Extract generic
- [ ] Clone Service pattern (shopee/lazada/tiktok) - Unify
- [ ] Ads Analytics pattern (shopee/tiktok) - Merge

**Frontend (3 major patterns):**

- [ ] Analytics composables (~95% identical) - Merge with platform param
- [ ] Analytics views (~90% identical) - Create shared component
- [ ] Loading/Error pattern (100+ locations) - Extract `useAsyncOperation()`

#### 5. Frontend Silent Error Catching [HIGH]

**9 instances** of `.catch(() => {})`:

- [ ] `src/views/analytics/TiktokAnalytics.vue` (line 285)
- [ ] `src/views/analytics/ShopeeAnalytics.vue` (line 285)
- [ ] `src/store/app.ts` (line 91)
- [ ] `src/views/analytics/TiktokAdsAnalytics.vue` (line 144)
- [ ] `src/views/analytics/ShopeeAdsAnalytics.vue` (line 150)
- [ ] `src/main.ts` (lines 26, 42)
- [ ] `src/components/RouteMapper/useRouteMappingData.ts` (line 135)
- [ ] `src/components/OrderManager/composables/useOrderManager.ts` (line 100)

### LOW Priority

#### 6. Hardcoded Values [LOW]

- [ ] Timeout values in `src/services/api.ts` (4 instances)
- [ ] Port numbers (3000) in multiple files
- [ ] Move to config/environment variables

---

## 📈 TRACKING PROGRESS

### Backend Progress

| Task                | Total | Fixed | Remaining | % Complete  |
| ------------------- | ----- | ----- | --------- | ----------- |
| Files > 300 lines   | 9     | 0     | 9         | 0%          |
| camelCase JSON      | 100+  | 9     | 90+       | ~10%        |
| Missing WithContext | 30+   | 0     | 30+       | 0%          |
| False positives     | 8     | 8     | 0         | **✅ 100%** |
| Duplicate code      | 3     | 0     | 3         | 0%          |

**Overall Backend**: ~15% complete

### Frontend Progress

| Task                 | Total | Fixed | Remaining | % Complete  |
| -------------------- | ----- | ----- | --------- | ----------- |
| Files > 300 lines    | 53    | 0     | 53        | 0%          |
| Wholesale snake_case | 6     | 6     | 0         | **✅ 100%** |
| Silent error catch   | 9     | 0     | 9         | 0%          |
| Duplicate code       | 3     | 0     | 3         | 0%          |
| Hardcoded values     | 20+   | 0     | 20+       | 0%          |

**Overall Frontend**: ~10% complete

### Documentation Progress

| Task                      | Status  |
| ------------------------- | ------- |
| README.md consolidation   | ✅ 100% |
| AI Workflow Decision Tree | ✅ 100% |
| Validation Checklist      | ✅ 100% |
| Naming Convention Update  | ✅ 100% |
| Testing & Docker Policies | ✅ 100% |
| CODE_REVIEW_PROGRESS.md   | ✅ 100% |

**Overall Documentation**: ✅ 100% complete

---

## 🎯 NEXT STEPS (Recommended Order)

### Phase 1: Complete snake_case Migration (HIGH)

1. Backend: Fix remaining 90+ files with camelCase JSON tags
2. Frontend: Update TypeScript interfaces to match
3. Run full build & test to verify

### Phase 2: Add WithContext to DB Operations (MEDIUM)

1. Add context.Context parameter to all DB functions
2. Update callers to pass context
3. Test for performance/cancellation

### Phase 3: Refactor Large Files (HIGH)

1. Backend: Split 9 files > 300 lines
2. Frontend: Refactor 53 files > 300 lines (focus on top 20 first)

### Phase 4: Code Cleanup (MEDIUM-LOW)

1. Remove duplicate code patterns
2. Fix silent error catching
3. Move hardcoded values to config

---

## 📝 CHANGELOG

| Tanggal    | Aksi                                         | Oleh |
| ---------- | -------------------------------------------- | ---- |
| 2026-01-29 | Initial review dan dokumentasi               | AI   |
| 2026-01-29 | Fixed false positive patterns (8/8)          | AI   |
| 2026-01-29 | Fixed 9 backend files camelCase→snake_case   | AI   |
| 2026-01-29 | Fixed 6 frontend wholesale components        | AI   |
| 2026-01-29 | Consolidated README.md + agentscopy.md       | AI   |
| 2026-01-29 | Updated naming convention to Hybrid approach | AI   |
| 2026-01-29 | Updated testing & docker policies            | AI   |
| 2026-01-29 | Updated this tracking document               | AI   |

---

## 📚 APPENDIX

### A. Full List of Frontend Files > 300 Lines (53 files)

| Lines | File                                                                              |
| ----- | --------------------------------------------------------------------------------- |
| 484   | `src/views/analytics/ShopeeAnalytics.vue`                                         |
| 462   | `src/components/analytics/ml/ProductScoreTable.vue`                               |
| 459   | `src/views/analytics/TiktokAnalytics.vue`                                         |
| 459   | `src/composables/useAnalytics.ts`                                                 |
| 447   | `src/composables/useTiktokAnalytics.ts`                                           |
| 416   | `src/components/analytics/ml/PortfolioHealthCard.vue`                             |
| 406   | `src/components/content/InventoryContent/InventoryContent.vue`                    |
| 401   | `src/components/ProductManager/LazadaProductManager/LazadaProductManager.vue`     |
| 393   | `src/components/analytics/ShopeeAdsUpload.vue`                                    |
| 386   | `src/components/AddProduct/CategorySelector.vue`                                  |
| 383   | `src/services/api.ts`                                                             |
| 380   | `src/components/ProductManager/FilterPanel.vue`                                   |
| 379   | `src/composables/useScriptMonitorLogic.ts`                                        |
| 378   | `src/composables/useMLAnalytics.ts`                                               |
| 378   | `src/components/analytics/ShopeeAdsDashboard.vue`                                 |
| 376   | `src/components/content/InventoryContent/composables/useMarketplaceAllocation.ts` |
| 368   | `src/components/OrderManager/composables/useOrderManager.ts`                      |
| 365   | `src/components/content/InventoryContent/HeaderFilterDropdown.vue`                |
| 365   | `src/components/ProductManager/TiktokProductManager/TiktokProductManager.vue`     |
| 363   | `src/components/content/InventoryContent/InventoryTable.vue`                      |
| 361   | `src/components/content/InventoryContent/FilterDropdown.vue`                      |
| 359   | `src/components/content/InventoryContent/composables/useInventoryStockUpdate.ts`  |
| 354   | `src/components/analytics/ShopeeAdsDataTable.vue`                                 |
| 353   | `src/composables/useRouteCacheManager.ts`                                         |
| 352   | `src/components/content/SettingsContent/RouteMonitoringTab.vue`                   |
| 351   | `src/components/content/InventoryContent/CloneProductModal.vue`                   |
| 347   | `src/components/ProductManager/ShopeeProductManager/ShopeeProductManager.vue`     |
| 343   | `src/components/content/SettingsContent/RouteStatesSection.vue`                   |
| 338   | `src/components/analytics/TiktokAdsDataTable.vue`                                 |
| 337   | `src/components/content/SettingsContent/RouteManagementTable.vue`                 |
| 337   | `src/components/content/InventoryContent/composables/useInventoryConfig.ts`       |
| 331   | `src/components/AddProduct/ImageUploader.vue`                                     |
| 329   | `src/components/content/SettingsContent/RouteExcludedRoutesList.vue`              |
| 328   | `src/store/app.ts`                                                                |
| 327   | `src/components/analytics/AnalyticsSettingsModal.vue`                             |
| 326   | `src/components/content/SettingsContent/RouteFlowControls.vue`                    |
| 325   | `src/components/layout/sidebar/SidebarMenu.vue`                                   |
| 325   | `src/components/content/SettingsContent/RouteConfigModal.vue`                     |
| 324   | `src/components/SpreadsheetLinkField.vue`                                         |
| 323   | `src/components/content/InventoryContent/InventoryLockPanel.vue`                  |
| 320   | `src/views/Dashboard.vue`                                                         |
| 318   | `src/components/AddProduct/AddProductModal.vue`                                   |
| 316   | `src/components/RouteMapper/useRouteMappingData.ts`                               |
| 314   | `src/components/content/SettingsContent/RouteExcludedRoutesManager.vue`           |
| 314   | `src/components/AddProduct/VariantBuilder.vue`                                    |
| 313   | `src/components/analytics/TiktokAnalyticsSettingsModal.vue`                       |
| 312   | `src/views/DevPreview.vue`                                                        |
| 310   | `src/components/content/SettingsContent/WebhookSettings.vue`                      |
| 308   | `src/components/content/SettingsContent/ConfigTab.vue`                            |
| 306   | `src/components/content/SettingsContent/RoutePerformanceRow.vue`                  |
| 305   | `src/views/LoginView.vue`                                                         |
| 303   | `src/components/content/SettingsContent/PlatformCard.vue`                         |
| 301   | `src/components/content/InventoryContent/composables/useInventoryWholesale.ts`    |

---

**CATATAN PENTING**:

- README.md sekarang menjadi **SINGLE SOURCE OF TRUTH** untuk semua aturan development
- agentscopy.md akan dihapus setelah merge selesai
- Semua perbaikan di masa depan harus mengikuti aturan di README.md versi baru
- Naming convention sekarang menggunakan **Hybrid Approach** - lebih natural dan practical

**Last Updated**: 2026-01-29 (Session 3 - README Consolidation Complete)

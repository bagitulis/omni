# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- **Task 6 — Backend token refresh lifecycle safeguards**: Added canonical credential lifecycle helpers with store-scoped active-row refresh locking, version compare-and-swap token updates, disabled-row stale refresh protection, refresh failure status/audit recording that preserves previous tokens, and focused multi-store/concurrency/disconnect/no-secret tests. (credential_connection.go, credential_repository_refresh.go, credential_lifecycle_service.go, credential_lifecycle_service_test.go)
- **Task 5 — Backend OAuth state binding and callback URL safety**: Hardened OAuth initiation and callback flow with signed expiring state claims bound to tenant/platform/attempt/user/session/CSRF nonce/redirect intent, single-use attempt validation, wrong-platform and wrong-scope rejection, duplicate-store reconnection checks, and a non-empty TikTok callback URL path. Added focused replay/expiry/cancel/redaction tests and sanitized OAuth logging. (oauth_handler.go, oauth_repository.go, oauth.go, oauth_connection_attempt.go, oauth_service.go, tiktok_oauth.go, oauth_state_security_test.go)
- **Dry-run credentials tool**: Created `backend/cmd/dry-run-credentials/main.go` — a standalone Go CLI that connects to PostgreSQL, queries active tenants, inspects platform_configs table structure (key-value vs structured), and prints a redacted count-by-platform report. Uses database/sql + lib/pq with env-var config and no secret exposure. Added github.com/lib/pq dependency (v1.12.3). (main.go, go.mod, go.sum)
- **Task 4 — Backend canonical credential model + repository**: Hardened the canonical credential repository contract with explicit tenant/platform/store scope, disabled-connection filtering, explicit OAuth attempt tenant/platform ownership, and deterministic app-config upsert handling. Added masked store identifier helpers, Fernet-backed encryption round-trip coverage, redaction tests, cross-tenant denial tests, incomplete-record coverage, and separate audit/OAuth attempt tests. Registered canonical credential models in tenant migration. (credential_connection.go, credential_app_config.go, oauth_connection_attempt.go, credential_audit_event.go, credential_repository.go, credential_repository_test.go, migration.go)
- **Task 4 validation hardening**: Added exact validation for empty app config platform and OAuth attempt platform/attempt_id in CredentialRepository with focused tests proving `platform is required` and `attempt_id is required` errors. (credential_repository.go, credential_repository_test.go)
### Added
- **Task 8 — Frontend analytics test suite**: Created `frontend/src/lib/analyticsHelpers.test.ts` (22 tests covering formatMonthYear, formatDate, getPriceDiffStatus, formatPriceDiff, getReconciliationHealth, exportToCSV), `frontend/src/api/analytics.test.ts` (29 tests covering all 16 Shopee+TikTok API functions with endpoint/params/error assertions), and `frontend/src/hooks/useAnalytics.test.ts` (32 tests covering all 7 hooks + analyticsKeys with platform dispatch verification). Also fixed `formatDate` to return "-" for invalid date strings (e.g. `new Date("not-a-date")`). All 82 tests pass with zero lint errors. (analyticsHelpers.test.ts, analytics.test.ts [api], useAnalytics.test.ts)

### Fixed
- **Shopee shipping labels**: Restricted `shipping_document_type` resolution to Shopee's accepted document types and removed obsolete fallback values (`THERMAL_WAYBILL`, `NORMAL_WAYBILL`) so label creation no longer sends invalid SDK parameters. Added regression coverage for accepted fallbacks, Shopee parameter priority, and invalid requested types. (shipping_label.go, shipping_label_test.go)
- **Credentials inventory dry-run test**: Fixed the Postgres fixture column list, corrected key-value row-shape detection so tenant-scoped key-value rows are not misclassified as mixed, and aligned structured backfill assertions with the service contract. (credentials_inventory_service.go, credentials_inventory_service_test.go)
- **Docker-dependent tests panic on Windows**: Replaced non-functional deferred panic recovery with proper `isDockerAvailable()` check that runs `docker ps` via `os/exec` before calling testcontainers. ~28 Docker-dependent tests across repositories, webhooks, and master_product now SKIP gracefully (not FAIL) with "Docker is not available on this platform". Removed `"strings"` import, added `"os/exec"`. (database.go)
- **Shopee report page**: Added missing `toCsvRows` and `isShopeeShippingOrder` helper functions that were omitted during initial page creation. (ShopeeReportPage.tsx)
- **Month/year validation missing in analytics handlers**: Added shared `parseMonthYear` helper that validates month (1-12) and year (2000-2100) query params, returning 400 Bad Request on invalid values. Applied to 5 handler methods per file in Shopee and TikTok analytics handlers. Removed unused `strconv` and `time` imports. (shopee_analytics_handler.go, tiktok_analytics.go)
- **Force-resync Delete error handling in escrow sync**: Added `.Error` checking on all `Delete()` calls in `SyncMonthWithProgress` force-resync cleanup, returning wrapped errors on failure. (shopee_escrow_sync.go, tiktok_escrow_sync.go)
- **PostgreSQL GROUP BY in reconciliation queries**: Fixed `Group("sku, item_name")` to use full SQL expressions (`COALESCE(model_sku, sku, '')`, `COALESCE(seller_sku, '')`) instead of aliases, fixing `column must appear in GROUP BY` error on PostgreSQL. (shopee_analytics.go, tiktok_analytics.go)

### Added
- **Task 8 — Shared frontend report layer**: Created `frontend/src/types/analytics.ts` (15 interfaces + 2 type aliases matching backend DTOs), `frontend/src/api/analytics.ts` (16 API functions for Shopee/TikTok), `frontend/src/hooks/useAnalytics.ts` (9 TanStack Query hooks with platform dispatch), `frontend/src/lib/analyticsHelpers.ts` (8 utility helpers including CSV export), and 10 shared components under `frontend/src/components/analytics/common/` (StatCard, ReconciliationSummaryCards, ReconciliationTable, ShippingFeeSummaryCards, ShippingFeeTable, SyncProgressCard, ReportSettingsModal, ReportFilters, ReportToolbar, ReportPageHeader). TypeScript and build pass with zero errors. (analytics.ts, index.ts, analytics.ts [api], useAnalytics.ts, analyticsHelpers.ts, 10 components)

### Added
- **Task 5 — Shopee report services**: Created `backend/internal/services/analytics/` package with `ShopeeAnalyticsService` (8 methods matching handler interface) and `ShopeeEscrowSyncService` (implements `jobs.EscrowSyncService`), plus helpers and 10 service-level tests. (shopee_analytics.go, shopee_escrow_sync.go, helpers.go, shopee_analytics_test.go)
### Added
- **Task 7A — Escrow sync job handler**: Created `backend/internal/services/jobs/escrow_sync_handler.go` with `EscrowSyncService` interface, `EscrowSyncHandler` struct, `HandleShopeeEscrowSync`/`HandleTiktokEscrowSync` methods (JobHandler-compatible), and DI setters for platform service injection. (escrow_sync_handler.go)
### Added
- **Task 6 — TikTok report services**: Created backend/internal/services/analytics/tiktok_analytics.go with TiktokAnalyticsService (8 methods matching handler interface) and tiktok_escrow_sync.go with TiktokEscrowSyncService (implements jobs.EscrowSyncService), plus tiktok_analytics_test.go with 11 service-level tests. Uses TikTok-specific model types (TiktokEscrowSync, TiktokEscrowOrder, TiktokEscrowItem) and DTOs (TiktokReconciliationResultDTO, TiktokShippingFeeResultDTO). No ads/ML/unified dependencies.

### Added
- **Report routes**: Restored `backend/internal/routes/report_routes.go` with `RegisterShopeeAnalyticsRoutes` and `RegisterTiktokAnalyticsRoutes` (16 report endpoints, no ads/ML/unified). (report_routes.go)
- **Analytics DTO types**: Added `backend/internal/dto/analytics_dto.go` with all report DTO types (AnalyticsSettings, SyncStatus, Reconciliation, ShippingFee, SKU groups for Shopee/TikTok). (analytics_dto.go)
- **Shopee analytics handler**: Created `backend/internal/handlers/shopee_analytics_handler.go` with `ShopeeAnalyticsService` interface, `ShopeeAnalyticsHandler` struct, and 8 handler methods (GetSettings, SaveSettings, GetSyncStatus, SyncEscrow, DeleteSyncData, GetReconciliation, GetShippingFeeAnalysis, RepopulateItems). (shopee_analytics_handler.go)
- **TikTok analytics handler**: Created `backend/internal/handlers/tiktok_analytics.go` with `TiktokAnalyticsService` interface, `TiktokAnalyticsHandler` struct, and 8 handler methods using TikTok-specific DTOs. (tiktok_analytics.go)
- **Handler unit tests**: Added `shopee_analytics_handler_test.go` (14 tests) and `tiktok_analytics_test.go` (11 tests) with manual mock test doubles covering missing tenant (401), valid tenant (200), service errors (500), and invalid body (400). (shopee_analytics_handler_test.go, tiktok_analytics_test.go)
- **Task 7B — Backend wiring for analytics routes and escrow sync**: Added `ShopeeAnalyticsHandler`/`TiktokAnalyticsHandler` fields to App struct with init in `initHandlers()`; updated handler `getService()` to create real analytics services via `config.GetTenantDBByID`; added `EscrowSyncServiceFactory` type and factory support to `EscrowSyncHandler`; wired escrow sync handlers with `JobExecutor.RegisterHandler` in `startBackgroundJobExecutor()`; registered `/api/analytics/shopee/*` and `/api/analytics/tiktok/*` routes in main.go. Build passes, 2917 tests pass, no ads/ML/unified imports. (app.go, app_handlers.go, main.go, escrow_sync_handler.go, shopee_analytics_handler.go, tiktok_analytics.go)

### Added
- **Booking Tab**: Added Booking tab as first tab in order manager with dedicated BookingOrdersTable component (Booking SN, Order SN, Booking/Match Status, Recipient, Items, Courier, Created/Updated, Actions). Tab icon uses FileTextOutlined. (OrderStatusTabs.constants.ts, BookingOrdersTable.tsx, OrderStatusTabs.tsx)
- **useOrders enabled option**: Added `enabled` option to useOrders hook to conditionally skip queries (used for booking tab). (useOrders.ts)
- **Booking sync error handling**: Booking sync catches partial failure gracefully (console.warn only, no error toast). (useOrderSync.ts)
|- **BookingDetailDrawer**: Added BookingDetailDrawer component with collapsible sections (Order Info, Recipient, Items, Courier, Timeline, Amounts, Platform Info, Memo) for viewing booking order details. Includes BookingDetailDrawerSections sub-components for each section. Added unit tests for render states and interaction.

### Fixed
- **Handler tests**: Fixed 3 pre-existing test failures in `backend/internal/handlers/` — CSRF HttpOnly assertion, inventory platform-status nil DB handling, webhook processor configuration and timestamp validation. (csrf_handler_test.go, inventory_stats_test.go, webhooks_test.go)

### Changed
- **OrdersPage**: BookingOrdersTable rendered when active tab is "booking"; OrdersBulkActionsBar hidden for booking tab. (OrdersPage.tsx)
- **useOrdersLogic**: Added "booking" to ORDER_MANAGER_VISIBLE_TABS; skips useOrders call for booking tab. (useOrdersLogic.ts)
|- **OrdersPage**: Wired BookingDetailDrawer with platform prop; booking table onViewDetail triggers drawer open with selected booking_sn. (OrdersPage.tsx)

### Fixed
- **Shopee Booking Orders frontend audit**: Replaced parent order console stubs with order detail navigation, improved booking loading/empty/error copy with retry actions, supported pagination totals, added external refresh trigger, and aligned booking item ID types with backend responses. (BookingOrdersTable.tsx, BookingDetailDrawerSections.tsx, booking.ts)
- **Shopee Booking Orders backend audit**: Rejected non-Shopee booking sync platforms, added list pagination metadata, exposed parent_order detail metadata, made booking+items persistence transactional per booking, and preserved item_id/model_id/quantity in Shopee booking item mapping. (backend booking handlers/services/platform SDK)
- **BookingDetailDrawer tests**: Fixed text matching assertion — corrected `findAllByText` to `findByText` for unique text queries in BookingDetailDrawer unit tests. (BookingDetailDrawer.test.tsx)
### Changed
- **rules-master maintenance**: Added rules-master-maintenance and sisyphus-e2e-evaluate blocks to prompt_blocks. Removed stale extension-based-testing block. Added to sisyphus + oracle prompt_compose. (rules-master/rules.json, opencode-configs/opencode-profiles.json)

### Fixed
- **rules-master validation**: --check now exits 1 when declared AGENTS.md target is missing (was warning + exit 0). Removed dead duplicate compose_prompt code, duplicated target init/filter blocks, and stale compose_json-disabled comments. Documented go_rules as absent, svelte-frontend-rules as unmanaged, and .agents/skills vs .opencode/skills split. (rules-master/sync_rules.py, notepads, evidence)

- **sync_rules.py hardening**: compose_prompt unknown block now exits 1 (was warning+skip). matches_target() extracted as shared function for all 3 target loops (gen, inject, compose). compose_json "not found" now triggers has_warnings+any_changed (was silently ignored). (rules-master/sync_rules.py)

### Added

- `GET /api/dev/audit-logs` endpoint for cross-tenant audit log viewing in developer panel (supports action, user_id, start_date, end_date, page, page_size filters)

- **Shopee Booking Orders handlers+routes**: Created `shopee_booking_handler.go` with SyncBookingOrders/GetBookingOrders/GetBookingOrderDetail. Modified `order_sync.go` SyncByCategory to route category=booking to booking sync service. Registered GET /booking and GET /booking/:booking_sn routes before catch-all in `additional_routes.go`. (handlers/shopee_booking_handler.go, handlers/order_sync.go, routes/additional_routes.go)

- **Shopee Booking models**: Registered `ShopeeBooking` and `ShopeeBookingItem` in tenant migration for GORM AutoMigrate

- **Shopee Booking repository**: Created `GormBookingRepository` with UpsertBookings, ReplaceBookingItems, ListBookings, GetBookingDetail. Added ItemCount/HasParentOrder to Booking DTO.
|- **Shopee Booking client methods**: Created `GetBookingList` and `GetBookingDetails` on `ShopeeAPIClient` in `platform/shopee_client_bookings.go` with token refresh, cursor pagination, and batch detail fetching (max 50).
- **Shopee Booking sync service**: Added `BookingSyncService` with Shopee manager adapter, booking detail hydration, idempotent repository upsert, item replacement, read/detail orchestration, and partial failure reporting.
- **Shopee Booking sync tests**: Added unit tests for `rawToBooking` transform and `SyncBookings` partial failure handling with mock manager/repository

- **Shopee Booking frontend API layer**: Created `types/booking.ts` with snake_case interfaces (`Booking`, `BookingItem`, `BookingListResponse`, `BookingDetailResponse`). Added `getBookingOrders`, `getBookingOrderDetail`, `syncBookingOrders` to `api/orders.ts`. Added `"booking"` tab to `orderTabMapping.ts` (endpoint, sync category, syncable flag). Created `bookingTransforms.ts` with pure display formatting functions. Added unit tests for all mapping and transform functions. (frontend/src/types/booking.ts, api/orders.ts, api/orderTabMapping.ts, pages/orders/utils/bookingTransforms.ts, test files)
- **Shopee Booking handler tests**: Added unit tests for `GetBookingOrders`, `GetBookingOrderDetail`, and `SyncByCategory` booking routing — covers missing tenant, non-Shopee platform rejection, missing booking_sn, and valid request parsing. No DB dependency. (handlers/shopee_booking_handler_test.go)
### Fixed

- `system.tenants` table missing `deactivated_at` column causing `ListTenants` and `DeactivateTenant` to fail with SQLSTATE 42703 — added migration in `MigrateSystemDatabase` to add column
- `shopee_skus` query in platform status function referenced non-existent `sku_id` column causing SQL error (SQLSTATE 42703) and 500 on `/api/monitoring/health/detailed` — replaced with `COALESCE(CAST(model_id AS TEXT), '')`
- Nil pointer dereference in `GetDetailedHealthMonitoring` when `MonitoringHandler.db` is nil (created via `NewSimpleMonitoringHandler()`) — added nil guard in `checkDatabase()`
- MonitoringHandler initialized with nil DB (`NewSimpleMonitoringHandler()`) causing health endpoint to always show "Disconnected" — now uses `NewMonitoringHandler(db, nil)` with actual DB connection

- Bulk handler tests updated to match new `items: [{user_id, tenant_id}]` request struct (previously referenced old `user_ids` field)
- `DevLoginInfo` and `DevLogin` handlers panic with nil pointer when `tenantService` is nil (added nil guards, test assertions updated)

- Bulk reset passwords and bulk disable users now support cross-tenant operations (backend accepts `items: [{user_id, tenant_id}]` instead of flat `user_ids` + single `tenant_id`)
- Frontend BulkActionModal passes per-user tenant_id to backend for correct multi-tenant bulk operations

### Added
- `GET /api/dev/tenants` endpoint returning all tenants (active and inactive) with user counts, created_at, and deactivated_at for developer panel

- `UsersTab` component for developer panel with cross-tenant user search (debounced 300ms, min 2 chars), results table with row selection, bulk action bar
- `ResetPasswordModal` component with password validation (min 8 chars, upper + lower + number)
- `BulkActionModal` component for bulk reset passwords and bulk disable (uses TypeToConfirmModal for disable confirmation, shows success/failure result summary)
- `SettingsTab` component for developer panel displaying read-only environment info (App Version, Environment, DB Driver, Go Version, Node Env, Build Time, API Base URL, Dev Login, Active Tenants)
- `GET /api/dev/environment` backend handler returning non-sensitive environment info (no secrets exposed)
- `developer_environment.go` split file for environment handler (SRP compliance)
- `TenantsTab` component for developer panel with tenant list table (Name, Status, Users, Created, Actions columns)
- `CreateTenantModal` component with name validation (regex `^[a-z][a-z0-9_]{2,49}$`), error display in modal
- Deactivation flow using `TypeToConfirmModal` with tenant name confirmation
- `SystemTab` component for developer panel with system health dashboard (DB status, memory, goroutines, uptime, components) and audit trail viewer (table with filters, pagination)
- `POST /api/dev/users/bulk-reset-password` endpoint for bulk password reset across tenants (max 50 users, partial success support)
- `POST /api/dev/users/bulk-disable` endpoint for bulk user disable via account locking (max 50 users, partial success support)
- `developer_handler_bulk.go` split file for bulk operations (SRP compliance)
- Tests for bulk operations: empty user_ids, exceeds max, weak password, success, missing fields, partial failure
- `GET /api/dev/users/search` endpoint for cross-tenant user search by username or email
- `SearchUsers` handler in developer_handler.go with query validation, timeout, and 50-tenant cap
- `UserSearchResult` struct with tenant_id/tenant_name fields (no password exposure)
- Tests for SearchUsers: missing query, too-short query, valid query, empty results
- `TenantService.CreateTenant()` method for creating new tenants (validates name, creates schema, runs migrations, registers in system.tenants)
- `POST /api/dev/tenants` endpoint in developer handler for tenant creation with 400/409/500 error handling
- `TenantValidationError` and `TenantDuplicateError` typed errors for CreateTenant
- `CreateTenantResult` struct for tenant creation response
- Tenant name validation regex (`^[a-z][a-z0-9_]{2,49}$`) with SQL injection prevention
- Rollback mechanism: drops schema on migration or registration failure
- Tests for CreateTenant handler (missing name, invalid name, duplicate, success)
- Tests for tenant name pattern validation (17 cases covering valid/invalid inputs)
- `TenantService.DeactivateTenant()` method for soft-deleting tenants (sets `is_active=false`)
- `DELETE /api/dev/tenants/:id` endpoint in developer handler for tenant deactivation
- `ErrTenantNotFound` sentinel error for tenant service
- Tests for DeactivateTenant handler (missing ID, not found, success)
- Test for `ErrTenantNotFound` error wrapping in tenant service

### Changed

- `GetAvailableTenants()` now queries `system.tenants` table with `is_active=true` filter (previously queried `information_schema.schemata`)
### Changed

- Refactored `DeveloperPanelPage` into thin tabbed shell with URL-synced tabs (`?tab=overview|tenants|users|system|settings`)
- Extracted overview content into `frontend/src/pages/developer/tabs/OverviewTab.tsx`

### Fixed

- Seed password changed from blocklisted `password123` to compliant `DevOwner@2024`
- `ResetPassword` service method now validates password strength before hashing (prevents weak passwords via admin reset)

### Added
- `POST /api/dev/reset-password` endpoint for developer panel password reset (cross-tenant)
- `developer_handler_test.go` with validation, weak password, and success test cases

### Added
- `TypeToConfirmModal` shared component for destructive action confirmation (type-to-confirm pattern)
- Developer panel API interfaces (`TenantDetail`, `CrossTenantUser`, `SystemHealth`, `AuditLogEntry`, `BulkOperationResult`, `EnvironmentInfo`) and function stubs in `frontend/src/api/developer.ts`

- Retry-After header respect in Shopee, Lazada, and TikTok SDK clients (429 responses)
- Minimum 1s exponential backoff for rate limit (429) responses across all 3 platform clients
- Max retries increased from 3 to 5 for all platform SDK clients
- Global concurrency semaphore (max 20 jobs) and per-tenant limit (max 3 jobs) in MultiTenantExecutor
- Panic recovery in job goroutines with automatic job failure marking

### Security

- Webhook body size limit (1MB) on all webhook handlers (Shopee, Lazada, TikTok)
- Fail-closed webhook behavior: reject with 503/400 when processor is nil or tenant_id is missing
- TikTok webhook timestamp replay protection (±5 min tolerance window)
- Webhook signature errors now return 401 Unauthorized instead of 400 Bad Request

### Changed

- Server context propagation to job executors: MultiTenantExecutor now receives a cancellable server context for coordinated shutdown
- Graceful shutdown sequence: executor Stop() called before HTTP server shutdown, timeout increased to 60s
- Job execution uses server context instead of context.Background() for proper cancellation propagation

### Fixed

- Safe type assertions in token_manager.go singleflight callbacks (comma-ok pattern prevents panics)

- Safe type assertions in token_manager.go singleflight callbacks (comma-ok pattern prevents panics)

### Added

- Secure HTTP client (`httputils.SecureGet`, `SecureGetAPI`) with host allowlist, HTTPS enforcement, and response body size limits
- SSRF protection: all outbound HTTP GET calls now validated against AllowedHosts whitelist
- ListUsers API now supports pagination (page/limit query params, default page=1 limit=20)
- SSE ticket auth system: POST /api/auth/sse-ticket exchanges JWT for short-lived one-time ticket
- Last owner protection: cannot delete or demote the last owner of a tenant
- UserRepository: Count, FindPaginated, CountByRole methods
- User Management frontend: full CRUD page with table, create/edit modals, role-based actions
- Admin role now has users.create and users.delete permissions
- Sidebar shows User Management menu item for authorized roles (admin+)
- Route /users protected by users.list permission guard
- Tenant-scoped User Management: UserHandler now builds per-request UserRepository from tenant schema instead of SystemDB, enabling proper tenant isolation (`/api/users` returns users of requested tenant, not system schema)
- Developer cross-tenant override: middleware/tenant.go honors `x-tenant-id` header for `developer` role only, enabling tester to switch tenants; non-developer users remain locked to JWT tenant
- Data migration scripts/fix-user-hierarchy.sql: promote `yumna` admin→owner in bertigamart tenant; create `nusseyba` owner in nusseyba tenant
- Developer Panel: dedicated /developer page (developer role only) with cross-tenant overview — tenant list, user counts by role (owner/admin/user), health status, and "Manage Users" button that switches tenant and navigates to /users
- Backend GET /api/dev/overview endpoint (developer-only, RequireRole guard) returns all tenants with user counts grouped by role
- Sidebar shows "Developer Panel" menu item for developer role only
- Route /developer protected by ProtectedRoute role="developer"

### Fixed

- SSRF vulnerability in TikTok OAuth token exchange (`http.Get` replaced with secure client)
- SSRF vulnerability in Lazada OAuth token exchange (`http.Get` replaced with secure client)
- SSRF vulnerability in TikTok handler `doTiktokTokenRequest` (`http.Get` replaced with secure client)
- Unbounded response body read in `DownloadImage` (now capped at 50MB via `io.LimitReader`)
- Disable dead Price Sync button with "coming soon" message instead of empty callback
- Add retry logic with timeout to image download service (retry on 5xx/network errors, no retry on 4xx)
- Fix silent error swallowing in inventory sync_service (json.Marshal) with proper zerolog error logging
- Fix silent error swallowing in route scanner_service (filepath.Walk) with proper zerolog warning
- SSE ticket handler now returns `{success, data: {ticket}}` (was `{success, ticket}` flat, caused frontend to fall back to JWT-in-URL)
- Missing `strconv` import in backend/internal/handlers/tiktok/orders.go
- UserManagementPage race condition: useEffect now gated on accessToken, wrapped in try/catch
- ListUsers response schema: nested as `{success, data: {users, total, page, limit}}` to match frontend contract
- SSE /notifications/stream moved outside Auth middleware group — EventSource cannot send Authorization header, ticket-based auth is now handled inside the handler via ValidateSSETicket()
- DeveloperPanelPage handleManageUsers now uses correct /auth/switch-tenant response shape ({token, tenant_id} via response.Success wrapper)

## [1.0.0] - 2026-05-02

### Added

- Multi-platform e-commerce integration (Shopee, Lazada, TikTok Shop)
- Smart PostgreSQL backup/restore with change detection and SHA-256 checksums
- React 19 frontend with Ant Design 5
- JWT authentication with refresh tokens and rate limiting
- Multi-tenant schema isolation (tenant_{id} pattern)
- Automated build system with 22+ error pattern recovery
- MCP server implementations for platform APIs

### Security

- SHA-256 backup integrity verification
- Pre-restore safety backups
- Data encryption for sensitive fields (Fernet)
- CSRF protection and CORS configuration
- Account lockout after failed login attempts

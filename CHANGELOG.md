# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **rules-master maintenance**: Added rules-master-maintenance and sisyphus-e2e-evaluate blocks to prompt_blocks. Removed stale extension-based-testing block. Added to sisyphus + oracle prompt_compose. (rules-master/rules.json, opencode-configs/opencode-profiles.json)

### Fixed
- **rules-master validation**: --check now exits 1 when declared AGENTS.md target is missing (was warning + exit 0). Removed dead duplicate compose_prompt code, duplicated target init/filter blocks, and stale compose_json-disabled comments. Documented go_rules as absent, svelte-frontend-rules as unmanaged, and .agents/skills vs .opencode/skills split. (rules-master/sync_rules.py, notepads, evidence)

### Added

- `GET /api/dev/audit-logs` endpoint for cross-tenant audit log viewing in developer panel (supports action, user_id, start_date, end_date, page, page_size filters)

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

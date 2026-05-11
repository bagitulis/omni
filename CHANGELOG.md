# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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

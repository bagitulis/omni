# PRD: User & Tenant Management Panel

## Overview

**Product**: OMNI Multi-Tenant E-Commerce Management System  
**Feature**: User Management Admin Panel + Security Hardening + Bug Fixes  
**Author**: Sisyphus (AI Agent)  
**Date**: 2026-05-11  
**Status**: Approved for Implementation

## Problem Statement

1. **No User Management UI** — Backend CRUD endpoints exist (`/api/users`) with RBAC, but no frontend page to manage users within a tenant.
2. **Sidebar shows all items to all roles** — No permission-based menu filtering.
3. **Permission mismatch** — Backend allows admin to create "user" role accounts, but frontend `permissionService` doesn't grant `users.create` to admin.
4. **Pre-existing bugs** — JWT in SSE URL, dead buttons, missing error handling, no pagination.
5. **Missing safety guards** — No "last owner" protection, no pagination on user list.

## Goals

- Provide a complete User Management page accessible to developer/owner/admin roles
- Enforce role hierarchy visually (UI matches backend RBAC rules)
- Fix all identified pre-existing bugs and security issues
- Add pagination to prevent performance issues at scale
- Protect against destructive edge cases (last owner deletion, self-deletion)

## Non-Goals

- Tenant creation/deletion UI (tenants are managed at infrastructure level)
- User self-registration (users are created by admins/owners)
- Token revocation on user deletion (accept 15min natural expiry)
- OAuth/SSO user management

## User Roles & Permissions

| Action | developer | owner | admin | user |
|--------|-----------|-------|-------|------|
| View user list | ✅ | ✅ | ✅ | ❌ |
| Create user (any role below) | ✅ | ✅ | ✅ (user only) | ❌ |
| Edit user (any role below) | ✅ | ✅ | ✅ (user only) | ❌ |
| Delete user (any role below) | ✅ | ✅ | ✅ (user only) | ❌ |
| Unlock locked account | ✅ | ✅ | ✅ | ❌ |
| See User Management menu | ✅ | ✅ | ✅ | ❌ |

**Hierarchy**: developer > owner > admin > user  
**Rule**: Actor can only manage users with strictly lower role level.

## Technical Design

### Phase 1: Permission Fix + Sidebar

**1.1 Fix `permissionService.ts`**
- Grant admin: `users.list`, `users.create`, `users.edit`, `users.delete` (backend already allows, restricted by CanManageRole)

**1.2 Sidebar Role Filtering (`Sidebar.tsx`)**
- Use `usePermission` hook to filter menu items
- Add "User Management" item (icon: `team`, path: `/users`)
- Only visible when `canManageUsers` is true

**1.3 Route (`App.tsx`)**
- Add `/users` route with `<ProtectedRoute role="admin">` guard

### Phase 2: User Management Page

**2.1 API Layer (`frontend/src/api/users.ts`)**
- `getUsers(params?: {page, limit})` → GET /api/users
- `createUser(data)` → POST /api/users
- `updateUser(id, data)` → PUT /api/users/:id
- `deleteUser(id)` → DELETE /api/users/:id
- `unlockUser(id)` → POST /api/users/:id/unlock

**2.2 Page (`frontend/src/pages/users/UserManagementPage.tsx`)**
- Ant Design Table with pagination
- Columns: Username, Email, Role (color badge), Status, Created, Actions
- Search/filter by username or email
- "Add User" button (hidden if no permission)
- Loading/empty states

**2.3 Column Factory (`frontend/src/pages/users/utils/userColumns.tsx`)**
- Role badges: developer=purple, owner=gold, admin=blue, user=green
- Status: Active (green), Locked (red with lock icon + unlock time)
- Actions: Edit, Delete, Unlock — conditionally rendered per hierarchy

**2.4 Modals**
- `CreateUserModal` — username, email, password (min 8, show requirement), role dropdown (filtered by actor level)
- `EditUserModal` — username, email, role dropdown (filtered)
- Delete: `Modal.confirm()` with warning text

### Phase 3: Backend Hardening

**3.1 Last Owner Protection**
- In `DeleteUser` and `UpdateUser` (role change): count owners in tenant
- If count=1 and target is the last owner → return 400 "Cannot remove the last owner"

**3.2 Pagination on ListUsers**
- Accept `?page=1&limit=20` query params
- Return: `{users: [], total: int, page: int, limit: int}`
- Default: page=1, limit=20, max limit=100

### Phase 4: Bug Fixes

**4.1 CRITICAL: SSE JWT Exposure**
- New endpoint: `POST /api/auth/sse-ticket`
- Returns: `{ticket: "short-lived-uuid"}` (30s TTL, stored in Redis)
- New endpoint: `GET /api/notifications/stream?ticket=xxx`
- Backend validates ticket, deletes after use (one-time)
- Frontend: exchange JWT for ticket before connecting SSE

**4.2 HIGH: Price Sync Dead Button**
- Disable button + add Ant Design Tooltip "Coming soon"

**4.3 MEDIUM: TikTok Wholesale 501**
- Return proper JSON: `{"success": false, "error": "TikTok wholesale not yet available"}`

**4.4 MEDIUM: Image Download No Timeout**
- Add `http.Client{Timeout: 30 * time.Second}`
- Add single retry on timeout

**4.5 LOW: Silent Error Swallowing**
- `sync_service.go:142` — log json.Marshal error
- `scanner_service.go:177` — log filepath.Walk error

## Edge Cases

| # | Case | Handling |
|---|------|----------|
| 1 | Self-deletion | Backend returns 400 + frontend hides delete on own row |
| 2 | Last owner | Backend count check + frontend shows error from API |
| 3 | Role downgrade | Confirmation modal with warning |
| 4 | Concurrent edits | Last-write-wins (acceptable at this scale) |
| 5 | Developer cross-tenant | Uses current tenant context from JWT/header selector |
| 6 | 1000+ users | Pagination (page/limit) |
| 7 | Locked account | Show "Locked until [time]" + Unlock button |
| 8 | Duplicate email | Global uniqueness — show API error in form |
| 9 | Password requirements | Inline hint "Min 8 characters" |
| 10 | Deleted user active session | Natural JWT expiry (15min) — acceptable risk |

## Success Criteria

- [ ] User Management page renders with correct data
- [ ] Role-based actions work (admin can only manage "user" targets)
- [ ] Sidebar shows/hides based on role
- [ ] Create/Edit/Delete/Unlock all functional
- [ ] Last owner cannot be deleted or demoted
- [ ] Pagination works (page navigation)
- [ ] SSE no longer exposes JWT in URL
- [ ] All dead buttons disabled with feedback
- [ ] No TypeScript errors, no console errors
- [ ] Build passes (frontend + backend)

## Files Affected

### New Files (5)
- `frontend/src/pages/users/UserManagementPage.tsx`
- `frontend/src/pages/users/utils/userColumns.tsx`
- `frontend/src/pages/users/components/CreateUserModal.tsx`
- `frontend/src/pages/users/components/EditUserModal.tsx`
- `frontend/src/api/users.ts`

### Modified Files (~12)
- `frontend/src/lib/permissionService.ts` — add admin permissions
- `frontend/src/components/layout/Sidebar.tsx` — role filtering + new menu item
- `frontend/src/App.tsx` — add /users route
- `frontend/src/contexts/NotificationContext.tsx` — SSE ticket auth
- `frontend/src/pages/products/UnifiedProductsPage.tsx` — disable dead button
- `frontend/src/api/auth.ts` — add getSSETicket
- `backend/internal/services/user_management_service.go` — pagination + last owner check
- `backend/internal/handlers/user_handler.go` — pagination params
- `backend/internal/handlers/auth_handler.go` — SSE ticket endpoint
- `backend/internal/handlers/wholesale_extended_handler.go` — proper error response
- `backend/internal/services/master_product/image_service.go` — timeout
- `backend/internal/services/inventory/sync_service.go` — error logging
- `backend/internal/services/route/scanner_service.go` — error logging
- `backend/internal/routes/routes.go` — SSE ticket route

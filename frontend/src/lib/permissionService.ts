/**
 * Permission Service
 * Ported from Vue permissionService.ts — same role/permission mappings.
 * Uses functional exports instead of class-based singleton.
 */

export type UserRole = "developer" | "owner" | "admin" | "user";

/**
 * Role → Permission mappings (exact port from Vue)
 */
export const ROLE_PERMISSIONS: Record<UserRole, Set<string>> = {
  developer: new Set([
    "users.list",
    "users.create",
    "users.update",
    "users.delete",
    "users.changePassword",
    "tenant.switch",
    "tenant.create",
    "tenant.delete",
    "store.manage",
    "store.settings",
    "reports.view",
    "reports.export",
    "audit.view",
    "system.settings",
  ]),
  owner: new Set([
    "users.list",
    "users.create",
    "users.update",
    "users.delete",
    "users.changePassword",
    "store.manage",
    "store.settings",
    "reports.view",
    "reports.export",
    "audit.view",
  ]),
  admin: new Set([
    "users.list",
    "users.update",
    "users.changePassword",
    "store.manage",
    "reports.view",
    "audit.view",
  ]),
  user: new Set(["users.update.self", "users.changePassword", "store.manage"]),
};

/** All valid roles for type-safe checks */
const VALID_ROLES: ReadonlySet<string> = new Set<string>([
  "developer",
  "owner",
  "admin",
  "user",
]);

/** Check if a string is a valid UserRole */
export function isValidRole(role: string): role is UserRole {
  return VALID_ROLES.has(role);
}

/** Check if role has a specific permission */
export function hasPermission(role: string, permission: string): boolean {
  if (!isValidRole(role)) return false;
  return ROLE_PERMISSIONS[role].has(permission);
}

/** Check if user has a specific role (exact match) */
export function hasRole(userRole: string, role: string): boolean {
  return userRole === role;
}

/** Check if user can manage other users */
export function canManageUsers(role: string): boolean {
  return hasPermission(role, "users.list");
}

/** Check if user can create users */
export function canCreateUser(role: string): boolean {
  return hasPermission(role, "users.create");
}

/** Check if user can delete users */
export function canDeleteUser(role: string): boolean {
  return hasPermission(role, "users.delete");
}

/** Check if user can view audit logs */
export function canViewAudit(role: string): boolean {
  return hasPermission(role, "audit.view");
}

/** Check if user can switch tenants (developer only) */
export function canSwitchTenant(role: string): boolean {
  return hasPermission(role, "tenant.switch");
}

/** Check if user can access admin features (developer, owner, admin) */
export function canAccessAdmin(role: string | undefined): boolean {
  if (!role) return false;
  return ["developer", "owner", "admin"].includes(role);
}

/** Get all permissions for a role */
export function getAllPermissions(role: string): string[] {
  if (!isValidRole(role)) return [];
  return Array.from(ROLE_PERMISSIONS[role]);
}

/** Get role display name */
export function getRoleDisplayName(role: string): string {
  const names: Record<UserRole, string> = {
    developer: "Developer",
    owner: "Owner",
    admin: "Administrator",
    user: "User",
  };
  if (!isValidRole(role)) return role;
  return names[role];
}

/** Get role description */
export function getRoleDescription(role: string): string {
  const descriptions: Record<UserRole, string> = {
    developer: "Full system access, can switch tenants",
    owner: "Full access to own tenant",
    admin: "Can manage users, store, and view reports",
    user: "Can manage own data only",
  };
  if (!isValidRole(role)) return "";
  return descriptions[role];
}

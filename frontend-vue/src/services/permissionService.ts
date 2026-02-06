export type UserRole = "developer" | "owner" | "admin" | "user";

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

class PermissionService {
  /**
   * Check if role has specific permission
   */
  hasPermission(role: UserRole, permission: string): boolean {
    const permissions = ROLE_PERMISSIONS[role] || new Set();
    return permissions.has(permission);
  }

  /**
   * Check if user can manage other users
   */
  canManageUsers(userRole: UserRole): boolean {
    return this.hasPermission(userRole, "users.list");
  }

  /**
   * Check if user can create users
   */
  canCreateUser(userRole: UserRole): boolean {
    return this.hasPermission(userRole, "users.create");
  }

  /**
   * Check if user can delete users
   */
  canDeleteUser(userRole: UserRole): boolean {
    return this.hasPermission(userRole, "users.delete");
  }

  /**
   * Check if user can view audit logs
   */
  canViewAudit(userRole: UserRole): boolean {
    return this.hasPermission(userRole, "audit.view");
  }

  /**
   * Get all permissions for role
   */
  getAllPermissions(role: UserRole): string[] {
    return Array.from(ROLE_PERMISSIONS[role] || new Set());
  }

  /**
   * Get role display name
   */
  getRoleDisplayName(role: UserRole): string {
    const names: Record<UserRole, string> = {
      developer: "Developer",
      owner: "Owner",
      admin: "Administrator",
      user: "User",
    };
    return names[role] || role;
  }

  /**
   * Get role description
   */
  getRoleDescription(role: UserRole): string {
    const descriptions: Record<UserRole, string> = {
      developer: "Full system access, can switch tenants",
      owner: "Full access to own tenant",
      admin: "Can manage users, store, and view reports",
      user: "Can manage own data only",
    };
    return descriptions[role] || "";
  }

  /**
   * Check if user can switch tenants (developer only)
   */
  canSwitchTenant(userRole: UserRole): boolean {
    return this.hasPermission(userRole, "tenant.switch");
  }

  /**
   * Check if user can access admin features
   * Admin, owner, and developer roles can access admin panel
   */
  canAccessAdmin(userRole: UserRole | undefined): boolean {
    if (!userRole) return false;
    return ["developer", "owner", "admin"].includes(userRole);
  }
}

export default new PermissionService();

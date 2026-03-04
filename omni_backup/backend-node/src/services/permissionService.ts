import { getLogger } from "../utils/logger";

getLogger("PermissionService"); // Initialize logger for module

export type UserRole = "developer" | "owner" | "admin" | "user" | "service";

/**
 * Permission definitions for each role
 * developer: Full system access, can switch tenants
 * owner: Full access to own tenant
 * admin: Can manage users, view reports, manage store
 * user: Can only manage own data
 * service: API-only access for automation (n8n, etc)
 */
export const ROLE_PERMISSIONS: Record<UserRole, Set<string>> = {
  developer: new Set([
    // Full user management
    "users.list",
    "users.create",
    "users.update",
    "users.delete",
    "users.changePassword",
    // Tenant management
    "tenant.switch",
    "tenant.create",
    "tenant.delete",
    "tenant.access.all",
    // Store management
    "store.manage",
    "store.settings",
    // Reporting
    "reports.view",
    "reports.export",
    // Audit
    "audit.view",
    // System
    "system.settings",
  ]),
  owner: new Set([
    // User management
    "users.list",
    "users.create",
    "users.update",
    "users.delete",
    "users.changePassword",
    // Store management
    "store.manage",
    "store.settings",
    // Reporting
    "reports.view",
    "reports.export",
    // Audit
    "audit.view",
  ]),
  admin: new Set([
    // User management (limited)
    "users.list",
    "users.update",
    "users.changePassword",
    // Store management
    "store.manage",
    // Reporting
    "reports.view",
    // Audit
    "audit.view",
  ]),
  user: new Set([
    // Self management only
    "users.update.self",
    "users.changePassword",
    // Store management
    "store.manage",
  ]),
  service: new Set([
    // Inventory (read + write for stock updates)
    "inventory.read",
    "inventory.write",
    // Orders (read only)
    "orders.read",
    "orders.sync",
    // Analytics (read only)
    "analytics.read",
    "reports.view",
    // Products (read only)
    "products.read",
    // Job Queue (full access for automation)
    "jobs.read",
    "jobs.write",
    "jobs.execute",
    // Auto Functions (execute)
    "autoFunctions.read",
    "autoFunctions.execute",
    // Google Sheets (read + write)
    "sheets.read",
    "sheets.write",
    // Monitoring (read only)
    "monitoring.read",
    // Settings (read only)
    "settings.read",
    // Tenant access (can access all tenants via header)
    "tenant.access.all",
  ]),
};

/**
 * Permission Service - Role-based access control
 */
export class PermissionService {
  hasPermission(role: UserRole, permission: string): boolean {
    const permissions = ROLE_PERMISSIONS[role] || new Set();
    return permissions.has(permission);
  }

  canManageUser(userRole: UserRole, targetRole: UserRole): boolean {
    if (userRole === "developer" || userRole === "owner") return true;
    if (userRole === "admin")
      return targetRole === "user" || targetRole === "admin";
    return false;
  }

  canUpdateUser(
    userRole: UserRole,
    targetUserId: string,
    currentUserId: string
  ): boolean {
    if (userRole === "developer" || userRole === "owner") return true;
    if (userRole === "admin") return true;
    return targetUserId === currentUserId;
  }

  canDeleteUser(userRole: UserRole, targetRole: UserRole): boolean {
    if (userRole === "developer" || userRole === "owner") return true;
    if (userRole === "admin") return targetRole === "user";
    return false;
  }

  getAllPermissions(role: UserRole): string[] {
    return Array.from(ROLE_PERMISSIONS[role] || new Set());
  }

  isServiceAccount(role: UserRole): boolean {
    return role === "service";
  }

  canAccessAllTenants(role: UserRole): boolean {
    const permissions = ROLE_PERMISSIONS[role] || new Set();
    return permissions.has("tenant.access.all");
  }
}

export default new PermissionService();

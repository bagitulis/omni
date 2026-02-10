import { useMemo } from "react";
import { useAuthStore } from "@/stores/authStore";
import {
  hasPermission,
  hasRole,
  canAccessAdmin,
  canManageUsers,
  canSwitchTenant,
  getAllPermissions,
  getRoleDisplayName,
} from "@/lib/permissionService";

/**
 * React hook for permission checks.
 * Wraps permissionService functions with current user's role from authStore.
 */
export function usePermission() {
  const user = useAuthStore((state) => state.user);
  const userRole = user?.role ?? "";

  return useMemo(
    () => ({
      /** Check if current user has a specific permission */
      hasPermission: (permission: string) =>
        hasPermission(userRole, permission),
      /** Check if current user has a specific role */
      hasRole: (role: string) => hasRole(userRole, role),
      /** Whether the current user is a developer */
      isDeveloper: userRole === "developer",
      /** Whether the current user is an owner */
      isOwner: userRole === "owner",
      /** Whether the current user is an admin */
      isAdmin: userRole === "admin",
      /** Whether the current user can access admin features */
      canAccessAdmin: canAccessAdmin(userRole),
      /** Whether the current user can manage other users */
      canManageUsers: canManageUsers(userRole),
      /** Whether the current user can switch tenants */
      canSwitchTenant: canSwitchTenant(userRole),
      /** All permissions for current user's role */
      permissions: getAllPermissions(userRole),
      /** Current user's role */
      userRole,
      /** Current user's role display name */
      roleDisplayName: getRoleDisplayName(userRole),
    }),
    [userRole],
  );
}

/**
 * Convenience hook: returns true if user has the specified permission.
 */
export function useHasPermission(permission: string): boolean {
  const user = useAuthStore((state) => state.user);
  const userRole = user?.role ?? "";
  return useMemo(
    () => hasPermission(userRole, permission),
    [userRole, permission],
  );
}

/**
 * Convenience hook: returns true if user has the specified role.
 */
export function useHasRole(role: string): boolean {
  const user = useAuthStore((state) => state.user);
  const userRole = user?.role ?? "";
  return useMemo(() => hasRole(userRole, role), [userRole, role]);
}

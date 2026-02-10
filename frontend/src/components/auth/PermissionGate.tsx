import { useAuthStore } from "@/stores/authStore";
import { hasPermission, hasRole } from "@/lib/permissionService";

interface PermissionGateProps {
  /** Permission required to render children */
  permission?: string;
  /** Role required to render children */
  role?: string;
  /** Content to render when user has the permission/role */
  children: React.ReactNode;
  /** Optional fallback when permission is denied */
  fallback?: React.ReactNode;
}

/**
 * Conditional render component based on user permission or role.
 * Renders children only if the current user has the required permission/role.
 *
 * Usage:
 *   <PermissionGate permission="users.create">
 *     <Button>Create User</Button>
 *   </PermissionGate>
 *
 *   <PermissionGate role="admin" fallback={<span>Admin only</span>}>
 *     <AdminPanel />
 *   </PermissionGate>
 */
export function PermissionGate({
  permission,
  role,
  children,
  fallback = null,
}: PermissionGateProps) {
  const user = useAuthStore((state) => state.user);
  const userRole = user?.role ?? "";

  // If no permission or role specified, always render children
  if (!permission && !role) {
    return <>{children}</>;
  }

  // Check permission if specified
  if (permission && !hasPermission(userRole, permission)) {
    return <>{fallback}</>;
  }

  // Check role if specified
  if (role && !hasRole(userRole, role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

import { Navigate, Outlet, useLocation } from "react-router-dom";
import { Result, Button } from "antd";
import { useAuthStore } from "@/stores/authStore";
import { hasPermission } from "@/lib/permissionService";

interface ProtectedRouteProps {
  /** Optional permission required to access this route */
  permission?: string;
  /** Optional role required to access this route */
  role?: string;
  /** Children to render (defaults to Outlet for nested routes) */
  children?: React.ReactNode;
}

/**
 * Route guard component.
 * - If not authenticated → redirect to /login?returnUrl=currentPath
 * - If authenticated but missing permission/role → show Access Denied
 * - Otherwise → render children or <Outlet />
 */
export function ProtectedRoute({
  permission,
  role,
  children,
}: ProtectedRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const location = useLocation();

  // Not authenticated → redirect to login with return URL
  if (!isAuthenticated) {
    const returnUrl = encodeURIComponent(location.pathname + location.search);
    return <Navigate to={`/login?returnUrl=${returnUrl}`} replace />;
  }

  // Check permission if specified
  if (permission) {
    const userRole = user?.role ?? "";
    if (!hasPermission(userRole, permission)) {
      return <AccessDenied />;
    }
  }

  // Check role if specified
  if (role) {
    if (user?.role !== role) {
      return <AccessDenied />;
    }
  }

  // Render children or Outlet for nested routes
  return children ? <>{children}</> : <Outlet />;
}

/** Inline Access Denied view — keeps it simple without extra file */
function AccessDenied() {
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
      }}
    >
      <Result
        status="403"
        title="Access Denied"
        subTitle="You do not have permission to access this page."
        extra={
          <Button type="primary" href="/">
            Back to Dashboard
          </Button>
        }
      />
    </div>
  );
}

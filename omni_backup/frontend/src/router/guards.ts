import type {
  Router,
  NavigationGuardNext,
  RouteLocationNormalized,
  NavigationFailure,
} from "vue-router";
import { useRouteState } from "../composables/useRouteState";
import { useNavigationQueue } from "../composables/useNavigationQueue";
import { usePerformanceMonitor } from "../composables/usePerformanceMonitor";
import { useRouteFlowController } from "../composables/useRouteFlowController";

// Initialize managers
const routeStateManager = useRouteState();
const navigationQueue = useNavigationQueue();
const performanceMonitor = usePerformanceMonitor();
const routeFlowController = useRouteFlowController();

/**
 * Check if user has admin role
 */
function hasAdminRole(): boolean {
  const userRole = localStorage.getItem("userRole") || "user";
  return ["developer", "owner", "admin"].includes(userRole);
}

/**
 * Setup before navigation guard
 */
export async function beforeNavigationGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
): Promise<void> {
  try {
    const { useAuthStore } = await import("../store/authStore");
    const authStore = useAuthStore();

    const requiresAuth = to.meta.requiresAuth !== false;
    const requiresAdminRole = to.meta.requiresAdminRole === true;
    const isLoggedIn = authStore.isAuthenticated;

    // Redirect to login if route requires auth and user not logged in
    if (requiresAuth && !isLoggedIn) {
      return next("/login");
    }

    // Check admin role access
    if (requiresAdminRole && !hasAdminRole()) {
      return next("/");
    }

    // Redirect to dashboard if trying to access login/register while logged in
    if ((to.path === "/login" || to.path === "/register") && isLoggedIn) {
      return next("/");
    }

    // /order-manager now handles all platforms directly without redirect
    // No redirect needed - OrderManager component handles ?type= query param directly

    // Handle /product-manager with ?type= query param - redirect to platform-specific page
    if (to.path === "/product-manager" && to.query.type) {
      return next({
        path: "/product-manager/shopee",
        query: to.query,
        replace: true,
      });
    }

    // Handle /operation with query params - redirect to platform-specific page
    if (to.path === "/operation" && Object.keys(to.query).length > 0) {
      return next({
        path: "/operation/shopee",
        query: to.query,
        replace: true,
      });
    }

    const toRouteName = to.name as string;
    if (!toRouteName) {
      return next();
    }

    // Track route flow
    const flowType = routeFlowController.getFlowType(toRouteName, to.fullPath);
    const flowId = routeFlowController.recordFlowStart(
      toRouteName,
      to.fullPath,
      flowType
    );

    // Add to navigation queue only when required
    let navId: string | null = null;
    if (flowType === "queued") {
      navId = navigationQueue.addToQueue(toRouteName, "normal");
      navigationQueue.startProcessing(navId);
    }

    // Initialize route state
    routeStateManager.setLoading(toRouteName);
    performanceMonitor.initializeMetrics(toRouteName);

    const startTime = performance.now();
    const cleanup = (failed?: boolean, failureMessage?: string) => {
      if (navId) navigationQueue.finishProcessing(navId);
      routeFlowController.completeFlow(
        flowId,
        !failed,
        performance.now() - startTime,
        failureMessage || null
      );
    };

    // Attach metadata to route
    (to as any)._performanceStartTime = startTime;
    (to as any)._routeName = toRouteName;
    (to as any)._cleanup = cleanup;
    (to as any)._flowType = flowType;

    next();
  } catch (error: any) {
    console.error("[Router Guard] Error in beforeEach:", error);
    navigationQueue.clearQueue();
    next();
  }
}

/**
 * After navigation handler
 */
export function afterNavigationHandler(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  failure: NavigationFailure | void | undefined
): void {
  try {
    const toRouteName = (to as any)._routeName;
    const startTime = (to as any)._performanceStartTime;
    const cleanup = (to as any)._cleanup;

    if (cleanup) {
      cleanup(Boolean(failure), failure?.message);
    }

    if (!toRouteName || !startTime) {
      return;
    }

    // Calculate response time
    const endTime = performance.now();
    const responseTime = Math.round(endTime - startTime);

    // Record performance
    if (failure) {
      routeStateManager.setError(
        toRouteName,
        failure.message || "Navigation failed"
      );
      performanceMonitor.recordError(toRouteName, responseTime);
    } else {
      routeStateManager.setSuccess(toRouteName);
      performanceMonitor.recordSuccess(toRouteName, responseTime);
    }
  } catch (error: any) {
    console.error("[Router Cleanup] Error in afterEach:", error);
  }
}

/**
 * Setup all navigation guards on router
 */
export function setupNavigationGuards(router: Router): void {
  router.beforeEach(beforeNavigationGuard);
  router.afterEach(afterNavigationHandler);
}

/**
 * Attach managers to router instance
 */
export function attachManagersToRouter(router: Router): void {
  router.routeStateManager = routeStateManager;
  router.navigationQueue = navigationQueue;
  router.performanceMonitor = performanceMonitor;
}

export { routeStateManager, navigationQueue, performanceMonitor };

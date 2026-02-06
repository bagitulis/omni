import { useRouter } from 'vue-router';
import { useRouteState } from './useRouteState';
import { useNavigationQueue } from './useNavigationQueue';
import { usePerformanceMonitor } from './usePerformanceMonitor';

/**
 * Route Manager Facade
 * Provides unified interface to all route management features
 * Single Responsibility: Provide easy-to-use API for components
 */

export function useRouteManager() {
  const router = useRouter();

  // Get managers from router instance
  const routeStateManager = (router as any).routeStateManager || useRouteState();
  const navigationQueue = (router as any).navigationQueue || useNavigationQueue();
  const performanceMonitor = (router as any).performanceMonitor || usePerformanceMonitor();

  /**
   * Get current route loading state
   */
  const isRouteLoading = (routeName?: string): boolean => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return routeStateManager.isLoading(targetRoute);
  };

  /**
   * Get current route error
   */
  const getRouteError = (routeName?: string): string | null => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return routeStateManager.getError(targetRoute);
  };

  /**
   * Check if current route has error
   */
  const hasRouteError = (routeName?: string): boolean => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return routeStateManager.hasError(targetRoute);
  };

  /**
   * Get current route state
   */
  const getRouteState = (routeName?: string) => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return routeStateManager.getState(targetRoute);
  };

  /**
   * Get performance metrics for current route
   */
  const getRouteMetrics = (routeName?: string) => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return performanceMonitor.getMetrics(targetRoute);
  };

  /**
   * Get formatted performance info
   */
  const getFormattedMetrics = (routeName?: string): string => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return performanceMonitor.getFormattedMetrics(targetRoute);
  };

  /**
   * Get health status of route
   */
  const getRouteHealth = (routeName?: string) => {
    const targetRoute = routeName || (router.currentRoute.value.name as string);
    return performanceMonitor.getHealthStatus(targetRoute);
  };

  /**
   * Get queue info
   */
  const getQueueInfo = () => {
    return {
      size: navigationQueue.getQueueSize(),
      queued: navigationQueue.getAllQueued(),
      isProcessing: navigationQueue.processingRoute.value ?? null,
    };
  };

  /**
   * Get all monitoring data
   */
  const getAllMonitoringData = () => {
    return {
      routeStates: routeStateManager.getAllStates.value,
      metrics: performanceMonitor.getAllMetrics.value,
      errorStats: performanceMonitor.getErrorStats.value,
      slowRoutes: performanceMonitor.getSlowRoutes(),
      queueInfo: getQueueInfo(),
    };
  };

  return {
    // State queries
    isRouteLoading,
    getRouteError,
    hasRouteError,
    getRouteState,

    // Performance queries
    getRouteMetrics,
    getFormattedMetrics,
    getRouteHealth,

    // Queue info
    getQueueInfo,

    // Full monitoring data
    getAllMonitoringData,

    // Direct manager access (if needed)
    routeStateManager,
    navigationQueue,
    performanceMonitor,
  };
}

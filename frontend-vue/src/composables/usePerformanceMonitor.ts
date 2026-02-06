import { ref, Ref, computed } from 'vue';

/**
 * Performance Monitoring
 * Tracks response times and metrics per route
 * Single Responsibility: Monitor performance metrics only
 */

interface RouteMetrics {
  routeName: string;
  requestCount: number;
  totalTime: number;
  minTime: number;
  maxTime: number;
  averageTime: number;
  lastRequestTime: number;
  lastRequestTimestamp: number;
  errorCount: number;
  successCount: number;
}

interface MetricsMap {
  [routeName: string]: RouteMetrics;
}

// Excluded routes from monitoring (long-running operations)
const EXCLUDED_ROUTES = new Set([
  'OrderManager',
  'OrderManagerShopee',
  'OrderManagerLazada',
  'OrderManagerTiktok',
  // Add load config routes if they exist
]);

export function usePerformanceMonitor() {
  const metricsMap: Ref<MetricsMap> = ref({});

  /**
   * Check if route should be monitored
   */
  const shouldMonitor = (routeName: string): boolean => {
    return !EXCLUDED_ROUTES.has(routeName);
  };

  /**
   * Initialize metrics for a route
   */
  const initializeMetrics = (routeName: string): void => {
    if (!shouldMonitor(routeName)) return;

    if (!metricsMap.value[routeName]) {
      metricsMap.value[routeName] = {
        routeName,
        requestCount: 0,
        totalTime: 0,
        minTime: Infinity,
        maxTime: 0,
        averageTime: 0,
        lastRequestTime: 0,
        lastRequestTimestamp: 0,
        errorCount: 0,
        successCount: 0,
      };
    }
  };

  /**
   * Record successful request
   * @param routeName - Route name
   * @param responseTime - Response time in milliseconds
   */
  const recordSuccess = (routeName: string, responseTime: number): void => {
    if (!shouldMonitor(routeName)) return;
    if (!metricsMap.value[routeName]) {
      initializeMetrics(routeName);
    }

    const metrics = metricsMap.value[routeName];
    metrics.requestCount++;
    metrics.successCount++;
    metrics.totalTime += responseTime;
    metrics.minTime = Math.min(metrics.minTime, responseTime);
    metrics.maxTime = Math.max(metrics.maxTime, responseTime);
    metrics.lastRequestTime = responseTime;
    metrics.lastRequestTimestamp = Date.now();
    metrics.averageTime = Math.round(metrics.totalTime / metrics.requestCount);
  };

  /**
   * Record failed request
   * @param routeName - Route name
   * @param responseTime - Response time in milliseconds
   */
  const recordError = (routeName: string, responseTime: number = 0): void => {
    if (!shouldMonitor(routeName)) return;
    if (!metricsMap.value[routeName]) {
      initializeMetrics(routeName);
    }

    const metrics = metricsMap.value[routeName];
    metrics.requestCount++;
    metrics.errorCount++;
    if (responseTime > 0) {
      metrics.totalTime += responseTime;
      metrics.lastRequestTime = responseTime;
      metrics.averageTime = Math.round(metrics.totalTime / metrics.requestCount);
    }
    metrics.lastRequestTimestamp = Date.now();
  };

  /**
   * Get metrics for a route
   */
  const getMetrics = (routeName: string): RouteMetrics | null => {
    return metricsMap.value[routeName] || null;
  };

  /**
   * Get all metrics
   */
  const getAllMetrics = computed(() => {
    return Object.values(metricsMap.value);
  });

  /**
   * Get metrics sorted by average response time (slowest first)
   */
  const getMetricsBySpeed = computed(() => {
    return [...Object.values(metricsMap.value)].sort(
      (a, b) => b.averageTime - a.averageTime
    );
  });

  /**
   * Get slow routes (average > threshold)
   */
  const getSlowRoutes = (thresholdMs: number = 1000): RouteMetrics[] => {
    return Object.values(metricsMap.value).filter(
      m => m.averageTime > thresholdMs && m.requestCount > 0
    );
  };

  /**
   * Get error stats
   */
  const getErrorStats = computed(() => {
    const stats = {
      totalErrors: 0,
      totalRequests: 0,
      errorRate: 0,
    };

    Object.values(metricsMap.value).forEach(metrics => {
      stats.totalErrors += metrics.errorCount;
      stats.totalRequests += metrics.requestCount;
    });

    stats.errorRate = stats.totalRequests > 0 
      ? Math.round((stats.totalErrors / stats.totalRequests) * 100)
      : 0;

    return stats;
  });

  /**
   * Get formatted metrics for display
   */
  const getFormattedMetrics = (routeName: string): string => {
    const metrics = metricsMap.value[routeName];
    if (!metrics) return 'No data';

    const successRate = metrics.requestCount > 0
      ? Math.round((metrics.successCount / metrics.requestCount) * 100)
      : 0;

    return `Avg: ${metrics.averageTime}ms | Last: ${metrics.lastRequestTime}ms | Success: ${successRate}% | Requests: ${metrics.requestCount}`;
  };

  /**
   * Clear metrics for a route
   */
  const clearMetrics = (routeName: string): void => {
    if (metricsMap.value[routeName]) {
      delete metricsMap.value[routeName];
    }
  };

  /**
   * Clear all metrics
   */
  const clearAllMetrics = (): void => {
    metricsMap.value = {};
  };

  /**
   * Export metrics as JSON
   */
  const exportMetrics = (): string => {
    return JSON.stringify(Object.values(metricsMap.value), null, 2);
  };

  /**
   * Get health status based on metrics
   */
  const getHealthStatus = (routeName: string): 'healthy' | 'warning' | 'error' => {
    const metrics = metricsMap.value[routeName];
    if (!metrics || metrics.requestCount === 0) return 'healthy';

    const errorRate = (metrics.errorCount / metrics.requestCount) * 100;
    
    if (errorRate > 20) return 'error';
    if (errorRate > 10 || metrics.averageTime > 3000) return 'warning';
    return 'healthy';
  };

  return {
    shouldMonitor,
    initializeMetrics,
    recordSuccess,
    recordError,
    getMetrics,
    getAllMetrics,
    getMetricsBySpeed,
    getSlowRoutes,
    getErrorStats,
    getFormattedMetrics,
    clearMetrics,
    clearAllMetrics,
    exportMetrics,
    getHealthStatus,
  };
}

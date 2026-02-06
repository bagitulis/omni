import { computed } from 'vue';
import { useRouteManager } from './useRouteManager';

/**
 * Script Monitor Display Logic
 * Formats monitoring data for display in Script Monitor component
 * Single Responsibility: Format data for UI display only
 */

interface RoutePerformanceDisplay {
  routeName: string;
  status: 'healthy' | 'warning' | 'error';
  averageTime: number;
  lastTime: number;
  requestCount: number;
  successRate: number;
  errorCount: number;
  responseTimeStatus: 'fast' | 'normal' | 'slow' | 'very-slow';
  responseTimeFormatted: string;
}

interface RouteStateDisplay {
  routeName: string;
  state: string;
  icon: string;
  error: string | null;
  timestamp: string;
}

interface QueueDisplay {
  queueSize: number;
  isProcessing: string | null;
  queued: Array<{
    routeName: string;
    waitTime: number;
    priority: 'normal' | 'high';
  }>;
}

interface ErrorStatsDisplay {
  totalErrors: number;
  totalRequests: number;
  errorRate: string;
  slowRoutes: Array<{
    name: string;
    avgTime: string;
    slowness: string;
  }>;
}

interface HealthDisplay {
  total: number;
  healthy: number;
  warning: number;
  error: number;
  healthyPercent: number;
  status: 'good' | 'fair' | 'poor';
}

interface ChartData {
  labels: string[];
  datasets: Array<{
    label: string;
    data: number[];
    backgroundColor: string[];
  }>;
}

export function useScriptMonitorDisplay() {
  const routeManager = useRouteManager();

  /**
   * Format route state for display
   */
  const getStateIcon = (state: string): string => {
    const icons: Record<string, string> = {
      'idle': '✅',
      'loading': '⏳',
      'error': '❌',
    };
    return icons[state] || '❓';
  };

  /**
   * Format response time with color coding
   */
  const formatResponseTime = (ms: number): { value: string; status: 'fast' | 'normal' | 'slow' | 'very-slow' } => {
    if (ms < 500) return { value: `${ms}ms`, status: 'fast' };
    if (ms < 1000) return { value: `${ms}ms`, status: 'normal' };
    if (ms < 3000) return { value: `${ms}ms`, status: 'slow' };
    return { value: `${ms}ms`, status: 'very-slow' };
  };

  /**
   * Get route performance summary
   */
  const getRoutePerformanceSummary = computed((): RoutePerformanceDisplay[] => {
    const allMetrics = routeManager.performanceMonitor.getAllMetrics.value;
    
    return allMetrics
      .filter((m: any) => m.requestCount > 0) // Only routes with requests
      .map((m: any) => {
        const successRate = Math.round((m.successCount / m.requestCount) * 100);
        const responseTimeInfo = formatResponseTime(m.averageTime);
        
        return {
          routeName: m.routeName,
          status: routeManager.performanceMonitor.getHealthStatus(m.routeName),
          averageTime: m.averageTime,
          lastTime: m.lastRequestTime,
          requestCount: m.requestCount,
          successRate,
          errorCount: m.errorCount,
          responseTimeStatus: responseTimeInfo.status,
          responseTimeFormatted: responseTimeInfo.value,
        };
      })
      .sort((a: RoutePerformanceDisplay, b: RoutePerformanceDisplay) => b.averageTime - a.averageTime); // Sort by slowest first
  });

  /**
   * Get route state summary
   */
  const getRouteStateSummary = computed((): RouteStateDisplay[] => {
    const allStates = routeManager.routeStateManager.getAllStates.value;
    
    return Object.entries(allStates)
      .map(([routeName, data]: [string, any]) => ({
        routeName,
        state: data.state,
        icon: getStateIcon(data.state),
        error: data.error,
        timestamp: new Date(data.timestamp).toLocaleTimeString(),
      }))
      .sort((a, b) => a.routeName.localeCompare(b.routeName));
  });

  /**
   * Get queue status
   */
  const getQueueStatus = computed((): QueueDisplay => {
    const queueInfo = routeManager.getQueueInfo();
    return {
      queueSize: queueInfo.size,
      isProcessing: queueInfo.isProcessing as string | null,
      queued: queueInfo.queued.map((nav: any) => ({
        routeName: nav.routeName,
        waitTime: Math.round(Date.now() - nav.timestamp),
        priority: nav.priority,
      })),
    };
  });

  /**
   * Get error statistics
   */
  const getErrorStatistics = computed((): ErrorStatsDisplay => {
    const errorStats = routeManager.performanceMonitor.getErrorStats.value;
    const slowRoutes = routeManager.performanceMonitor.getSlowRoutes(1000);

    return {
      totalErrors: errorStats.totalErrors,
      totalRequests: errorStats.totalRequests,
      errorRate: `${errorStats.errorRate}%`,
      slowRoutes: slowRoutes.map((route: any) => ({
        name: route.routeName,
        avgTime: `${route.averageTime}ms`,
        slowness: ((route.averageTime - 1000) / 1000).toFixed(1) + 'x slower than threshold',
      })),
    };
  });

  /**
   * Get health summary
   */
  const getHealthSummary = computed((): HealthDisplay => {
    const allMetrics = routeManager.performanceMonitor.getAllMetrics.value;
    
    const healthCount = {
      healthy: 0,
      warning: 0,
      error: 0,
    };

    allMetrics.forEach((m: any) => {
      const health = routeManager.performanceMonitor.getHealthStatus(m.routeName);
      healthCount[health as keyof typeof healthCount]++;
    });

    const totalRoutes = allMetrics.length || 1;
    const healthyPercent = Math.round((healthCount.healthy / totalRoutes) * 100);

    return {
      total: totalRoutes,
      healthy: healthCount.healthy,
      warning: healthCount.warning,
      error: healthCount.error,
      healthyPercent,
      status: healthyPercent >= 80 ? 'good' : healthyPercent >= 50 ? 'fair' : 'poor',
    };
  });

  /**
   * Export all monitoring data
   */
  const exportMonitoringData = (): string => {
    const data = {
      timestamp: new Date().toISOString(),
      performance: getRoutePerformanceSummary.value,
      states: getRouteStateSummary.value,
      queue: getQueueStatus.value,
      errors: getErrorStatistics.value,
      health: getHealthSummary.value,
      rawMetrics: routeManager.performanceMonitor.exportMetrics(),
    };

    return JSON.stringify(data, null, 2);
  };

  /**
   * Get chart data for performance visualization
   */
  const getChartData = computed((): ChartData => {
    const allMetrics = routeManager.performanceMonitor.getAllMetrics.value;

    return {
      labels: allMetrics.map((m: any) => m.routeName),
      datasets: [
        {
          label: 'Average Response Time (ms)',
          data: allMetrics.map((m: any) => m.averageTime),
          backgroundColor: allMetrics.map((m: any) => {
            if (m.averageTime < 500) return '#27ae60'; // green
            if (m.averageTime < 1000) return '#f39c12'; // orange
            return '#e74c3c'; // red
          }),
        },
      ],
    };
  });

  return {
    getStateIcon,
    formatResponseTime,
    getRoutePerformanceSummary,
    getRouteStateSummary,
    getQueueStatus,
    getErrorStatistics,
    getHealthSummary,
    getChartData,
    exportMonitoringData,
  };
}

/**
 * useMonitoring Composable
 * Single Responsibility: Fetch monitoring metrics from backend
 * Max 150 lines
 */

import { ref, onMounted, onUnmounted } from "vue";
import api from "@/services/api";

export interface MonitoringMetrics {
  tenantId: string;
  metrics: {
    cache: {
      hits: number;
      misses: number;
      hitRate: number;
    };
    queue: {
      enqueued: number;
      completed: number;
      failed: number;
      avgProcessingTime: number;
    };
  };
  queue: {
    queued: number;
    byPriority: {
      high: number;
      medium: number;
      low: number;
    };
  };
  alerts: {
    total: number;
    bySeverity: {
      critical: number;
      warning: number;
      info: number;
    };
    recent: any[];
  };
}

export function useMonitoring() {
  const metrics = ref<MonitoringMetrics | null>(null);
  const isLoading = ref(false);
  const error = ref("");
  const autoRefresh = ref(false);
  let refreshInterval: any = null;

  const fetchMetrics = async () => {
    isLoading.value = true;
    error.value = "";
    try {
      const data = await api.get("/monitoring/summary");
      metrics.value = data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to fetch metrics";
      console.error("Monitoring fetch error:", err);
    } finally {
      isLoading.value = false;
    }
  };

  const fetchAlerts = async (limit = 10) => {
    try {
      const data = await api.get(`/monitoring/alerts?limit=${limit}`);
      return data.data;
    } catch (err: any) {
      console.error("Failed to fetch alerts:", err);
      return null;
    }
  };

  const clearMetrics = async () => {
    if (!confirm("Clear all monitoring metrics?")) return false;
    try {
      await api.post("/monitoring/clear");
      await fetchMetrics(); // Refresh
      return true;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to clear metrics";
      return false;
    }
  };

  const configureAlerts = async (config: any) => {
    try {
      const data = await api.post("/monitoring/configure-alerts", config);
      return data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to configure alerts";
      return null;
    }
  };

  const startAutoRefresh = (intervalMs = 5000) => {
    autoRefresh.value = true;
    fetchMetrics(); // Initial fetch
    refreshInterval = setInterval(fetchMetrics, intervalMs);
  };

  const stopAutoRefresh = () => {
    autoRefresh.value = false;
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  };

  const exportMetrics = () => {
    if (!metrics.value) return;
    const blob = new Blob([JSON.stringify(metrics.value, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `monitoring-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  onMounted(() => {
    fetchMetrics();
  });

  onUnmounted(() => {
    stopAutoRefresh();
  });

  return {
    metrics,
    isLoading,
    error,
    autoRefresh,
    fetchMetrics,
    fetchAlerts,
    clearMetrics,
    configureAlerts,
    startAutoRefresh,
    stopAutoRefresh,
    exportMetrics,
  };
}

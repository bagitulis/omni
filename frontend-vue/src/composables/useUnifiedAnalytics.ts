import { ref, computed } from "vue";
import { useApi } from "./useApi";

export interface KPIData {
  total_products: number;
  avg_roas: number;
  actions: {
    scale_up: number;
    maintain: number;
    reduce: number;
    stop: number;
  };
}

export interface PlatformSummary {
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  avg_roas: number;
  products_count: number;
}

export interface UnifiedSummary {
  combined: PlatformSummary;
  tiktok: PlatformSummary;
  shopee: PlatformSummary;
}

export interface ClassifiedProduct {
  product_id: string;
  product_name: string;
  total_cost: number;
  total_revenue: number;
  roas: number;
  action: string;
  action_label: string;
  confidence_level: string;
  trend_direction: string;
  source: string;
}

export interface ClassifiedProducts {
  scale_up: ClassifiedProduct[];
  maintain: ClassifiedProduct[];
  reduce: ClassifiedProduct[];
  stop: ClassifiedProduct[];
}

export interface CacheStatus {
  view_name: string;
  last_refresh: string;
  refresh_time_ms: number;
  row_count: number;
}

export function useUnifiedAnalytics() {
  const api = useApi();
  const loading = ref(false);
  const error = ref<string | null>(null);

  const kpi = ref<KPIData | null>(null);
  const summary = ref<UnifiedSummary | null>(null);
  const classifiedProducts = ref<ClassifiedProducts | null>(null);
  const cacheStatus = ref<CacheStatus[]>([]);

  const totalProducts = computed(() => kpi.value?.total_products || 0);
  const avgRoas = computed(() => kpi.value?.avg_roas || 0);

  const actionCounts = computed(
    () =>
      kpi.value?.actions || {
        scale_up: 0,
        maintain: 0,
        reduce: 0,
        stop: 0,
      },
  );

  async function fetchKPI(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.client.get("/analytics/unified/kpi");
      // API returns { success: true, data: { ... } }
      kpi.value = response.data.data || response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to fetch KPI";
    } finally {
      loading.value = false;
    }
  }

  async function fetchSummary(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.client.get("/analytics/unified/summary");
      // API returns { success: true, data: { combined, tiktok, shopee } }
      summary.value = response.data.data || response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to fetch summary";
    } finally {
      loading.value = false;
    }
  }

  async function fetchClassifiedProducts(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.client.get("/analytics/products/classified");
      // API returns { success: true, data: { scale_up, maintain, reduce, stop }, counts: {...} }
      classifiedProducts.value = response.data.data || response.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to fetch products";
    } finally {
      loading.value = false;
    }
  }

  async function fetchCacheStatus(): Promise<void> {
    try {
      const response = await api.client.get("/analytics/cache/status");
      // API returns { success: true, data: { metadata: [...] } }
      const data = response.data.data || response.data;
      cacheStatus.value = data.metadata || [];
    } catch (err: any) {
      console.error("Failed to fetch cache status:", err);
    }
  }

  async function refreshCache(): Promise<void> {
    loading.value = true;
    try {
      await api.client.post("/analytics/cache/refresh");
      await fetchCacheStatus();
    } catch (err: any) {
      error.value = err.response?.data?.error || "Failed to refresh cache";
    } finally {
      loading.value = false;
    }
  }

  async function fetchAll(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      await Promise.all([
        fetchKPI(),
        fetchSummary(),
        fetchClassifiedProducts(),
        fetchCacheStatus(),
      ]);
    } catch (err: any) {
      error.value = "Failed to fetch data";
    } finally {
      loading.value = false;
    }
  }

  function formatCurrency(value: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  }

  function formatRoas(value: number): string {
    return `${value.toFixed(2)}x`;
  }

  function formatNumber(value: number): string {
    return new Intl.NumberFormat("id-ID").format(value);
  }

  return {
    // State
    loading,
    error,
    kpi,
    summary,
    classifiedProducts,
    cacheStatus,

    // Computed
    totalProducts,
    avgRoas,
    actionCounts,

    // Actions
    fetchKPI,
    fetchSummary,
    fetchClassifiedProducts,
    fetchCacheStatus,
    refreshCache,
    fetchAll,

    // Formatters
    formatCurrency,
    formatRoas,
    formatNumber,
  };
}

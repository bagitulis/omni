import { defineStore } from "pinia";
import axios from "axios";
import { ref, computed } from "vue";

// Types - snake_case to match Go backend JSON response
export interface AdsSummary {
  total_cost: number;
  total_revenue: number;
  total_clicks: number;
  total_impressions: number;
  total_conversions: number;
  avg_roas: number;
  avg_acos: number;
  avg_ctr: number;
  product_count: number;
}

export interface TrendDataPoint {
  label: string;
  period_start: string;
  spend: number;
  gmv: number;
  roas: number;
  order_count: number;
}

export interface ProductPerformance {
  product_id: string;
  product_name: string;
  thumb_url?: string;
  spend: number;
  gmv: number;
  roas: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversions: number;
  sold: number;
  score: {
    composite_score: number;
    category: string;
    action: string;
  };
}

export const useAnalyticsStore = defineStore("analytics", () => {
  // State
  const platform = ref<"shopee" | "tiktok">("shopee");

  // Date handling (Default last 30 days)
  const endDate = ref(new Date());
  const startDate = ref(
    new Date(new Date().setDate(new Date().getDate() - 30)),
  );

  const loading = ref(false);
  const error = ref<string | null>(null);

  const summary = ref<AdsSummary | null>(null);
  const trends = ref<TrendDataPoint[]>([]);
  const products = ref<ProductPerformance[]>([]);

  // Computed
  const formattedStartDate = computed(
    () => startDate.value.toISOString().split("T")[0],
  );
  const formattedEndDate = computed(
    () => endDate.value.toISOString().split("T")[0],
  );

  // Actions
  async function fetchAll() {
    loading.value = true;
    error.value = null;
    try {
      await Promise.all([fetchSummary(), fetchTrends(), fetchPerformance()]);
    } catch (e: any) {
      error.value = e.response?.data?.error || e.message;
      console.error("Failed to fetch analytics data", e);
    } finally {
      loading.value = false;
    }
  }

  async function fetchSummary() {
    const response = await axios.get(`/api/ads/${platform.value}/summary`, {
      params: {
        startDate: formattedStartDate.value,
        endDate: formattedEndDate.value,
      },
    });
    if (response.data.success) {
      summary.value = response.data.data;
    }
  }

  async function fetchTrends() {
    const response = await axios.get(`/api/ads/${platform.value}/trends`, {
      params: {
        startDate: formattedStartDate.value,
        endDate: formattedEndDate.value,
      },
    });
    if (response.data.success) {
      trends.value = response.data.data || [];
    }
  }

  async function fetchPerformance() {
    const response = await axios.get(`/api/ads/${platform.value}/performance`, {
      params: {
        startDate: formattedStartDate.value,
        endDate: formattedEndDate.value,
      },
    });
    if (response.data.success) {
      products.value = response.data.data || [];
    }
  }

  async function uploadReport(file: File) {
    loading.value = true;
    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await axios.post(
        `/api/ads/${platform.value}/upload`,
        formData,
        {
          headers: { "Content-Type": "multipart/form-data" },
        },
      );
      if (response.data.success) {
        // Refresh data after upload
        await fetchAll();
        return true;
      }
      return false;
    } catch (e: any) {
      error.value = e.response?.data?.error || "Upload failed";
      throw e;
    } finally {
      loading.value = false;
    }
  }

  function setDateRange(start: Date, end: Date) {
    startDate.value = start;
    endDate.value = end;
    fetchAll();
  }

  function setPlatform(p: "shopee" | "tiktok") {
    platform.value = p;
    fetchAll(); // Refetch when platform changes
  }

  return {
    platform,
    startDate,
    endDate,
    loading,
    error,
    summary,
    trends,
    products,
    fetchAll,
    uploadReport,
    setDateRange,
    setPlatform,
  };
});

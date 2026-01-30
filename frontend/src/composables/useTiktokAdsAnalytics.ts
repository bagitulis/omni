/**
 * TikTok Ads Analytics Composable
 * Handle API calls for TikTok Ads data
 * Single Responsibility: API communication only
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";

const api = useApi();

// Types - Using snake_case to match backend API responses
export interface UploadBatchResult {
  success: boolean;
  batch_id: string;
  file_name: string;
  period_start: string;
  period_end: string;
  total_rows: number;
  inserted_rows: number;
  skipped_rows: number;
  updated_rows: number;
  errors: string[];
}

export interface DashboardSummary {
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  avg_roi: number;
  total_impressions: number;
  total_clicks: number;
  avg_ctr: number;
  avg_conversion_rate: number;
  top_products: ProductPerformance[];
  creative_type_comparison: CreativeTypeStats[];
}

export interface ProductPerformance {
  product_id: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
}

export interface CreativeTypeStats {
  creative_type: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
  cost_per_order: number;
}

export interface CreativeData {
  id: number;
  campaign_id: string;
  campaign_name: string;
  product_id: string;
  creative_type: string;
  video_title: string | null;
  cost: number;
  orders_sku: number;
  gross_revenue: number;
  roi: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversion_rate: number;
  period_start: string;
  period_end: string;
}

export interface UploadBatch {
  id: string;
  file_name: string;
  period_start: string;
  period_end: string;
  total_rows: number;
  inserted_rows: number;
  skipped_rows: number;
  status: string;
  created_at: string;
}

export interface UploadResult {
  success: boolean;
  processedRows: number;
  insertedRows: number;
  skippedRows: number;
  errorRows: number;
}

export function useTiktokAdsAnalytics() {
  // State
  const loading = ref(false);
  const uploading = ref(false);
  const error = ref<string | null>(null);
  const dashboard = ref<DashboardSummary | null>(null);
  const creativeData = ref<CreativeData[]>([]);
  const uploadHistory = ref<UploadBatch[]>([]);
  const totalRecords = ref(0);

  // Computed
  const hasData = computed(() => creativeData.value.length > 0);

  // API Methods
  async function uploadFile(
    file: File,
    mode: "skip" | "update" = "skip",
  ): Promise<UploadBatchResult | null> {
    uploading.value = true;
    error.value = null;

    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("mode", mode);

      const response = await api.client.post<{
        success: boolean;
        data: UploadBatchResult;
      }>("/analytics/tiktok-ads/upload", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      if (response.data?.success) {
        // Refresh data after upload
        await fetchDashboard();
        await fetchUploadHistory();
        return response.data.data;
      }

      return null;
    } catch (err) {
      error.value = err instanceof Error ? err.message : "Upload failed";
      return null;
    } finally {
      uploading.value = false;
    }
  }

  async function fetchDashboard(
    periodStart?: string,
    periodEnd?: string,
  ): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const params = new URLSearchParams();
      if (periodStart) params.append("periodStart", periodStart);
      if (periodEnd) params.append("periodEnd", periodEnd);

      const url = `/analytics/tiktok-ads/dashboard${params.toString() ? `?${params}` : ""}`;

      const response = await api.client.get<{
        success: boolean;
        data: DashboardSummary;
      }>(url);

      if (response.data?.success) {
        dashboard.value = response.data.data;
      }
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to fetch dashboard";
    } finally {
      loading.value = false;
    }
  }

  async function fetchCreativeData(
    options: {
      periodStart?: string;
      periodEnd?: string;
      productId?: string;
      creativeType?: string;
      orderBy?: string;
      orderDir?: string;
      limit?: number;
      offset?: number;
    } = {},
  ): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const params = new URLSearchParams();
      Object.entries(options).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, String(value));
        }
      });

      const url = `/analytics/tiktok-ads/data${params.toString() ? `?${params}` : ""}`;

      const response = await api.client.get<{
        success: boolean;
        data: CreativeData[];
        pagination: { total: number };
      }>(url);

      if (response.data?.success) {
        creativeData.value = response.data.data;
        totalRecords.value = response.data.pagination?.total || 0;
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : "Failed to fetch data";
    } finally {
      loading.value = false;
    }
  }

  async function fetchUploadHistory(limit = 20): Promise<void> {
    try {
      const response = await api.client.get<{
        success: boolean;
        data: UploadBatch[];
      }>(`/analytics/tiktok-ads/uploads?limit=${limit}`);

      if (response.data?.success) {
        uploadHistory.value = response.data.data;
      }
    } catch (err) {
      console.error("Failed to fetch upload history:", err);
    }
  }

  // Formatters
  function formatCurrency(value: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  }

  function formatNumber(value: number): string {
    return new Intl.NumberFormat("id-ID").format(value);
  }

  function formatPercent(value: number): string {
    return `${(value * 100).toFixed(2)}%`;
  }

  function formatRoi(value: number): string {
    return value.toFixed(2);
  }

  return {
    // State
    loading,
    uploading,
    error,
    dashboard,
    creativeData,
    uploadHistory,
    totalRecords,
    hasData,

    // Methods
    uploadFile,
    fetchDashboard,
    fetchCreativeData,
    fetchUploadHistory,
    getDataWithCursor,

    // Formatters
    formatCurrency,
    formatNumber,
    formatPercent,
    formatRoi,
  };
}

// Cursor pagination method
async function getDataWithCursor(
  periodLabel: string = "",
  cursor?: string,
  limit: number = 100,
) {
  const params: Record<string, string> = {
    limit: limit.toString(),
  };

  if (periodLabel) {
    params.periodLabel = periodLabel;
  }

  if (cursor) {
    params.cursor = cursor;
  }

  return api.get("/analytics/tiktok-ads/data/cursor", { params });
}

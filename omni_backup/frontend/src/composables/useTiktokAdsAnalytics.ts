/**
 * TikTok Ads Analytics Composable
 * Handle API calls for TikTok Ads data
 * Single Responsibility: API communication only
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";

const api = useApi();

// Types
export interface UploadBatchResult {
  success: boolean;
  batchId: string;
  fileName: string;
  periodStart: string;
  periodEnd: string;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  updatedRows: number;
  errors: string[];
}

export interface DashboardSummary {
  totalCost: number;
  totalRevenue: number;
  totalOrders: number;
  avgRoi: number;
  totalImpressions: number;
  totalClicks: number;
  avgCtr: number;
  avgConversionRate: number;
  topProducts: ProductPerformance[];
  creativeTypeComparison: CreativeTypeStats[];
}

export interface ProductPerformance {
  productId: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
}

export interface CreativeTypeStats {
  creativeType: string;
  cost: number;
  revenue: number;
  orders: number;
  roi: number;
  costPerOrder: number;
}

export interface CreativeData {
  id: number;
  campaignId: string;
  campaignName: string;
  productId: string;
  creativeType: string;
  videoTitle: string | null;
  cost: number;
  ordersSku: number;
  grossRevenue: number;
  roi: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversionRate: number;
  periodStart: string;
  periodEnd: string;
}

export interface UploadBatch {
  id: string;
  fileName: string;
  periodStart: string;
  periodEnd: string;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  status: string;
  createdAt: string;
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
    mode: "skip" | "update" = "skip"
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
    periodEnd?: string
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
    } = {}
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

    // Formatters
    formatCurrency,
    formatNumber,
    formatPercent,
    formatRoi,
  };
}

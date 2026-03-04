/**
 * Shopee Ads Analytics Composable
 * Handle API calls for Shopee Ads data
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
  periodLabel: string;
  totalRows: number;
  insertedRows: number;
  skippedRows: number;
  updatedRows: number;
  errors: string[];
}

export interface DashboardSummary {
  totalCost: number;
  totalRevenue: number;
  totalDirectRevenue: number;
  totalOrders: number;
  avgRoas: number;
  avgDirectRoas: number;
  totalImpressions: number;
  totalClicks: number;
  avgCtr: number;
  avgConversionRate: number;
  topProducts: ProductPerformance[];
  biddingModeComparison: BiddingModeStats[];
}

export interface ProductPerformance {
  productId: string;
  productName: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
}

export interface BiddingModeStats {
  biddingMode: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
  costPerOrder: number;
  productCount: number;
}

export interface ProductData {
  id: number;
  productId: string;
  productName: string;
  biddingMode: string | null;
  cost: number;
  revenue: number;
  directRevenue: number;
  conversions: number;
  roas: number;
  directRoas: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversionRate: number;
  periodStart: string;
  periodEnd: string;
  periodLabel: string;
}

export interface UploadBatch {
  id: string;
  filename: string;
  periodLabel: string;
  recordCount: number;
  uploadedAt: string;
}

export function useShopeeAdsAnalytics() {
  // State
  const loading = ref(false);
  const uploading = ref(false);
  const error = ref<string | null>(null);
  const dashboard = ref<DashboardSummary | null>(null);
  const productData = ref<ProductData[]>([]);
  const uploadHistory = ref<UploadBatch[]>([]);
  const totalRecords = ref(0);

  // Computed
  const hasData = computed(() => productData.value.length > 0);

  const dashboardSummary = computed(() => dashboard.value);

  const topProducts = computed(() => dashboard.value?.topProducts || []);

  // API Methods
  async function uploadFile(
    file: File,
    periodLabel: string,
  ): Promise<UploadBatchResult | null> {
    uploading.value = true;
    error.value = null;

    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("periodLabel", periodLabel);

      const response = await api.client.post<{
        success: boolean;
        data: UploadBatchResult;
      }>("/analytics/shopee-ads/upload", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      if (response.data?.success) {
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

      const url = `/analytics/shopee-ads/dashboard${params.toString() ? `?${params}` : ""}`;

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

  async function fetchProductData(
    options: {
      periodStart?: string;
      periodEnd?: string;
      productId?: string;
      biddingMode?: string;
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

      const url = `/analytics/shopee-ads/data${params.toString() ? `?${params}` : ""}`;

      const response = await api.client.get<{
        success: boolean;
        data: ProductData[];
        pagination: { total: number };
      }>(url);

      if (response.data?.success) {
        productData.value = response.data.data;
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
      }>(`/analytics/shopee-ads/uploads?limit=${limit}`);

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

  function formatRoas(value: number): string {
    return value.toFixed(2);
  }

  return {
    // State
    loading,
    uploading,
    error,
    dashboard,
    dashboardSummary,
    topProducts,
    productData,
    uploadHistory,
    totalRecords,
    hasData,

    // Methods
    uploadFile,
    fetchDashboard,
    fetchProductData,
    fetchUploadHistory,

    // Formatters
    formatCurrency,
    formatNumber,
    formatPercent,
    formatRoas,
  };
}

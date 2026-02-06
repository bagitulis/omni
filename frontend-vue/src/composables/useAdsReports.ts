/**
 * Ads Reports Composable
 * Handle API calls for HTML reports from notebooks
 * Single Responsibility: Reports API communication
 */

import { ref } from "vue";
import { useApi } from "./useApi";

const api = useApi();

export interface ReportInfo {
  filename: string;
  type: "full" | "executive";
  platform: "shopee" | "tiktok";
  period: string;
  created_at: string;
  size: number;
}

export function useAdsReports() {
  const loading = ref(false);
  const error = ref<string | null>(null);
  const shopeeReports = ref<ReportInfo[]>([]);
  const tiktokReports = ref<ReportInfo[]>([]);

  async function fetchShopeeReports(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.client.get<{
        success: boolean;
        data: ReportInfo[];
      }>("/reports/shopee/ads");
      if (response.data?.success) {
        shopeeReports.value = response.data.data;
      }
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to fetch reports";
    } finally {
      loading.value = false;
    }
  }

  async function fetchTiktokReports(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.client.get<{
        success: boolean;
        data: ReportInfo[];
      }>("/reports/tiktok/ads");
      if (response.data?.success) {
        tiktokReports.value = response.data.data;
      }
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to fetch reports";
    } finally {
      loading.value = false;
    }
  }

  function getReportUrl(
    platform: "shopee" | "tiktok",
    type: "full" | "executive",
  ): string {
    const baseUrl = api.client.defaults.baseURL || "";
    return `${baseUrl}/reports/${platform}/latest?type=${type}`;
  }

  function getReportFileUrl(
    platform: "shopee" | "tiktok",
    filename: string,
  ): string {
    const baseUrl = api.client.defaults.baseURL || "";
    return `${baseUrl}/reports/${platform}/${filename}`;
  }

  return {
    loading,
    error,
    shopeeReports,
    tiktokReports,
    fetchShopeeReports,
    fetchTiktokReports,
    getReportUrl,
    getReportFileUrl,
  };
}

import apiClient from "./client";
import type {
  ShopeeAdsDashboardData,
  ShopeeAdsProductData,
  TiktokAdsDashboardData,
  TiktokAdsCreativeData,
  AdsPagination,
} from "@/types/ads";

// --- Shopee Ads Analytics ---

export async function getShopeeAdsDashboard() {
  const response = await apiClient.get<ShopeeAdsDashboardData>(
    "/analytics/shopee-ads/dashboard",
  );
  return response;
}

export interface ShopeeAdsDataParams {
  limit?: number;
  offset?: number;
  orderBy?: string;
  orderDir?: "asc" | "desc";
  productId?: string;
  biddingMode?: string;
}

export async function getShopeeAdsData(params?: ShopeeAdsDataParams) {
  const response = await apiClient.get<{
    items: ShopeeAdsProductData[];
    pagination: AdsPagination;
  }>("/analytics/shopee-ads/data", { params });
  return {
    success: response.success,
    data: response.data?.items ?? (response.data as unknown as ShopeeAdsProductData[]) ?? [],
    pagination: response.data?.pagination ?? { total: 0, limit: 50, offset: 0 },
  };
}

// --- TikTok Ads Analytics ---

export async function getTiktokAdsDashboard() {
  const response = await apiClient.get<TiktokAdsDashboardData>(
    "/analytics/tiktok-ads/dashboard",
  );
  return response;
}

export interface TiktokAdsDataParams {
  limit?: number;
  offset?: number;
  orderBy?: string;
  orderDir?: "asc" | "desc";
  productId?: string;
  creativeType?: string;
}

export async function getTiktokAdsData(params?: TiktokAdsDataParams) {
  const response = await apiClient.get<{
    items: TiktokAdsCreativeData[];
    pagination: AdsPagination;
  }>("/analytics/tiktok-ads/data", { params });
  return {
    success: response.success,
    data: response.data?.items ?? (response.data as unknown as TiktokAdsCreativeData[]) ?? [],
    pagination: response.data?.pagination ?? { total: 0, limit: 50, offset: 0 },
  };
}

// --- Ads Reports ---

export interface ReportInfo {
  filename: string;
  type: "full" | "executive";
  platform: "shopee" | "tiktok";
  period: string;
  created_at: string;
  size: number;
}

export async function getShopeeAdsReports() {
  const response = await apiClient.get<ReportInfo[]>(
    "/analytics/ml/reports/shopee/list",
  );
  return response;
}

export async function getTiktokAdsReports() {
  const response = await apiClient.get<ReportInfo[]>(
    "/analytics/ml/reports/tiktok/list",
  );
  return response;
}

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
  const response = await apiClient.get<{
    success: boolean;
    data: ShopeeAdsDashboardData;
  }>("/analytics/shopee-ads/dashboard");
  return response.data;
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
    success: boolean;
    data: ShopeeAdsProductData[];
    pagination: AdsPagination;
  }>("/analytics/shopee-ads/data", { params });
  return response.data;
}

// --- TikTok Ads Analytics ---

export async function getTiktokAdsDashboard() {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsDashboardData;
  }>("/analytics/tiktok-ads/dashboard");
  return response.data;
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
    success: boolean;
    data: TiktokAdsCreativeData[];
    pagination: AdsPagination;
  }>("/analytics/tiktok-ads/data", { params });
  return response.data;
}

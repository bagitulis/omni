import apiClient from "./client";
import type {
  ShopeeAdsDashboardData,
  ShopeeAdsProductData,
  ShopeeAdsUploadBatch,
  TiktokAdsDashboardData,
  TiktokAdsCreativeData,
  TiktokAdsUploadBatch,
  ShopeeAdsSummary,
  ShopeeAdsTrend,
  ShopeeAdsProductPerformance,
  TiktokAdsSummary,
  TiktokAdsTrend,
  TiktokAdsProductPerformance,
  TiktokAdsPrediction,
  AdsReport,
  AdsReportLatest,
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

export async function getShopeeAdsUploads(limit = 20) {
  const response = await apiClient.get<{
    success: boolean;
    data: ShopeeAdsUploadBatch[];
  }>("/analytics/shopee-ads/uploads", { params: { limit } });
  return response.data;
}

export async function uploadShopeeAds(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const response = await apiClient.post<{
    success: boolean;
    message: string;
    data: any;
  }>("/analytics/shopee-ads/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
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

export async function getTiktokAdsUploads(limit = 20) {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsUploadBatch[];
  }>("/analytics/tiktok-ads/uploads", { params: { limit } });
  return response.data;
}

export async function uploadTiktokAds(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const response = await apiClient.post<{
    success: boolean;
    message: string;
    data: any;
  }>("/analytics/tiktok-ads/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return response.data;
}

// --- Ads Management (Shopee) ---

export async function uploadShopeeAdsFile(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const response = await apiClient.post<{
    success: boolean;
    message: string;
    totalRows: number;
    period: string;
  }>("/ads/shopee/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return response.data;
}

export async function getShopeeAds(period?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: any; // Use specific type if known, currently generic data based on handler
  }>("/ads/shopee", { params: { period } });
  return response.data;
}

export async function getShopeeAdsSummary(period?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: ShopeeAdsSummary;
  }>("/ads/shopee/summary", { params: { period } });
  return response.data;
}

export async function getShopeeAdsTrends(startDate?: string, endDate?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: ShopeeAdsTrend[];
  }>("/ads/shopee/trends", { params: { startDate, endDate } });
  return response.data;
}

export async function getShopeeAdsPerformance(
  startDate?: string,
  endDate?: string,
) {
  const response = await apiClient.get<{
    success: boolean;
    data: ShopeeAdsProductPerformance[];
  }>("/ads/shopee/performance", { params: { startDate, endDate } });
  return response.data;
}

// --- Ads Management (TikTok) ---

export async function uploadTiktokAdsFile(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const response = await apiClient.post<{
    success: boolean;
    message: string;
    totalRows: number;
    period: string;
  }>("/ads/tiktok/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return response.data;
}

export async function getTiktokAds(period?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: any; // Use specific type if known
  }>("/ads/tiktok", { params: { period } });
  return response.data;
}

export async function getTiktokAdsSummary(period?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsSummary;
  }>("/ads/tiktok/summary", { params: { period } });
  return response.data;
}

export async function getTiktokAdsTrends(startDate?: string, endDate?: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsTrend[];
  }>("/ads/tiktok/trends", { params: { startDate, endDate } });
  return response.data;
}

export async function getTiktokAdsPerformance(
  startDate?: string,
  endDate?: string,
) {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsProductPerformance[];
  }>("/ads/tiktok/performance", { params: { startDate, endDate } });
  return response.data;
}

export async function getTiktokAdsPredictions() {
  const response = await apiClient.get<{
    success: boolean;
    data: TiktokAdsPrediction[];
  }>("/ads/tiktok/predictions");
  return response.data;
}

// --- Reports ---

export async function getShopeeAdsReports(
  startDate?: string,
  endDate?: string,
) {
  const response = await apiClient.get<{
    success: boolean;
    data: AdsReport;
  }>("/reports/shopee/ads", { params: { startDate, endDate } });
  return response.data;
}

export async function getShopeeAdsReportsLatest() {
  const response = await apiClient.get<{
    success: boolean;
    data: AdsReportLatest;
  }>("/reports/shopee/ads/latest");
  return response.data;
}

export async function getShopeeAdsReportByDate(date: string) {
  const response = await apiClient.get<{
    success: boolean;
    data: AdsReport;
  }>(`/reports/shopee/ads/${date}`);
  return response.data;
}

export async function getTiktokAdsReports(
  startDate?: string,
  endDate?: string,
) {
  const response = await apiClient.get<{
    success: boolean;
    data: AdsReport;
  }>("/reports/tiktok/ads", { params: { startDate, endDate } });
  return response.data;
}

export async function getTiktokAdsReportsLatest() {
  const response = await apiClient.get<{
    success: boolean;
    data: AdsReportLatest;
  }>("/reports/tiktok/ads/latest");
  return response.data;
}

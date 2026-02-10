import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as adsApi from "@/api/ads";
import type { ShopeeAdsDataParams, TiktokAdsDataParams } from "@/api/ads";

// --- Shopee Ads Analytics Hooks ---

export function useShopeeAdsDashboard() {
  return useQuery({
    queryKey: ["shopee-ads-dashboard"],
    queryFn: () => adsApi.getShopeeAdsDashboard(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useShopeeAdsData(params?: ShopeeAdsDataParams) {
  return useQuery({
    queryKey: ["shopee-ads-data", params],
    queryFn: () => adsApi.getShopeeAdsData(params),
    placeholderData: (previousData) => previousData,
  });
}

export function useShopeeAdsUploads(limit = 20) {
  return useQuery({
    queryKey: ["shopee-ads-uploads", limit],
    queryFn: () => adsApi.getShopeeAdsUploads(limit),
  });
}

export function useUploadShopeeAds() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adsApi.uploadShopeeAds(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shopee-ads-dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["shopee-ads-data"] });
      queryClient.invalidateQueries({ queryKey: ["shopee-ads-uploads"] });
    },
  });
}

// --- TikTok Ads Analytics Hooks ---

export function useTiktokAdsDashboard() {
  return useQuery({
    queryKey: ["tiktok-ads-dashboard"],
    queryFn: () => adsApi.getTiktokAdsDashboard(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useTiktokAdsData(params?: TiktokAdsDataParams) {
  return useQuery({
    queryKey: ["tiktok-ads-data", params],
    queryFn: () => adsApi.getTiktokAdsData(params),
    placeholderData: (previousData) => previousData,
  });
}

export function useTiktokAdsUploads(limit = 20) {
  return useQuery({
    queryKey: ["tiktok-ads-uploads", limit],
    queryFn: () => adsApi.getTiktokAdsUploads(limit),
  });
}

export function useUploadTiktokAds() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adsApi.uploadTiktokAds(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-data"] });
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-uploads"] });
    },
  });
}

// --- Ads Management Hooks (Shopee) ---

export function useUploadShopeeAdsFile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adsApi.uploadShopeeAdsFile(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads-shopee"] });
    },
  });
}

export function useShopeeAds(period?: string) {
  return useQuery({
    queryKey: ["ads-shopee", period],
    queryFn: () => adsApi.getShopeeAds(period),
  });
}

export function useShopeeAdsSummary(period?: string) {
  return useQuery({
    queryKey: ["ads-shopee-summary", period],
    queryFn: () => adsApi.getShopeeAdsSummary(period),
  });
}

export function useShopeeAdsTrends(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["ads-shopee-trends", startDate, endDate],
    queryFn: () => adsApi.getShopeeAdsTrends(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

export function useShopeeAdsPerformance(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["ads-shopee-performance", startDate, endDate],
    queryFn: () => adsApi.getShopeeAdsPerformance(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

// --- Ads Management Hooks (TikTok) ---

export function useUploadTiktokAdsFile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => adsApi.uploadTiktokAdsFile(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads-tiktok"] });
    },
  });
}

export function useTiktokAds(period?: string) {
  return useQuery({
    queryKey: ["ads-tiktok", period],
    queryFn: () => adsApi.getTiktokAds(period),
  });
}

export function useTiktokAdsSummary(period?: string) {
  return useQuery({
    queryKey: ["ads-tiktok-summary", period],
    queryFn: () => adsApi.getTiktokAdsSummary(period),
  });
}

export function useTiktokAdsTrends(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["ads-tiktok-trends", startDate, endDate],
    queryFn: () => adsApi.getTiktokAdsTrends(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

export function useTiktokAdsPerformance(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["ads-tiktok-performance", startDate, endDate],
    queryFn: () => adsApi.getTiktokAdsPerformance(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

export function useTiktokAdsPredictions() {
  return useQuery({
    queryKey: ["ads-tiktok-predictions"],
    queryFn: () => adsApi.getTiktokAdsPredictions(),
  });
}

// --- Reports Hooks ---

export function useShopeeAdsReports(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["reports-shopee-ads", startDate, endDate],
    queryFn: () => adsApi.getShopeeAdsReports(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

export function useShopeeAdsReportsLatest() {
  return useQuery({
    queryKey: ["reports-shopee-ads-latest"],
    queryFn: () => adsApi.getShopeeAdsReportsLatest(),
  });
}

export function useShopeeAdsReportByDate(date: string) {
  return useQuery({
    queryKey: ["reports-shopee-ads-date", date],
    queryFn: () => adsApi.getShopeeAdsReportByDate(date),
    enabled: !!date,
  });
}

export function useTiktokAdsReports(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: ["reports-tiktok-ads", startDate, endDate],
    queryFn: () => adsApi.getTiktokAdsReports(startDate, endDate),
    enabled: !!startDate && !!endDate,
  });
}

export function useTiktokAdsReportsLatest() {
  return useQuery({
    queryKey: ["reports-tiktok-ads-latest"],
    queryFn: () => adsApi.getTiktokAdsReportsLatest(),
  });
}

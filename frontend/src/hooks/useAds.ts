import { useQuery } from "@tanstack/react-query";
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

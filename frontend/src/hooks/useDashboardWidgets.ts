import { useQuery } from "@tanstack/react-query";
import {
  getWalletData,
  getShippingFeeData,
  getSyncStatus,
} from "@/api/dashboard";

export function useWalletData(platform: string) {
  return useQuery({
    queryKey: ["wallet", platform],
    queryFn: () => getWalletData(platform),
    staleTime: 60 * 1000,
  });
}

export function useShippingFeeData(platform: string) {
  return useQuery({
    queryKey: ["shipping-fee", platform],
    queryFn: () => getShippingFeeData(platform),
    staleTime: 60 * 1000,
  });
}

export function useSyncStatus(platform: string) {
  return useQuery({
    queryKey: ["sync-status", platform],
    queryFn: () => getSyncStatus(platform),
    staleTime: 30 * 1000, // Sync status updates more frequently
    refetchInterval: 30 * 1000,
  });
}

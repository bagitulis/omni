import { useQuery } from "@tanstack/react-query";
import {
  getWalletData,
  getShippingFeeData,
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


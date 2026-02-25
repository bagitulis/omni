import { useQuery } from "@tanstack/react-query";
import {
  getPortfolioHealth,
  getMLProducts,
  type MLProductsParams,
} from "@/api/mlAnalytics";

export function usePortfolioHealth(platform: string = "tiktok") {
  return useQuery({
    queryKey: ["ml", "portfolio-health", platform],
    queryFn: () => getPortfolioHealth(platform),
    staleTime: 60 * 1000, // 1 minute (matches backend cache)
  });
}

export function useMLProducts(params: MLProductsParams = {}) {
  return useQuery({
    queryKey: ["ml", "products", params],
    queryFn: () => getMLProducts(params),
    staleTime: 60 * 1000,
  });
}


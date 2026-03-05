import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import {
  getDbProducts,
  syncPlatformProducts,
  type GetDbProductsParams,
} from "@/api/productManager";
import type { ProductManagerPlatform } from "@/types/product_manager";

export function useDbProducts(
  platform: ProductManagerPlatform,
  params: GetDbProductsParams,
  options?: { autoRefresh?: boolean; refetchInterval?: number },
) {
  const autoRefresh = options?.autoRefresh ?? true;
  const refetchInterval = options?.refetchInterval ?? 15_000;

  return useQuery({
    queryKey: ["product_manager", platform, params],
    queryFn: () => getDbProducts(platform, params),
    staleTime: 5_000,
    refetchInterval: autoRefresh ? refetchInterval : false,
    refetchIntervalInBackground: false,
  });
}

export function useSyncPlatformProducts(platform: ProductManagerPlatform) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => syncPlatformProducts(platform),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["product_manager", platform],
      });
    },
    onError: (error: Error) => {
      message.error(error.message || `Failed to sync ${platform} products`);
    },
  });
}

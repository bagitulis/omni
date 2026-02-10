import {
  useQuery,
  useMutation,
  useQueryClient,
  UseQueryResult,
  UseMutationResult,
} from "@tanstack/react-query";
import {
  cloneProduct,
  batchClone,
  getCloneStatus,
  getProductData,
  getAvailableTargets,
  getClonePreview,
} from "@/api/clone";
import {
  CloneRequest,
  BatchCloneRequest,
  CloneResult,
  BatchCloneResult,
  ProductData,
  CloneTargetsResult,
  ConflictResult,
} from "@/types/clone";

// =============================================================================
// Mutation Hooks
// =============================================================================

/**
 * Hook for single product clone mutation
 */
export function useCloneProduct(): UseMutationResult<
  CloneResult,
  Error,
  CloneRequest,
  unknown
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: CloneRequest) => cloneProduct(req),
    onSuccess: () => {
      // Invalidate clone status queries when clone completes
      queryClient.invalidateQueries({ queryKey: ["clone"] });
    },
  });
}

/**
 * Hook for batch product clone mutation
 */
export function useBatchClone(): UseMutationResult<
  BatchCloneResult,
  Error,
  BatchCloneRequest,
  unknown
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: BatchCloneRequest) => batchClone(req),
    onSuccess: () => {
      // Invalidate clone status queries when batch completes
      queryClient.invalidateQueries({ queryKey: ["clone"] });
    },
  });
}

// =============================================================================
// Query Hooks
// =============================================================================

/**
 * Hook for polling clone status with auto-refetch when in progress
 */
export function useCloneStatus(id: string): UseQueryResult<CloneResult, Error> {
  return useQuery({
    queryKey: ["clone", "status", id],
    queryFn: () => getCloneStatus(id),
    enabled: Boolean(id), // Only fetch if ID is provided
    staleTime: 5 * 1000, // 5 seconds - poll frequently while in progress
    refetchInterval: 2 * 1000, // Poll every 2 seconds for active clones
  });
}

/**
 * Hook for fetching product data for clone preview/form
 */
export function useProductData(
  platform: string,
  sku: string,
): UseQueryResult<ProductData, Error> {
  return useQuery({
    queryKey: ["clone", "product-data", platform, sku],
    queryFn: () => getProductData(platform, sku),
    enabled: Boolean(platform && sku), // Only fetch if both params provided
    staleTime: 10 * 60 * 1000, // 10 minutes - product data rarely changes
  });
}

/**
 * Hook for fetching available clone targets for a SKU
 */
export function useAvailableTargets(
  sku: string,
): UseQueryResult<CloneTargetsResult, Error> {
  return useQuery({
    queryKey: ["clone", "available-targets", sku],
    queryFn: () => getAvailableTargets(sku),
    enabled: Boolean(sku), // Only fetch if SKU provided
    staleTime: 10 * 60 * 1000, // 10 minutes - target status rarely changes
  });
}

/**
 * Hook for fetching clone preview with conflict detection
 */
export function useClonePreview(params: {
  source_platform: string;
  target_platform: string;
  source_item_id: string;
  sku?: string;
}): UseQueryResult<ConflictResult, Error> {
  const allParamsPresent = Boolean(
    params.source_platform && params.target_platform && params.source_item_id,
  );

  return useQuery({
    queryKey: ["clone", "preview", params],
    queryFn: () => getClonePreview(params),
    enabled: allParamsPresent, // Only fetch if all required params present
    staleTime: 5 * 60 * 1000, // 5 minutes - preview data changes less frequently
  });
}

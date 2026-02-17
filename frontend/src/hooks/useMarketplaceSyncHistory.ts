import { useQuery } from "@tanstack/react-query";
import { getMarketplaceSyncHistory } from "@/api/marketplaceSyncHistory";
import type { MarketplaceSyncHistoryFilter } from "@/types/shared";

/**
 * TanStack Query hook for fetching marketplace sync history
 * @param filter - Optional filters for platform, operation, status, SKU search, date range, and pagination
 * @returns Query result with entries, loading state, and error handling
 */
export function useMarketplaceSyncHistory(
  filter: MarketplaceSyncHistoryFilter = {},
) {
  return useQuery({
    queryKey: ["marketplace-sync-history", filter],
    queryFn: () => getMarketplaceSyncHistory(filter),
    staleTime: 30 * 1000, // 30 seconds
  });
}

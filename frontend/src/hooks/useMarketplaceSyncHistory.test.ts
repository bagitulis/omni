import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

const getMarketplaceSyncHistoryMock = vi.fn();

vi.mock("@/api/marketplaceSyncHistory", () => ({
  getMarketplaceSyncHistory: (...args: unknown[]) =>
    getMarketplaceSyncHistoryMock(...args),
}));

import { useMarketplaceSyncHistory } from "./useMarketplaceSyncHistory";

describe("useMarketplaceSyncHistory", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey for default filters", () => {
    useMarketplaceSyncHistory();

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
      queryFn: () => void;
      staleTime: number;
    };

    expect(queryOptions.queryKey).toEqual(["marketplace-sync-history", {}]);
    expect(queryOptions.staleTime).toBe(30 * 1000);
  });

  it("calls useQuery with correct queryKey for custom filters", () => {
    const filter = {
      platform: "shopee" as const,
      status: "success" as const,
      page: 2,
      page_size: 20,
    };

    useMarketplaceSyncHistory(filter);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };

    expect(queryOptions.queryKey).toEqual(["marketplace-sync-history", filter]);
  });

  it("calls API client via queryFn", () => {
    useMarketplaceSyncHistory();

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => void;
    };

    queryOptions.queryFn();

    expect(getMarketplaceSyncHistoryMock).toHaveBeenCalledWith({});
  });

  it("passes filter to API client via queryFn", () => {
    const filter = {
      platform: "tiktok" as const,
      operation: "stock_update" as const,
      sku_search: "TEST-SKU",
      date_from: "2025-01-01",
      date_to: "2025-01-31",
    };

    useMarketplaceSyncHistory(filter);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => void;
    };

    queryOptions.queryFn();

    expect(getMarketplaceSyncHistoryMock).toHaveBeenCalledWith(filter);
  });

  it("ensures query key stability for identical filters", () => {
    const filter1 = { platform: "lazada" as const, page: 1 };
    const filter2 = { platform: "lazada" as const, page: 1 };

    useMarketplaceSyncHistory(filter1);
    useMarketplaceSyncHistory(filter2);

    const queryKey1 = (
      useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] }
    ).queryKey;
    const queryKey2 = (
      useQueryMock.mock.calls[1]?.[0] as { queryKey: unknown[] }
    ).queryKey;

    // Query keys should be structurally equal for cache deduplication
    expect(JSON.stringify(queryKey1)).toBe(JSON.stringify(queryKey2));
  });
});

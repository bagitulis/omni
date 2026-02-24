import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("@/api/productManager", () => ({
  getDbProducts: vi.fn(),
  syncPlatformProducts: vi.fn(),
}));

import { useDbProducts, useSyncPlatformProducts } from "./useProductManager";
import * as productManagerApi from "@/api/productManager";
import type { ProductManagerPlatform } from "@/types/product_manager";

describe("useDbProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with correct queryKey including platform and params", () => {
    const platform: ProductManagerPlatform = "shopee";
    const params = { page: 1, limit: 20 };
    useDbProducts(platform, params);

    expect(useQueryMock).toHaveBeenCalledTimes(1);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["product_manager", "shopee", params]);
  });

  it("calls getDbProducts with platform and params in queryFn", async () => {
    vi.mocked(productManagerApi.getDbProducts).mockResolvedValue({
      products: [],
      total: 0,
    } as unknown as ReturnType<
      typeof productManagerApi.getDbProducts
    > extends Promise<infer T>
      ? T
      : never);

    const platform: ProductManagerPlatform = "lazada";
    const params = { search: "test" } as unknown as import("@/api/productManager").GetDbProductsParams;

    useDbProducts(platform, params);

    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(productManagerApi.getDbProducts).toHaveBeenCalledWith(
      "lazada",
      params,
    );
  });

  it("uses 5 second staleTime", () => {
    useDbProducts("shopee", {});
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };
    expect(opts.staleTime).toBe(5_000);
  });

  it("enables auto-refresh by default with 15s interval", () => {
    useDbProducts("shopee", {});
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      refetchInterval: number | false;
      refetchIntervalInBackground: boolean;
    };
    expect(opts.refetchInterval).toBe(15_000);
    expect(opts.refetchIntervalInBackground).toBe(false);
  });

  it("disables auto-refresh when autoRefresh=false", () => {
    useDbProducts("shopee", {}, { autoRefresh: false });
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      refetchInterval: number | false;
    };
    expect(opts.refetchInterval).toBe(false);
  });

  it("uses custom refetchInterval when provided", () => {
    useDbProducts("shopee", {}, { autoRefresh: true, refetchInterval: 30_000 });
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      refetchInterval: number | false;
    };
    expect(opts.refetchInterval).toBe(30_000);
  });
});

describe("useSyncPlatformProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useSyncPlatformProducts("shopee");
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls syncPlatformProducts with platform in mutationFn", async () => {
    vi.mocked(productManagerApi.syncPlatformProducts).mockResolvedValue({
      success: true,
    });

    useSyncPlatformProducts("tiktok");
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: () => Promise<unknown>;
    };
    await opts.mutationFn();
    expect(productManagerApi.syncPlatformProducts).toHaveBeenCalledWith(
      "tiktok",
    );
  });

  it("invalidates product_manager queries for platform on success", async () => {
    useSyncPlatformProducts("lazada");
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: () => void;
    };
    opts.onSuccess?.();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["product_manager", "lazada"],
    });
  });

  it("uses correct platform for different calls", async () => {
    useSyncPlatformProducts("shopee");
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: () => void;
    };
    opts.onSuccess?.();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["product_manager", "shopee"],
    });
  });
});

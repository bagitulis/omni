import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

const getProductsMock = vi.fn();

vi.mock("@/api/products", () => ({
  getProducts: (...args: unknown[]) => getProductsMock(...args),
}));

import { useUnifiedProducts } from "./useUnifiedProducts";
import type { ProductFilterValues } from "@/types/shared";

const defaultFilters: ProductFilterValues = {
  search: "",
  platform: "all",
  status: "all",
  category: "all",
};

describe("useUnifiedProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey including filters, page, pageSize", () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };

    expect(queryOptions.queryKey).toEqual([
      "unified-products",
      defaultFilters,
      1,
      20,
    ]);
  });

  it("calls useQuery with staleTime of 30 seconds", () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };

    expect(queryOptions.staleTime).toBe(30_000);
  });

  it("queryFn calls getProducts with page and pageSize as limit", async () => {
    const filters: ProductFilterValues = {
      search: "shoes",
      platform: "shopee",
      status: "active",
      category: "all",
    };

    useUnifiedProducts(filters, 2, 50);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [],
      meta: { total: 0, page: 2, page_size: 50 },
    });

    await queryOptions.queryFn();

    expect(getProductsMock).toHaveBeenCalledWith({
      page: 2,
      limit: 50,
      search: "shoes",
      status: "active",
      platform: "shopee",
      linked_only: true,
    });
  });

  it("queryFn passes undefined for status when status is 'all'", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [],
      meta: { total: 0, page: 1, page_size: 20 },
    });

    await queryOptions.queryFn();

    expect(getProductsMock).toHaveBeenCalledWith({
      page: 1,
      limit: 20,
      search: undefined,
      status: undefined,
      platform: undefined,
      linked_only: true,
    });
  });

  it("queryFn keeps MasterProduct archived status", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<{
        products: Array<{ status: string }>;
        total: number;
      }>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [
        {
          id: 1,
          tenant_id: "t1",
          title: "Archived Product",
          description: "",
          images: [],
          status: "archived",
          created_at: "2025-01-01T00:00:00Z",
          updated_at: "2025-01-01T00:00:00Z",
          skus: [],
        },
      ],
      meta: { total: 1, page: 1, page_size: 20 },
    });

    const result = await queryOptions.queryFn();

    expect(result.products[0].status).toBe("archived");
  });

  it("queryFn returns total from meta.total when available", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<{ products: unknown[]; total: number }>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [],
      meta: { total: 99, page: 1, page_size: 20 },
    });

    const result = await queryOptions.queryFn();

    expect(result.total).toBe(99);
  });

  it("queryFn falls back to products.length when meta is absent", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<{ products: unknown[]; total: number }>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [
        {
          id: 2,
          tenant_id: "t1",
          title: "Product A",
          description: "",
          images: [],
          status: "active",
          created_at: "2025-01-01T00:00:00Z",
          updated_at: "2025-01-01T00:00:00Z",
          skus: [],
        },
      ],
    });

    const result = await queryOptions.queryFn();

    // meta is absent, so total falls back to products.length (1)
    expect(result.total).toBe(1);
  });

  it("queryFn maps failed sync_status to error in platform_summary", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<{
        products: Array<{
          platform_summary: { shopee: string; tiktok: string; lazada: string };
        }>;
        total: number;
      }>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [
        {
          id: 3,
          tenant_id: "t1",
          title: "Product B",
          description: "",
          images: [],
          status: "active",
          created_at: "2025-01-01T00:00:00Z",
          updated_at: "2025-01-01T00:00:00Z",
          skus: [
            {
              id: 10,
              tenant_id: "t1",
              master_product_id: 3,
              seller_sku: "SKU-B",
              variant_name: "Default",
              variant_data: {},
              price: 5000,
              stock: 10,
              created_at: "2025-01-01T00:00:00Z",
              updated_at: "2025-01-01T00:00:00Z",
              platform_links: [
                {
                  id: 1,
                  master_product_id: 3,
                  platform: "shopee",
                  sync_status: "synced",
                },
                {
                  id: 2,
                  master_product_id: 3,
                  platform: "tiktok",
                  sync_status: "failed",
                },
              ],
            },
          ],
        },
      ],
      meta: { total: 1, page: 1, page_size: 20 },
    });

    const result = await queryOptions.queryFn();
    const summary = result.products[0].platform_summary;

    expect(summary.shopee).toBe("linked");
    expect(summary.tiktok).toBe("error");
    expect(summary.lazada).toBe("not_linked");
  });

  it("queryFn treats outdated as linked and error as error", async () => {
    useUnifiedProducts(defaultFilters, 1, 20);

    const queryOptions = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<{
        products: Array<{
          platform_summary: { shopee: string; tiktok: string; lazada: string };
          skus: Array<{
            platform_links: Array<{ sync_status: string }>;
          }>;
        }>;
      }>;
    };

    getProductsMock.mockResolvedValue({
      success: true,
      data: [
        {
          id: 4,
          tenant_id: "t1",
          title: "Product C",
          description: "",
          images: [],
          status: "active",
          created_at: "2025-01-01T00:00:00Z",
          updated_at: "2025-01-01T00:00:00Z",
          skus: [
            {
              id: 11,
              tenant_id: "t1",
              master_product_id: 4,
              seller_sku: "SKU-C",
              variant_name: "Default",
              variant_data: {},
              price: 5000,
              stock: 10,
              created_at: "2025-01-01T00:00:00Z",
              updated_at: "2025-01-01T00:00:00Z",
              platform_links: [
                {
                  id: 3,
                  master_product_id: 4,
                  platform: "shopee",
                  sync_status: "outdated",
                },
                {
                  id: 4,
                  master_product_id: 4,
                  platform: "tiktok",
                  sync_status: "error",
                },
              ],
            },
          ],
        },
      ],
      meta: { total: 1, page: 1, page_size: 20 },
    });

    const result = await queryOptions.queryFn();
    const summary = result.products[0].platform_summary;
    const links = result.products[0].skus[0].platform_links;

    expect(summary.shopee).toBe("linked");
    expect(summary.tiktok).toBe("error");
    expect(links[0].sync_status).toBe("outdated");
    expect(links[1].sync_status).toBe("error");
  });
});

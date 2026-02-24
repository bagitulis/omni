import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getDbProducts,
  syncPlatformProducts,
  syncSelectedProducts,
  getFilterPreferences,
  saveFilterPreferences,
} from "./productManager";

const { mockGet, mockPost, mockPatch, mockClientGet, mockClientPost } =
  vi.hoisted(() => ({
    mockGet: vi.fn(),
    mockPost: vi.fn(),
    mockPatch: vi.fn(),
    mockClientGet: vi.fn(),
    mockClientPost: vi.fn(),
  }));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: vi.fn(),
    patch: mockPatch,
    delete: vi.fn(),
    client: { get: mockClientGet, post: mockClientPost },
  },
}));

beforeEach(() => vi.clearAllMocks());

describe("getDbProducts", () => {
  it("returns product data for a platform", async () => {
    const mockData = {
      success: true,
      products: [{ id: 1 }],
      total: 1,
      offset: 0,
      limit: 20,
      count: 1,
    };
    mockClientGet.mockResolvedValue({ data: mockData });

    const result = await getDbProducts("shopee", { offset: 0, limit: 20 });

    expect(result).toEqual(mockData);
  });

  it("calls correct platform endpoint with params", async () => {
    mockClientGet.mockResolvedValue({
      data: {
        success: true,
        products: [],
        total: 0,
        offset: 0,
        limit: 20,
        count: 0,
      },
    });

    await getDbProducts("tiktok", { offset: 10, limit: 50 });

    expect(mockClientGet).toHaveBeenCalledWith("/tiktok/db/products", {
      params: { offset: 10, limit: 50 },
    });
  });

  it("throws when data.success is false", async () => {
    mockClientGet.mockResolvedValue({
      data: {
        success: false,
        products: [],
        total: 0,
        offset: 0,
        limit: 20,
        count: 0,
      },
    });

    await expect(getDbProducts("shopee", {})).rejects.toThrow(
      "Failed to fetch products",
    );
  });
});

describe("syncPlatformProducts", () => {
  it("returns sync response on success", async () => {
    const mockData = { success: true, data: { message: "Synced", synced: 5 } };
    mockClientPost.mockResolvedValue({ data: mockData });

    const result = await syncPlatformProducts("lazada");

    expect(result).toEqual(mockData);
  });

  it("calls correct platform sync endpoint", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: true, data: {} },
    });

    await syncPlatformProducts("shopee");

    expect(mockClientPost).toHaveBeenCalledWith("/shopee/sync/products");
  });

  it("throws when data.success is false", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: false, error: "Sync failed" },
    });

    await expect(syncPlatformProducts("tiktok")).rejects.toThrow("Sync failed");
  });
});

describe("syncSelectedProducts", () => {
  it("returns sync result on success", async () => {
    const mockData = { synced: 3, failed: 0, details: ["SKU-1", "SKU-2"] };
    mockPost.mockResolvedValue({ success: true, data: mockData });

    const result = await syncSelectedProducts([1, 2, 3]);

    expect(result).toEqual(mockData);
  });

  it("posts product_ids to correct endpoint", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { synced: 2, failed: 0, details: [] },
    });

    await syncSelectedProducts([10, 20]);

    expect(mockPost).toHaveBeenCalledWith("/products/master/sync-selected", {
      product_ids: [10, 20],
    });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Sync error" });

    await expect(syncSelectedProducts([1])).rejects.toThrow("Sync error");
  });
});

describe("getFilterPreferences", () => {
  it("returns normalized preferences with column_visibility as object", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {
        platform: "shopee",
        tab: "products",
        visible_columns: ["name", "price", "stock"],
        column_filters: {},
        search_query: "test query",
        locked_columns: [],
      },
    });

    const result = await getFilterPreferences("shopee", "products");

    expect(result).not.toBeNull();
    expect(result!.search).toBe("test query");
    expect(result!.column_visibility).toEqual({
      name: true,
      price: true,
      stock: true,
    });
    expect(result!.page_size).toBe(50);
  });

  it("returns null when response.success is false", async () => {
    mockGet.mockResolvedValue({ success: false, data: null });

    const result = await getFilterPreferences("shopee", "products");

    expect(result).toBeNull();
  });

  it("returns null when response.data is falsy", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    const result = await getFilterPreferences("tiktok", "orders");

    expect(result).toBeNull();
  });

  it("defaults search to empty string when search_query is absent", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {
        visible_columns: [],
        column_filters: {},
        search_query: "",
        locked_columns: [],
      },
    });

    const result = await getFilterPreferences("lazada", "products");

    expect(result!.search).toBe("");
    expect(result!.column_visibility).toEqual({});
  });
});

describe("saveFilterPreferences", () => {
  it("posts transformed payload to backend", async () => {
    mockPost.mockResolvedValue({ success: true });

    await saveFilterPreferences({
      platform: "shopee",
      page: "products",
      preferences: {
        search: "my search",
        column_visibility: { name: true, price: false, stock: true },
        page_size: 50,
      },
    });

    expect(mockPost).toHaveBeenCalledWith("/filter-preferences", {
      platform: "shopee",
      page: "products",
      visible_columns: expect.arrayContaining(["name", "stock"]),
      column_filters: {},
      search_query: "my search",
      locked_columns: [],
    });

    // price=false should NOT be in visible_columns
    const call = mockPost.mock.calls[0];
    expect(call[1].visible_columns).not.toContain("price");
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Save failed" });

    await expect(
      saveFilterPreferences({
        platform: "shopee",
        page: "products",
        preferences: { search: "", column_visibility: {}, page_size: 20 },
      }),
    ).rejects.toThrow("Save failed");
  });
});

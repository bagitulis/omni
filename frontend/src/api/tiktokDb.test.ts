import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockGet, mockPost, mockDelete } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockDelete: vi.fn(),
}));

vi.mock("./client", () => ({
  default: { get: mockGet, post: mockPost, delete: mockDelete },
}));

vi.mock("@/lib/constants", () => ({
  API_TIMEOUT: { EXTRA_LONG: 120000 },
}));

import {
  searchProducts,
  getAllProducts,
  getProductDetail,
  getProductListFromDB,
  getMasterProductsFromDB,
  searchProductsInDB,
  getProductsByStatus,
  getProductById,
  getStatistics,
  deleteProduct,
} from "./tiktokDb";

describe("searchProducts", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns products on success with default params", async () => {
    const products = [{ product_id: "p1", title: "Product 1" }];
    mockPost.mockResolvedValue({ success: true, data: products });
    const result = await searchProducts();
    expect(result).toEqual(products);
    expect(mockPost).toHaveBeenCalledWith(
      "/tiktok/products/search",
      {
        status: "ACTIVATE",
        page_size: 100,
        limit: undefined,
        sync_to_db: true,
      },
      { timeout: 120000 },
    );
  });

  it("passes custom params", async () => {
    mockPost.mockResolvedValue({ success: true, data: [] });
    await searchProducts("INACTIVE", 50, 200);
    expect(mockPost).toHaveBeenCalledWith(
      "/tiktok/products/search",
      { status: "INACTIVE", page_size: 50, limit: 200, sync_to_db: true },
      { timeout: 120000 },
    );
  });

  it("returns empty array when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });
    const result = await searchProducts();
    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Search failed" });
    await expect(searchProducts()).rejects.toThrow("Search failed");
  });
});

describe("getAllProducts", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns all products on success", async () => {
    const products = [{ product_id: "p2" }];
    mockPost.mockResolvedValue({ success: true, data: products });
    const result = await getAllProducts();
    expect(result).toEqual(products);
    expect(mockPost).toHaveBeenCalledWith("/tiktok/products/get-all", {
      status: "ACTIVATE",
      page_size: 100,
      sync_to_db: true,
    });
  });

  it("returns empty array when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });
    const result = await getAllProducts();
    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Get all failed" });
    await expect(getAllProducts()).rejects.toThrow("Get all failed");
  });
});

describe("getProductDetail", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product detail on success", async () => {
    const detail = { product_id: "p3", title: "Detail" };
    mockGet.mockResolvedValue({ success: true, data: detail });
    const result = await getProductDetail("p3");
    expect(result).toEqual(detail);
    expect(mockGet).toHaveBeenCalledWith("/tiktok/products/p3", {
      params: { sync_to_db: true },
    });
  });

  it("returns null when data is absent", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    const result = await getProductDetail("p3");
    expect(result).toBeNull();
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Detail error" });
    await expect(getProductDetail("p3")).rejects.toThrow("Detail error");
  });
});

describe("getProductListFromDB", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product list from DB", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ product_id: "p4" }] });
    const result = await getProductListFromDB();
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/products/list");
  });

  it("returns empty array when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    const result = await getProductListFromDB();
    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "DB list error" });
    await expect(getProductListFromDB()).rejects.toThrow("DB list error");
  });
});

describe("getMasterProductsFromDB", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns master products", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ product_id: "m1" }] });
    const result = await getMasterProductsFromDB();
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/products/master");
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Master error" });
    await expect(getMasterProductsFromDB()).rejects.toThrow("Master error");
  });
});

describe("searchProductsInDB", () => {
  beforeEach(() => vi.clearAllMocks());

  it("searches with default field", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ product_id: "s1" }] });
    const result = await searchProductsInDB("test query");
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/search", {
      params: { q: "test query", field: "name" },
    });
  });

  it("searches with custom field", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });
    await searchProductsInDB("sku-001", "sku");
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/search", {
      params: { q: "sku-001", field: "sku" },
    });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Search DB error" });
    await expect(searchProductsInDB("query")).rejects.toThrow(
      "Search DB error",
    );
  });
});

describe("getProductsByStatus", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns products by status", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ product_id: "s2" }] });
    const result = await getProductsByStatus("ACTIVATE");
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/products/status/ACTIVATE");
  });

  it("returns empty array when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    const result = await getProductsByStatus("DEACTIVATE");
    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Status error" });
    await expect(getProductsByStatus("UNKNOWN")).rejects.toThrow(
      "Status error",
    );
  });
});

describe("getProductById", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product by ID", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { product_id: "prod-123" },
    });
    const result = await getProductById("prod-123");
    expect(result).toEqual({ product_id: "prod-123" });
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/products/prod-123");
  });

  it("returns null when data is absent", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    const result = await getProductById("prod-456");
    expect(result).toBeNull();
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "ID error" });
    await expect(getProductById("prod-789")).rejects.toThrow("ID error");
  });
});

describe("getStatistics", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns statistics on success", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { total_products: 200, active_products: 150 },
    });
    const result = await getStatistics();
    expect(result).toEqual({ total_products: 200, active_products: 150 });
    expect(mockGet).toHaveBeenCalledWith("/tiktok/db/statistics");
  });

  it("returns empty object when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    const result = await getStatistics();
    expect(result).toEqual({});
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Stats error" });
    await expect(getStatistics()).rejects.toThrow("Stats error");
  });
});

describe("deleteProduct", () => {
  beforeEach(() => vi.clearAllMocks());

  it("deletes product successfully (returns void)", async () => {
    mockDelete.mockResolvedValue({ success: true });
    const result = await deleteProduct("prod-del");
    expect(result).toBeUndefined();
    expect(mockDelete).toHaveBeenCalledWith("/tiktok/db/products/prod-del");
  });

  it("throws on failure", async () => {
    mockDelete.mockResolvedValue({ success: false, error: "Delete failed" });
    await expect(deleteProduct("prod-fail")).rejects.toThrow("Delete failed");
  });
});

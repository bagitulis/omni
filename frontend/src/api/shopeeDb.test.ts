import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: { get: mockGet, post: mockPost },
}));

vi.mock("@/lib/constants", () => ({
  API_TIMEOUT: { EXTRA_LONG: 120000 },
}));

import {
  getProductList,
  getProductBaseList,
  getProductModelList,
  getProductBase,
  getProductModels,
  getProductVariations,
  getProductFull,
  searchBySku,
  getProductsByStatus,
  getDatabaseStats,
  getUnprocessedItems,
  getItemsWithoutModels,
  getSyncLogs,
  syncShopeeProducts,
  formatProductForDisplay,
  formatPrice,
  formatNumber,
} from "./shopeeDb";

describe("getProductList", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product list on success", async () => {
    const products = [{ item_id: 1, item_name: "Product A" }];
    mockGet.mockResolvedValue({ success: true, data: products });
    const result = await getProductList();
    expect(result).toEqual(products);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/list", {
      params: { limit: 100, offset: 0 },
    });
  });

  it("passes status param when provided", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });
    await getProductList(50, 10, "ACTIVE");
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/list", {
      params: { limit: 50, offset: 10, status: "ACTIVE" },
    });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "DB error" });
    await expect(getProductList()).rejects.toThrow("DB error");
  });
});

describe("getProductBaseList", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns base product list", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });
    const result = await getProductBaseList();
    expect(result).toEqual([]);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/base", {
      params: { limit: 100, offset: 0 },
    });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not found" });
    await expect(getProductBaseList()).rejects.toThrow("Not found");
  });
});

describe("getProductModelList", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns model list", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ item_id: 2 }] });
    const result = await getProductModelList();
    expect(result).toHaveLength(1);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Failed" });
    await expect(getProductModelList()).rejects.toThrow("Failed");
  });
});

describe("getProductBase", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns single product", async () => {
    mockGet.mockResolvedValue({ success: true, data: { item_id: 42 } });
    const result = await getProductBase(42);
    expect(result).toEqual({ item_id: 42 });
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/base/42");
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not found" });
    await expect(getProductBase(99)).rejects.toThrow("Not found");
  });
});

describe("getProductModels", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product models", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ item_id: 5 }] });
    const result = await getProductModels(5);
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/models/5", { params: undefined });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Models error" });
    await expect(getProductModels(5)).rejects.toThrow("Models error");
  });
});

describe("getProductVariations", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns product variations", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });
    const result = await getProductVariations(10);
    expect(result).toEqual([]);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/variations/10", { params: undefined });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Var error" });
    await expect(getProductVariations(10)).rejects.toThrow("Var error");
  });
});

describe("getProductFull", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns full product details", async () => {
    mockGet.mockResolvedValue({ success: true, data: { item_id: 7 } });
    const result = await getProductFull(7);
    expect(result).toEqual({ item_id: 7 });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Full error" });
    await expect(getProductFull(7)).rejects.toThrow("Full error");
  });
});

describe("searchBySku", () => {
  beforeEach(() => vi.clearAllMocks());

  it("throws when SKU is empty string", async () => {
    await expect(searchBySku("")).rejects.toThrow("SKU cannot be empty");
  });

  it("throws when SKU is whitespace only", async () => {
    await expect(searchBySku("   ")).rejects.toThrow("SKU cannot be empty");
  });

  it("searches with trimmed SKU", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ item_id: 3 }] });
    const result = await searchBySku("  SKU-001  ");
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/search", {
      params: { sku: "SKU-001" },
    });
  });

  it("throws on API failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Search failed" });
    await expect(searchBySku("SKU-X")).rejects.toThrow("Search failed");
  });
});

describe("getProductsByStatus", () => {
  beforeEach(() => vi.clearAllMocks());

  it("throws when status is empty string", async () => {
    await expect(getProductsByStatus("")).rejects.toThrow("Status is required");
  });

  it("returns products by status", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ item_id: 9 }] });
    const result = await getProductsByStatus("ACTIVE");
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/products/status/ACTIVE", { params: undefined });
  });

  it("throws on API failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Status error" });
    await expect(getProductsByStatus("INACTIVE")).rejects.toThrow(
      "Status error",
    );
  });
});

describe("getDatabaseStats", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns stats on success", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { total_products: 100 },
    });
    const result = await getDatabaseStats();
    expect(result).toEqual({ total_products: 100 });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Stats error" });
    await expect(getDatabaseStats()).rejects.toThrow("Stats error");
  });
});

describe("getUnprocessedItems", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns unprocessed items", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });
    const result = await getUnprocessedItems();
    expect(result).toEqual([]);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/sync/unprocessed", {
      params: undefined,
    });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Unprocessed error" });
    await expect(getUnprocessedItems()).rejects.toThrow("Unprocessed error");
  });
});

describe("getItemsWithoutModels", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns items without models", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ item_id: 11 }] });
    const result = await getItemsWithoutModels();
    expect(result).toHaveLength(1);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "No models error" });
    await expect(getItemsWithoutModels()).rejects.toThrow("No models error");
  });
});

describe("getSyncLogs", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns sync logs", async () => {
    mockGet.mockResolvedValue({ success: true, data: [{ id: 1 }] });
    const result = await getSyncLogs(50);
    expect(result).toHaveLength(1);
    expect(mockGet).toHaveBeenCalledWith("/shopee/db/sync/logs", {
      params: { limit: 50 },
    });
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Logs error" });
    await expect(getSyncLogs()).rejects.toThrow("Logs error");
  });
});

describe("syncShopeeProducts", () => {
  beforeEach(() => vi.clearAllMocks());

  it("triggers product sync successfully", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { message: "Synced", processed: 50 },
    });
    const result = await syncShopeeProducts();
    expect(result).toEqual({ message: "Synced", processed: 50 });
    expect(mockPost).toHaveBeenCalledWith(
      "/shopee/sync/products",
      undefined,
      expect.objectContaining({ timeout: 120000 }),
    );
  });

  it("throws on sync failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Sync failed" });
    await expect(syncShopeeProducts()).rejects.toThrow("Sync failed");
  });
});

describe("formatProductForDisplay", () => {
  it("formats product with all fields present", () => {
    const result = formatProductForDisplay({
      item_id: 1,
      item_name: "Test Product",
      item_sku: "SKU-001",
      current_price: 25000,
      original_price: 30000,
      seller_stock: 10,
      item_status: "ACTIVE",
      has_model: true,
      category_id: 5,
      brand: "Brand A",
      description: "A great product",
      image_urls: ["img1.jpg", "img2.jpg"],
      updated_at: "2024-01-01T00:00:00Z",
    });
    expect(result.item_id).toBe(1);
    expect(result.name).toBe("Test Product");
    expect(result.sku).toBe("SKU-001");
    expect(result.price).toBe(25000);
    expect(result.stock).toBe(10);
    expect(result.images).toEqual(["img1.jpg", "img2.jpg"]);
  });

  it("uses price fallback when current_price is absent", () => {
    const result = formatProductForDisplay({
      item_id: 2,
      price: 15000,
    });
    expect(result.price).toBe(15000);
  });

  it("uses stock fallback when seller_stock is absent", () => {
    const result = formatProductForDisplay({
      item_id: 3,
      stock: 5,
    });
    expect(result.stock).toBe(5);
  });

  it("parses comma-separated image_urls string", () => {
    const result = formatProductForDisplay({
      item_id: 4,
      image_urls: "img1.jpg,img2.jpg,img3.jpg",
    });
    expect(result.images).toEqual(["img1.jpg", "img2.jpg", "img3.jpg"]);
  });

  it("returns empty images for empty image_urls", () => {
    const result = formatProductForDisplay({ item_id: 5, image_urls: "" });
    expect(result.images).toEqual([]);
  });

  it("uses default values for missing optional fields", () => {
    const result = formatProductForDisplay({ item_id: 6 });
    expect(result.name).toBe("N/A");
    expect(result.sku).toBe("N/A");
    expect(result.price).toBe(0);
    expect(result.stock).toBe(0);
    expect(result.status).toBe("UNKNOWN");
    expect(result.brand).toBe("N/A");
    expect(result.description).toBe("");
  });
});

describe("formatPrice", () => {
  it("formats price in IDR currency", () => {
    const result = formatPrice(50000);
    expect(result).toContain("50.000");
  });

  it("formats zero as zero IDR", () => {
    const result = formatPrice(0);
    expect(result).toContain("0");
  });
});

describe("formatNumber", () => {
  it("formats number with id-ID locale", () => {
    const result = formatNumber(1000000);
    expect(result).toContain("1.000.000");
  });

  it("formats zero", () => {
    const result = formatNumber(0);
    expect(result).toBe("0");
  });
});

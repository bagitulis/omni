import { describe, it, expect, vi, beforeEach } from "vitest";
import { updatePrice, updatePriceBatch } from "./pricing";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  default: {
    get: vi.fn(),
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

const mockBatchResult = {
  total: 1,
  success: 1,
  failed: 0,
  results: [
    {
      sku: "SKU-001",
      price: 50000,
      success: true,
      platforms: { shopee: { success: true, item_id: "123" } },
    },
  ],
};

describe("updatePriceBatch", () => {
  it("returns batch result on success", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockBatchResult });

    const result = await updatePriceBatch([{ sku: "SKU-001", price: 50000 }]);

    expect(result).toEqual(mockBatchResult);
  });

  it("posts to correct endpoint with items array", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockBatchResult });

    const items = [
      { sku: "SKU-001", price: 50000 },
      { sku: "SKU-002", price: 30000 },
    ];
    await updatePriceBatch(items);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-price-batch", {
      items,
    });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Price update failed",
    });

    await expect(
      updatePriceBatch([{ sku: "SKU-001", price: 50000 }]),
    ).rejects.toThrow("Price update failed");
  });

  it("throws when response.data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });

    await expect(
      updatePriceBatch([{ sku: "SKU-001", price: 50000 }]),
    ).rejects.toThrow("Failed to update prices");
  });

  it("throws default message when error is not provided", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(
      updatePriceBatch([{ sku: "SKU-001", price: 50000 }]),
    ).rejects.toThrow("Failed to update prices");
  });
});

describe("updatePrice", () => {
  it("delegates to updatePriceBatch with single item", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockBatchResult });

    const item = { sku: "SKU-001", price: 50000 };
    const result = await updatePrice(item);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-price-batch", {
      items: [item],
    });
    expect(result).toEqual(mockBatchResult);
  });

  it("passes platform restriction through to batch", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockBatchResult });

    const item = { sku: "SKU-001", price: 50000, platforms: ["tiktok"] };
    await updatePrice(item);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-price-batch", {
      items: [item],
    });
  });

  it("throws when underlying batch call fails", async () => {
    mockPost.mockResolvedValue({ success: false, error: "API error" });

    await expect(updatePrice({ sku: "SKU-001", price: 50000 })).rejects.toThrow(
      "API error",
    );
  });
});

import { beforeEach, describe, expect, it, vi } from "vitest";
import { autoMapSkus, refreshProductImages, syncProduct } from "./products";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    post: mockPost,
  },
}));

describe("products api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("syncProduct sends target_platform and returns response data", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { skus_synced: 3, target_platform: "shopee" },
    });

    const result = await syncProduct(12, "shopee");

    expect(mockPost).toHaveBeenCalledWith("/master-products/12/sync", {
      target_platform: "shopee",
    });
    expect(result).toEqual({ skus_synced: 3, target_platform: "shopee" });
  });

  it("syncProduct throws raw backend error when success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "shopee API error [code=E1001]: invalid access token",
    });

    await expect(syncProduct(12, "shopee")).rejects.toThrow(
      "shopee API error [code=E1001]: invalid access token",
    );
  });

  it("refreshProductImages posts force flag and returns data", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        master_product_id: 12,
        previous_image_count: 0,
        image_count: 2,
        updated: true,
        forced: true,
        images: ["/uploads/a.webp", "/uploads/b.webp"],
      },
    });

    const result = await refreshProductImages(12, true);

    expect(mockPost).toHaveBeenCalledWith(
      "/master-products/12/images/refresh",
      {
        force: true,
      },
    );
    expect(result.image_count).toBe(2);
    expect(result.images).toHaveLength(2);
  });

  it("autoMapSkus posts to auto-link with skus payload", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        mapped_count: 1,
        skipped_count: 1,
        mappings: [
          {
            master_sku_id: 101,
            seller_sku: "SKU-001",
            platform: "shopee",
            platform_item_id: "123456",
            platform_sku_id: "98765",
            sync_status: "synced",
          },
        ],
        errors: ["no platform match found for SKU SKU-002"],
      },
    });

    const result = await autoMapSkus(["SKU-001", "SKU-002"]);

    expect(mockPost).toHaveBeenCalledWith(
      "/master-products/mapping/auto-link",
      {
        skus: ["SKU-001", "SKU-002"],
      },
    );
    expect(result.mapped_count).toBe(1);
    expect(result.skipped_count).toBe(1);
  });
});

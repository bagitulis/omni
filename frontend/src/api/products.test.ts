import { beforeEach, describe, expect, it, vi } from "vitest";
import { refreshProductImages, syncProduct } from "./products";

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
});

import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: vi.fn(),
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

import { batchCheckSku, checkPlatformStatus } from "./inventoryCheck";

describe("inventoryCheck", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("batchCheckSku", () => {
    it("returns results array on success", async () => {
      const mockResults = [
        { sku: "SKU-001", status: "active" },
        { sku: "SKU-002", status: "inactive" },
      ];
      mockPost.mockResolvedValue({
        success: true,
        data: { total: 2, checked: 2, results: mockResults },
      });

      const result = await batchCheckSku(["SKU-001", "SKU-002"]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/batch-check-sku", {
        skus: ["SKU-001", "SKU-002"],
      });
      expect(result).toEqual(mockResults);
    });

    it("returns empty array when results is missing", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { total: 0, checked: 0 },
      });

      const result = await batchCheckSku(["SKU-001"]);
      expect(result).toEqual([]);
    });

    it("throws with API error message on failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "SKU check service unavailable",
      });

      await expect(batchCheckSku(["SKU-001"])).rejects.toThrow(
        "SKU check service unavailable",
      );
    });

    it("throws generic error when no error message provided", async () => {
      mockPost.mockResolvedValue({ success: false });

      await expect(batchCheckSku(["SKU-001"])).rejects.toThrow(
        "Failed to check SKU status",
      );
    });

    it("handles empty skus array", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { total: 0, checked: 0, results: [] },
      });

      const result = await batchCheckSku([]);
      expect(result).toEqual([]);
    });

    it("propagates network errors", async () => {
      mockPost.mockRejectedValue(new Error("Network error"));

      await expect(batchCheckSku(["SKU-001"])).rejects.toThrow("Network error");
    });
  });

  describe("checkPlatformStatus", () => {
    it("delegates to batchCheckSku and returns same results", async () => {
      const mockResults = [{ sku: "SKU-001", status: "active" }];
      mockPost.mockResolvedValue({
        success: true,
        data: { total: 1, checked: 1, results: mockResults },
      });

      const result = await checkPlatformStatus(["SKU-001"]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/batch-check-sku", {
        skus: ["SKU-001"],
      });
      expect(result).toEqual(mockResults);
    });

    it("throws when batchCheckSku fails", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Platform check failed",
      });

      await expect(checkPlatformStatus(["SKU-001"])).rejects.toThrow(
        "Platform check failed",
      );
    });
  });
});

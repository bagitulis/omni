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

import { updatePrice, updatePriceBatch } from "./inventoryPrice";

describe("inventoryPrice", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("updatePrice", () => {
    it("updates price for a single SKU", async () => {
      const mockResult = { sku: "SKU-001", success: true, price: 15000 };
      mockPost.mockResolvedValue({ success: true, data: mockResult });

      const result = await updatePrice("SKU-001", 15000);

      expect(mockPost).toHaveBeenCalledWith("/inventory/update-price", {
        sku: "SKU-001",
        price: 15000,
        platforms: undefined,
      });
      expect(result).toEqual(mockResult);
    });

    it("includes platforms when provided", async () => {
      const mockResult = { sku: "SKU-001", success: true, price: 15000 };
      mockPost.mockResolvedValue({ success: true, data: mockResult });

      await updatePrice("SKU-001", 15000, ["shopee", "lazada"]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/update-price", {
        sku: "SKU-001",
        price: 15000,
        platforms: ["shopee", "lazada"],
      });
    });

    it("throws on API failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Price update failed",
      });

      await expect(updatePrice("SKU-001", 15000)).rejects.toThrow(
        "Price update failed",
      );
    });

    it("throws when data is null despite success", async () => {
      mockPost.mockResolvedValue({ success: true, data: null });

      await expect(updatePrice("SKU-001", 15000)).rejects.toThrow(
        "Updated price response is empty",
      );
    });
  });

  describe("updatePriceBatch", () => {
    it("handles legacy array response format", async () => {
      const legacyPayload = [
        { sku: "SKU-001", success: true },
        { sku: "SKU-002", success: true },
        { sku: "SKU-003", success: false },
      ];
      mockPost.mockResolvedValue({ success: true, data: legacyPayload });

      const result = await updatePriceBatch([
        { sku: "SKU-001", price: 10000 },
        { sku: "SKU-002", price: 12000 },
        { sku: "SKU-003", price: 9000 },
      ]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/update-price-batch", {
        items: [
          { sku: "SKU-001", price: 10000 },
          { sku: "SKU-002", price: 12000 },
          { sku: "SKU-003", price: 9000 },
        ],
      });
      expect(result.total).toBe(3);
      expect(result.successful).toBe(2);
      expect(result.failed).toBe(1);
      expect(result.skipped).toBe(0);
      expect(result.results).toEqual(legacyPayload);
    });

    it("handles object response format with results array", async () => {
      const mockPayload = {
        total: 2,
        successful: 2,
        failed: 0,
        skipped: 0,
        results: [
          { sku: "SKU-001", success: true },
          { sku: "SKU-002", success: true },
        ],
      };
      mockPost.mockResolvedValue({ success: true, data: mockPayload });

      const result = await updatePriceBatch([
        { sku: "SKU-001", price: 10000 },
        { sku: "SKU-002", price: 12000 },
      ]);

      expect(result.total).toBe(2);
      expect(result.successful).toBe(2);
      expect(result.failed).toBe(0);
    });

    it("handles object response with data array (alternate field)", async () => {
      const mockPayload = {
        total: 1,
        data: [{ sku: "SKU-001", success: true }],
      };
      mockPost.mockResolvedValue({ success: true, data: mockPayload });

      const result = await updatePriceBatch([{ sku: "SKU-001", price: 10000 }]);

      expect(result.results).toHaveLength(1);
      expect(result.results[0]).toEqual({ sku: "SKU-001", success: true });
    });

    it("computes successful/failed from results when not in response", async () => {
      const mockPayload = {
        results: [
          { sku: "SKU-001", success: true },
          { sku: "SKU-002", success: false },
        ],
      };
      mockPost.mockResolvedValue({ success: true, data: mockPayload });

      const result = await updatePriceBatch([
        { sku: "SKU-001", price: 10000 },
        { sku: "SKU-002", price: 12000 },
      ]);

      expect(result.successful).toBe(1);
      expect(result.failed).toBe(1);
    });

    it("throws when API returns failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Batch price update failed",
      });

      await expect(
        updatePriceBatch([{ sku: "SKU-001", price: 10000 }]),
      ).rejects.toThrow("Batch price update failed");
    });

    it("handles empty items array", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { total: 0, successful: 0, failed: 0, skipped: 0, results: [] },
      });

      const result = await updatePriceBatch([]);
      expect(result.total).toBe(0);
      expect(result.results).toEqual([]);
    });
  });
});

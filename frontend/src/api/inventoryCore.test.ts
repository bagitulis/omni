import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockGet, mockPut } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPut: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: vi.fn(),
    put: mockPut,
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

import {
  throwIfFailed,
  getInventory,
  getInventoryBySku,
  updateInventoryRecord,
  getInventoryStats,
  getSyncHistory,
} from "./inventoryCore";

describe("inventoryCore", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("throwIfFailed", () => {
    it("does nothing when response is successful", () => {
      expect(() =>
        throwIfFailed({ success: true, data: {} }, "Should not throw"),
      ).not.toThrow();
    });

    it("throws with API error message when failed", () => {
      expect(() =>
        throwIfFailed({ success: false, error: "API Error" }, "Fallback"),
      ).toThrow("API Error");
    });

    it("throws with fallback message when no error in response", () => {
      expect(() =>
        throwIfFailed({ success: false }, "Fallback message"),
      ).toThrow("Fallback message");
    });
  });

  describe("getInventory", () => {
    it("returns inventory list result on success", async () => {
      const mockRecords = [
        { sku: "SKU-001", stock: 10 },
        { sku: "SKU-002", stock: 5 },
      ];
      mockGet.mockResolvedValue({
        success: true,
        data: mockRecords,
        total: 2,
        offset: 0,
        limit: 100,
      });

      const result = await getInventory({ offset: 0, limit: 100 });

      expect(mockGet).toHaveBeenCalledWith("/inventory/list", {
        params: { offset: 0, limit: 100 },
      });
      expect(result.records).toEqual(mockRecords);
      expect(result.total).toBe(2);
      expect(result.offset).toBe(0);
      expect(result.limit).toBe(100);
    });

    it("uses default values when response fields are missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: [] });

      const result = await getInventory();
      expect(result.records).toEqual([]);
      expect(result.total).toBe(0);
      expect(result.offset).toBe(0);
      expect(result.limit).toBe(100);
    });

    it("throws on API failure", async () => {
      mockGet.mockResolvedValue({
        success: false,
        error: "Failed to list inventory",
      });

      await expect(getInventory()).rejects.toThrow("Failed to list inventory");
    });

    it("calls without params when not provided", async () => {
      mockGet.mockResolvedValue({ success: true, data: [] });

      await getInventory();
      expect(mockGet).toHaveBeenCalledWith("/inventory/list", {
        params: undefined,
      });
    });
  });

  describe("getInventoryBySku", () => {
    it("returns inventory record by SKU", async () => {
      const mockRecord = { sku: "SKU-001", stock: 10 };
      mockGet.mockResolvedValue({ success: true, data: mockRecord });

      const result = await getInventoryBySku("SKU-001");

      expect(mockGet).toHaveBeenCalledWith("/inventory/SKU-001");
      expect(result).toEqual(mockRecord);
    });

    it("throws when response failed", async () => {
      mockGet.mockResolvedValue({
        success: false,
        error: "SKU not found",
      });

      await expect(getInventoryBySku("MISSING")).rejects.toThrow(
        "SKU not found",
      );
    });

    it("throws when data is empty despite success", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });

      await expect(getInventoryBySku("SKU-001")).rejects.toThrow(
        "Inventory item response is empty",
      );
    });
  });

  describe("updateInventoryRecord", () => {
    it("updates and returns inventory record", async () => {
      const mockRecord = { sku: "SKU-001", stock: 20 };
      mockPut.mockResolvedValue({ success: true, data: mockRecord });

      const result = await updateInventoryRecord("SKU-001", { stock: 20 });

      expect(mockPut).toHaveBeenCalledWith("/inventory/SKU-001", { stock: 20 });
      expect(result).toEqual(mockRecord);
    });

    it("throws on update failure", async () => {
      mockPut.mockResolvedValue({
        success: false,
        error: "Update failed",
      });

      await expect(
        updateInventoryRecord("SKU-001", { stock: 20 }),
      ).rejects.toThrow("Update failed");
    });

    it("throws when data is null despite success", async () => {
      mockPut.mockResolvedValue({ success: true, data: null });

      await expect(
        updateInventoryRecord("SKU-001", { stock: 20 }),
      ).rejects.toThrow("Updated inventory record response is empty");
    });
  });

  describe("getInventoryStats", () => {
    it("returns inventory stats on success", async () => {
      const mockStats = { total: 100, low_stock: 5, out_of_stock: 2 };
      mockGet.mockResolvedValue({ success: true, data: mockStats });

      const result = await getInventoryStats();

      expect(mockGet).toHaveBeenCalledWith("/inventory/stats");
      expect(result).toEqual(mockStats);
    });

    it("throws on API failure", async () => {
      mockGet.mockResolvedValue({
        success: false,
        error: "Stats unavailable",
      });

      await expect(getInventoryStats()).rejects.toThrow("Stats unavailable");
    });

    it("throws when data is null despite success", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });

      await expect(getInventoryStats()).rejects.toThrow(
        "Inventory stats response is empty",
      );
    });
  });

  describe("getSyncHistory", () => {
    it("returns sync history array on success", async () => {
      const mockHistory = [
        { id: "1", status: "success", created_at: "2024-01-01" },
      ];
      mockGet.mockResolvedValue({ success: true, data: mockHistory });

      const result = await getSyncHistory();

      expect(mockGet).toHaveBeenCalledWith("/inventory/sync/history");
      expect(result).toEqual(mockHistory);
    });

    it("returns empty array when data is null", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });

      const result = await getSyncHistory();
      expect(result).toEqual([]);
    });

    it("throws on API failure", async () => {
      mockGet.mockResolvedValue({
        success: false,
        error: "History fetch failed",
      });

      await expect(getSyncHistory()).rejects.toThrow("History fetch failed");
    });
  });
});

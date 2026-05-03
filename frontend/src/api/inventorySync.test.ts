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

import {
  updateStock,
  updateStockBatch,
  syncInventory,
  syncToSheets,
} from "./inventorySync";

describe("inventorySync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("updateStock", () => {
    it("updates stock for a single SKU successfully", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { success: true },
      });

      await expect(
        updateStock("SKU-001", ["shopee"], 10),
      ).resolves.toBeUndefined();

      expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock", {
        sku: "SKU-001",
        platforms: ["shopee"],
        stock: 10,
      });
    });

    it("omits stock from payload when not provided", async () => {
      mockPost.mockResolvedValue({ success: true, data: { success: true } });

      await updateStock("SKU-001", ["shopee"]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock", {
        sku: "SKU-001",
        platforms: ["shopee"],
      });
    });

    it("throws on API response failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Stock update failed",
      });

      await expect(updateStock("SKU-001")).rejects.toThrow(
        "Stock update failed",
      );
    });

    it("throws when syncResult.success is false", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { success: false, error: "Platform sync error" },
      });

      await expect(updateStock("SKU-001")).rejects.toThrow(
        "Platform sync error",
      );
    });

    it("throws generic error when syncResult fails without message", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: { success: false },
      });

      await expect(updateStock("SKU-001")).rejects.toThrow("Stock sync failed");
    });

    it("collects errors from platforms field", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: {
          success: false,
          platforms: {
            shopee: { success: false, error: "Shopee API error" },
            lazada: { success: true },
          },
        },
      });

      await expect(updateStock("SKU-001")).rejects.toThrow("Shopee API error");
    });
  });

  describe("updateStockBatch", () => {
    it("handles legacy string array input", async () => {
      mockPost.mockResolvedValue({ success: true, data: [] });

      const result = await updateStockBatch(["SKU-001", "SKU-002"], ["shopee"]);

      expect(result).toEqual({
        total: 0,
        succeeded: 0,
        failed: 0,
        results: [],
      });
      expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock-batch", {
        skus: ["SKU-001", "SKU-002"],
        platforms: ["shopee"],
        items: [
          { sku: "SKU-001", platforms: ["shopee"] },
          { sku: "SKU-002", platforms: ["shopee"] },
        ],
      });
    });

    it("handles object array input", async () => {
      mockPost.mockResolvedValue({ success: true, data: [] });

      await updateStockBatch([
        { sku: "SKU-001", stock: 10, platforms: ["shopee"] },
        { sku: "SKU-002", stock: 5 },
      ]);

      const callArg = mockPost.mock.calls[0][1];
      expect(callArg.items).toEqual([
        { sku: "SKU-001", stock: 10, platforms: ["shopee"] },
        { sku: "SKU-002", stock: 5 },
      ]);
    });

    it("throws on API failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Batch update failed",
      });

      await expect(updateStockBatch(["SKU-001"])).rejects.toThrow(
        "Batch update failed",
      );
    });

    it("returns structured result with failures (array response)", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: [
          { success: true },
          { success: false, error: "SKU-002 sync failed" },
        ],
      });

      const result = await updateStockBatch(["SKU-001", "SKU-002"]);
      expect(result.total).toBe(2);
      expect(result.succeeded).toBe(1);
      expect(result.failed).toBe(1);
      expect(result.results[1].error).toBe("SKU-002 sync failed");
    });

    it("returns structured result from object response with failures", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: {
          results: [{ success: false, error: "sync error" }, { success: true }],
        },
      });

      const result = await updateStockBatch(["SKU-001", "SKU-002"]);
      expect(result.total).toBe(2);
      expect(result.succeeded).toBe(1);
      expect(result.failed).toBe(1);
      expect(result.results[0].error).toBe("sync error");
    });

    it("handles response with data array (all success)", async () => {
      mockPost.mockResolvedValue({
        success: true,
        data: {
          data: [{ success: true }, { success: true }],
        },
      });

      const result = await updateStockBatch(["SKU-001", "SKU-002"]);
      expect(result.total).toBe(2);
      expect(result.succeeded).toBe(2);
      expect(result.failed).toBe(0);
    });
  });

  describe("syncInventory", () => {
    it("syncs from sheets successfully", async () => {
      mockPost.mockResolvedValue({ success: true, data: { status: "success", total_records: 5 } });

      const result = await syncInventory("sheet-id", "Sheet1");
      expect(result.status).toBe("success");
      expect(result.total_records).toBe(5);

      expect(mockPost).toHaveBeenCalledWith("/inventory/sync/from-sheets", {
        spreadsheet_id: "sheet-id",
        sheet_name: "Sheet1",
      });
    });

    it("works without optional parameters", async () => {
      mockPost.mockResolvedValue({ success: true, data: { status: "success" } });

      const result = await syncInventory();
      expect(result.status).toBe("success");

      expect(mockPost).toHaveBeenCalledWith("/inventory/sync/from-sheets", {
        spreadsheet_id: undefined,
        sheet_name: undefined,
      });
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Sync failed",
      });

      await expect(syncInventory()).rejects.toThrow("Sync failed");
    });

    it("throws generic error when no message", async () => {
      mockPost.mockResolvedValue({ success: false });

      await expect(syncInventory()).rejects.toThrow("Failed to sync inventory");
    });
  });

  describe("syncToSheets", () => {
    it("exports to sheets successfully", async () => {
      mockPost.mockResolvedValue({ success: true, data: { status: "success", total_records: 10 } });

      const result = await syncToSheets();
      expect(result.status).toBe("success");
      expect(result.total_records).toBe(10);
      expect(mockPost).toHaveBeenCalledWith("/inventory/sync/to-sheets");
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Export failed",
      });

      await expect(syncToSheets()).rejects.toThrow("Export failed");
    });

    it("throws generic error when no message", async () => {
      mockPost.mockResolvedValue({ success: false });

      await expect(syncToSheets()).rejects.toThrow(
        "Failed to export to sheets",
      );
    });
  });
});

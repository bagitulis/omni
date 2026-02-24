import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockClientGet, mockPost } = vi.hoisted(() => ({
  mockClientGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: vi.fn(),
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: mockClientGet, post: vi.fn() },
  },
}));

import {
  getAvailableColumns,
  getSelectedColumns,
  saveSelectedColumns,
} from "./inventoryColumns";

describe("inventoryColumns", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getAvailableColumns", () => {
    it("returns normalized column names from data.columns", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          columns: [
            { name: "SKU" },
            { label: "Product Name" },
            { key: "stock" },
          ],
        },
      });

      const result = await getAvailableColumns();
      expect(result).toEqual(["SKU", "Product Name", "stock"]);
    });

    it("falls back to data.data when columns is missing", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          data: [{ name: "sku" }, { name: "price" }],
        },
      });

      const result = await getAvailableColumns();
      expect(result).toEqual(["sku", "price"]);
    });

    it("deduplicates column names", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          columns: [{ name: "sku" }, { name: "sku" }, { name: "price" }],
        },
      });

      const result = await getAvailableColumns();
      expect(result).toEqual(["sku", "price"]);
    });

    it("filters out empty column entries", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          columns: [{ name: "sku" }, {}, { name: "" }],
        },
      });

      const result = await getAvailableColumns();
      expect(result).toEqual(["sku"]);
    });

    it("throws when success is false", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: false,
          error: "Unauthorized",
        },
      });

      await expect(getAvailableColumns()).rejects.toThrow("Unauthorized");
    });

    it("throws generic error when no error message", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: false },
      });

      await expect(getAvailableColumns()).rejects.toThrow(
        "Failed to fetch available columns",
      );
    });

    it("calls the correct endpoint", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: true, columns: [] },
      });

      await getAvailableColumns();
      expect(mockClientGet).toHaveBeenCalledWith(
        "/inventory/columns/available",
      );
    });
  });

  describe("getSelectedColumns", () => {
    it("returns selected_columns array on success", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          selected_columns: ["sku", "price", "stock"],
        },
      });

      const result = await getSelectedColumns();
      expect(result).toEqual(["sku", "price", "stock"]);
    });

    it("falls back to data.data array when selected_columns missing", async () => {
      mockClientGet.mockResolvedValue({
        data: {
          success: true,
          data: ["sku", "price"],
        },
      });

      const result = await getSelectedColumns();
      expect(result).toEqual(["sku", "price"]);
    });

    it("returns empty array when neither array is present", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: true },
      });

      const result = await getSelectedColumns();
      expect(result).toEqual([]);
    });

    it("throws when success is false", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: false },
      });

      await expect(getSelectedColumns()).rejects.toThrow(
        "Failed to fetch selected columns",
      );
    });

    it("calls the correct endpoint", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: true, selected_columns: [] },
      });

      await getSelectedColumns();
      expect(mockClientGet).toHaveBeenCalledWith("/inventory/columns/selected");
    });
  });

  describe("saveSelectedColumns", () => {
    it("posts selected columns successfully", async () => {
      mockPost.mockResolvedValue({ success: true });

      await saveSelectedColumns(["sku", "price"]);

      expect(mockPost).toHaveBeenCalledWith("/inventory/columns/selected", {
        selected_columns: ["sku", "price"],
        columns: ["sku", "price"],
      });
    });

    it("throws on API error", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Failed to save",
      });

      await expect(saveSelectedColumns(["sku"])).rejects.toThrow(
        "Failed to save",
      );
    });

    it("throws generic error when no error message", async () => {
      mockPost.mockResolvedValue({ success: false });

      await expect(saveSelectedColumns(["sku"])).rejects.toThrow(
        "Failed to save column selection",
      );
    });
  });
});

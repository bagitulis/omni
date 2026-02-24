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

// inventoryCore is imported inside inventoryConfig — no need to mock it (it uses apiClient too)

import {
  normalizeSelectedColumns,
  normalizeInventoryConfig,
  getInventoryConfig,
  updateInventoryConfig,
} from "./inventoryConfig";

describe("inventoryConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("normalizeSelectedColumns", () => {
    it("serializes array of strings to JSON string", () => {
      const result = normalizeSelectedColumns(["sku", "price", "stock"]);
      expect(result).toBe('["sku","price","stock"]');
    });

    it("filters non-string values from array", () => {
      const result = normalizeSelectedColumns(["sku", 42, null, "price"]);
      expect(result).toBe('["sku","price"]');
    });

    it("returns string as-is if already a string", () => {
      const result = normalizeSelectedColumns('["sku","price"]');
      expect(result).toBe('["sku","price"]');
    });

    it("returns empty string for null", () => {
      const result = normalizeSelectedColumns(null);
      expect(result).toBe("");
    });

    it("returns empty string for undefined", () => {
      const result = normalizeSelectedColumns(undefined);
      expect(result).toBe("");
    });

    it("handles empty array", () => {
      const result = normalizeSelectedColumns([]);
      expect(result).toBe("[]");
    });
  });

  describe("normalizeInventoryConfig", () => {
    it("returns null for null input", () => {
      expect(normalizeInventoryConfig(null)).toBeNull();
    });

    it("returns null for undefined input", () => {
      expect(normalizeInventoryConfig(undefined)).toBeNull();
    });

    it("normalizes full config object with defaults", () => {
      const result = normalizeInventoryConfig({
        id: "cfg-1",
        tenant_id: "tenant-1",
        spreadsheet_id: "sheet-id",
        sheet_name: "Sheet1",
        selected_columns: ["sku", "price"],
        all_columns: "all",
        header_row: 1,
        data_start_row: 2,
        key_column: "sku",
        auto_sync: true,
        sync_interval_seconds: 300,
        last_sync_timestamp: "2024-01-01T00:00:00Z",
        last_headers_hash: "abc123",
        last_sync_status: "success",
        low_stock_threshold: 10,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      });

      expect(result).not.toBeNull();
      expect(result!.id).toBe("cfg-1");
      expect(result!.tenant_id).toBe("tenant-1");
      expect(result!.header_row).toBe(1);
      expect(result!.auto_sync).toBe(true);
      expect(result!.selected_columns).toBe('["sku","price"]');
    });

    it("uses key_column_name as fallback for key_column", () => {
      const result = normalizeInventoryConfig({
        key_column: undefined,
        key_column_name: "product_id",
      });

      expect(result!.key_column).toBe("product_id");
    });

    it("fills in default values for missing fields", () => {
      const result = normalizeInventoryConfig({});

      expect(result!.id).toBe("");
      expect(result!.header_row).toBe(1);
      expect(result!.data_start_row).toBe(2);
      expect(result!.sync_interval_seconds).toBe(300);
      expect(result!.auto_sync).toBe(false);
      expect(result!.low_stock_threshold).toBeUndefined();
    });
  });

  describe("getInventoryConfig", () => {
    it("returns normalized config on success", async () => {
      mockGet.mockResolvedValue({
        success: true,
        data: {
          id: "cfg-1",
          tenant_id: "tenant-1",
          spreadsheet_id: "sheet-id",
          sheet_name: "Sheet1",
        },
      });

      const result = await getInventoryConfig();

      expect(mockGet).toHaveBeenCalledWith("/inventory/config");
      expect(result).not.toBeNull();
      expect(result!.id).toBe("cfg-1");
    });

    it("returns null when data is null", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });

      const result = await getInventoryConfig();
      expect(result).toBeNull();
    });

    it("throws with API error message on failure", async () => {
      mockGet.mockResolvedValue({
        success: false,
        error: "Config not found",
      });

      await expect(getInventoryConfig()).rejects.toThrow("Config not found");
    });

    it("throws generic error when no error message", async () => {
      mockGet.mockResolvedValue({ success: false });

      await expect(getInventoryConfig()).rejects.toThrow(
        "Failed to fetch inventory config",
      );
    });
  });

  describe("updateInventoryConfig", () => {
    it("returns updated config on success", async () => {
      const updatedConfig = {
        id: "cfg-1",
        tenant_id: "tenant-1",
        spreadsheet_id: "new-sheet-id",
        sheet_name: "NewSheet",
        selected_columns: "",
        all_columns: "",
        header_row: 1,
        data_start_row: 2,
        key_column: "",
        auto_sync: false,
        sync_interval_seconds: 300,
        last_sync_timestamp: null,
        last_headers_hash: "",
        last_sync_status: "",
        created_at: "",
        updated_at: "",
      };
      mockPut.mockResolvedValue({ success: true, data: updatedConfig });

      const result = await updateInventoryConfig({
        spreadsheet_id: "new-sheet-id",
      });

      expect(mockPut).toHaveBeenCalledWith("/inventory/config", {
        spreadsheet_id: "new-sheet-id",
      });
      expect(result).toEqual(updatedConfig);
    });

    it("throws when API returns failure", async () => {
      mockPut.mockResolvedValue({
        success: false,
        error: "Validation failed",
      });

      await expect(updateInventoryConfig({})).rejects.toThrow(
        "Validation failed",
      );
    });

    it("throws when data is null despite success", async () => {
      mockPut.mockResolvedValue({ success: true, data: null });

      await expect(updateInventoryConfig({})).rejects.toThrow(
        "Updated inventory config response is empty",
      );
    });
  });
});

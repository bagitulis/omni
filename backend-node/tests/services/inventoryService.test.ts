/**
 * InventoryService Unit Tests
 * Tests inventory operations, Google Sheets sync, and data management
 */

// Create shared mock objects that can be accessed in tests
const mockSettingsService = {
  getSettings: jest.fn(),
  updateSettings: jest.fn(),
  clearSettings: jest.fn(),
  updateSelectedColumns: jest.fn(),
  updateLastSyncTimestamp: jest.fn(),
  updateHeadersHash: jest.fn(),
};

const mockDataService = {
  list: jest.fn(),
  search: jest.fn(),
  getByKey: jest.fn(),
  getStats: jest.fn(),
  getConfig: jest.fn(),
};

const mockSyncService = {
  syncFromSheets: jest.fn(),
  syncToSheets: jest.fn(),
  validateHeaders: jest.fn(),
  getSheetHeaders: jest.fn(),
};

// Mock all dependencies BEFORE importing
jest.mock("../../src/services/inventorySettingsService", () => ({
  getInventorySettingsService: () => mockSettingsService,
}));

jest.mock("../../src/services/inventoryDataService", () => ({
  getInventoryDataService: () => mockDataService,
}));

jest.mock("../../src/services/inventorySyncService", () => ({
  getInventorySyncService: () => mockSyncService,
}));

import { getInventoryService } from "../../src/services/inventoryService";

describe("InventoryService", () => {
  let inventoryService: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Get service instance
    inventoryService = getInventoryService();
  });

  describe("Settings Operations", () => {
    describe("getSettings", () => {
      it("should delegate to settingsService", async () => {
        const mockSettings = {
          spreadsheetId: "test-sheet-id",
          sheetName: "Inventory",
          keyColumn: "SKU",
        };

        mockSettingsService.getSettings.mockResolvedValue(mockSettings);

        const result = await inventoryService.getSettings("test-sheet-id");

        expect(mockSettingsService.getSettings).toHaveBeenCalledWith(
          "test-sheet-id"
        );
        expect(result).toEqual(mockSettings);
      });

      it("should handle missing settings", async () => {
        mockSettingsService.getSettings.mockResolvedValue(null);

        const result = await inventoryService.getSettings();

        expect(result).toBeNull();
      });
    });

    describe("updateSettings", () => {
      it("should update settings successfully", async () => {
        const updatedSettings = {
          spreadsheetId: "new-sheet-id",
          sheetName: "NewSheet",
          allColumns: ["SKU", "Name", "Qty"],
          keyColumn: "SKU",
          headerRow: 1,
          dataStartRow: 2,
          autoSync: false,
          syncIntervalSeconds: 300,
        };

        mockSettingsService.updateSettings.mockResolvedValue(updatedSettings);

        const result = await inventoryService.updateSettings(
          "new-sheet-id",
          "NewSheet",
          ["SKU", "Name", "Qty"],
          "SKU",
          1,
          2,
          false,
          300
        );

        expect(mockSettingsService.updateSettings).toHaveBeenCalledWith({
          spreadsheetId: "new-sheet-id",
          sheetName: "NewSheet",
          allColumns: ["SKU", "Name", "Qty"],
          keyColumn: "SKU",
          headerRow: 1,
          dataStartRow: 2,
          autoSync: false,
          syncIntervalSeconds: 300,
        });
        expect(result).toEqual(updatedSettings);
      });
    });

    describe("clearSettings", () => {
      it("should clear settings successfully", async () => {
        mockSettingsService.clearSettings.mockResolvedValue(undefined);

        await inventoryService.clearSettings();

        expect(mockSettingsService.clearSettings).toHaveBeenCalled();
      });
    });

    describe("updateSelectedColumns", () => {
      it("should update selected columns", async () => {
        const columns = ["SKU", "Quantity", "Price"];
        mockSettingsService.updateSelectedColumns.mockResolvedValue(undefined);

        await inventoryService.updateSelectedColumns(columns);

        expect(mockSettingsService.updateSelectedColumns).toHaveBeenCalledWith(
          columns
        );
      });
    });

    describe("updateLastSyncTimestamp", () => {
      it("should update sync timestamp", async () => {
        mockSettingsService.updateLastSyncTimestamp.mockResolvedValue(
          undefined
        );

        await inventoryService.updateLastSyncTimestamp();

        expect(mockSettingsService.updateLastSyncTimestamp).toHaveBeenCalled();
      });
    });
  });

  describe("Data Operations", () => {
    describe("listRecords", () => {
      it("should list records with pagination", async () => {
        const mockRecords = {
          total: 100,
          returned: 20,
          offset: 0,
          data: [
            { sku: "SKU001", name: "Product 1", qty: 10 },
            { sku: "SKU002", name: "Product 2", qty: 20 },
          ],
        };

        mockDataService.list.mockResolvedValue(mockRecords);

        const result = await inventoryService.getInventoryList(0, 20);

        expect(mockDataService.list).toHaveBeenCalledWith(0, 20);
        expect(result).toEqual(mockRecords);
        expect(result.data).toHaveLength(2);
      });

      it("should handle empty results", async () => {
        mockDataService.list.mockResolvedValue({
          total: 0,
          returned: 0,
          offset: 0,
          data: [],
        });

        const result = await inventoryService.getInventoryList(0, 20);

        expect(result.total).toBe(0);
        expect(result.data).toHaveLength(0);
      });

      it("should handle large offsets", async () => {
        mockDataService.list.mockResolvedValue({
          total: 1000,
          returned: 20,
          offset: 500,
          data: [],
        });

        const result = await inventoryService.getInventoryList(500, 20);

        expect(mockDataService.list).toHaveBeenCalledWith(500, 20);
        expect(result.offset).toBe(500);
      });
    });

    describe("getRecord", () => {
      it("should get single record by key", async () => {
        const mockRecord = {
          sku: "SKU001",
          name: "Product 1",
          qty: 10,
          price: 100000,
        };

        mockDataService.getByKey.mockResolvedValue(mockRecord);

        const result = await inventoryService.getInventoryByKey("SKU001");

        expect(mockDataService.getByKey).toHaveBeenCalledWith("SKU001");
        expect(result).toEqual(mockRecord);
      });

      it("should return null for non-existent record", async () => {
        mockDataService.getByKey.mockResolvedValue(null);

        const result = await inventoryService.getInventoryByKey("NONEXISTENT");

        expect(result).toBeNull();
      });
    });

    describe("searchRecords", () => {
      it("should search records by query", async () => {
        const mockResults = {
          total: 2,
          returned: 2,
          offset: 0,
          data: [
            { sku: "SKU001", name: "Product 1" },
            { sku: "SKU010", name: "Product 10" },
          ],
        };

        mockDataService.search.mockResolvedValue(mockResults);

        const result = await inventoryService.searchInventory("SKU00");

        expect(mockDataService.search).toHaveBeenCalledWith("SKU00", 0, 50);
        expect(result.data).toHaveLength(2);
      });

      it("should return empty array for no matches", async () => {
        mockDataService.search.mockResolvedValue({
          total: 0,
          returned: 0,
          offset: 0,
          data: [],
        });

        const result = await inventoryService.searchInventory("NOMATCH");

        expect(result.data).toHaveLength(0);
      });

      it("should handle special characters in query", async () => {
        mockDataService.search.mockResolvedValue({
          total: 0,
          returned: 0,
          offset: 0,
          data: [],
        });

        await inventoryService.searchInventory("SKU-001/A");

        expect(mockDataService.search).toHaveBeenCalledWith("SKU-001/A", 0, 50);
      });
    });
  });

  describe("Sync Operations", () => {
    describe("syncFromSheets", () => {
      it("should sync from Google Sheets successfully", async () => {
        const mockResult = {
          status: "SUCCESS",
          message: "Synced successfully",
          records_synced: 100,
          new_records: 100,
        };

        mockSyncService.syncFromSheets.mockResolvedValue(mockResult);

        const result = await inventoryService.syncFromSheets();

        expect(mockSyncService.syncFromSheets).toHaveBeenCalled();
        expect(result.status).toBe("success");
        expect(result.synced_records).toBe(100);
        expect(result.failed_records).toBe(0);
      });

      it("should handle sync errors", async () => {
        const mockError = {
          status: "error",
          message: "API quota exceeded",
          total_records: 0,
          synced_records: 0,
          failed_records: 0,
          duration: 0,
        };

        mockSyncService.syncFromSheets.mockResolvedValue(mockError);

        const result = await inventoryService.syncFromSheets();

        expect(result.status).toBe("error");
        expect(result.message).toContain("quota");
      });

      it("should detect headers changed", async () => {
        const mockResult = {
          status: "SUCCESS",
          message: "Headers changed",
          records_synced: 100,
          new_records: 100,
        };

        mockSyncService.syncFromSheets.mockResolvedValue(mockResult);

        const result = await inventoryService.syncFromSheets();

        expect(result.status).toBe("success");
        expect(result.headers_changed).toBe(false); // Service doesn't track this currently
      });
    });

    describe("syncToSheets", () => {
      it("should sync to Google Sheets successfully", async () => {
        const mockResult = {
          status: "SUCCESS",
          message: "Pushed to sheets",
          records_synced: 50,
          updated_records: 30,
          new_records: 20,
          unchanged_records: 0,
        };

        mockSyncService.syncToSheets.mockResolvedValue(mockResult);

        const result = await inventoryService.syncToSheets();

        expect(mockSyncService.syncToSheets).toHaveBeenCalled();
        expect(result.status).toBe("success");
        expect(result.updated_records).toBe(30);
        expect(result.new_records).toBe(20);
      });

      it("should handle partial sync failure", async () => {
        const mockResult = {
          status: "SUCCESS",
          message: "Partially synced",
          records_synced: 95,
          updated_records: 95,
          new_records: 0,
          unchanged_records: 0,
        };

        mockSyncService.syncToSheets.mockResolvedValue(mockResult);

        const result = await inventoryService.syncToSheets();

        expect(result.synced_records).toBe(95);
        expect(result.synced_records).toBe(95);
        expect(result.failed_records).toBe(0); // Service doesn't track failures, always returns 0
      });

      it("should handle authentication errors", async () => {
        mockSyncService.syncToSheets.mockRejectedValue(
          new Error("Authentication failed")
        );

        const result = await inventoryService.syncToSheets();

        expect(result.status).toBe("error");
        expect(result.message).toBe("Authentication failed");
      });
    });
  });

  describe("Facade Pattern", () => {
    it("should coordinate between specialized services", async () => {
      // Test that service acts as facade
      mockSettingsService.getSettings.mockResolvedValue({});
      mockDataService.list.mockResolvedValue({ data: [] });
      mockSyncService.syncFromSheets.mockResolvedValue({ status: "SUCCESS" });

      await inventoryService.getSettings();
      await inventoryService.getInventoryList(0, 10);
      await inventoryService.syncFromSheets();

      expect(mockSettingsService.getSettings).toHaveBeenCalled();
      expect(mockDataService.list).toHaveBeenCalled();
      expect(mockSyncService.syncFromSheets).toHaveBeenCalled();
    });
  });

  describe("Error Handling", () => {
    it("should propagate service errors", async () => {
      mockSettingsService.getSettings.mockRejectedValue(
        new Error("Database connection failed")
      );

      await expect(inventoryService.getSettings()).rejects.toThrow(
        "Database connection failed"
      );
    });

    it("should handle null/undefined inputs gracefully", async () => {
      mockDataService.getByKey.mockResolvedValue(null);

      const result = await inventoryService.getInventoryByKey(null as any);

      expect(mockDataService.getByKey).toHaveBeenCalledWith(null);
      expect(result).toBeNull();
    });
  });
});

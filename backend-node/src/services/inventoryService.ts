import { getInventorySettingsService } from "./inventorySettingsService";
import { getInventoryDataService } from "./inventoryDataService";
import { getInventorySyncService } from "./inventorySyncService";
import { getPrisma } from "./prismaClient";
import { v4 as uuidv4 } from "uuid";
import { tenantContext } from "../utils/tenantContext";

/**
 * Inventory Service - Coordinator/Facade Pattern
 *
 * RESPONSIBILITY: Coordinate between specialized services and maintain backward compatibility
 * - Delegates settings operations to InventorySettingsService
 * - Delegates data queries to InventoryDataService
 * - Delegates sync operations to InventorySyncService
 * - Provides unified interface for routes
 *
 * This follows the FACADE pattern to maintain backward compatibility while
 * improving internal architecture through service separation.
 */

export interface InventorySettingsData {
  spreadsheetId?: string;
  sheetName?: string;
  selectedColumns?: string[];
  allColumns?: string[];
  headerRow?: number;
  dataStartRow?: number;
  keyColumn?: string;
  autoSync?: boolean;
  syncIntervalSeconds?: number;
  lastSyncTimestamp?: Date;
  lastHeadersHash?: string;
  lastSyncStatus?: string;
}

export interface SyncResult {
  status: "success" | "error";
  message: string;
  total_records: number;
  synced_records: number;
  updated_records?: number;
  new_records?: number;
  unchanged_records?: number;
  failed_records: number;
  duration: number;
  headers?: string[];
  headers_changed?: boolean;
}

export interface ListResult {
  total: number;
  returned: number;
  offset: number;
  data: any[];
}

class InventoryService {
  private _settingsService: ReturnType<
    typeof getInventorySettingsService
  > | null = null;
  private _dataService: ReturnType<typeof getInventoryDataService> | null =
    null;
  private _syncService: ReturnType<typeof getInventorySyncService> | null =
    null;

  private get settingsService() {
    if (!this._settingsService)
      this._settingsService = getInventorySettingsService();
    return this._settingsService;
  }

  private get dataService() {
    if (!this._dataService) this._dataService = getInventoryDataService();
    return this._dataService;
  }

  private get syncService() {
    if (!this._syncService) this._syncService = getInventorySyncService();
    return this._syncService;
  }

  // ========== SETTINGS OPERATIONS (Delegated to InventorySettingsService) ==========

  async getSettings(spreadsheetId?: string): Promise<InventorySettingsData> {
    return this.settingsService.getSettings(spreadsheetId);
  }

  async updateSettings(
    spreadsheetId: string,
    sheetName: string,
    allColumns: string[],
    keyColumn: string = "SKU",
    headerRow: number = 1,
    dataStartRow: number = 2,
    autoSync: boolean = false,
    syncIntervalSeconds: number = 300
  ): Promise<InventorySettingsData> {
    return this.settingsService.updateSettings({
      spreadsheetId,
      sheetName,
      allColumns,
      keyColumn,
      headerRow,
      dataStartRow,
      autoSync,
      syncIntervalSeconds,
    });
  }

  async clearSettings(): Promise<void> {
    return this.settingsService.clearSettings();
  }

  async updateSelectedColumns(columns: string[]): Promise<void> {
    return this.settingsService.updateSelectedColumns(columns);
  }

  async updateLastSyncTimestamp(): Promise<void> {
    return this.settingsService.updateLastSyncTimestamp();
  }

  async updateHeadersHash(headersHash: string): Promise<void> {
    return this.settingsService.updateHeadersHash(headersHash);
  }

  // ========== SYNC OPERATIONS (Delegated to InventorySyncService) ==========

  async syncFromSheets(
    spreadsheetId?: string,
    sheetName?: string
  ): Promise<SyncResult> {
    try {
      const result = await this.syncService.syncFromSheets(
        spreadsheetId,
        sheetName
      );

      return {
        status: result.status === "SUCCESS" ? "success" : "error",
        message: result.message,
        total_records: result.records_synced || 0,
        synced_records: result.new_records || 0,
        failed_records: 0,
        duration: 0,
        headers_changed: false,
      };
    } catch (error: any) {
      return {
        status: "error",
        message: error.message,
        total_records: 0,
        synced_records: 0,
        failed_records: 0,
        duration: 0,
      };
    }
  }

  async syncToSheets(
    spreadsheetId?: string,
    sheetName?: string,
    lockedColumns: string[] = []
  ): Promise<SyncResult> {
    try {
      const result = await this.syncService.syncToSheets(
        spreadsheetId,
        sheetName,
        lockedColumns
      );

      return {
        status: result.status === "SUCCESS" ? "success" : "error",
        message: result.message,
        total_records: result.records_synced || 0,
        synced_records: result.records_synced || 0,
        updated_records: result.updated_records || 0,
        new_records: result.new_records || 0,
        unchanged_records: result.unchanged_records || 0,
        failed_records: 0,
        duration: 0,
        headers_changed: false,
      };
    } catch (error: any) {
      return {
        status: "error",
        message: error.message,
        total_records: 0,
        synced_records: 0,
        failed_records: 0,
        duration: 0,
      };
    }
  }

  async validateHeaders(): Promise<any> {
    try {
      const settings = await this.settingsService.getSettings();
      const headers = await this.syncService.getSheetHeaders();

      if (!headers || headers.length === 0) {
        return {
          status: "error",
          is_valid: false,
          message: "No headers found in sheet",
        };
      }

      const validation = await this.syncService.validateHeaders(
        headers,
        settings
      );

      return {
        status: validation.is_valid ? "success" : "error",
        is_valid: validation.is_valid,
        message: validation.message,
        missing_columns: validation.missing_columns,
        extra_columns: validation.extra_columns,
      };
    } catch (error: any) {
      return {
        status: "error",
        is_valid: false,
        message: error.message,
      };
    }
  }

  // ========== DATA OPERATIONS (Delegated to InventoryDataService) ==========

  async getInventoryList(
    offset: number = 0,
    limit: number = 50
  ): Promise<ListResult> {
    return this.dataService.list(offset, limit);
  }

  async searchInventory(
    query: string,
    offset: number = 0,
    limit: number = 50
  ): Promise<ListResult> {
    return this.dataService.search(query, offset, limit);
  }

  async getInventoryByKey(key: string): Promise<any> {
    return this.dataService.getByKey(key);
  }

  async getStats(): Promise<any> {
    return this.dataService.getStats();
  }

  async getConfig(): Promise<any> {
    return this.dataService.getConfig();
  }

  // ========== CRUD OPERATIONS (Keep in facade for consistency) ==========

  async createInventory(
    keyValue: string,
    data: Record<string, any>,
    keyColumnName: string = "SKU"
  ): Promise<any> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const inventory = await prisma.inventoryRecord.create({
        data: {
          id: uuidv4(),
          tenantId,
          keyColumnName,
          keyValue,
          data: JSON.stringify(data),
          createdAt: new Date(),
          updatedAt: new Date(),
        },
      });

      return inventory;
    } catch (error) {
      console.error("❌ Create inventory error:", error);
      throw error;
    }
  }

  async updateInventory(
    keyValueOrId: string,
    data: Record<string, any>
  ): Promise<any> {
    try {
      const prisma = getPrisma();
      const settings = await this.getSettings();

      // Try to find by ID first, then by keyValue
      let record = await prisma.inventoryRecord.findUnique({
        where: { id: keyValueOrId },
      });

      // If not found by ID, try by keyValue
      if (!record && settings.keyColumn) {
        record = await prisma.inventoryRecord.findFirst({
          where: {
            keyValue: keyValueOrId,
            keyColumnName: settings.keyColumn,
          },
        });
      }

      if (!record) {
        throw new Error(`Record not found: ${keyValueOrId}`);
      }

      const inventory = await prisma.inventoryRecord.update({
        where: { id: record.id },
        data: {
          data: JSON.stringify(data),
          updatedAt: new Date(),
        },
      });

      return inventory;
    } catch (error) {
      console.error("❌ Update inventory error:", error);
      throw error;
    }
  }

  async deleteInventory(id: string): Promise<void> {
    try {
      const prisma = getPrisma();

      await prisma.inventoryRecord.delete({
        where: { id },
      });
    } catch (error) {
      console.error("❌ Delete inventory error:", error);
      throw error;
    }
  }

  // ========== HELPER METHODS ==========

  async getSyncHistory(_limit: number = 20): Promise<any[]> {
    try {
      // Placeholder: Sync history would come from sync job tracking
      // This functionality requires proper sync job table implementation
      return [];
    } catch (error) {
      console.error("❌ Get sync history error:", error);
      return [];
    }
  }
}

let instance: InventoryService;

export function getInventoryService(): InventoryService {
  if (!instance) {
    instance = new InventoryService();
  }
  return instance;
}

// Note: Do NOT export singleton instance directly
// Always use getInventoryService() to avoid module-load errors

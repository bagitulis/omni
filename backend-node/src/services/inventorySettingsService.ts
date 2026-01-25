import { getPrisma } from "./prismaClient";
import { InventoryTenantService } from "./inventory/inventoryTenantService";

/**
 * Inventory Settings Service with Multi-Tenant Support
 *
 * SINGLE RESPONSIBILITY: Manage inventory configuration/settings only
 * - Get/update/delete settings (tenant-scoped)
 * - Manage selected columns
 * - Update sync timestamps
 * - All operations automatically filtered by tenantId
 *
 * Does NOT handle:
 * - Data synchronization logic (delegated to InventorySyncService)
 * - Data queries (delegated to InventoryDataService)
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

class InventorySettingsService {
  private tenantService: InventoryTenantService | null = null;

  /**
   * Get tenant service lazily (not in constructor)
   * This avoids calling getPrisma() during module load
   */
  private getTenantService(): InventoryTenantService {
    if (!this.tenantService) {
      this.tenantService = new InventoryTenantService(getPrisma());
    }
    return this.tenantService;
  }

  async getSettings(_spreadsheetId?: string): Promise<InventorySettingsData> {
    try {
      // Get tenant-scoped settings
      const settings = await this.getTenantService().getSettings();

      if (!settings) {
        return {
          spreadsheetId: "",
          sheetName: "",
          selectedColumns: [],
          allColumns: [],
          headerRow: 1,
          dataStartRow: 2,
          keyColumn: "SKU",
          autoSync: false,
          syncIntervalSeconds: 300,
          lastSyncTimestamp: undefined,
          lastHeadersHash: "",
          lastSyncStatus: "pending",
        };
      }

      return {
        spreadsheetId: settings.spreadsheetId || undefined,
        sheetName: settings.sheetName || undefined,
        selectedColumns: settings.selectedColumns
          ? Array.isArray(settings.selectedColumns)
            ? settings.selectedColumns
            : settings.selectedColumns.split(",")
          : [],
        allColumns: settings.allColumns
          ? Array.isArray(settings.allColumns)
            ? settings.allColumns
            : settings.allColumns.split(",")
          : [],
        headerRow: settings.headerRow,
        dataStartRow: settings.dataStartRow,
        keyColumn: settings.keyColumn || undefined,
        autoSync: settings.autoSync,
        syncIntervalSeconds: settings.syncIntervalSeconds,
        lastSyncTimestamp: settings.lastSyncTimestamp || undefined,
        lastHeadersHash: settings.lastHeadersHash || undefined,
        lastSyncStatus: settings.lastSyncStatus || undefined,
      };
    } catch (error) {
      throw error;
    }
  }

  async updateSettings(
    settings: Partial<InventorySettingsData>
  ): Promise<InventorySettingsData> {
    try {
      const updated = await this.getTenantService().upsertSettings({
        spreadsheetId: settings.spreadsheetId || undefined,
        sheetName: settings.sheetName || undefined,
        selectedColumns: settings.selectedColumns?.join(",") || undefined,
        allColumns: settings.allColumns?.join(",") || undefined,
        keyColumn: settings.keyColumn || undefined,
        headerRow: settings.headerRow || undefined,
        dataStartRow: settings.dataStartRow || undefined,
        autoSync: settings.autoSync || undefined,
        syncIntervalSeconds: settings.syncIntervalSeconds || undefined,
      });

      return {
        spreadsheetId: updated.spreadsheetId || undefined,
        sheetName: updated.sheetName || undefined,
        selectedColumns: updated.selectedColumns
          ? updated.selectedColumns.split(",")
          : [],
        allColumns: updated.allColumns ? updated.allColumns.split(",") : [],
        headerRow: updated.headerRow,
        dataStartRow: updated.dataStartRow,
        keyColumn: updated.keyColumn || undefined,
        autoSync: updated.autoSync,
        syncIntervalSeconds: updated.syncIntervalSeconds,
      };
    } catch (error) {
      throw error;
    }
  }

  async clearSettings(): Promise<void> {
    try {
      await this.getTenantService().deleteSettings();
    } catch (error) {
      throw error;
    }
  }

  async updateSelectedColumns(columns: string[]): Promise<void> {
    try {
      await this.getTenantService().updateSettings({
        selectedColumns: columns.join(","),
      });
    } catch (error) {
      throw error;
    }
  }

  async updateLastSyncTimestamp(): Promise<void> {
    try {
      await this.getTenantService().updateSettings({
        lastSyncTimestamp: new Date(),
        lastSyncStatus: "success",
      });
    } catch (error) {
      throw error;
    }
  }

  async updateHeadersHash(hash: string): Promise<void> {
    try {
      await this.getTenantService().updateSettings({
        lastHeadersHash: hash,
      });
    } catch (error) {
      throw error;
    }
  }
}

let instance: InventorySettingsService;

export function getInventorySettingsService(): InventorySettingsService {
  if (!instance) {
    instance = new InventorySettingsService();
  }
  return instance;
}

// Note: Do NOT export singleton instance directly
// Always use getInventorySettingsService() to avoid module-load errors

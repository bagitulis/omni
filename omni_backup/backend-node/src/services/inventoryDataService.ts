import { getPrisma } from "./prismaClient";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

const logger = getLogger("InventoryDataService");

/**
 * Inventory Data Service
 *
 * SINGLE RESPONSIBILITY: Query and retrieve inventory data only
 * - List inventory items with pagination
 * - Search inventory
 * - Get statistics
 * - Get inventory configuration
 *
 * All operations are tenant-scoped using tenantContext
 *
 * Does NOT handle:
 * - Settings management (delegated to InventorySettingsService)
 * - Data synchronization (delegated to InventorySyncService)
 */

export interface ListResult {
  total: number;
  returned: number;
  offset: number;
  data: any[];
}

export interface StatsResult {
  total_records: number;
  total_columns: number;
  columns: any[];
  last_sync: any;
  db_size_kb: number;
}

class InventoryDataService {
  async list(offset: number = 0, limit: number = 50): Promise<ListResult> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const total = await prisma.inventoryRecord.count({
        where: { tenantId },
      });
      const records = await prisma.inventoryRecord.findMany({
        where: { tenantId },
        skip: offset,
        take: limit,
        orderBy: { createdAt: "desc" },
      });

      return {
        total,
        returned: records.length,
        offset: offset + records.length,
        data: records,
      };
    } catch (error) {
      logger.error(`List inventory error: ${error}`);
      throw error;
    }
  }

  async search(
    query: string,
    offset: number = 0,
    limit: number = 50
  ): Promise<ListResult> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const searchQuery = {
        tenantId,
        OR: [
          { keyValue: { contains: query, mode: "insensitive" } },
          { data: { contains: query, mode: "insensitive" } },
        ],
      };

      const total = await prisma.inventoryRecord.count({ where: searchQuery });
      const records = await prisma.inventoryRecord.findMany({
        where: searchQuery,
        skip: offset,
        take: limit,
        orderBy: { createdAt: "desc" },
      });

      return {
        total,
        returned: records.length,
        offset: offset + records.length,
        data: records,
      };
    } catch (error) {
      logger.error(`Search inventory error: ${error}`);
      throw error;
    }
  }

  async getStats(): Promise<StatsResult> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const total = await prisma.inventoryRecord.count({
        where: { tenantId },
      });

      // Get settings for headers (tenant-scoped)
      const settings = await prisma.inventorySettings.findUnique({
        where: { tenantId },
      });
      const allColumns = settings?.allColumns?.split(",") || [];

      const columns = allColumns.map((name: string, index: number) => ({
        column_name: name,
        column_type: "text",
        is_key: name === (settings?.keyColumn || "SKU"),
        spreadsheet_column: String.fromCharCode(65 + index),
      }));

      return {
        total_records: total,
        total_columns: columns.length,
        columns,
        last_sync: {
          sync_timestamp: settings?.lastSyncTimestamp,
          status: settings?.lastSyncStatus,
        },
        db_size_kb: 0,
      };
    } catch (error) {
      logger.error(`Get stats error: ${error}`);
      throw error;
    }
  }

  async getConfig(): Promise<any> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const settings = await prisma.inventorySettings.findUnique({
        where: { tenantId },
      });

      const config = {
        spreadsheet_id: settings?.spreadsheetId || "",
        sheet_name: settings?.sheetName || "",
        selected_columns: settings?.selectedColumns?.split(",") || [],
        key_column: settings?.keyColumn || "SKU",
        header_row: settings?.headerRow || 1,
        data_start_row: settings?.dataStartRow || 2,
        auto_sync: settings?.autoSync || false,
        sync_interval: settings?.syncIntervalSeconds || 300,
      };

      return {
        status: "SUCCESS",
        data: config,
      };
    } catch (error) {
      logger.error(`Get config error: ${error}`);
      throw error;
    }
  }

  async getByKey(keyValue: string): Promise<any> {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const record = await prisma.inventoryRecord.findFirst({
        where: { tenantId, keyValue },
      });

      if (!record) {
        return null;
      }

      // Parse data field if it's JSON string
      let data = record.data;
      if (typeof data === "string") {
        try {
          data = JSON.parse(data);
        } catch {
          // If parsing fails, keep as string
        }
      }

      return {
        ...record,
        data,
      };
    } catch (error) {
      logger.error(`Get by key error for ${keyValue}: ${error}`);
      throw error;
    }
  }
}

let instance: InventoryDataService;

export function getInventoryDataService(): InventoryDataService {
  if (!instance) {
    instance = new InventoryDataService();
  }
  return instance;
}

// Note: Do NOT export singleton instance directly
// Always use getInventoryDataService() to avoid module-load issues

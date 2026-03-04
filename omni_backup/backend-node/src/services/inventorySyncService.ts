import { getPrisma } from "./prismaClient";
import { getGoogleSheetsService } from "./googleSheetsService";
import { getInventorySettingsService } from "./inventorySettingsService";
// import { InventorySyncStrategy } from "./inventorySyncStrategy";
import {
  InventoryValidationService,
  HeaderValidationResult,
} from "./inventoryValidationService";
// import { InventoryBatchUpdateService } from "./inventoryBatchUpdateService";
import {
  getSnapshotService,
  getDeltaSyncExporter,
  ColumnConfig,
} from "./deltasync";
import { tenantContext } from "../utils/tenantContext";
import { CheckboxDataConverter } from "../utils/CheckboxDataConverter";

/**
 * Inventory Sync Service
 *
 * SINGLE RESPONSIBILITY: Orchestrate sync between database and Google Sheets
 * Delegates to:
 * - InventorySyncStrategy: Smart sync logic (legacy)
 * - InventoryBatchUpdateService: Batch update execution with error handling
 * - InventoryValidationService: Header validation
 * - DeltaSyncExporter: Cell-level delta sync (NEW - quota optimized)
 * - SnapshotService: Sheet state snapshots for diff comparison
 */

export interface SyncResult {
  status: "SUCCESS" | "FAILED";
  message: string;
  records_synced?: number;
  updated_records?: number;
  unchanged_records?: number;
  new_records?: number;
  skipped_records?: number;
  errors?: Array<{ rowIndex: number; error: string }>;
  timestamp: string;
}

export { HeaderValidationResult };

class InventorySyncService {
  async syncFromSheets(
    spreadsheetId?: string,
    sheetName?: string
  ): Promise<SyncResult> {
    try {
      const prisma = getPrisma();
      const sheetsService = getGoogleSheetsService();
      const settingsService = getInventorySettingsService();

      // Get settings
      let settings = await settingsService.getSettings();
      if (spreadsheetId) settings.spreadsheetId = spreadsheetId;
      if (sheetName) settings.sheetName = sheetName;

      if (!settings?.spreadsheetId || !settings?.sheetName) {
        throw new Error("Spreadsheet not configured");
      }

      const sheetData = await sheetsService.getSheetData(
        settings.spreadsheetId,
        settings.sheetName
      );

      if (!sheetData || sheetData.length === 0) {
        throw new Error("Sheet is empty");
      }

      const headers = sheetData[0];
      const keyColumn = settings.keyColumn || headers[0];
      const keyColumnIndex = headers.indexOf(keyColumn);

      if (keyColumnIndex === -1) {
        throw new Error(`Key column "${keyColumn}" not found in sheet`);
      }

      // Validate headers – continue sync but warn and auto-align stored columns when mismatch
      const headerValidation = await this.validateHeaders(headers, settings);
      let headerWarning: string | null = null;

      if (!headerValidation.is_valid) {
        const missing = headerValidation.missing_columns || [];
        const extra = headerValidation.extra_columns || [];

        headerWarning = `Headers mismatch detected. Missing: ${
          missing.join(", ") || "-"
        }. Extra: ${
          extra.join(", ") || "-"
        }. Stored schema will be realigned to current sheet headers.`;

        // Realign stored schema to current sheet headers to avoid repeated mismatch
        await settingsService.updateSettings({
          allColumns: headers,
          selectedColumns: headers,
        });
      }

      // Process data rows (skip header)
      let newCount = 0;
      let updatedCount = 0;
      const tenantId = tenantContext.getTenantId();

      for (let rowIndex = 1; rowIndex < sheetData.length; rowIndex++) {
        const row = sheetData[rowIndex];
        if (!row[keyColumnIndex]) continue;

        const keyValue = String(row[keyColumnIndex]).trim();
        if (!keyValue) continue;

        const recordData: any = {};
        for (let colIndex = 0; colIndex < headers.length; colIndex++) {
          recordData[headers[colIndex]] = row[colIndex] || "";
        }

        // Use tenant-scoped query
        const existingRecord = await prisma.inventoryRecord.findFirst({
          where: {
            tenantId,
            keyColumnName: keyColumn,
            keyValue: keyValue,
          },
        });

        if (existingRecord) {
          await prisma.inventoryRecord.update({
            where: { id: existingRecord.id },
            data: {
              data: JSON.stringify(recordData),
              keyValue,
              keyColumnName: keyColumn,
              updatedAt: new Date(),
            },
          });
          updatedCount++;
        } else {
          await prisma.inventoryRecord.create({
            data: {
              tenantId,
              data: JSON.stringify(recordData),
              keyValue,
              keyColumnName: keyColumn,
              createdAt: new Date(),
              updatedAt: new Date(),
            },
          });
          newCount++;
        }
      }

      // 📸 Save snapshot after successful sync from sheet
      const snapshotService = getSnapshotService();
      await snapshotService.saveSnapshot(sheetData, headers, keyColumn);
      console.log("📸 Snapshot saved after sync from sheets");

      await settingsService.updateLastSyncTimestamp();

      const messageBase = `Successfully synced ${newCount} new + ${updatedCount} updated records from Google Sheets`;
      const message = headerWarning
        ? `${messageBase}. ${headerWarning}`
        : messageBase;

      return {
        status: "SUCCESS",
        message,
        records_synced: newCount + updatedCount,
        new_records: newCount,
        updated_records: updatedCount,
        timestamp: new Date().toISOString(),
      };
    } catch (error: any) {
      const message = `Sync from sheets failed: ${error.message}`;
      console.error(`❌ ${message}`);
      return { status: "FAILED", message, timestamp: new Date().toISOString() };
    }
  }

  /**
   * Sync to Google Sheets using Delta Sync (Cell-Level Updates)
   * Only updates changed cells - minimizes API quota usage
   */
  async syncToSheets(
    spreadsheetIdParam?: string,
    sheetNameParam?: string,
    lockedColumns: string[] = []
  ): Promise<SyncResult> {
    try {
      const prisma = getPrisma();
      const settingsService = getInventorySettingsService();
      const deltaSyncExporter = getDeltaSyncExporter();

      let settings = await settingsService.getSettings();
      if (spreadsheetIdParam) settings.spreadsheetId = spreadsheetIdParam;
      if (sheetNameParam) settings.sheetName = sheetNameParam;

      if (!settings?.spreadsheetId || !settings?.sheetName) {
        throw new Error("Spreadsheet not configured");
      }

      // Get tenant-scoped records
      const tenantId = tenantContext.getTenantId();
      const records = await prisma.inventoryRecord.findMany({
        where: { tenantId },
      });

      if (records.length === 0) throw new Error("No records to sync");

      // Prepare database records for diff
      const dbRecords = records.map((r) => {
        const data = typeof r.data === "string" ? JSON.parse(r.data) : r.data;
        const transformedData =
          CheckboxDataConverter.transformRowForSheetsExport(data);

        // Find SKU value
        const skuKey = Object.keys(transformedData).find(
          (k) => k.toLowerCase() === "sku" || k.toLowerCase() === "skuid"
        );
        const sku = skuKey ? String(transformedData[skuKey] || "").trim() : "";

        return { sku, data: transformedData };
      });

      // Get headers/columns config
      const selectedColumns = Array.isArray(settings.selectedColumns)
        ? settings.selectedColumns
        : typeof settings.selectedColumns === "string" &&
            settings.selectedColumns
          ? (settings.selectedColumns as string).split(",")
          : [];

      // Build column configs (all active by default, respect locked)
      const columnConfigs: ColumnConfig[] = selectedColumns.map((col) => ({
        name: col,
        isActive: true,
        isLocked: lockedColumns.includes(col),
      }));

      // Execute delta sync
      const result = await deltaSyncExporter.exportDelta(
        settings.spreadsheetId,
        settings.sheetName,
        dbRecords,
        selectedColumns,
        columnConfigs,
        lockedColumns
      );

      await settingsService.updateLastSyncTimestamp();

      const message = result.success
        ? `✅ Delta sync: ${result.stats.cellsUpdated} cells, ${result.stats.rowsAppended} new rows (${result.stats.apiCalls} API calls)`
        : result.message;

      return {
        status: result.success ? "SUCCESS" : "FAILED",
        message,
        records_synced: result.stats.cellsUpdated + result.stats.rowsAppended,
        updated_records: result.stats.cellsUpdated,
        unchanged_records: result.stats.unchanged,
        new_records: result.stats.rowsAppended,
        skipped_records: result.stats.skipped,
        errors:
          result.errors.length > 0
            ? result.errors.map((e, i) => ({ rowIndex: i, error: e }))
            : undefined,
        timestamp: new Date().toISOString(),
      };
    } catch (error: any) {
      const message = `Sync to sheets failed: ${error.message}`;
      console.error(`❌ ${message}`);
      return { status: "FAILED", message, timestamp: new Date().toISOString() };
    }
  }

  async validateHeaders(
    headers: string[],
    settings: any
  ): Promise<HeaderValidationResult> {
    const validationService = new InventoryValidationService();
    return validationService.validateHeaders(headers, settings);
  }

  async getSheetHeaders(): Promise<string[]> {
    try {
      const sheetsService = getGoogleSheetsService();
      const settingsService = getInventorySettingsService();
      const settings = await settingsService.getSettings();

      if (!settings?.spreadsheetId || !settings?.sheetName) {
        throw new Error("Spreadsheet not configured");
      }

      const rows = await sheetsService.getSheetData(
        settings.spreadsheetId,
        settings.sheetName
      );

      return rows?.[0] || [];
    } catch (error: any) {
      console.error("❌ Get sheet headers error:", error);
      throw error;
    }
  }
}

let instance: InventorySyncService;

export function getInventorySyncService(): InventorySyncService {
  if (!instance) {
    instance = new InventorySyncService();
  }
  return instance;
}

// Note: Do NOT export singleton instance directly
// Always use getInventorySyncService() to avoid module-load issues

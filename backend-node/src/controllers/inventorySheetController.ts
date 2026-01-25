/**
 * Inventory Import/Export Controller
 * Handles inventory data sync with registered Google Sheets
 *
 * Single Responsibility: Inventory sheet operations
 */

import { Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getSpreadsheetRegistry } from "../services/spreadsheetRegistryService";
import { getGoogleSheetsDataService } from "../services/googleSheetsDataImportExportService";
import { backendLogger } from "../utils/backendLogger";

const logger = getLogger("InventorySheetController");

/**
 * POST /api/inventory/export-to-sheet
 * Export inventory data to registered sheet
 */
export async function exportInventoryToSheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { sheetId, data } = req.body;
    // Note: "system" here is a user identifier for logging (automated process),
    // NOT related to system.db tenant - per AGENTS.MD tenant architecture
    const userId = (req as any).user?.id || "automated";

    if (!sheetId || !data || !Array.isArray(data)) {
      res.status(400).json({
        success: false,
        error: "sheetId and data array are required",
      });
      return;
    }

    const registry = getSpreadsheetRegistry();
    const sheet = registry.getSpreadsheet(sheetId);

    if (!sheet) {
      res.status(404).json({
        success: false,
        error: "Sheet not found",
      });
      return;
    }

    // Check if locked
    if (registry.isLocked(sheetId)) {
      res.status(409).json({
        success: false,
        error: "Sheet is currently locked for editing",
      });
      return;
    }

    // Lock sheet for operation
    registry.lockSpreadsheet(sheetId, userId);

    try {
      const dataService = getGoogleSheetsDataService();

      // Add headers
      const headers = Object.keys(data[0] || {});
      const values = [
        headers,
        ...data.map((item: any) => headers.map((h) => item[h] || "")),
      ];

      // Get first sheet name
      const targetSheet = sheet.sheets[0]?.name || "Sheet1";

      // Clear and write
      const cleared = await dataService.clearRange(
        sheet.spreadsheetId,
        `${targetSheet}!A:Z`,
      );

      if (!cleared) {
        throw new Error("Failed to clear sheet");
      }

      const written = await dataService.writeSheetData(
        sheet.spreadsheetId,
        targetSheet,
        "A1",
        values,
      );

      if (!written) {
        throw new Error("Failed to write data");
      }

      // Update last used
      registry.recordUsage(sheetId);

      backendLogger.success(
        "InventorySheet",
        `✅ Exported ${data.length} items to ${sheet.spreadsheetName}`,
      );

      res.json({
        success: true,
        message: `Exported ${data.length} items successfully`,
        stats: {
          rows: data.length,
          columns: headers.length,
        },
      });
    } finally {
      // Always unlock
      registry.unlockSpreadsheet(sheetId, userId);
    }
  } catch (error: any) {
    logger.error(`Error exporting inventory: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/inventory/import-from-sheet
 * Import inventory data from registered sheet
 */
export async function importInventoryFromSheetController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { sheetId, sheetName } = req.body;

    if (!sheetId) {
      res.status(400).json({
        success: false,
        error: "sheetId is required",
      });
      return;
    }

    const registry = getSpreadsheetRegistry();
    const sheet = registry.getSpreadsheet(sheetId);

    if (!sheet) {
      res.status(404).json({
        success: false,
        error: "Sheet not found",
      });
      return;
    }

    try {
      const dataService = getGoogleSheetsDataService();
      const targetSheet = sheetName || sheet.sheets[0]?.name || "Sheet1";

      // Read data
      const rawData = await dataService.readSheetData(
        sheet.spreadsheetId,
        targetSheet,
        "A1:Z1000",
      );

      if (!rawData || rawData.length === 0) {
        res.status(400).json({
          success: false,
          error: "No data found in sheet",
        });
        return;
      }

      // Parse into objects
      const headers = rawData[0] as string[];
      const items = rawData.slice(1).map((row: any[]) => {
        const item: Record<string, any> = {};
        headers.forEach((header, index) => {
          item[header.trim()] = row[index] || "";
        });
        return item;
      });

      // Update last used
      registry.recordUsage(sheetId);

      backendLogger.success(
        "InventorySheet",
        `✅ Imported ${items.length} items from ${sheet.spreadsheetName}`,
      );

      res.json({
        success: true,
        data: items,
        stats: {
          rows: items.length,
          columns: headers.length,
          headers,
        },
      });
    } catch (error: any) {
      throw error;
    }
  } catch (error: any) {
    logger.error(`Error importing inventory: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/inventory/sync-status
 * Check sync status for sheet
 */
export async function checkSyncStatusController(
  req: Request,
  res: Response,
): Promise<void> {
  try {
    const { sheetId } = req.body;

    if (!sheetId) {
      res.status(400).json({
        success: false,
        error: "sheetId is required",
      });
      return;
    }

    const registry = getSpreadsheetRegistry();
    const sheet = registry.getSpreadsheet(sheetId);

    if (!sheet) {
      res.status(404).json({
        success: false,
        error: "Sheet not found",
      });
      return;
    }

    res.json({
      success: true,
      data: {
        spreadsheetName: sheet.spreadsheetName,
        isLocked: registry.isLocked(sheetId),
        lastUsedAt: sheet.lastUsedAt,
        syncEnabled: sheet.syncSettings.autoSync,
        syncInterval: sheet.syncSettings.syncInterval,
      },
    });
  } catch (error: any) {
    logger.error(`Error checking sync status: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

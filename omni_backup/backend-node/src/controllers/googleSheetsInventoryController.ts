/**
 * Google Sheets Inventory Controller
 * SRP: Handle inventory-related Google Sheets operations
 */

import { Request, Response } from "express";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getLogger } from "../utils/logger";

const logger = getLogger("GoogleSheetsInventory");

/**
 * GET /api/google/sheets/data
 * Fetch all inventory data from Google Sheets
 */
export async function getAllInventorySheetsDataController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("📊 Fetching all inventory data from Google Sheets");

    const service = getGoogleSheetsService();
    const { getInventorySettingsService } = await import(
      "../services/inventorySettingsService"
    );
    const settingsService = getInventorySettingsService();

    if (!service.isAuthenticated()) {
      logger.warn("❌ Not authenticated");
      res.status(401).json({ success: false, error: "Not authenticated" });
      return;
    }

    const settings = await settingsService.getSettings();

    if (!settings?.spreadsheetId || !settings?.sheetName) {
      logger.warn("❌ Inventory sheet not configured");
      res.status(400).json({
        success: false,
        error:
          "Inventory sheet not configured. Please set up Google Sheets integration first.",
      });
      return;
    }

    const rows = await service.getSheetData(
      settings.spreadsheetId,
      settings.sheetName
    );

    if (!rows || rows.length === 0) {
      logger.warn("⚠️ No data found in configured sheet");
      res.json({
        success: true,
        data: [],
        message: "No data found in configured sheet",
      });
      return;
    }

    const headers = rows[0] || [];
    const dataStartRow = settings.dataStartRow || 2;
    const dataRows = rows.slice(dataStartRow - 1);

    const transformedData = dataRows.map((row: any[]) => {
      const obj: any = {};
      headers.forEach((header: string, colIndex: number) => {
        obj[header] = row[colIndex] || null;
      });
      return obj;
    });

    logger.info(
      `✅ Fetched ${transformedData.length} items from Google Sheets (${settings.sheetName})`
    );

    res.json({
      success: true,
      data: transformedData,
      count: transformedData.length,
      sheet_name: settings.sheetName,
      spreadsheet_id: settings.spreadsheetId,
    });
  } catch (error: any) {
    logger.error(`❌ Error fetching Google Sheets data: ${error.message}`);
    res
      .status(500)
      .json({
        success: false,
        error: error.message || "Failed to fetch data from Google Sheets",
      });
  }
}

/**
 * GET /api/google/sheets/refresh
 * Refresh list of available spreadsheets
 */
export async function refreshSpreadsheetsController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("🔄 Refreshing spreadsheet list");

    const service = getGoogleSheetsService();

    if (!service.isAuthenticated()) {
      logger.warn("❌ Not authenticated");
      res.status(401).json({ success: false, error: "Not authenticated" });
      return;
    }

    const spreadsheets = await service.refreshSpreadsheets();
    logger.info(
      `✅ Refreshed spreadsheet list | Count: ${spreadsheets.length}`
    );

    res.json({
      success: true,
      data: spreadsheets,
      count: spreadsheets.length,
      message: `Refreshed ${spreadsheets.length} spreadsheets`,
    });
  } catch (error: any) {
    logger.error(`❌ Error refreshing spreadsheets: ${error.message}`);
    res.status(500).json({ success: false, error: error.message });
  }
}

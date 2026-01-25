/**
 * Google Sheets Data Controller
 * SRP: Handle basic spreadsheet and worksheet operations
 * Inventory operations delegated to googleSheetsInventoryController
 */

import { Request, Response } from "express";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getLogger } from "../utils/logger";

// Re-export inventory controllers for backward compatibility
export {
  getAllInventorySheetsDataController,
  refreshSpreadsheetsController,
} from "./googleSheetsInventoryController";

const logger = getLogger("GoogleSheetsData");

/**
 * GET /api/google/columns/:spreadsheetId/:sheetName
 */
export async function getColumnHeadersController(req: Request, res: Response): Promise<void> {
  try {
    const { spreadsheetId, sheetName } = req.params;

    if (!spreadsheetId || !sheetName) {
      logger.warn("❌ Missing spreadsheetId or sheetName");
      res.status(400).json({ success: false, error: "Missing spreadsheetId or sheetName" });
      return;
    }

    logger.info(`🔄 Getting column headers for ${sheetName} in spreadsheet ${spreadsheetId}`);

    const service = getGoogleSheetsService();
    const headers = await service.getColumnHeaders(spreadsheetId, decodeURIComponent(sheetName));

    logger.info(`✅ Retrieved ${headers.length} columns`);
    res.json({ success: true, data: { headers, count: headers.length } });
  } catch (error: any) {
    logger.error(`❌ Error getting column headers: ${error.message}`);
    res.status(500).json({ success: false, error: error.message });
  }
}

/**
 * GET /api/google/sheets/list
 */
export async function getSheetsListController(_req: Request, res: Response): Promise<void> {
  try {
    logger.info("🔄 Retrieving list of available spreadsheets");

    const service = getGoogleSheetsService();

    if (!service.isAuthenticated()) {
      logger.warn("❌ Not authenticated");
      res.status(401).json({ success: false, error: "Not authenticated", action: "Redirect to /api/google/auth/initiate" });
      return;
    }

    const spreadsheets = await service.getSpreadsheets();

    if (spreadsheets.length === 0) {
      logger.warn("⚠️ No spreadsheets found - may be due to invalid token");
    }

    logger.info(`✅ Retrieved ${spreadsheets.length} spreadsheets`);

    try {
      const { googleSheetsSettingsService } = await import("../services/googleSheetsSettingsService");
      await googleSheetsSettingsService.updateAvailableSpreadsheets(spreadsheets);
    } catch (dbError) {
      logger.warn(`⚠️ Could not save spreadsheets to database: ${(dbError as any).message}`);
    }

    res.json({ success: true, data: spreadsheets, count: spreadsheets.length });
  } catch (error: any) {
    const errorMsg = error.message || String(error);
    logger.error(`❌ Error retrieving spreadsheets: ${errorMsg}`);

    const isTokenError = errorMsg.includes("invalid_grant") || errorMsg.includes("Invalid credentials");

    res.status(isTokenError ? 401 : 500).json({
      success: false,
      error: errorMsg,
      action: isTokenError ? "Re-authenticate at /api/google/auth/initiate" : undefined,
    });
  }
}

/**
 * GET /api/google/worksheets/:spreadsheetId
 */
export async function getWorksheetListController(req: Request, res: Response): Promise<void> {
  try {
    const spreadsheetId = req.params.spreadsheetId;

    logger.info(`🔄 Retrieving worksheets for spreadsheet: ${spreadsheetId}`);

    const service = getGoogleSheetsService();

    if (!service.isAuthenticated()) {
      logger.warn("❌ Not authenticated");
      res.status(401).json({ success: false, error: "Not authenticated", action: "Redirect to /api/google/auth/initiate" });
      return;
    }

    const worksheets = await service.getWorksheets(spreadsheetId);

    if (worksheets.length === 0) {
      logger.warn(`⚠️ No worksheets found for spreadsheet ${spreadsheetId}`);
    }

    logger.info(`✅ Retrieved ${worksheets.length} worksheets`);

    res.json({ success: true, data: worksheets, count: worksheets.length });
  } catch (error: any) {
    const errorMsg = error.message || String(error);
    logger.error(`❌ Error retrieving worksheets: ${errorMsg}`);

    const isTokenError = errorMsg.includes("invalid_grant") || errorMsg.includes("Invalid credentials");

    res.status(isTokenError ? 401 : 500).json({
      success: false,
      error: errorMsg,
      action: isTokenError ? "Re-authenticate at /api/google/auth/initiate" : undefined,
    });
  }
}

/**
 * POST /api/google/spreadsheets/create
 */
export async function createSpreadsheetController(req: Request, res: Response): Promise<void> {
  try {
    const data = req.body;

    if (!data || !data.title) {
      logger.warn("❌ Title is required");
      res.status(400).json({ success: false, error: "Title is required" });
      return;
    }

    logger.info(`🔄 Creating new Google Spreadsheet with title: '${data.title}'`);

    const service = getGoogleSheetsService();
    const newSpreadsheet = await service.createSpreadsheet(data.title);

    if (newSpreadsheet) {
      logger.info(`✅ Spreadsheet '${data.title}' created successfully | ID: ${newSpreadsheet.id}`);
      res.json({ success: true, data: newSpreadsheet, message: `Spreadsheet '${data.title}' created successfully` });
    } else {
      logger.error(`❌ Failed to create spreadsheet '${data.title}'`);
      res.status(400).json({ success: false, error: "Failed to create spreadsheet" });
    }
  } catch (error: any) {
    logger.error(`❌ Error creating spreadsheet: ${error.message}`);
    res.status(500).json({ success: false, error: error.message });
  }
}

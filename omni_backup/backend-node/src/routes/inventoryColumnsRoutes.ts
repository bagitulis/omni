/**
 * Inventory Columns Routes
 * Handles column management for inventory
 * Single Responsibility: Column configuration only
 */

import { Router, Request, Response } from "express";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getInventoryService } from "../services/inventoryService";
import { getLogger } from "../utils/logger";

const logger = getLogger("InventoryColumns");
export const inventoryColumnsRouter = Router();

/**
 * GET /api/inventory/columns/available
 * Get available columns from the inventory sheet
 */
inventoryColumnsRouter.get(
  "/available",
  async (_req: Request, res: Response) => {
    try {
      const settings = await getInventoryService().getSettings();

      if (!settings.spreadsheetId || !settings.sheetName) {
        logger.warn("⚠️ Inventory spreadsheet not configured yet");
        return res.json({
          status: "SUCCESS",
          columns: [],
          total: 0,
        });
      }

      logger.info(`🔄 Fetching available columns from: ${settings.sheetName}`);

      const googleSheets = getGoogleSheetsService();
      const headers = await googleSheets.getColumnHeaders(
        settings.spreadsheetId,
        settings.sheetName
      );

      const columns = headers.map((name: string, index: number) => ({
        name,
        type: "text",
        is_key: false,
        spreadsheet_column: String.fromCharCode(65 + index),
        column_position: index + 1,
      }));

      logger.info(`✅ Found ${columns.length} columns`);

      return res.json({
        status: "SUCCESS",
        columns,
        total: columns.length,
      });
    } catch (error) {
      logger.error(`❌ Get available columns error: ${error}`);
      return res.status(500).json({
        status: "FAILED",
        message: (error as any).message || "Failed to fetch columns",
        columns: [],
      });
    }
  }
);

/**
 * GET /api/inventory/columns/selected
 * Get currently selected columns
 */
inventoryColumnsRouter.get(
  "/selected",
  async (_req: Request, res: Response) => {
    try {
      const settings = await getInventoryService().getSettings();
      const selectedColumns = settings.selectedColumns || [];

      logger.info(`✅ Selected: ${selectedColumns.join(", ")}`);

      return res.json({
        status: "SUCCESS",
        selected_columns: selectedColumns,
      });
    } catch (error) {
      logger.error(`❌ Get selected columns error: ${error}`);
      return res.status(500).json({
        status: "FAILED",
        message: (error as any).message || "Failed to get selected columns",
        selected_columns: [],
      });
    }
  }
);

/**
 * POST /api/inventory/columns/selected
 * Update selected columns
 */
inventoryColumnsRouter.post(
  "/selected",
  async (req: Request, res: Response) => {
    try {
      const { selected_columns } = req.body;

      if (!Array.isArray(selected_columns)) {
        return res.status(400).json({
          status: "FAILED",
          message: "selected_columns must be an array",
        });
      }

      await getInventoryService().updateSelectedColumns(selected_columns);

      logger.info(`✅ Updated: ${selected_columns.join(", ")}`);

      return res.json({
        status: "SUCCESS",
        selected_columns,
        message: "Selected columns updated",
      });
    } catch (error) {
      logger.error(`❌ Update selected columns error: ${error}`);
      return res.status(500).json({
        status: "FAILED",
        message: (error as any).message || "Failed to update selected columns",
      });
    }
  }
);

export default inventoryColumnsRouter;

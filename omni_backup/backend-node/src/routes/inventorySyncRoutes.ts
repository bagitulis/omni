/**
 * Inventory Sync Routes
 * Handles inventory synchronization with Google Sheets
 * Single Responsibility: Sync operations only
 */

import { Router, Request, Response } from "express";
import { getInventoryService } from "../services/inventoryService";
import { getLogger } from "../utils/logger";

const logger = getLogger("InventorySync");
export const inventorySyncRouter = Router();

/**
 * POST /api/inventory/sync/from-sheets
 * Sync inventory data FROM Google Sheets TO database
 */
inventorySyncRouter.post(
  "/from-sheets",
  async (req: Request, res: Response) => {
    try {
      let { spreadsheet_id, sheet_name } = req.body;

      // If parameters not provided, try to use saved settings from database
      if (!spreadsheet_id || !sheet_name) {
        logger.info(
          "Parameters not provided, attempting to use saved settings..."
        );
        const settings = await getInventoryService().getSettings();

        if (!settings?.spreadsheetId || !settings?.sheetName) {
          return res.status(400).json({
            status: "FAILED",
            message:
              "Missing required parameters: spreadsheet_id and sheet_name. Please configure in settings first.",
          });
        }

        spreadsheet_id = settings.spreadsheetId;
        sheet_name = settings.sheetName;
        logger.info(`Using saved settings: ${sheet_name}`);
      }

      const result = await getInventoryService().syncFromSheets(
        spreadsheet_id,
        sheet_name
      );

      logger.info(
        `✅ Synced ${result.synced_records} records from sheets in ${result.duration}ms`
      );

      return res.json(result);
    } catch (error) {
      logger.error(`❌ Sync from sheets error: ${error}`);
      return res.status(500).json({
        status: "FAILED",
        message: (error as any).message || "Failed to sync from sheets",
        total_records: 0,
        synced_records: 0,
        failed_records: 0,
        duration: 0,
      });
    }
  }
);

/**
 * POST /api/inventory/sync/to-sheets
 * Sync inventory data FROM database TO Google Sheets
 * Uses Delta Sync - only updates changed cells (minimizes API quota)
 *
 * Body params:
 * - spreadsheet_id: string (required)
 * - sheet_name: string (required)
 * - locked_columns: string[] (optional) - columns to skip during export
 */
inventorySyncRouter.post("/to-sheets", async (req: Request, res: Response) => {
  try {
    const { spreadsheet_id, sheet_name, locked_columns } = req.body;

    if (!spreadsheet_id || !sheet_name) {
      return res.status(400).json({
        status: "FAILED",
        message: "Missing required parameters: spreadsheet_id and sheet_name",
      });
    }

    // Parse locked_columns - can be array or comma-separated string
    const lockedColumns: string[] = Array.isArray(locked_columns)
      ? locked_columns
      : typeof locked_columns === "string"
        ? locked_columns
            .split(",")
            .map((c) => c.trim())
            .filter(Boolean)
        : [];

    const result = await getInventoryService().syncToSheets(
      spreadsheet_id,
      sheet_name,
      lockedColumns
    );

    logger.info(
      `✅ Delta sync: ${result.updated_records} cells + ${result.new_records} rows (${result.unchanged_records} unchanged)`
    );

    return res.json(result);
  } catch (error) {
    logger.error(`❌ Sync to sheets error: ${error}`);
    return res.status(500).json({
      status: "FAILED",
      message: (error as any).message || "Failed to sync to sheets",
      total_records: 0,
      synced_records: 0,
      duration: 0,
    });
  }
});

/**
 * GET /api/inventory/sync/history
 * Get sync history
 */
inventorySyncRouter.get("/history", async (_req: Request, res: Response) => {
  try {
    const history = await getInventoryService().getSyncHistory();

    logger.info(`✅ Retrieved ${history.length} sync history records`);

    return res.json({
      status: "SUCCESS",
      data: history,
      total: history.length,
    });
  } catch (error) {
    logger.error(`❌ Get sync history error: ${error}`);
    return res.status(500).json({
      status: "FAILED",
      message: (error as any).message || "Failed to get sync history",
      data: [],
    });
  }
});

export default inventorySyncRouter;

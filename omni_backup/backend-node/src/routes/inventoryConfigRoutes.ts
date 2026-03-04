/**
 * Inventory Config Routes
 * Handles inventory configuration
 * Single Responsibility: Config management only
 */

import { Router, Request, Response } from "express";
import { getInventoryService } from "../services/inventoryService";
import { getLogger } from "../utils/logger";

const logger = getLogger("InventoryConfig");
export const inventoryConfigRouter = Router();

/**
 * GET /api/inventory/config
 * Get current inventory configuration
 */
inventoryConfigRouter.get("/", async (_req: Request, res: Response) => {
  try {
    const settings = await getInventoryService().getSettings();

    const configData = {
      spreadsheet_id: settings.spreadsheetId || "",
      sheet_name: settings.sheetName || "",
      selected_columns: settings.selectedColumns || [],
      header_row: settings.headerRow || 1,
      data_start_row: settings.dataStartRow || 2,
      key_column: settings.keyColumn || "",
      auto_sync: settings.autoSync || false,
      sync_interval_seconds: settings.syncIntervalSeconds || 300,
      last_sync_timestamp: settings.lastSyncTimestamp?.toISOString() || null,
    };

    logger.info("✅ Config retrieved");

    return res.json({
      status: "SUCCESS",
      data: configData,
    });
  } catch (error) {
    logger.error(`❌ Get inventory config error: ${error}`);
    return res.status(500).json({
      status: "FAILED",
      message: (error as any).message || "Failed to get config",
      data: null,
    });
  }
});

export default inventoryConfigRouter;

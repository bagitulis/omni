/**
 * Inventory Data Routes
 * Handles inventory CRUD operations and statistics
 * Single Responsibility: Data management only
 */

import { Router, Request, Response } from "express";
import { getInventoryService } from "../services/inventoryService";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getLogger } from "../utils/logger";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import {
  sendSuccess,
  sendPaginated,
  sendBadRequest,
  sendNotFound,
  handleError,
} from "../utils/apiResponse";

const logger = getLogger("InventoryData");
export const inventoryDataRouter = Router();

// Apply authentication middleware to all routes
inventoryDataRouter.use(authMiddleware);
inventoryDataRouter.use(requireAuth);

/**
 * GET /api/inventory/list
 * Get inventory records with pagination and search
 */
inventoryDataRouter.get("/list", async (req: Request, res: Response) => {
  try {
    const offset = parseInt(req.query.offset as string) || 0;
    const limit = parseInt(req.query.limit as string) || 100;
    const search = req.query.search as string;

    const result = search
      ? await getInventoryService().searchInventory(search, offset, limit)
      : await getInventoryService().getInventoryList(offset, limit);

    return sendPaginated(res, result.data, result.total, offset, limit);
  } catch (error) {
    logger.error(`Get inventory list error: ${error}`);
    return handleError(res, error, "Failed to fetch inventory list");
  }
});

/**
 * GET /api/inventory/stats
 * Get inventory statistics
 */
inventoryDataRouter.get("/stats", async (_req: Request, res: Response) => {
  try {
    const settings = await getInventoryService().getSettings();

    const emptyStats = {
      total_records: 0,
      total_columns: 0,
      columns: [],
      last_sync: null,
      db_size_kb: 0,
    };

    if (!settings.spreadsheetId || !settings.sheetName) {
      return sendSuccess(res, emptyStats);
    }

    const googleSheets = getGoogleSheetsService();
    const sheetData = await googleSheets.getSheetData(
      settings.spreadsheetId,
      settings.sheetName
    );

    if (!sheetData || sheetData.length === 0) {
      return sendSuccess(res, emptyStats);
    }

    const headers = sheetData[0] as string[];
    const rows = sheetData.slice(1) as unknown[][];

    const columns = headers.map((name: string, index: number) => ({
      column_name: name,
      column_type: "text",
      is_key: settings.keyColumn === name,
      spreadsheet_column: String.fromCharCode(65 + index),
      column_position: index + 1,
    }));

    return sendSuccess(res, {
      total_records: rows.length,
      total_columns: headers.length,
      columns: columns,
      last_sync: settings.lastSyncTimestamp?.toISOString() || null,
      db_size_kb: 0,
    });
  } catch (error) {
    logger.error(`Get inventory stats error: ${error}`);
    return handleError(res, error, "Failed to get stats");
  }
});

/**
 * POST /api/inventory
 * Create new inventory record
 */
inventoryDataRouter.post("/", async (req: Request, res: Response) => {
  try {
    const data = req.body;

    if (!data || typeof data !== "object") {
      return sendBadRequest(res, "Invalid request body");
    }

    const settings = await getInventoryService().getSettings();
    const result = await getInventoryService().createInventory(
      data.keyValue || "",
      data,
      settings.keyColumn || "SKU"
    );

    logger.info(`Created inventory record`);
    return sendSuccess(res, result, 201);
  } catch (error) {
    logger.error(`Create inventory error: ${error}`);
    return handleError(res, error, "Failed to create inventory record");
  }
});

/**
 * GET /api/inventory/:keyValue
 * Get single inventory record by key value
 */
inventoryDataRouter.get("/:keyValue", async (req: Request, res: Response) => {
  try {
    const { keyValue } = req.params;
    const settings = await getInventoryService().getSettings();

    if (!settings.keyColumn) {
      return sendBadRequest(res, "Key column not configured");
    }

    const record = await getInventoryService().getInventoryByKey(keyValue);

    if (!record) {
      return sendNotFound(res, "Inventory record not found");
    }

    return sendSuccess(res, record);
  } catch (error) {
    logger.error(`Get inventory record error: ${error}`);
    return handleError(res, error, "Failed to get inventory record");
  }
});

/**
 * PUT /api/inventory/:keyValue
 * Update inventory record
 */
inventoryDataRouter.put("/:keyValue", async (req: Request, res: Response) => {
  try {
    const { keyValue } = req.params;
    const data = req.body;
    const settings = await getInventoryService().getSettings();

    if (!settings.keyColumn) {
      return sendBadRequest(res, "Key column not configured");
    }

    const result = await getInventoryService().updateInventory(keyValue, data);

    logger.info(`Updated inventory record: ${keyValue}`);
    return sendSuccess(res, result);
  } catch (error) {
    logger.error(`Update inventory error: ${error}`);
    return handleError(res, error, "Failed to update inventory record");
  }
});

/**
 * DELETE /api/inventory/:keyValue
 * Delete inventory record
 */
inventoryDataRouter.delete(
  "/:keyValue",
  async (req: Request, res: Response) => {
    try {
      const { keyValue } = req.params;
      const settings = await getInventoryService().getSettings();

      if (!settings.keyColumn) {
        return sendBadRequest(res, "Key column not configured");
      }

      const result = await getInventoryService().deleteInventory(keyValue);

      logger.info(`Deleted inventory record: ${keyValue}`);
      return sendSuccess(res, result);
    } catch (error) {
      logger.error(`Delete inventory error: ${error}`);
      return handleError(res, error, "Failed to delete inventory record");
    }
  }
);

export default inventoryDataRouter;

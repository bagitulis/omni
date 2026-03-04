/**
 * Inventory Sheet Routes
 * Handles export/import of inventory data to/from registered Google Sheets
 */

import { Router} from "express";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import {
  exportInventoryToSheetController,
  importInventoryFromSheetController,
  checkSyncStatusController,
} from "../controllers/inventorySheetController";

const router = Router();

/**
 * Export inventory to registered sheet
 * POST /api/inventory/export-to-sheet
 */
router.post(
  "/export-to-sheet",
  authMiddleware,
  requireAuth,
  exportInventoryToSheetController
);

/**
 * Import inventory from registered sheet
 * POST /api/inventory/import-from-sheet
 */
router.post(
  "/import-from-sheet",
  authMiddleware,
  requireAuth,
  importInventoryFromSheetController
);

/**
 * Check sync status
 * POST /api/inventory/sync-status
 */
router.post(
  "/sync-status",
  authMiddleware,
  requireAuth,
  checkSyncStatusController
);

export default router;

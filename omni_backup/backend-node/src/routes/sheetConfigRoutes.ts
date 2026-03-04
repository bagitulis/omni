import express from "express";
import {
  saveSheetConfig,
  loadSheetConfigs,
  deleteSheetConfig,
} from "../controllers/sheetConfigController";
import { authMiddleware } from "../middleware/authMiddleware";
import { tenantMiddleware } from "../middleware/tenantMiddleware";

const router = express.Router();

// Protect all routes with auth
router.use(authMiddleware);
router.use(tenantMiddleware);

/**
 * POST /google/sheet-config/save
 * Save sheet configuration
 */
router.post("/save", saveSheetConfig);

/**
 * GET /google/sheet-config/list
 * Load sheet configurations
 */
router.get("/list", loadSheetConfigs);

/**
 * DELETE /google/sheet-config/:sheetId
 * Delete sheet configuration
 */
router.delete("/:sheetId", deleteSheetConfig);

export default router;

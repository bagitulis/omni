import { Router, Request, Response } from "express";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { noCacheMiddleware } from "../middleware/cacheControl";
import { getAuthStatusController } from "../controllers/googleSheetsAuthController";
import {
  getSheetsListController,
  getWorksheetListController,
  createSpreadsheetController,
  getColumnHeadersController,
  refreshSpreadsheetsController,
  getAllInventorySheetsDataController,
} from "../controllers/googleSheetsDataController";
import {
  updateDetailedSettingsController,
  getDetailedSettingsController,
  testConnectionController,
  saveSpreadsheetLinksController,
  getSavedLinksController,
  validateSpreadsheetLinkController,
} from "../controllers/googleSheetsSettingsController";
import { getQuotaManager } from "../services/quota";

const router = Router();

// ==================== AUTHENTICATION (SERVICE ACCOUNT) ====================

/**
 * GET /api/google/auth/status
 * Get Google Sheets service account status
 */
router.get("/auth/status", noCacheMiddleware, getAuthStatusController);

// ==================== SHEETS MANAGEMENT (require auth) ====================

/**
 * GET /api/google/sheets/list
 * Get list of available spreadsheets
 */
router.get(
  "/sheets/list",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getSheetsListController
);

/**
 * GET /api/google/sheets/data
 * Fetch all inventory data from currently configured Google Sheet
 */
router.get(
  "/sheets/data",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getAllInventorySheetsDataController
);

/**
 * GET /api/google/sheets/worksheets/:spreadsheetId
 * Get list of worksheets in a spreadsheet
 */
router.get(
  "/sheets/worksheets/:spreadsheetId",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getWorksheetListController
);

/**
 * GET /api/google/sheets/columns/:spreadsheetId/:sheetName
 * Get column headers from a specific sheet
 */
router.get(
  "/sheets/columns/:spreadsheetId/:sheetName",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getColumnHeadersController
);

/**
 * POST /api/google/sheets/create
 * Create new Google Spreadsheet
 */
router.post(
  "/sheets/create",
  authMiddleware,
  requireAuth,
  createSpreadsheetController
);

// ==================== SETTINGS (require auth) ====================

/**
 * GET /api/google/settings/detailed
 * Get current detailed Google Sheets settings
 */
router.get(
  "/settings/detailed",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getDetailedSettingsController
);

/**
 * POST /api/google/settings/update-detailed
 * Update detailed Google Sheets settings (wallet, shipping, inventory, order)
 */
router.post(
  "/settings/update-detailed",
  authMiddleware,
  requireAuth,
  updateDetailedSettingsController
);

/**
 * GET /api/google/settings/test
 * Test current Google Sheets connection
 */
router.get(
  "/settings/test",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  testConnectionController
);

/**
 * GET /api/google/sheets/refresh
 * Refresh list of available spreadsheets
 */
router.get(
  "/sheets/refresh",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  refreshSpreadsheetsController
);

/**
 * POST /api/google/settings/save-links
 * Save spreadsheet links (URLs) for different purposes
 */
router.post(
  "/settings/save-links",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  saveSpreadsheetLinksController
);

/**
 * GET /api/google/settings/saved-links
 * Get saved spreadsheet links
 */
router.get(
  "/settings/saved-links",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  getSavedLinksController
);

/**
 * POST /api/google/settings/validate-link
 * Validate spreadsheet link and detect sheets
 */
router.post(
  "/settings/validate-link",
  noCacheMiddleware,
  authMiddleware,
  requireAuth,
  validateSpreadsheetLinkController
);

// ==================== QUOTA MANAGEMENT ====================

/**
 * GET /api/google/quota/status
 * Get current quota usage status across all service accounts
 * Public endpoint (no auth) - for monitoring purposes
 */
router.get(
  "/quota/status",
  noCacheMiddleware,
  (_req: Request, res: Response) => {
    try {
      const quotaManager = getQuotaManager();
      const status = quotaManager.getStatus();

      res.json({
        success: true,
        data: status,
      });
    } catch (error: any) {
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
);

/**
 * GET /api/google/quota/stats
 * Get detailed quota statistics
 * Public endpoint (no auth) - for monitoring purposes
 */
router.get("/quota/stats", noCacheMiddleware, (_req: Request, res: Response) => {
  try {
    const quotaManager = getQuotaManager();
    const stats = quotaManager.getDetailedStats();

    res.json({
      success: true,
      data: stats,
    });
  } catch (error: any) {
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

export default router;

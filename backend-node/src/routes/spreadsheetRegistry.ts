/**
 * Spreadsheet Registry Routes
 * Routes for managing spreadsheet registrations
 */

import { Router } from "express";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import {
  registerSpreadsheetController,
  getSpreadsheetController,
  getAllSpreadsheetsController,
  unlockSpreadsheetController,
  lockSpreadsheetController,
  deleteSpreadsheetController,
} from "../controllers/spreadsheetRegistryController";

const router = Router();

/**
 * POST /api/google/registry/register
 * Register new spreadsheet by URL
 */
router.post(
  "/register",
  authMiddleware,
  requireAuth,
  registerSpreadsheetController
);

/**
 * GET /api/google/registry
 * Get all spreadsheets for user
 */
router.get("/", authMiddleware, requireAuth, getAllSpreadsheetsController);

/**
 * GET /api/google/registry/:id
 * Get specific spreadsheet
 */
router.get("/:id", authMiddleware, requireAuth, getSpreadsheetController);

/**
 * POST /api/google/registry/:id/unlock
 * Unlock spreadsheet for editing
 */
router.post(
  "/:id/unlock",
  authMiddleware,
  requireAuth,
  unlockSpreadsheetController
);

/**
 * POST /api/google/registry/:id/lock
 * Lock spreadsheet after editing
 */
router.post(
  "/:id/lock",
  authMiddleware,
  requireAuth,
  lockSpreadsheetController
);

/**
 * DELETE /api/google/registry/:id
 * Delete spreadsheet registration
 */
router.delete("/:id", authMiddleware, requireAuth, deleteSpreadsheetController);

export default router;

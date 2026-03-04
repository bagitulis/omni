import { Router } from "express";
import {
  getDebugLogsController,
  clearDebugLogsController,
  getResourceUsageController,
} from "../controllers/settingsController";

const router = Router();

/**
 * GET /api/settings/debug-logs
 * Get debug logs with optional filtering by level
 */
router.get("/debug-logs", getDebugLogsController);

/**
 * POST /api/settings/debug-logs/clear
 * Clear all debug logs
 */
router.post("/debug-logs/clear", clearDebugLogsController);

/**
 * GET /api/settings/resource-usage
 * Get system and backend resource usage statistics
 */
router.get("/resource-usage", getResourceUsageController);

export default router;

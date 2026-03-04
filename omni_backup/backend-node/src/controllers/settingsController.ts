import { Request, Response } from "express";
import { getDebugLogs, clearDebugLogs } from "../services/logsService";
import { getResourceUsage } from "../services/resourceService";
import { getLogger } from "../utils/logger";

const logger = getLogger("SettingsController");

/**
 * GET /api/settings/debug-logs
 * Get debug logs with optional filtering
 */
export async function getDebugLogsController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const level = req.query.level as string | undefined;
    const limit = parseInt(req.query.limit as string) || 500;

    logger.info(
      `Getting debug logs | Level: ${level || "all"} | Limit: ${limit}`
    );

    const data = getDebugLogs(level, limit);

    res.json({
      success: true,
      data,
    });
  } catch (error: any) {
    logger.error("Error getting debug logs:", error.message);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/settings/debug-logs/clear
 * Clear all debug logs
 */
export async function clearDebugLogsController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("Clearing debug logs");

    const clearedCount = clearDebugLogs();

    logger.info(`Cleared ${clearedCount} log entries`);

    res.json({
      success: true,
      message: `Cleared ${clearedCount} log entries`,
    });
  } catch (error: any) {
    logger.error("Error clearing debug logs:", error.message);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * GET /api/settings/resource-usage
 * Get system and backend resource usage
 */
export async function getResourceUsageController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("Getting resource usage");

    const resourceData = getResourceUsage();

    res.json({
      success: true,
      data: resourceData,
    });
  } catch (error: any) {
    logger.error("Error getting resource usage:", error.message);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * Auto Function Routes
 * SRP: Handle CRUD for auto-function configuration
 * Actions delegated to autoFunctionActionRoutes
 */

import { Router } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { getAutoFunctionScheduler } from "../services/AutoFunctionScheduler";
import { getLogger } from "../utils/logger";
import { tenantContext } from "../utils/tenantContext";
import { autoFunctionActionRouter } from "./autoFunctionActionRoutes";

const logger = getLogger("AutoFunctionRoutes");
export const autoFunctionRouter = Router();

// Mount action routes
autoFunctionRouter.use("/", autoFunctionActionRouter);

/**
 * GET / - Get all auto-function configs
 */
autoFunctionRouter.get("/", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const scheduler = getAutoFunctionScheduler();
    const configs = scheduler.getAllConfigs(tenantId);

    res.json({
      success: true,
      data: { configs, total: configs.length },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get auto-function configs: ${error.message}`);
    res.status(500).json({ success: false, error: error.message });
  }
});

/**
 * GET /:name - Get specific auto-function config
 */
autoFunctionRouter.get("/:name", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const name = req.params.name;
    const scheduler = getAutoFunctionScheduler();
    const config = scheduler.getConfig(tenantId, name);

    if (!config) {
      return res.status(404).json({
        success: false,
        error: `Auto-function config not found: ${name}`,
      });
    }

    return res.json({ success: true, data: config });
  } catch (error: any) {
    logger.error(`❌ Failed to get auto-function config: ${error.message}`);
    return res.status(500).json({ success: false, error: error.message });
  }
});

/**
 * POST / - Create new auto-function
 */
autoFunctionRouter.post("/", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const { name, enabled, intervalMinutes, startTime, endTime } = req.body;

    if (!name || typeof enabled !== "boolean" || typeof intervalMinutes !== "number") {
      return res.status(400).json({
        success: false,
        error: "Invalid request: name, enabled (boolean), and intervalMinutes (number) required",
      });
    }

    if (intervalMinutes < 1) {
      return res.status(400).json({ success: false, error: "intervalMinutes must be at least 1" });
    }

    const scheduler = getAutoFunctionScheduler();
    const config = scheduler.createOrUpdateConfig(tenantId, name, enabled, intervalMinutes, startTime, endTime);

    if (!config) {
      return res.status(500).json({ success: false, error: `Failed to create auto-function config: ${name}` });
    }

    logger.info(`✅ Created new auto-function: ${name}`);
    return res.json({ success: true, data: config });
  } catch (error: any) {
    logger.error(`❌ Failed to create auto-function: ${error.message}`);
    return res.status(500).json({ success: false, error: error.message });
  }
});

/**
 * PUT /:name - Create or update auto-function config
 */
autoFunctionRouter.put("/:name", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const name = req.params.name;
    const { enabled, intervalMinutes, startTime, endTime } = req.body;

    if (typeof enabled !== "boolean" || typeof intervalMinutes !== "number") {
      return res.status(400).json({
        success: false,
        error: "Invalid request: enabled (boolean) and intervalMinutes (number) required",
      });
    }

    if (intervalMinutes < 1) {
      return res.status(400).json({ success: false, error: "intervalMinutes must be at least 1 minute" });
    }

    if (intervalMinutes > 525600) {
      return res.status(400).json({ success: false, error: "intervalMinutes cannot exceed 1 year (525600 minutes)" });
    }

    if ((startTime && !endTime) || (!startTime && endTime)) {
      return res.status(400).json({
        success: false,
        error: "Both startTime and endTime must be provided together, or neither",
      });
    }

    if (startTime && endTime) {
      const timeRegex = /^([0-1][0-9]|2[0-3]):[0-5][0-9]$/;
      if (!timeRegex.test(startTime) || !timeRegex.test(endTime)) {
        return res.status(400).json({
          success: false,
          error: "startTime and endTime must be in HH:mm format (00:00 to 23:59)",
        });
      }

      if (startTime >= endTime) {
        return res.status(400).json({ success: false, error: "startTime must be before endTime" });
      }
    }

    const scheduler = getAutoFunctionScheduler();
    const config = scheduler.createOrUpdateConfig(tenantId, name, enabled, intervalMinutes, startTime, endTime);

    if (!config) {
      return res.status(500).json({ success: false, error: `Failed to create/update auto-function config: ${name}` });
    }

    return res.json({ success: true, data: config });
  } catch (error: any) {
    logger.error(`❌ Failed to update auto-function config: ${error.message}`);
    return res.status(500).json({ success: false, error: error.message });
  }
});

/**
 * DELETE /:name - Delete auto-function
 */
autoFunctionRouter.delete("/:name", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const name = req.params.name;
    const scheduler = getAutoFunctionScheduler();
    const config = scheduler.getConfig(tenantId, name);

    if (!config) {
      return res.status(404).json({ success: false, error: `Auto-function not found: ${name}` });
    }

    const success = scheduler.deleteConfig(tenantId, name);

    if (!success) {
      return res.status(500).json({ success: false, error: `Failed to delete auto-function: ${name}` });
    }

    logger.info(`✅ Deleted auto-function: ${name}`);
    return res.json({ success: true, data: { deleted: true } });
  } catch (error: any) {
    logger.error(`❌ Failed to delete auto-function: ${error.message}`);
    return res.status(500).json({ success: false, error: error.message });
  }
});

export default autoFunctionRouter;

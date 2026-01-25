/**
 * Route Execution Config Routes
 * API endpoints for managing manual trigger mode configuration
 * Single Responsibility: HTTP handling for route execution configs
 */

import { Router } from "express";
import { getRouteExecutionConfigService } from "../services/routeExecutionConfigService";
import { getLogger } from "../utils/logger";

const logger = getLogger("RouteExecutionConfigRoutes");
export const routeExecutionConfigRouter = Router();

/**
 * GET /api/route-config - Get all route execution configs
 */
routeExecutionConfigRouter.get("/", (_req, res) => {
  try {
    const service = getRouteExecutionConfigService();
    const configs = service.getAll();

    res.json({
      success: true,
      data: configs,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get route configs: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /api/route-config/:routeKey - Get config by route key
 */
routeExecutionConfigRouter.get("/:routeKey", (req, res) => {
  try {
    const { routeKey } = req.params;
    const service = getRouteExecutionConfigService();
    const config = service.getByKey(routeKey);

    if (!config) {
      return res.status(404).json({
        success: false,
        error: "Route config not found",
      });
    }

    return res.json({
      success: true,
      data: config,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get route config: ${error.message}`);
    return res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /api/route-config/:routeKey/mode - Get execution mode for a route
 */
routeExecutionConfigRouter.get("/:routeKey/mode", (req, res) => {
  try {
    const { routeKey } = req.params;
    const service = getRouteExecutionConfigService();
    const mode = service.getExecutionMode(routeKey);
    const shouldQueue = service.shouldUseQueue(routeKey);

    res.json({
      success: true,
      data: {
        routeKey,
        executionMode: mode,
        shouldUseQueue: shouldQueue,
      },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get execution mode: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /api/route-config - Create new route config
 */
routeExecutionConfigRouter.post("/", (req, res) => {
  try {
    const {
      routeKey,
      routeName,
      description,
      executionMode,
      priority,
      icon,
      category,
    } = req.body;

    if (!routeKey || !routeName || !executionMode) {
      return res.status(400).json({
        success: false,
        error: "Missing required fields: routeKey, routeName, executionMode",
      });
    }

    const service = getRouteExecutionConfigService();
    const config = service.create({
      routeKey,
      routeName,
      description,
      executionMode,
      priority,
      icon,
      category,
    });

    if (!config) {
      return res.status(500).json({
        success: false,
        error: "Failed to create route config",
      });
    }

    return res.json({
      success: true,
      data: config,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to create route config: ${error.message}`);
    return res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * PUT /api/route-config/:routeKey - Update route config
 */
routeExecutionConfigRouter.put("/:routeKey", (req, res) => {
  try {
    const { routeKey } = req.params;
    const updates = req.body;

    const service = getRouteExecutionConfigService();
    const config = service.update(routeKey, updates);

    if (!config) {
      return res.status(404).json({
        success: false,
        error: "Route config not found",
      });
    }

    return res.json({
      success: true,
      data: config,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to update route config: ${error.message}`);
    return res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /api/route-config/:routeKey/toggle - Toggle execution mode
 */
routeExecutionConfigRouter.post("/:routeKey/toggle", (req, res) => {
  try {
    const { routeKey } = req.params;
    const service = getRouteExecutionConfigService();
    const config = service.toggleMode(routeKey);

    if (!config) {
      return res.status(404).json({
        success: false,
        error: "Route config not found",
      });
    }

    return res.json({
      success: true,
      data: config,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to toggle route mode: ${error.message}`);
    return res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * DELETE /api/route-config/:routeKey - Delete route config
 */
routeExecutionConfigRouter.delete("/:routeKey", (req, res) => {
  try {
    const { routeKey } = req.params;
    const service = getRouteExecutionConfigService();
    const deleted = service.delete(routeKey);

    res.json({
      success: true,
      data: { deleted },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to delete route config: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

export default routeExecutionConfigRouter;

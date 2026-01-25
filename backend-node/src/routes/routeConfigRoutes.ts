/**
 * Route Configuration Routes
 * Mengelola configuration untuk semua routes: enable/disable, caching, queue, timing
 * Single Responsibility: HTTP request handling & routing
 */

import { Router, Request, Response } from "express";
import { routeConfigService } from "../services/routeConfigService";
import { getLogger } from "../utils/logger";

const logger = getLogger("RouteConfigRoutes");
const router = Router();

/**
 * GET /api/routes-config/all
 */
router.get("/all", async (_req, res) => {
  try {
    const routes = await routeConfigService.getAllRoutes();
    res.json({
      success: true,
      data: {
        total: routes.length,
        routes,
      },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to fetch routes: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /api/routes-config/categories
 */
router.get("/categories", async (_req, res) => {
  try {
    const categories = await routeConfigService.getCategories();
    res.json({
      success: true,
      data: { categories },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to fetch categories: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /api/routes-config/by-category/:category
 */
router.get("/by-category/:category", async (req, res) => {
  try {
    const routes = await routeConfigService.getRoutesByCategory(req.params.category);
    res.json({
      success: true,
      data: {
        category: req.params.category,
        total: routes.length,
        routes,
      },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to fetch routes by category: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /api/routes-config/:id
 */
router.get("/:id", async (req, res) => {
  try {
    const route = await routeConfigService.getRouteById(req.params.id);
    if (!route) {
      return res.status(404).json({
        success: false,
        error: "Route not found",
      });
    }
    return res.json({ success: true, data: route });
  } catch (error: any) {
    logger.error(`❌ Failed to fetch route: ${error.message}`);
    return res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /api/routes-config
 */
router.post("/", async (req: Request, res: Response) => {
  try {
    if (!req.body.routePath) {
      return res.status(400).json({
        success: false,
        error: "routePath is required",
      });
    }
    const route = await routeConfigService.createRoute(req.body);
    return res.status(201).json({ success: true, data: route });
  } catch (error: any) {
    const statusCode = error.message.includes("already exists") ? 400 : 500;
    return res.status(statusCode).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * PATCH /api/routes-config/:id
 */
router.patch("/:id", async (req: Request, res: Response) => {
  try {
    const route = await routeConfigService.updateRoute(req.params.id, req.body);
    res.json({ success: true, data: route });
  } catch (error: any) {
    const statusCode = error.message.includes("not found") ? 404 : 500;
    res.status(statusCode).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * DELETE /api/routes-config/:id
 */
router.delete("/:id", async (req: Request, res: Response) => {
  try {
    const route = await routeConfigService.deleteRoute(req.params.id);
    res.json({ success: true, data: route });
  } catch (error: any) {
    const statusCode = error.message.includes("not found") ? 404 : 500;
    res.status(statusCode).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /api/routes-config/bulk-update
 */
router.post("/bulk-update", async (req: Request, res: Response) => {
  try {
    const { routeIds, updates } = req.body;
    const routes = await routeConfigService.bulkUpdateRoutes(routeIds, updates);
    res.json({
      success: true,
      data: {
        updated: routes.length,
        routes,
      },
    });
  } catch (error: any) {
    res.status(400).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /api/routes-config/apply-preset/:preset
 */
router.post("/apply-preset/:preset", async (req: Request, res: Response) => {
  try {
    const { routeIds } = req.body;
    const routes = await routeConfigService.applyPreset(
      routeIds,
      req.params.preset as any
    );
    res.json({
      success: true,
      data: {
        preset: req.params.preset,
        applied: routes.length,
        routes,
      },
    });
  } catch (error: any) {
    const statusCode = error.message.includes("Unknown preset") ? 400 : 500;
    res.status(statusCode).json({
      success: false,
      error: error.message,
    });
  }
});

export default router;

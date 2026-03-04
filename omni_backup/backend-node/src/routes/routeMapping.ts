/**
 * Route Mapping Routes
 * Routes untuk menampilkan frontend route map detail
 */

import express, { Router } from "express";
import { RouteMappingController } from "../controllers/routeMappingController";

const router: Router = express.Router();

/**
 * GET /api/route-mapping/analyze
 * Analyze semua frontend routes dan map ke backend endpoints
 */
router.get("/analyze", async (req, res) => {
  await RouteMappingController.analyze(req, res);
});

/**
 * GET /api/route-mapping/by-category
 * Get routes grouped by category
 * Query params: category (optional)
 */
router.get("/by-category", async (req, res) => {
  await RouteMappingController.getByCategory(req, res);
});

/**
 * GET /api/route-mapping/by-component
 * Get routes grouped by component
 * Query params: component (optional)
 */
router.get("/by-component", async (req, res) => {
  await RouteMappingController.getByComponent(req, res);
});

/**
 * GET /api/route-mapping/detailed-mapping
 * Get detailed mapping dengan format yang comprehensive
 */
router.get("/detailed-mapping", async (req, res) => {
  await RouteMappingController.getDetailedMapping(req, res);
});

/**
 * GET /api/route-mapping/statistics
 * Get statistics tentang route usage dan patterns
 */
router.get("/statistics", async (req, res) => {
  await RouteMappingController.getStatistics(req, res);
});

/**
 * GET /api/route-mapping/endpoint/:endpoint
 * Get detailed info tentang specific endpoint
 */
router.get("/endpoint/:endpoint", async (req, res) => {
  await RouteMappingController.getEndpointInfo(req, res);
});

/**
 * GET /api/route-mapping/component/:componentName
 * Get detailed info tentang specific component
 */
router.get("/component/:componentName", async (req, res) => {
  await RouteMappingController.getComponentInfo(req, res);
});
/**
 * POST /api/route-mapping/enable-auto-detection
 * Enable auto-detection dari source files
 * Scan backend routes dan frontend components, detect connections
 */
router.post("/enable-auto-detection", async (req, res) => {
  await RouteMappingController.enableAutoDetection(req, res);
});
export default router;

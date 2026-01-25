/**
 * Route Mapping Controller
 * Handles endpoints untuk menampilkan frontend route map detail
 */

import { Request, Response } from "express";
import { getRouteMappingService } from "../services/routeMappingService";
import { getLogger } from "../utils/logger";

const logger = getLogger("RouteMappingController");

export class RouteMappingController {
  /**
   * GET /api/route-mapping/analyze
   * Analyze semua frontend routes dan map ke backend endpoints dengan detail
   */
  static async analyze(_req: Request, res: Response): Promise<void> {
    try {
      const mapper = getRouteMappingService();
      const report = mapper.analyze();

      logger.info(
        `Route mapping analysis complete: ${report.totalRoutes} routes`
      );

      res.json({
        success: true,
        data: report,
      });
    } catch (error: any) {
      logger.error(`Route mapping analysis failed: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/by-category
   * Get routes grouped by category
   */
  static async getByCategory(req: Request, res: Response): Promise<void> {
    try {
      const mapper = getRouteMappingService();
      const result = mapper.getByCategory(req.query.category as string);

      res.json(result);
    } catch (error: any) {
      logger.error(`Failed to get routes by category: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/by-component
   * Get routes grouped by component/page
   */
  static async getByComponent(req: Request, res: Response): Promise<void> {
    try {
      const mapper = getRouteMappingService();
      const result = mapper.getByComponent(req.query.component as string);

      res.json(result);
    } catch (error: any) {
      logger.error(`Failed to get routes by component: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/detailed-mapping
   * Get detailed mapping dengan format yang comprehensive
   */
  static async getDetailedMapping(_req: Request, res: Response): Promise<void> {
    try {
      logger.info("Fetching detailed route mapping...");
      const mapper = getRouteMappingService();
      const report = mapper.getDetailedMapping();

      logger.info(
        `Generated detailed mapping: ${report.totalRoutes} routes, ${report.totalComponents} components`
      );

      res.json({
        success: true,
        total_routes: report.totalRoutes,
        total_components: report.totalComponents,
        total_categories: report.totalCategories,
        total_dynamic_routes: report.totalDynamicRoutes,
        total_called_routes: report.totalCalledRoutes,
        total_disconnected_routes: report.totalDisconnectedRoutes,
        total_unused_routes: report.totalUnusedRoutes,
        connection_rate: report.connectionRate,
        by_category: report.byCategory,
        category_labels: report.categoryLabels,
        categories: report.categories,
        category_stats: {
          connected: report.categories.connected.length,
          frontend_only: report.categories.frontendOnly.length,
          backend_only: report.categories.backendOnly.length,
          unused: report.categories.unused.length,
        },
        components: report.components,
        disconnected_routes: report.disconnectedRoutes,
        backend_only_routes: report.backendOnlyRoutes,
        unused_routes: report.unusedRoutes,
        button_to_endpoints: report.buttonToEndpoints,
        timestamp: report.timestamp,
      });
    } catch (error: any) {
      logger.error(`Failed to generate detailed mapping: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/statistics
   * Get statistics tentang route usage dan patterns
   */
  static async getStatistics(_req: Request, res: Response): Promise<void> {
    try {
      const mapper = getRouteMappingService();
      const stats = mapper.getStatistics();

      res.json(stats);
    } catch (error: any) {
      logger.error(`Failed to get route statistics: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/endpoint/:endpoint
   * Get detailed info tentang specific endpoint
   */
  static async getEndpointInfo(req: Request, res: Response): Promise<void> {
    try {
      const endpoint = req.params.endpoint;
      const mapper = getRouteMappingService();
      const report = mapper.analyze();

      // Search endpoint across all categories
      let foundEndpoint: any = null;

      for (const [category, routes] of Object.entries(report.byCategory)) {
        const found = (routes as any[]).find(
          (r) => r.endpoint === `/${endpoint}` || r.endpoint === endpoint
        );
        if (found) {
          foundEndpoint = { ...found, category };
          break;
        }
      }

      if (!foundEndpoint) {
        res.status(404).json({
          success: false,
          error: `Endpoint /${endpoint} not found`,
        });
        return;
      }

      res.json({
        success: true,
        endpoint: foundEndpoint,
      });
    } catch (error: any) {
      logger.error(`Failed to get endpoint info: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/route-mapping/component/:componentName
   * Get detailed info tentang specific component
   */
  static async getComponentInfo(req: Request, res: Response): Promise<void> {
    try {
      const componentName = req.params.componentName;
      const mapper = getRouteMappingService();
      const report = mapper.analyze();

      const component = report.components[componentName];

      if (!component) {
        res.status(404).json({
          success: false,
          error: `Component ${componentName} not found`,
        });
        return;
      }

      res.json({
        success: true,
        component: component,
      });
    } catch (error: any) {
      logger.error(`Failed to get component info: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/route-mapping/enable-auto-detection
   * Enable auto-detection dari source files
   */
  static async enableAutoDetection(
    _req: Request,
    res: Response
  ): Promise<void> {
    try {
      logger.info("Enabling auto-detection from source files...");
      const mapper = getRouteMappingService();
      mapper.enableAutoDetection();

      res.json({
        success: true,
        message: "Auto-detection enabled. Check console logs for details.",
      });
    } catch (error: any) {
      logger.error(`Failed to enable auto-detection: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}

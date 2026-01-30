/**
 * ML Analytics Controller
 * HTTP request/response handling for ML analytics endpoints
 * Single Responsibility: Handle HTTP layer, delegate to service
 */

import { Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";
import { MLAnalyticsService } from "../services/analytics/mlAnalyticsService";

const logger = getLogger("MLAnalyticsController");

export class MLAnalyticsController {
  /**
   * Validate tenant from request header
   */
  private validateTenant(req: Request, res: Response): string | null {
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!tenantId) {
      logger.warn("No tenant ID provided in request header");
      res.status(401).json({
        success: false,
        error: "Missing tenantId - authentication required",
        code: "TENANT_REQUIRED",
      });
      return null;
    }

    const dbManager = getDbManager();
    const tenants = dbManager.getAllTenants();

    if (!tenants.includes(tenantId)) {
      logger.warn(`Invalid tenant ID: ${tenantId}`);
      res.status(400).json({
        success: false,
        error: `Invalid tenant: ${tenantId}`,
        code: "INVALID_TENANT",
      });
      return null;
    }

    return tenantId;
  }

  /**
   * GET /api/analytics/ml/portfolio-health
   */
  async getPortfolioHealth(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const data = await service.getPortfolioHealth();

      res.json({
        success: true,
        data,
      });
    } catch (error) {
      logger.error(`Portfolio health failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }

  /**
   * GET /api/analytics/ml/products
   */
  async getProducts(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const options = {
        limit: parseInt(req.query.limit as string) || 20,
        cursor: req.query.cursor as string,
        sortBy: req.query.sort_by as string,
        sortDir: req.query.sort_dir as string,
        category: req.query.category as string,
        action: req.query.action as string,
        platform: req.query.platform as string,
      };

      const result = await service.getProducts(options);

      res.json({
        success: true,
        data: result.products,
        meta: result.meta,
      });
    } catch (error) {
      logger.error(`Get products failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }

  /**
   * GET /api/analytics/ml/product/:productId
   */
  async getProductDetail(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const productId = req.params.productId;
      if (!productId) {
        res.status(400).json({
          success: false,
          error: "Product ID is required",
        });
        return;
      }

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const data = await service.getProductDetail(productId);

      if (!data) {
        res.status(404).json({
          success: false,
          error: "Product not found",
        });
        return;
      }

      res.json({
        success: true,
        data,
      });
    } catch (error) {
      logger.error(`Get product detail failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }

  /**
   * GET /api/analytics/ml/alerts
   */
  async getAlerts(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const result = await service.getAlerts();

      res.json({
        success: true,
        data: result.alerts,
        meta: result.meta,
      });
    } catch (error) {
      logger.error(`Get alerts failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }

  /**
   * GET /api/analytics/ml/distribution
   */
  async getDistribution(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const data = await service.getDistribution();

      res.json({
        success: true,
        data,
      });
    } catch (error) {
      logger.error(`Get distribution failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }

  /**
   * POST /api/analytics/ml/budget-sim
   */
  async simulateBudget(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const { product_ids, budget_change_pct } = req.body;

      if (!product_ids || !Array.isArray(product_ids)) {
        res.status(400).json({
          success: false,
          error: "product_ids array is required",
        });
        return;
      }

      if (
        budget_change_pct === undefined ||
        typeof budget_change_pct !== "number"
      ) {
        res.status(400).json({
          success: false,
          error: "budget_change_pct is required",
        });
        return;
      }

      const prisma = getDbManager().getConnection(tenantId);
      const service = new MLAnalyticsService(prisma, tenantId);

      const data = await service.simulateBudget(product_ids, budget_change_pct);

      res.json({
        success: true,
        data,
      });
    } catch (error) {
      logger.error(`Budget simulation failed: ${error}`);
      res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
}

// Singleton instance
let controllerInstance: MLAnalyticsController | null = null;

export function getMLAnalyticsController(): MLAnalyticsController {
  if (!controllerInstance) {
    controllerInstance = new MLAnalyticsController();
  }
  return controllerInstance;
}

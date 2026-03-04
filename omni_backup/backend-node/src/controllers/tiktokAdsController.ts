/**
 * TikTok Ads Controller
 * Handle TikTok Ads analytics endpoints
 * Single Responsibility: HTTP request/response handling
 */

import { Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";
import { createTiktokAdsService } from "../services/analytics/tiktokAdsService";
import {
  UploadOptions,
  CreativeDataFilter,
} from "../services/analytics/tiktokAdsTypes";

const logger = getLogger("TiktokAdsController");

export class TiktokAdsController {
  /**
   * Validate tenant from request header
   */
  private validateTenant(req: Request, res: Response): string | null {
    const dbManager = getDbManager();
    const tenants = dbManager.getAllTenants();
    const headerTenantId = req.headers["x-tenant-id"] as string;

    if (!headerTenantId) {
      logger.warn("No tenant ID provided in request header");
      res.status(401).json({
        success: false,
        error: "Missing tenantId - authentication required",
        code: "TENANT_REQUIRED",
      });
      return null;
    }

    if (!tenants.includes(headerTenantId)) {
      logger.warn(`Invalid tenant ID: ${headerTenantId}`);
      res.status(400).json({
        success: false,
        error: `Invalid tenant: ${headerTenantId}`,
        code: "INVALID_TENANT",
      });
      return null;
    }

    return headerTenantId;
  }

  /**
   * Get Prisma client for tenant
   */
  private getPrisma(tenantId: string) {
    const dbManager = getDbManager();
    return dbManager.getConnection(tenantId);
  }

  /**
   * Upload Excel file
   * POST /api/analytics/tiktok-ads/upload
   */
  async uploadFile(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      // Access file from multer middleware
      const file = (req as Request & { file?: Express.Multer.File }).file;
      if (!file) {
        res.status(400).json({
          success: false,
          error: "No file uploaded",
        });
        return;
      }

      const mode = (req.body.mode as "skip" | "update") || "skip";
      // "automated" = fallback for logging when user not specified
      // NOT related to system.db tenant - per AGENTS.MD architecture
      const uploadedBy = req.body.uploadedBy || "automated";

      const options: UploadOptions = {
        mode,
        tenantId,
        uploadedBy,
      };

      const prisma = this.getPrisma(tenantId);
      const service = createTiktokAdsService(prisma, tenantId);

      const result = await service.uploadExcelFile(
        file.buffer,
        file.originalname,
        options,
      );

      logger.info(`Upload completed for tenant ${tenantId}`, {
        fileName: file.originalname,
        inserted: result.insertedRows,
        skipped: result.skippedRows,
      });

      res.json({
        success: result.success,
        data: result,
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error("Upload failed", { error: msg });
      res.status(500).json({
        success: false,
        error: msg,
      });
    }
  }

  /**
   * Get dashboard summary
   * GET /api/analytics/tiktok-ads/dashboard
   */
  async getDashboard(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const periodStart = req.query.periodStart
        ? new Date(req.query.periodStart as string)
        : undefined;
      const periodEnd = req.query.periodEnd
        ? new Date(req.query.periodEnd as string)
        : undefined;

      const prisma = this.getPrisma(tenantId);
      const service = createTiktokAdsService(prisma, tenantId);

      const summary = await service.getDashboardSummary(periodStart, periodEnd);

      res.json({
        success: true,
        data: summary,
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error("Get dashboard failed", { error: msg });
      res.status(500).json({
        success: false,
        error: msg,
      });
    }
  }

  /**
   * Get creative data with filters
   * GET /api/analytics/tiktok-ads/data
   */
  async getCreativeData(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const filter: CreativeDataFilter = {
        tenantId,
        periodStart: req.query.periodStart
          ? new Date(req.query.periodStart as string)
          : undefined,
        periodEnd: req.query.periodEnd
          ? new Date(req.query.periodEnd as string)
          : undefined,
        campaignId: req.query.campaignId as string,
        productId: req.query.productId as string,
        creativeType: req.query.creativeType as string,
        minRoi: req.query.minRoi ? Number(req.query.minRoi) : undefined,
        maxRoi: req.query.maxRoi ? Number(req.query.maxRoi) : undefined,
        limit: req.query.limit ? Number(req.query.limit) : 100,
        offset: req.query.offset ? Number(req.query.offset) : 0,
        orderBy: req.query.orderBy as "cost" | "revenue" | "roi" | "orders",
        orderDir: req.query.orderDir as "asc" | "desc",
      };

      const prisma = this.getPrisma(tenantId);
      const service = createTiktokAdsService(prisma, tenantId);

      const [data, count] = await Promise.all([
        service.getCreativeData(filter),
        service.getDataCount(filter.periodStart, filter.periodEnd),
      ]);

      res.json({
        success: true,
        data,
        pagination: {
          total: count,
          limit: filter.limit,
          offset: filter.offset,
        },
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error("Get creative data failed", { error: msg });
      res.status(500).json({
        success: false,
        error: msg,
      });
    }
  }

  /**
   * Get upload history
   * GET /api/analytics/tiktok-ads/uploads
   */
  async getUploadHistory(req: Request, res: Response): Promise<void> {
    try {
      const tenantId = this.validateTenant(req, res);
      if (!tenantId) return;

      const limit = req.query.limit ? Number(req.query.limit) : 20;

      const prisma = this.getPrisma(tenantId);
      const service = createTiktokAdsService(prisma, tenantId);

      const uploads = await service.getUploadHistory(limit);

      res.json({
        success: true,
        data: uploads,
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error("Get upload history failed", { error: msg });
      res.status(500).json({
        success: false,
        error: msg,
      });
    }
  }
}

// Singleton instance
let controllerInstance: TiktokAdsController | null = null;

export function getTiktokAdsController(): TiktokAdsController {
  if (!controllerInstance) {
    controllerInstance = new TiktokAdsController();
  }
  return controllerInstance;
}

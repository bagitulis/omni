/**
 * Shopee Ads Controller
 * Handle HTTP requests for Shopee Ads analytics
 * Single Responsibility: Request/Response handling only
 */

import { Request, Response } from "express";
import { getPrisma } from "../services/prismaClient";
import { ShopeeAdsService } from "../services/analytics/shopeeAdsService";
import { getLogger } from "../utils/logger";
import { ProductDataFilter } from "../services/analytics/shopeeAdsTypes";

// Extend Express Request type for user property
interface AuthenticatedRequest extends Request {
  user?: { username: string; role?: string };
}

const logger = getLogger("ShopeeAdsController");

class ShopeeAdsController {
  private getService(req: Request): ShopeeAdsService {
    const tenantId = req.tenantId;
    if (!tenantId) {
      throw new Error("Missing tenantId - authentication required");
    }
    const prisma = getPrisma(tenantId);
    return new ShopeeAdsService(prisma, tenantId);
  }

  async uploadFile(req: AuthenticatedRequest, res: Response): Promise<void> {
    try {
      if (!req.file) {
        res.status(400).json({ success: false, error: "No file uploaded" });
        return;
      }

      const service = this.getService(req);
      const mode = (req.body.mode as "skip" | "update") || "skip";
      const periodLabel = req.body.periodLabel as string | undefined;
      const uploadedBy = req.user?.username || "unknown";

      const result = await service.uploadCsvFile(
        req.file.buffer,
        req.file.originalname,
        { mode, uploadedBy, periodLabel },
      );

      if (result.success) {
        res.json({ success: true, data: result });
      } else {
        res.status(400).json({ success: false, error: result.errors[0] });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Upload failed";
      logger.error("Upload error", { error: message });
      res.status(500).json({ success: false, error: message });
    }
  }

  async getDashboard(req: Request, res: Response): Promise<void> {
    try {
      const service = this.getService(req);
      const { periodStart, periodEnd } = req.query;

      const dashboard = await service.getDashboard(
        periodStart as string,
        periodEnd as string,
      );

      res.json({ success: true, data: dashboard });
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Failed to get dashboard";
      logger.error("Dashboard error", { error: message });
      res.status(500).json({ success: false, error: message });
    }
  }

  async getProductData(req: Request, res: Response): Promise<void> {
    try {
      const service = this.getService(req);
      const filter: ProductDataFilter = {
        periodStart: req.query.periodStart as string,
        periodEnd: req.query.periodEnd as string,
        productId: req.query.productId as string,
        biddingMode: req.query.biddingMode as string,
        minRoas: req.query.minRoas
          ? parseFloat(req.query.minRoas as string)
          : undefined,
        maxRoas: req.query.maxRoas
          ? parseFloat(req.query.maxRoas as string)
          : undefined,
        limit: req.query.limit ? parseInt(req.query.limit as string) : 100,
        offset: req.query.offset ? parseInt(req.query.offset as string) : 0,
        orderBy: req.query.orderBy as string,
        orderDir: req.query.orderDir as "asc" | "desc",
      };

      const { data, total } = await service.getProductData(filter);

      res.json({
        success: true,
        data,
        pagination: { total, limit: filter.limit, offset: filter.offset },
      });
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Failed to get data";
      logger.error("Data error", { error: message });
      res.status(500).json({ success: false, error: message });
    }
  }

  async getUploadHistory(req: Request, res: Response): Promise<void> {
    try {
      const service = this.getService(req);
      const limit = req.query.limit ? parseInt(req.query.limit as string) : 20;

      const history = await service.getUploadHistory(limit);

      res.json({ success: true, data: history });
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Failed to get history";
      logger.error("History error", { error: message });
      res.status(500).json({ success: false, error: message });
    }
  }
}

// Singleton
let controllerInstance: ShopeeAdsController | null = null;

export function getShopeeAdsController(): ShopeeAdsController {
  if (!controllerInstance) {
    controllerInstance = new ShopeeAdsController();
  }
  return controllerInstance;
}

export { ShopeeAdsController };

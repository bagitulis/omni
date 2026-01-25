/**
 * TikTok Analytics Routes
 * Handles TikTok-specific analytics endpoints
 * Single Responsibility: TikTok Analytics API endpoints
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";
import { getTokenManager } from "../services/tokenManager";
import { TiktokEscrowSyncService } from "../services/analytics/tiktokEscrowSyncService";
import { TiktokPriceReconciliationService } from "../services/analytics/tiktokPriceReconciliationService";
import { TiktokShippingFeeAnalysisService } from "../services/analytics/tiktokShippingFeeAnalysisService";

const logger = getLogger("TiktokAnalyticsRoutes");
export const tiktokAnalyticsRouter = Router();

/**
 * GET /api/analytics/tiktok/sync-status
 * Get sync status for a month
 */
tiktokAnalyticsRouter.get(
  "/sync-status",
  async (req: Request, res: Response) => {
    try {
      const { month, year } = req.query;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!month || !year || !tenantId) {
        return res.status(400).json({
          success: false,
          error: "month, year, and x-tenant-id are required",
        });
      }

      const prisma = getDbManager().getConnection(tenantId);
      const tokenManager = await getTokenManager(tenantId);
      const tiktokApiClient = tokenManager.getTiktokClient();

      if (!tiktokApiClient) {
        return res.status(500).json({
          success: false,
          error: "TikTok API client not initialized",
        });
      }

      const syncService = new TiktokEscrowSyncService(
        prisma,
        tiktokApiClient,
        tenantId
      );

      const status = await syncService.getSyncStatus(
        parseInt(month as string),
        parseInt(year as string)
      );

      return res.json({
        success: true,
        data: {
          synced: status !== null,
          totalOrders: status?.totalOrders || 0,
          syncedAt: status?.syncedAt || null,
        },
      });
    } catch (error) {
      logger.error(`TikTok sync status failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * POST /api/analytics/tiktok/sync
 * Sync TikTok escrow data for a month
 */
tiktokAnalyticsRouter.post("/sync", async (req: Request, res: Response) => {
  try {
    const { month, year, forceResync } = req.body;
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!month || !year || !tenantId) {
      return res.status(400).json({
        success: false,
        error: "month, year, and x-tenant-id are required",
      });
    }

    // Prevent syncing current month
    const now = new Date();
    if (
      parseInt(month) === now.getMonth() + 1 &&
      parseInt(year) === now.getFullYear()
    ) {
      return res.status(400).json({
        success: false,
        error: "Cannot sync current month. Wait until month ends.",
      });
    }

    const prisma = getDbManager().getConnection(tenantId);
    const tokenManager = await getTokenManager(tenantId);
    const tiktokApiClient = tokenManager.getTiktokClient();

    if (!tiktokApiClient) {
      return res.status(500).json({
        success: false,
        error: "TikTok API client not initialized",
      });
    }

    const syncService = new TiktokEscrowSyncService(
      prisma,
      tiktokApiClient,
      tenantId
    );

    const result = await syncService.syncMonth(
      parseInt(month),
      parseInt(year),
      forceResync === true
    );

    return res.json({
      success: result.success,
      message: result.message,
      data: {
        totalOrders: result.totalOrders,
        totalItems: result.totalItems,
        failedOrders: result.failedOrders,
      },
    });
  } catch (error) {
    logger.error(`TikTok sync failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * DELETE /api/analytics/tiktok/sync
 * Delete sync data for a month
 */
tiktokAnalyticsRouter.delete("/sync", async (req: Request, res: Response) => {
  try {
    const { month, year } = req.query;
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!month || !year || !tenantId) {
      return res.status(400).json({
        success: false,
        error: "month, year, and x-tenant-id are required",
      });
    }

    const prisma = getDbManager().getConnection(tenantId);
    const tokenManager = await getTokenManager(tenantId);
    const tiktokApiClient = tokenManager.getTiktokClient();

    if (!tiktokApiClient) {
      return res.status(500).json({
        success: false,
        error: "TikTok API client not initialized",
      });
    }

    const syncService = new TiktokEscrowSyncService(
      prisma,
      tiktokApiClient,
      tenantId
    );

    await syncService.deleteMonthData(
      parseInt(month as string),
      parseInt(year as string)
    );

    return res.json({
      success: true,
      message: "Sync data deleted successfully",
    });
  } catch (error) {
    logger.error(`TikTok delete sync failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * GET /api/analytics/tiktok/reconciliation
 * Get TikTok price reconciliation analysis
 */
tiktokAnalyticsRouter.get(
  "/reconciliation",
  async (req: Request, res: Response) => {
    try {
      const { month, year } = req.query;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!month || !year || !tenantId) {
        return res.status(400).json({
          success: false,
          error: "month, year, and x-tenant-id are required",
        });
      }

      const prisma = getDbManager().getConnection(tenantId);
      const service = new TiktokPriceReconciliationService(prisma, tenantId);

      const result = await service.analyze(
        parseInt(month as string),
        parseInt(year as string)
      );

      return res.json({
        success: true,
        data: result,
      });
    } catch (error) {
      logger.error(`TikTok reconciliation failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * GET /api/analytics/tiktok/settings
 * Get TikTok analytics settings
 */
tiktokAnalyticsRouter.get("/settings", async (req: Request, res: Response) => {
  try {
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!tenantId) {
      return res.status(400).json({
        success: false,
        error: "x-tenant-id is required",
      });
    }

    const prisma = getDbManager().getConnection(tenantId);
    const service = new TiktokPriceReconciliationService(prisma, tenantId);
    const settings = await service.getSettings();

    return res.json({
      success: true,
      data: settings,
    });
  } catch (error) {
    logger.error(`TikTok get settings failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * POST /api/analytics/tiktok/settings
 * Save TikTok analytics settings
 */
tiktokAnalyticsRouter.post("/settings", async (req: Request, res: Response) => {
  try {
    const tenantId = req.headers["x-tenant-id"] as string;
    const { priceColumn, formulaDeduction, formulaMultiplier } = req.body;

    if (!tenantId) {
      return res.status(400).json({
        success: false,
        error: "x-tenant-id is required",
      });
    }

    const prisma = getDbManager().getConnection(tenantId);
    const service = new TiktokPriceReconciliationService(prisma, tenantId);

    await service.saveSettings({
      priceColumn,
      formulaDeduction,
      formulaMultiplier,
    });

    return res.json({
      success: true,
      message: "Settings saved successfully",
    });
  } catch (error) {
    logger.error(`TikTok save settings failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * GET /api/analytics/tiktok/shipping-fee
 * Get TikTok shipping fee analysis
 */
tiktokAnalyticsRouter.get(
  "/shipping-fee",
  async (req: Request, res: Response) => {
    try {
      const { month, year } = req.query;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!month || !year || !tenantId) {
        return res.status(400).json({
          success: false,
          error: "month, year, and x-tenant-id are required",
        });
      }

      const prisma = getDbManager().getConnection(tenantId);
      const service = new TiktokShippingFeeAnalysisService(prisma, tenantId);

      const result = await service.analyzeMonth(
        parseInt(month as string),
        parseInt(year as string)
      );

      return res.json({
        success: true,
        data: result,
      });
    } catch (error) {
      logger.error(`TikTok shipping fee analysis failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default tiktokAnalyticsRouter;

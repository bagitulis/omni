/**
 * Analytics Routes
 * Handles analytics endpoints for price reconciliation
 * Single Responsibility: Analytics API endpoints
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { getDbManager } from "../services/dbConnectionManager";
import { getTokenManager } from "../services/tokenManager";
import { ShopeeEscrowSyncService } from "../services/analytics/shopeeEscrowSyncService";
import { PriceReconciliationService } from "../services/analytics/priceReconciliationService";
import { ShippingFeeAnalysisService } from "../services/analytics/shippingFeeAnalysisService";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";

const logger = getLogger("AnalyticsRoutes");
export const analyticsRouter = Router();

/**
 * GET /api/analytics/shopee/sync-status
 * Get sync status for a month
 */
analyticsRouter.get(
  "/shopee/sync-status",
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
      const shopeeApiClient = tokenManager.getShopeeClient();

      if (!shopeeApiClient) {
        return res.status(500).json({
          success: false,
          error: "Shopee API client not initialized",
        });
      }

      const syncService = new ShopeeEscrowSyncService(
        prisma,
        shopeeApiClient,
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
      logger.error(`Sync status failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * POST /api/analytics/shopee/sync
 * Sync escrow data for a month
 */
analyticsRouter.post("/shopee/sync", async (req: Request, res: Response) => {
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
    const shopeeApiClient = tokenManager.getShopeeClient();

    if (!shopeeApiClient) {
      return res.status(500).json({
        success: false,
        error: "Shopee API client not initialized",
      });
    }

    // First get wallet transactions
    const orderManager = new ShopeeOrderManager(shopeeApiClient);
    const walletTransactions = await orderManager.wallet.getTransactions(
      parseInt(month),
      parseInt(year),
      "wallet_order_income"
    );

    if (!walletTransactions || walletTransactions.length === 0) {
      return res.json({
        success: true,
        message: "No transactions found for this period",
        data: { totalOrders: 0, totalItems: 0 },
      });
    }

    // Sync escrow data
    const syncService = new ShopeeEscrowSyncService(
      prisma,
      shopeeApiClient,
      tenantId
    );

    const result = await syncService.syncMonth(
      parseInt(month),
      parseInt(year),
      walletTransactions,
      forceResync === true
    );

    return res.json({
      success: result.success,
      message: result.message,
      data: {
        totalOrders: result.totalOrders,
        totalItems: result.totalItems,
      },
    });
  } catch (error) {
    logger.error(`Sync failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * DELETE /api/analytics/shopee/sync
 * Delete sync data for a month
 */
analyticsRouter.delete("/shopee/sync", async (req: Request, res: Response) => {
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
    const shopeeApiClient = tokenManager.getShopeeClient();

    if (!shopeeApiClient) {
      return res.status(500).json({
        success: false,
        error: "Shopee API client not initialized",
      });
    }

    const syncService = new ShopeeEscrowSyncService(
      prisma,
      shopeeApiClient,
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
    logger.error(`Shopee delete sync failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * GET /api/analytics/shopee/reconciliation
 * Get price reconciliation analysis
 */
analyticsRouter.get(
  "/shopee/reconciliation",
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
      const service = new PriceReconciliationService(prisma, tenantId);

      const result = await service.analyze(
        parseInt(month as string),
        parseInt(year as string)
      );

      return res.json({
        success: true,
        data: result,
      });
    } catch (error) {
      logger.error(`Reconciliation failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * GET /api/analytics/shopee/settings
 * Get analytics settings
 */
analyticsRouter.get("/shopee/settings", async (req: Request, res: Response) => {
  try {
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!tenantId) {
      return res.status(400).json({
        success: false,
        error: "x-tenant-id is required",
      });
    }

    const prisma = getDbManager().getConnection(tenantId);
    const service = new PriceReconciliationService(prisma, tenantId);
    const settings = await service.getSettings();

    return res.json({
      success: true,
      data: settings,
    });
  } catch (error) {
    logger.error(`Get settings failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * POST /api/analytics/shopee/settings
 * Save analytics settings
 */
analyticsRouter.post(
  "/shopee/settings",
  async (req: Request, res: Response) => {
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
      const service = new PriceReconciliationService(prisma, tenantId);

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
      logger.error(`Save settings failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * GET /api/analytics/shopee/shipping-fee
 * Get shipping fee analysis (only orders with differences)
 */
analyticsRouter.get(
  "/shopee/shipping-fee",
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
      const service = new ShippingFeeAnalysisService(prisma, tenantId);

      const result = await service.analyzeMonth(
        parseInt(month as string),
        parseInt(year as string)
      );

      return res.json({
        success: true,
        data: result,
      });
    } catch (error) {
      logger.error(`Shipping fee analysis failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default analyticsRouter;

/**
 * Wholesale Routes
 * RESPONSIBILITY: Define API routes for wholesale operations
 *
 * Routes:
 * - DELETE /api/wholesale/shopee/:itemId - Delete wholesale tiers
 * - PUT /api/wholesale/shopee/:itemId - Update wholesale tiers
 * - GET /api/wholesale/shopee/:itemId - Get wholesale info
 * - POST /api/wholesale/shopee/batch-delete - Batch delete by itemIds
 * - POST /api/wholesale/shopee/batch-delete-skus - Batch delete by SKUs (with dedupe)
 * - GET /api/wholesale/shopee/lookup/:sku - Lookup itemId by SKU
 * - GET /api/wholesale/settings - Get tenant settings
 * - PUT /api/wholesale/settings - Update tenant settings
 * - POST /api/wholesale/preview - Preview calculated tiers
 * - POST /api/wholesale/shopee/batch-update-skus - Batch update by SKUs (with formula)
 * - POST /api/wholesale/shopee/batch-mpq - Batch set MPQ for Shopee
 * - POST /api/wholesale/shopee/batch-wholesale-reset - Batch wholesale with MPQ reset
 * - POST /api/wholesale/tiktok/batch-mpq - Batch set MPQ for TikTok
 */

import { Router, Response } from "express";
import { WholesaleController } from "../controllers/wholesaleController";
import { WholesaleBatchController } from "../controllers/wholesaleBatchController";
import { WholesaleSettingsController } from "../controllers/wholesaleSettingsController";
import { MpqController } from "../controllers/mpqController";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { getDbManager } from "../services/dbConnectionManager";
import { getLogger } from "../utils/logger";

const router = Router();
const logger = getLogger("WholesaleRoutes");

/**
 * Create tenant-aware Shopee config and client
 */
async function getShopeeServices(tenantId: string) {
  const shopeeConfig = new ShopeeConfigManager(tenantId);
  await shopeeConfig.loadConfig();
  const shopeeClient = new ShopeeAPIClient(shopeeConfig);
  return { shopeeConfig, shopeeClient };
}

/**
 * Create tenant-aware TikTok config and client
 */
async function getTiktokServices(tenantId: string) {
  const tiktokConfig = new TiktokConfigManager(tenantId);
  await tiktokConfig.loadConfig();
  const tiktokClient = new TiktokAPIClient(tiktokConfig);
  return { tiktokConfig, tiktokClient };
}

/**
 * Get wholesale controller with tenant-aware Prisma and credentials
 */
async function getController(req: AuthRequest): Promise<WholesaleController> {
  if (!req.tenantId) {
    throw new Error("Missing tenantId in request - authentication required");
  }
  const tenantId = req.tenantId;
  const prisma = getDbManager().getConnection(tenantId);
  const { shopeeConfig, shopeeClient } = await getShopeeServices(tenantId);
  logger.info(
    `[${tenantId}] Created wholesale controller with tenant credentials`
  );
  return new WholesaleController(shopeeClient, shopeeConfig, prisma);
}

/**
 * Get MPQ controller with tenant-aware Prisma and credentials
 */
async function getMpqController(req: AuthRequest): Promise<MpqController> {
  if (!req.tenantId) {
    throw new Error("Missing tenantId in request - authentication required");
  }
  const tenantId = req.tenantId;
  const prisma = getDbManager().getConnection(tenantId);
  const { shopeeConfig, shopeeClient } = await getShopeeServices(tenantId);
  const { tiktokConfig, tiktokClient } = await getTiktokServices(tenantId);
  return new MpqController(
    shopeeClient,
    shopeeConfig,
    tiktokClient,
    tiktokConfig,
    prisma
  );
}

/**
 * Get settings controller with tenant-aware Prisma
 */
function getSettingsController(req: AuthRequest): WholesaleSettingsController {
  if (!req.tenantId) {
    throw new Error("Missing tenantId in request - authentication required");
  }
  const tenantId = req.tenantId;
  const prisma = getDbManager().getConnection(tenantId);
  return new WholesaleSettingsController(prisma);
}

/**
 * Get batch controller with tenant-aware Prisma and credentials
 */
async function getBatchController(
  req: AuthRequest
): Promise<WholesaleBatchController> {
  if (!req.tenantId) {
    throw new Error("Missing tenantId in request - authentication required");
  }
  const tenantId = req.tenantId;
  const prisma = getDbManager().getConnection(tenantId);
  const { shopeeConfig, shopeeClient } = await getShopeeServices(tenantId);
  return new WholesaleBatchController(shopeeClient, shopeeConfig, prisma);
}

/**
 * Load config on startup - DISABLED
 * Config now loaded inside each request handler with tenant context
 * This prevents startup errors when no tenant context is available
 */
// async function loadConfigs() {...}
// loadConfigs(); // REMOVED - causes startup error

// ============================================
// SHOPEE WHOLESALE ROUTES
// ============================================

/**
 * DELETE /api/wholesale/shopee/:itemId
 * Delete all wholesale tiers for a product
 */
router.delete("/shopee/:itemId", async (req: AuthRequest, res: Response) => {
  const controller = await getController(req);
  return controller.deleteShopeeWholesale(req, res);
});

/**
 * PUT /api/wholesale/shopee/:itemId
 * Update wholesale tiers for a product
 * Body: { tiers: [{ minCount, maxCount, unitPrice }] }
 */
router.put("/shopee/:itemId", async (req: AuthRequest, res: Response) => {
  const controller = await getController(req);
  return controller.updateShopeeWholesale(req, res);
});

/**
 * GET /api/wholesale/shopee/:itemId
 * Get current wholesale info for a product
 */
router.get("/shopee/:itemId", async (req: AuthRequest, res: Response) => {
  const controller = await getController(req);
  return controller.getShopeeWholesale(req, res);
});

/**
 * POST /api/wholesale/shopee/batch-delete
 * Batch delete wholesale for multiple products by itemIds
 * Body: { itemIds: number[] }
 */
router.post("/shopee/batch-delete", async (req: AuthRequest, res: Response) => {
  const controller = await getBatchController(req);
  return controller.batchDelete(req, res);
});

/**
 * POST /api/wholesale/shopee/batch-delete-skus
 * Batch delete wholesale by SKUs (with automatic deduplication)
 * IMPORTANT: Multiple SKUs may share same item_id - this dedupes automatically
 * Body: { skus: string[] }
 */
router.post(
  "/shopee/batch-delete-skus",
  async (req: AuthRequest, res: Response) => {
    const controller = await getBatchController(req);
    return controller.batchDeleteBySkus(req, res);
  }
);

/**
 * GET /api/wholesale/shopee/lookup/:sku
 * Lookup item_id from SKU
 */
router.get("/shopee/lookup/:sku", async (req: AuthRequest, res: Response) => {
  const controller = await getController(req);
  return controller.lookupItemBySku(req, res);
});

// ============================================
// WHOLESALE SETTINGS ROUTES
// ============================================

/**
 * GET /api/wholesale/settings
 * Get wholesale settings for current tenant
 */
router.get("/settings", async (req: AuthRequest, res: Response) => {
  const controller = getSettingsController(req);
  return controller.getSettings(req, res);
});

/**
 * PUT /api/wholesale/settings
 * Update wholesale settings for current tenant
 * Body: { adminFee?, minOrder1?, maxOrder1?, maxOrderTier3? }
 */
router.put("/settings", async (req: AuthRequest, res: Response) => {
  const controller = getSettingsController(req);
  return controller.updateSettings(req, res);
});

/**
 * POST /api/wholesale/preview
 * Preview calculated tiers for a given base price
 * Body: { basePrice: number, settings?: {...} }
 */
router.post("/preview", async (req: AuthRequest, res: Response) => {
  const controller = getSettingsController(req);
  return controller.previewTiers(req, res);
});

/**
 * POST /api/wholesale/shopee/batch-update-skus
 * Batch update wholesale by SKUs with auto-calculated tiers
 * Body: { items: [{ sku: string, price: number }] }
 */
router.post(
  "/shopee/batch-update-skus",
  async (req: AuthRequest, res: Response) => {
    const controller = await getBatchController(req);
    return controller.batchUpdateBySkus(req, res);
  }
);

// ============================================
// MPQ ROUTES (Shopee + TikTok)
// ============================================

/**
 * POST /api/wholesale/shopee/batch-mpq
 * Batch set MPQ mode for Shopee (delete wholesale, set price, set MPQ)
 * Body: { items: [{ sku, price }], mpq: number }
 */
router.post("/shopee/batch-mpq", async (req: AuthRequest, res: Response) => {
  const controller = await getMpqController(req);
  return controller.batchShopeeMpq(req, res);
});

/**
 * POST /api/wholesale/shopee/batch-wholesale-reset
 * Batch update wholesale with MPQ reset first (MPQ=1, then set tiers)
 * Body: { items: [{ sku, price }] }
 */
router.post(
  "/shopee/batch-wholesale-reset",
  async (req: AuthRequest, res: Response) => {
    const controller = await getMpqController(req);
    return controller.batchShopeeWholesaleWithReset(req, res);
  }
);

/**
 * POST /api/wholesale/tiktok/batch-mpq
 * Batch set MPQ for TikTok products
 * Body: { items: [{ sku, price }], mpq: number }
 */
router.post("/tiktok/batch-mpq", async (req: AuthRequest, res: Response) => {
  const controller = await getMpqController(req);
  return controller.batchTiktokMpq(req, res);
});

export default router;

import { Router, Response } from "express";
import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { PriceUpdateController } from "../controllers/priceUpdateController";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("PriceRoutes");
const router = Router();

// Apply authentication middleware to all routes
router.use(authMiddleware);
router.use(requireAuth);

/**
 * Create controller with tenant-specific DB connection and credentials
 * Creates NEW config managers per request with explicit tenantId to ensure
 * correct credentials are loaded for the requesting tenant
 */
async function createController(
  req: AuthRequest
): Promise<PriceUpdateController> {
  const tenantId = req.tenantId || tenantContext.getTenantId();
  const tenantPrisma = getDbManager().getConnection(tenantId);

  // Create tenant-specific config managers (NOT singletons!)
  const shopeeConfig = new ShopeeConfigManager(tenantId);
  const lazadaConfig = new LazadaConfigManager(tenantId);
  const tiktokConfig = new TiktokConfigManager(tenantId);

  // Load credentials for THIS tenant
  await Promise.all([
    shopeeConfig.loadConfig(),
    lazadaConfig.loadConfig(),
    tiktokConfig.loadConfig(),
  ]);

  logger.info(`[${tenantId}] Loaded platform credentials for price update`);

  // Create API clients with tenant-specific credentials
  const shopeeClient = new ShopeeAPIClient(shopeeConfig);
  const lazadaClient = new LazadaAPIClient(lazadaConfig);
  const tiktokClient = new TiktokAPIClient(tiktokConfig);

  return new PriceUpdateController(
    tenantPrisma,
    shopeeClient,
    lazadaClient,
    tiktokClient,
    shopeeConfig,
    lazadaConfig,
    tiktokConfig
  );
}

// ============================================
// PRICE UPDATE ROUTES
// ============================================

/**
 * POST /api/inventory/update-price
 * Update price for a single SKU across platforms
 *
 * Body: { sku: string, price: number, platforms?: string[] }
 */
router.post("/update-price", async (req: AuthRequest, res: Response) => {
  const controller = await createController(req);
  return controller.updateSinglePrice(req, res);
});

/**
 * POST /api/inventory/update-price-batch
 * Batch update price for multiple SKUs
 *
 * Body: { items: [{ sku: string, price: number, platforms?: string[] }] }
 */
router.post("/update-price-batch", async (req: AuthRequest, res: Response) => {
  const controller = await createController(req);
  return controller.updateBatchPrice(req, res);
});

export default router;

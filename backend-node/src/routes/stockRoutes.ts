import { Router, Response } from "express";
import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { StockUpdateController } from "../controllers/stockUpdateController";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("StockRoutes");
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
): Promise<StockUpdateController> {
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

  logger.info(`[${tenantId}] Loaded platform credentials for stock update`);

  // Create API clients with tenant-specific credentials
  const shopeeClient = new ShopeeAPIClient(shopeeConfig);
  const lazadaClient = new LazadaAPIClient(lazadaConfig);
  const tiktokClient = new TiktokAPIClient(tiktokConfig);

  return new StockUpdateController(
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
// STOCK UPDATE ROUTES
// ============================================

/**
 * POST /api/inventory/update-stock
 * Update stock for a single SKU across platforms
 *
 * Body: { sku: string, stock: number, platforms?: string[] }
 */
router.post("/update-stock", async (req: AuthRequest, res: Response) => {
  const controller = await createController(req);
  return controller.updateSingleStock(req, res);
});

/**
 * POST /api/inventory/update-stock-batch
 * Batch update stock for multiple SKUs
 *
 * Body: { items: [{ sku: string, stock: number, platforms?: string[] }] }
 */
router.post("/update-stock-batch", async (req: AuthRequest, res: Response) => {
  const controller = await createController(req);
  return controller.updateBatchStock(req, res);
});

/**
 * POST /api/inventory/lookup-platform-ids
 * Lookup platform product IDs for a SKU (debugging)
 *
 * Body: { sku: string }
 */
router.post("/lookup-platform-ids", async (req: AuthRequest, res: Response) => {
  const controller = await createController(req);
  return controller.lookupPlatformIds(req, res);
});

export default router;

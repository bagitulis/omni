/**
 * Products Route Hub
 * SRP: Main router that delegates to platform-specific routes
 * Individual routes split into separate files for maintainability
 * TENANT-AWARE: Creates fresh config managers per request with explicit tenantId
 */

import { Router, Response } from "express";
import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { LazadaProductController } from "../controllers/lazadaProductController";
import { TiktokProductController } from "../controllers/tiktokProductController";
import { TiktokProductCreateController } from "../controllers/tiktokProductCreateController";
import { ShopeeProductDataController } from "../controllers/shopeeProductDataController";
import { ShopeeProductSyncController } from "../controllers/shopeeProductSyncController";
import { ShopeeProductMasterController } from "../controllers/shopeeProductMasterController";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { getLogger } from "../utils/logger";

const router = Router();
const logger = getLogger("ProductRoutes");

router.use(authMiddleware);
router.use(requireAuth);

/**
 * Helper function to create controllers with tenant-specific DB connection and credentials
 * Creates NEW config managers per request with explicit tenantId
 */
export async function createProductControllers(req: AuthRequest) {
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

  logger.info(
    `[${tenantId}] Loaded platform credentials for product operations`,
  );

  // Create API clients with tenant-specific credentials
  const shopeeClient = new ShopeeAPIClient(shopeeConfig);
  const lazadaClient = new LazadaAPIClient(lazadaConfig);
  const tiktokClient = new TiktokAPIClient(tiktokConfig);

  return {
    lazada: new LazadaProductController(
      tenantPrisma,
      lazadaClient,
      lazadaConfig,
    ),
    tiktok: new TiktokProductController(
      tenantPrisma,
      tiktokClient,
      tiktokConfig,
    ),
    tiktokCreate: new TiktokProductCreateController(tiktokClient, tiktokConfig),
    shopeeData: new ShopeeProductDataController(
      tenantPrisma,
      shopeeClient,
      shopeeConfig,
    ),
    shopeeSync: new ShopeeProductSyncController(
      tenantPrisma,
      shopeeClient,
      shopeeConfig,
    ),
    shopeeMaster: new ShopeeProductMasterController(
      tenantPrisma,
      shopeeClient,
      shopeeConfig,
    ),
  };
}

// ============================================
// LAZADA PRODUCT ROUTES
// ============================================

router.get("/lazada/products", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.lazada.getProducts(req, res);
});

router.get("/lazada/db/products", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.lazada.getMasterProductsFromDb(req, res);
});

router.get(
  "/lazada/db/products/list",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.lazada.getProductListFromDb(req, res);
  },
);

router.get(
  "/lazada/product/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.lazada.getProductDetail(req, res);
  },
);

// ============================================
// TIKTOK PRODUCT ROUTES
// ============================================

router.post(
  "/tiktok/products/search",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.searchProducts(req, res);
  },
);

router.post(
  "/tiktok/products/get-all",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getAllProducts(req, res);
  },
);

router.get(
  "/tiktok/products/:productId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getProductDetail(req, res);
  },
);

router.get(
  "/tiktok/db/products/list",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getProductsFromDb(req, res);
  },
);

router.get(
  "/tiktok/db/products/master",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getMasterProductsFromDb(req, res);
  },
);

router.get(
  "/tiktok/db/products/:productId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getProductByIdFromDb(req, res);
  },
);

router.get("/tiktok/db/search", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.tiktok.searchProductsInDb(req, res);
});

router.get(
  "/tiktok/db/products/status/:status",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.getProductsByStatus(req, res);
  },
);

router.get("/tiktok/db/statistics", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.tiktok.getStatistics(req, res);
});

router.delete(
  "/tiktok/db/products/:productId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktok.deleteProductFromDb(req, res);
  },
);

// ============================================
// TIKTOK CREATE PRODUCT ROUTES
// ============================================

// Get category tree for product creation
router.get("/tiktok/categories", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.tiktokCreate.getCategories(req, res);
});

// Get category attributes (required fields)
router.get(
  "/tiktok/categories/:categoryId/attributes",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.getCategoryAttributes(req, res);
  },
);

// Search categories by keyword
router.get(
  "/tiktok/categories/search",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.searchCategories(req, res);
  },
);

// Recommend category based on title and images
router.post(
  "/tiktok/categories/recommend",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.recommendCategory(req, res);
  },
);

// Get popular categories
router.get(
  "/tiktok/categories/popular",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.getPopularCategories(req, res);
  },
);

// Get brands for a category
router.get("/tiktok/brands", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.tiktokCreate.getBrands(req, res);
});

// Get delivery options
router.get(
  "/tiktok/delivery-options",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.getDeliveryOptions(req, res);
  },
);

// Get warehouse list
router.get("/tiktok/warehouses", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.tiktokCreate.getWarehouses(req, res);
});

// Upload image to TikTok
router.post(
  "/tiktok/images/upload",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.uploadImage(req, res);
  },
);

// Create new product on TikTok Shop
router.post(
  "/tiktok/products/create",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.tiktokCreate.createProduct(req, res);
  },
);

// ============================================
// SHOPEE PRODUCT ROUTES
// ============================================

router.get("/shopee/products", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeData.getProducts(req, res);
});

router.get("/shopee/db/products", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeData.getMasterProductsFromDb(req, res);
});

router.get(
  "/shopee/db/products/master",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeMaster.getMasterProductList(req, res);
  },
);

router.get(
  "/shopee/product/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductDetail(req, res);
  },
);

router.get(
  "/shopee/db/products/list",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductList(req, res);
  },
);

router.get(
  "/shopee/db/products/base",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductBaseList(req, res);
  },
);

router.get(
  "/shopee/db/products/model",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductModelList(req, res);
  },
);

router.get(
  "/shopee/db/products/base/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductBase(req, res);
  },
);

router.get(
  "/shopee/db/products/models/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductModels(req, res);
  },
);

router.get(
  "/shopee/db/products/variations/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductVariations(req, res);
  },
);

router.get(
  "/shopee/db/products/full/:itemId",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeData.getProductFull(req, res);
  },
);

router.post(
  "/shopee/db/products/list/fetch-from-api",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeSync.syncProducts(req, res);
  },
);

router.post(
  "/shopee/db/products/base/batch-fetch",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeSync.batchSync(req, res);
  },
);

router.post(
  "/shopee/db/products/model/batch-fetch",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeSync.batchSync(req, res);
  },
);

router.get(
  "/shopee/db/sync/unprocessed",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeSync.getUnprocessedItems(req, res);
  },
);

router.get(
  "/shopee/db/sync/no-models",
  async (req: AuthRequest, res: Response) => {
    const controllers = await createProductControllers(req);
    return controllers.shopeeSync.getItemsWithoutModels(req, res);
  },
);

router.get("/shopee/db/sync/logs", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeSync.getSyncLogs(req, res);
});

// ============================================
// MASTER PRODUCT ROUTES
// ============================================

router.get("/master-product/list", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeMaster.getMasterProductList(req, res);
});

router.get("/master-product/stats", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeMaster.getMasterProductStats(req, res);
});

router.post("/master-product/sync", async (req: AuthRequest, res: Response) => {
  const controllers = await createProductControllers(req);
  return controllers.shopeeMaster.syncMasterProduct(req, res);
});

export default router;

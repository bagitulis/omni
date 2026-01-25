/**
 * Product Create Routes
 * Handles product creation routes for all platforms (TikTok, Shopee, Lazada)
 * Separated to comply with AGENTS.MD max 300 lines rule
 */

import { Router, Response } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { ShopeeProductCreateController } from "../controllers/shopeeProductCreateController";
import { LazadaProductCreateController } from "../controllers/lazadaProductCreateController";

const router = Router();

router.use(authMiddleware);
router.use(requireAuth);

/** Helper to create create-product controllers */
async function createControllers(req: AuthRequest) {
  const tenantId = req.tenantId || tenantContext.getTenantId();
  const prisma = getDbManager().getConnection(tenantId);

  const shopeeConfig = new ShopeeConfigManager(tenantId);
  const lazadaConfig = new LazadaConfigManager(tenantId);

  await Promise.all([shopeeConfig.loadConfig(), lazadaConfig.loadConfig()]);

  return {
    shopee: new ShopeeProductCreateController(prisma, new ShopeeAPIClient(shopeeConfig), shopeeConfig),
    lazada: new LazadaProductCreateController(prisma, new LazadaAPIClient(lazadaConfig), lazadaConfig),
  };
}

// ============================================
// SHOPEE CREATE PRODUCT ROUTES
// ============================================

router.get("/shopee/categories", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.getCategories(req, res);
});

router.get("/shopee/categories/:categoryId/attributes", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.getCategoryAttributes(req, res);
});

router.get("/shopee/categories/search", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.searchCategories(req, res);
});

router.get("/shopee/categories/recommend", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.recommendCategory(req, res);
});

router.get("/shopee/categories/popular", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.getPopularCategories(req, res);
});

router.get("/shopee/brands", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.getBrands(req, res);
});

router.get("/shopee/logistics", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.getLogistics(req, res);
});

router.post("/shopee/images/upload", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.uploadImage(req, res);
});

router.post("/shopee/products/create", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.shopee.createProduct(req, res);
});

// ============================================
// LAZADA CREATE PRODUCT ROUTES
// ============================================

router.get("/lazada/categories", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.getCategories(req, res);
});

router.get("/lazada/categories/:categoryId/attributes", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.getCategoryAttributes(req, res);
});

router.get("/lazada/categories/search", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.searchCategories(req, res);
});

router.get("/lazada/categories/recommend", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.recommendCategory(req, res);
});

router.get("/lazada/categories/popular", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.getPopularCategories(req, res);
});

router.get("/lazada/brands", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.getBrands(req, res);
});

router.post("/lazada/images/upload", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.uploadImage(req, res);
});

router.post("/lazada/products/create", async (req: AuthRequest, res: Response) => {
  const c = await createControllers(req);
  return c.lazada.createProduct(req, res);
});

export default router;

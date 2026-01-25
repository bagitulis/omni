/**
 * Product Clone Routes
 * Handles product cloning between platforms (Shopee ↔ Lazada ↔ TikTok)
 * Fetches product data from source platform for cloning to target platform
 * Max 300 lines - AGENTS.MD compliant
 */

import { Router, Response } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { getDbManager } from "../services/dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import { sendSuccess, sendBadRequest, handleError } from "../utils/apiResponse";

const router = Router();
const logger = getLogger("ProductCloneRoutes");

type Platform = "shopee" | "lazada" | "tiktok";

interface CloneProductData {
  title: string;
  description: string;
  images: string[];
  skus: Array<{
    sellerSku: string;
    price: number;
    stock: number;
    variantLabel?: string;
  }>;
  weight?: number;
  packageDimensions?: {
    length: number;
    width: number;
    height: number;
  };
  categoryId?: string;
  brandId?: string;
  attributes?: Record<string, any>;
}

router.use(authMiddleware);
router.use(requireAuth);

/**
 * GET /api/clone/product-data
 * Fetch product details from source platform for cloning
 * Query params: platform (source), sku
 */
router.get("/clone/product-data", async (req: AuthRequest, res: Response) => {
  try {
    const { platform, sku } = req.query;

    if (!platform || !sku) {
      return sendBadRequest(res, "platform and sku are required");
    }

    const sourcePlatform = (platform as string).toLowerCase() as Platform;
    const skuValue = sku as string;
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const prisma = getDbManager().getConnection(tenantId);

    logger.info(`[${tenantId}] Fetching product data for SKU: ${skuValue} from ${sourcePlatform}`);

    let productData: CloneProductData | null = null;

    switch (sourcePlatform) {
      case "shopee":
        productData = await fetchShopeeProduct(prisma, skuValue);
        break;
      case "lazada":
        productData = await fetchLazadaProduct(prisma, skuValue);
        break;
      case "tiktok":
        productData = await fetchTiktokProduct(prisma, skuValue);
        break;
      default:
        return sendBadRequest(res, `Invalid platform: ${platform}`);
    }

    if (!productData) {
      return res.status(404).json({
        success: false,
        error: `Product with SKU ${skuValue} not found in ${sourcePlatform}`,
      });
    }

    logger.info(`[${tenantId}] Found product: ${productData.title}`);

    return sendSuccess(res, {
      sourcePlatform,
      sku: skuValue,
      product: productData,
    });
  } catch (error) {
    logger.error("[Product Clone] Error fetching product data:", error);
    return handleError(res, error, "Failed to fetch product data for cloning");
  }
});

/**
 * GET /api/clone/available-targets
 * Get available target platforms for a given SKU (platforms where product doesn't exist)
 * Query params: sku
 */
router.get("/clone/available-targets", async (req: AuthRequest, res: Response) => {
  try {
    const { sku } = req.query;

    if (!sku) {
      return sendBadRequest(res, "sku is required");
    }

    const skuValue = sku as string;
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const prisma = getDbManager().getConnection(tenantId);

    // Check which platforms have this SKU
    const [shopeeExists, lazadaExists, tiktokExists] = await Promise.all([
      checkShopeeSkuExists(prisma, skuValue),
      checkLazadaSkuExists(prisma, skuValue),
      checkTiktokSkuExists(prisma, skuValue),
    ]);

    // Available sources are platforms where product EXISTS
    const sources: Platform[] = [];
    if (shopeeExists) sources.push("shopee");
    if (lazadaExists) sources.push("lazada");
    if (tiktokExists) sources.push("tiktok");

    // Available targets are platforms where product DOESN'T exist
    const targets: Platform[] = [];
    if (!shopeeExists) targets.push("shopee");
    if (!lazadaExists) targets.push("lazada");
    if (!tiktokExists) targets.push("tiktok");

    return sendSuccess(res, {
      sku: skuValue,
      sources,
      targets,
      status: {
        shopee: shopeeExists,
        lazada: lazadaExists,
        tiktok: tiktokExists,
      },
    });
  } catch (error) {
    logger.error("[Clone Targets] Error:", error);
    return handleError(res, error, "Failed to get available clone targets");
  }
});

// ============================================
// HELPER FUNCTIONS - Fetch product from DB
// ============================================

async function fetchShopeeProduct(prisma: any, sku: string): Promise<CloneProductData | null> {
  // Find SKU in ShopeeSku table
  const shopeeSku = await prisma.shopeeSku.findFirst({
    where: { sellerSku: sku },
    include: { product: true },
  });

  if (!shopeeSku || !shopeeSku.product) return null;

  const product = shopeeSku.product;

  // Get all SKUs for this product
  const allSkus = await prisma.shopeeSku.findMany({
    where: { productId: product.id },
  });

  // Parse image - database stores 'image' (single), convert to array
  let images: string[] = [];
  if (product.image) {
    try {
      const parsed = JSON.parse(product.image);
      images = Array.isArray(parsed) ? parsed : [product.image];
    } catch {
      images = [product.image];
    }
  }

  return {
    title: product.name || product.itemName || "",
    description: product.description || "",
    images,
    skus: allSkus.map((s: any) => ({
      sellerSku: s.sellerSku || "",
      price: s.price || 0,
      stock: s.quantity || 0,
      variantLabel: s.variantName || undefined,
    })),
    weight: product.weight || undefined,
    packageDimensions: product.dimensions ? JSON.parse(product.dimensions) : undefined,
    categoryId: product.categoryId ? String(product.categoryId) : undefined,
    brandId: product.brandId ? String(product.brandId) : undefined,
  };
}

async function fetchLazadaProduct(prisma: any, sku: string): Promise<CloneProductData | null> {
  // Find SKU in LazadaSku table
  const lazadaSku = await prisma.lazadaSku.findFirst({
    where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
    include: { product: true },
  });

  if (!lazadaSku || !lazadaSku.product) return null;

  const product = lazadaSku.product;

  // Get all SKUs for this product
  const allSkus = await prisma.lazadaSku.findMany({
    where: { productId: product.id },
  });

  return {
    title: product.name || "",
    description: product.description || "",
    images: product.images ? JSON.parse(product.images) : [],
    skus: allSkus.map((s: any) => ({
      sellerSku: s.sellerSku || s.skuId || "",
      price: s.price || 0,
      stock: s.stock || s.quantity || 0,
      variantLabel: s.variantName || undefined,
    })),
    weight: product.weight || undefined,
    packageDimensions: product.dimensions ? JSON.parse(product.dimensions) : undefined,
    categoryId: product.categoryId ? String(product.categoryId) : undefined,
    brandId: product.brandName || undefined,
  };
}

async function fetchTiktokProduct(prisma: any, sku: string): Promise<CloneProductData | null> {
  // Find SKU in TiktokSku table
  const tiktokSku = await prisma.tiktokSku.findFirst({
    where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
    include: { product: true },
  });

  if (!tiktokSku || !tiktokSku.product) return null;

  const product = tiktokSku.product;

  // Get all SKUs for this product
  const allSkus = await prisma.tiktokSku.findMany({
    where: { productId: product.id },
  });

  return {
    title: product.title || product.name || "",
    description: product.description || "",
    images: product.images ? JSON.parse(product.images) : [],
    skus: allSkus.map((s: any) => ({
      sellerSku: s.sellerSku || s.skuId || "",
      price: s.price || 0,
      stock: s.stock || s.availableStock || 0,
      variantLabel: s.variantName || undefined,
    })),
    weight: product.packageWeight || undefined,
    packageDimensions: product.packageDimensions
      ? JSON.parse(product.packageDimensions)
      : undefined,
    categoryId: product.categoryId ? String(product.categoryId) : undefined,
  };
}

// ============================================
// HELPER FUNCTIONS - Check SKU exists
// ============================================

async function checkShopeeSkuExists(prisma: any, sku: string): Promise<boolean> {
  const result = await prisma.shopeeSku.findFirst({
    where: { sellerSku: sku },
    select: { id: true },
  });
  return !!result;
}

async function checkLazadaSkuExists(prisma: any, sku: string): Promise<boolean> {
  const result = await prisma.lazadaSku.findFirst({
    where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
    select: { id: true },
  });
  return !!result;
}

async function checkTiktokSkuExists(prisma: any, sku: string): Promise<boolean> {
  const result = await prisma.tiktokSku.findFirst({
    where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
    select: { id: true },
  });
  return !!result;
}

export default router;

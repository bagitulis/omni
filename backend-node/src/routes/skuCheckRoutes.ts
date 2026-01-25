import { Router, Request, Response } from "express";
import { getPrisma } from "../services/prismaClient";
import { getLogger } from "../utils/logger";
import { sendSuccess, sendBadRequest, handleError } from "../utils/apiResponse";

const logger = getLogger("SkuCheckRoutes");
const router = Router();

/**
 * GET /api/products/sku-check/lazada
 * Check if SKU exists in Lazada products
 */
router.get("/lazada", async (req: Request, res: Response) => {
  try {
    const { sku } = req.query;
    if (!sku || typeof sku !== 'string') {
      return sendBadRequest(res, 'SKU parameter is required');
    }

    const prisma = getPrisma();
    const lazadaSku = await prisma.lazadaSku.findFirst({
      where: {
        OR: [
          { skuId: sku },
          { sellerSku: sku },
        ],
      },
      select: { id: true },
    });

    return sendSuccess(res, { exists: !!lazadaSku });
  } catch (error) {
    logger.error('[SKU Check] Lazada error:', error);
    return handleError(res, error, 'Failed to check SKU');
  }
});

/**
 * GET /api/products/sku-check/shopee
 * Check if SKU exists in Shopee products
 */
router.get("/shopee", async (req: Request, res: Response) => {
  try {
    const { sku } = req.query;
    if (!sku || typeof sku !== 'string') {
      return res.status(400).json({ error: 'SKU parameter is required' });
    }

    const prisma = getPrisma();
    
    // Try find by sellerSku first (string match)
    let shopeeSku = await prisma.shopeeSku.findFirst({
      where: {
        sellerSku: sku,
      },
      select: { id: true },
    });

    // If not found, try modelId (convert to BigInt)
    if (!shopeeSku) {
      try {
        const modelIdBig = BigInt(sku);
        shopeeSku = await prisma.shopeeSku.findFirst({
          where: {
            modelId: modelIdBig,
          },
          select: { id: true },
        });
      } catch (e) {
        // sku tidak valid sebagai BigInt, skip check ini
      }
    }

    return sendSuccess(res, { exists: !!shopeeSku });
  } catch (error) {
    logger.error('[SKU Check] Shopee error:', error);
    return handleError(res, error, 'Failed to check SKU');
  }
});

/**
 * GET /api/products/sku-check/tiktok
 * Check if SKU exists in TikTok products
 */
router.get("/tiktok", async (req: Request, res: Response) => {
  try {
    const { sku } = req.query;
    if (!sku || typeof sku !== 'string') {
      return sendBadRequest(res, 'SKU parameter is required');
    }

    const prisma = getPrisma();
    const tiktokSku = await prisma.tiktokSku.findFirst({
      where: {
        OR: [
          { skuId: sku },
          { sellerSku: sku },
        ],
      },
      select: { id: true },
    });

    return sendSuccess(res, { exists: !!tiktokSku });
  } catch (error) {
    logger.error('[SKU Check] TikTok error:', error);
    return handleError(res, error, 'Failed to check SKU');
  }
});

export default router;

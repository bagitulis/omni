import { Router, Request, Response } from "express";
import { getPrisma } from "../services/prismaClient";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import { sendSuccess, sendBadRequest, handleError } from "../utils/apiResponse";

const logger = getLogger("SkuBatchCheckRoutes");
const router = Router();

interface SkuCheckResult {
  sku: string;
  lazada: boolean;
  shopee: boolean;
  tiktok: boolean;
}

/**
 * POST /api/inventory/batch-check-sku
 * Check multiple SKUs across all platforms (batch operation)
 * Request body: { skus: ["SKU1", "SKU2", ...] }
 */
router.post(
  "/inventory/batch-check-sku",
  async (req: Request, res: Response) => {
    try {
      const { skus } = req.body;

      if (!Array.isArray(skus) || skus.length === 0) {
        return sendBadRequest(res, "skus array is required");
      }

      const prisma = getPrisma();
      const results: SkuCheckResult[] = [];

      // Process each SKU
      for (const sku of skus) {
        if (!sku || typeof sku !== "string") continue;

        const [lazadaExists, shopeeExists, tiktokExists] = await Promise.all([
          checkLazadaSku(prisma, sku),
          checkShopEeSku(prisma, sku),
          checkTiktokSku(prisma, sku),
        ]);

        results.push({
          sku,
          lazada: lazadaExists,
          shopee: shopeeExists,
          tiktok: tiktokExists,
        });
      }

      return sendSuccess(res, {
        total: skus.length,
        checked: results.length,
        results,
      });
    } catch (error) {
      logger.error("[Batch SKU Check] Error:", error);
      return handleError(res, error, "Failed to batch check SKUs");
    }
  }
);

/**
 * Helper: Check if SKU exists in Lazada
 */
async function checkLazadaSku(prisma: any, sku: string): Promise<boolean> {
  try {
    const result = await prisma.lazadaSku.findFirst({
      where: {
        OR: [{ skuId: sku }, { sellerSku: sku }],
      },
      select: { id: true },
    });
    return !!result;
  } catch (error) {
    logger.warn(`Error checking Lazada SKU ${sku}:`, error);
    return false;
  }
}

/**
 * Helper: Check if SKU exists in Shopee
 */
async function checkShopEeSku(prisma: any, sku: string): Promise<boolean> {
  try {
    // Try sellerSku first
    let result = await prisma.shopeeSku.findFirst({
      where: { sellerSku: sku },
      select: { id: true },
    });

    // Try modelId as BigInt
    if (!result) {
      try {
        const modelIdBig = BigInt(sku);
        result = await prisma.shopeeSku.findFirst({
          where: { modelId: modelIdBig },
          select: { id: true },
        });
      } catch (e) {
        // Not a valid BigInt
      }
    }

    return !!result;
  } catch (error) {
    logger.warn(`Error checking Shopee SKU ${sku}:`, error);
    return false;
  }
}

/**
 * Helper: Check if SKU exists in TikTok
 */
async function checkTiktokSku(prisma: any, sku: string): Promise<boolean> {
  try {
    const result = await prisma.tiktokSku.findFirst({
      where: {
        OR: [{ skuId: sku }, { sellerSku: sku }],
      },
      select: { id: true },
    });
    return !!result;
  } catch (error) {
    logger.warn(`Error checking TikTok SKU ${sku}:`, error);
    return false;
  }
}

/**
 * POST /api/inventory/batch-save-platform-status
 * Save L/S/T platform status ke database untuk setiap SKU
 * Clean data lama terlebih dahulu, kemudian insert data baru
 * Request body: { results: [{ sku, lazada, shopee, tiktok }, ...] }
 */
router.post(
  "/inventory/batch-save-platform-status",
  async (req: Request, res: Response) => {
    try {
      const { results } = req.body;

      if (!Array.isArray(results) || results.length === 0) {
        return sendBadRequest(res, "results array is required");
      }

      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      // STEP 1: Delete semua data lama di tabel InventorySkuPlatformStatus (tenant-scoped)
      const deletedCount = await prisma.inventorySkuPlatformStatus.deleteMany({
        where: { tenantId },
      });
      logger.info(
        `[${tenantId}] Deleted ${deletedCount.count} old records from inventory SKU platform status`
      );

      // STEP 2: Insert data baru (bulk create with tenantId)
      const createdRecords = await prisma.inventorySkuPlatformStatus.createMany(
        {
          data: results.map((result) => ({
            tenantId,
            sku: result.sku,
            lazada: result.lazada,
            shopee: result.shopee,
            tiktok: result.tiktok,
            checkedAt: new Date(),
          })),
        }
      );

      logger.info(
        `Created ${createdRecords.count} new platform status records`
      );

      return sendSuccess(res, {
        total: results.length,
        saved: createdRecords.count,
        deleted: deletedCount.count,
      });
    } catch (error) {
      logger.error("[Batch Save Status] Error:", error);
      return handleError(res, error, "Failed to save platform status");
    }
  }
);

/**
 * GET /api/inventory/platform-status
 * Load semua platform status yang sudah di-check sebelumnya
 * Return: { success, count, results: [{ sku, lazada, shopee, tiktok }, ...] }
 */
router.get(
  "/inventory/platform-status",
  async (_req: Request, res: Response) => {
    try {
      const prisma = getPrisma();
      const tenantId = tenantContext.getTenantId();

      const statuses = await prisma.inventorySkuPlatformStatus.findMany({
        where: { tenantId },
        select: {
          sku: true,
          lazada: true,
          shopee: true,
          tiktok: true,
          checkedAt: true,
        },
      });

      return sendSuccess(res, {
        count: statuses.length,
        lastChecked: statuses.length > 0 ? statuses[0].checkedAt : null,
        results: statuses.map((s) => ({
          sku: s.sku,
          lazada: s.lazada,
          shopee: s.shopee,
          tiktok: s.tiktok,
        })),
      });
    } catch (error) {
      logger.error("[Load Platform Status] Error:", error);
      return handleError(res, error, "Failed to load platform status");
    }
  }
);

export default router;

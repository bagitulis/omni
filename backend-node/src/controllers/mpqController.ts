/**
 * MPQ Controller
 * RESPONSIBILITY: Handle MPQ (Minimum Purchase Quantity) operations
 *
 * Endpoints:
 * - POST /api/wholesale/shopee/batch-mpq - Batch set MPQ for Shopee
 * - POST /api/wholesale/tiktok/batch-mpq - Batch set MPQ for TikTok
 * - POST /api/wholesale/shopee/batch-wholesale-reset - Batch wholesale with MPQ reset
 */

import { Request, Response } from "express";
import {
  ShopeeWholesaleService,
  WholesaleTier,
} from "../services/wholesale/shopeeWholesaleService";
import { TiktokMpqService } from "../services/wholesale/tiktokMpqService";
import { WholesaleSettingsService } from "../services/wholesale/wholesaleSettingsService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { JobHistoryManager } from "../services/jobHistoryManager";
import { getLogger } from "../utils/logger";
import { PrismaClient } from "@prisma/client";

const logger = getLogger("MpqController");

export class MpqController {
  private shopeeService: ShopeeWholesaleService;
  private tiktokService: TiktokMpqService;
  private settingsService: WholesaleSettingsService;
  private historyManager: JobHistoryManager;

  constructor(
    shopeeApiClient: ShopeeAPIClient,
    shopeeConfig: ShopeeConfigManager,
    tiktokApiClient: TiktokAPIClient,
    tiktokConfig: TiktokConfigManager,
    prisma?: PrismaClient
  ) {
    this.shopeeService = new ShopeeWholesaleService(
      shopeeApiClient,
      shopeeConfig,
      prisma
    );
    this.tiktokService = new TiktokMpqService(
      tiktokApiClient,
      tiktokConfig,
      prisma
    );
    this.settingsService = new WholesaleSettingsService(prisma);
    this.historyManager = new JobHistoryManager();
  }

  /**
   * POST /api/wholesale/shopee/batch-mpq
   * Batch set MPQ mode for Shopee (delete wholesale, set price, set MPQ)
   *
   * Body: { items: [{ sku, price }], mpq: number }
   */
  async batchShopeeMpq(req: Request, res: Response): Promise<Response> {
    try {
      const { items, mpq } = req.body;

      if (!Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
          success: false,
          error: "items array is required: [{ sku, price }]",
        });
      }

      if (!mpq || mpq < 1) {
        return res.status(400).json({
          success: false,
          error: "mpq must be >= 1",
        });
      }

      logger.info(`🔢 Batch Shopee MPQ for ${items.length} SKUs, MPQ=${mpq}`);

      // Build SKU → Price map
      const skuPriceMap = new Map<string, number>();
      for (const item of items) {
        if (item.sku && item.price > 0) {
          skuPriceMap.set(item.sku, item.price);
        }
      }

      const result = await this.shopeeService.batchSetMpqBySkus(
        skuPriceMap,
        mpq
      );

      // Log summary with errors if any
      if (result.failed > 0) {
        const failedResults = result.results.filter((r) => !r.success);
        logger.warn(
          `⚠️ Shopee MPQ partial: ${result.processed}/${result.uniqueItems} OK, ${result.failed} failed`
        );
        failedResults.forEach((r) => logger.warn(`   └─ ${r.error}`));
      } else {
        logger.info(`✅ Shopee MPQ complete: ${result.processed} items`);
      }

      // Log to history per-SKU (like stock/price update)
      const batchId = Date.now();
      for (const r of result.results) {
        const sku =
          r.message?.match(/SKUs?: ([^)]+)/)?.[1] || `item_${r.itemId}`;
        const jobId = `mpq_shopee_${batchId}_${r.itemId}`;
        this.historyManager.addDirectHistory(
          jobId,
          `mpq:${sku}:SHOPEE:MPQ=${mpq}`,
          r.success ? "completed" : "failed",
          r.success ? undefined : r.error
        );
      }

      return res.json({
        success: result.success,
        message: `MPQ=${mpq}: ${result.processed}/${result.uniqueItems} items`,
        data: {
          totalSkus: result.total,
          uniqueItems: result.uniqueItems,
          processed: result.processed,
          failed: result.failed,
          skipped: result.skipped,
          results: result.results,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Batch Shopee MPQ error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/wholesale/tiktok/batch-mpq
   * Batch set MPQ for TikTok
   *
   * Body: { items: [{ sku, price }], mpq: number }
   */
  async batchTiktokMpq(req: Request, res: Response): Promise<Response> {
    try {
      const { items, mpq } = req.body;

      if (!Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
          success: false,
          error: "items array is required: [{ sku, price }]",
        });
      }

      if (!mpq || mpq < 1) {
        return res.status(400).json({
          success: false,
          error: "mpq must be >= 1",
        });
      }

      logger.info(`🔢 Batch TikTok MPQ for ${items.length} SKUs, MPQ=${mpq}`);

      // Build SKU → Price map
      const skuPriceMap = new Map<string, number>();
      for (const item of items) {
        if (item.sku && item.price > 0) {
          skuPriceMap.set(item.sku, item.price);
          logger.info(`   📋 SKU: ${item.sku}, Price: ${item.price}`);
        } else {
          logger.warn(
            `   ⚠️ Invalid item: sku=${item.sku}, price=${item.price}`
          );
        }
      }

      const result = await this.tiktokService.batchUpdateMpq(skuPriceMap, mpq);

      // Log summary with errors if any
      if (result.failed > 0) {
        const failedResults = result.results.filter((r) => !r.success);
        logger.warn(
          `⚠️ TikTok MPQ partial: ${result.processed}/${result.uniqueProducts} OK, ${result.failed} failed`
        );
        failedResults.forEach((r) =>
          logger.warn(`   └─ [${r.productId}] ${r.error}`)
        );
      } else if (result.skipped.length > 0) {
        logger.warn(
          `⚠️ TikTok MPQ: ${
            result.processed
          } OK, skipped: ${result.skipped.join(", ")}`
        );
      } else {
        logger.info(`✅ TikTok MPQ complete: ${result.processed} products`);
      }

      // Log to history per-SKU (like stock/price update)
      const batchId = Date.now();
      for (const r of result.results) {
        const sku =
          r.message?.match(/SKUs?: ([^)]+)/)?.[1] || `product_${r.productId}`;
        const jobId = `mpq_tiktok_${batchId}_${r.productId}`;
        this.historyManager.addDirectHistory(
          jobId,
          `mpq:${sku}:TIKTOK:MPQ=${mpq}`,
          r.success ? "completed" : "failed",
          r.success ? undefined : r.error
        );
      }

      return res.json({
        success: result.success,
        message: `MPQ=${mpq}: ${result.processed}/${result.uniqueProducts} products`,
        data: {
          totalSkus: result.total,
          uniqueProducts: result.uniqueProducts,
          processed: result.processed,
          failed: result.failed,
          skipped: result.skipped,
          results: result.results,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Batch TikTok MPQ error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/wholesale/shopee/batch-wholesale-reset
   * Batch update wholesale with MPQ reset first
   * Step 1: Reset MPQ to 1 for each item
   * Step 2: Set wholesale tiers
   *
   * Body: { items: [{ sku, price }] }
   */
  async batchShopeeWholesaleWithReset(
    req: Request,
    res: Response
  ): Promise<Response> {
    try {
      const tenantId = (req as any).tenantId;
      if (!tenantId) {
        return res
          .status(401)
          .json({
            success: false,
            error: "Missing tenantId - authentication required",
          });
      }
      const { items } = req.body;

      if (!Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
          success: false,
          error: "items array is required: [{ sku, price }]",
        });
      }

      logger.info(`📦 Batch wholesale (with reset) for ${items.length} SKUs`);

      // Get tenant settings
      const settings = await this.settingsService.getSettings(tenantId);

      // Build SKU → Price map
      const skuPriceMap = new Map<string, number>();
      for (const item of items) {
        if (item.sku && item.price > 0) {
          skuPriceMap.set(item.sku, item.price);
        }
      }

      // Create tier calculator
      const calculateTiers = (basePrice: number): WholesaleTier[] => {
        const tiers = this.settingsService.calculateTiers(basePrice, settings);
        return tiers.map((t) => ({
          minCount: t.minCount,
          maxCount: t.maxCount,
          unitPrice: t.unitPrice,
        }));
      };

      const result = await this.shopeeService.batchUpdateWholesaleWithMpqReset(
        skuPriceMap,
        calculateTiers
      );

      // Log to history per-SKU (like stock/price update)
      const batchId = Date.now();
      for (const r of result.results) {
        const sku =
          r.message?.match(/SKUs?: ([^)]+)/)?.[1] || `item_${r.itemId}`;
        const jobId = `wholesale_shopee_${batchId}_${r.itemId}`;
        this.historyManager.addDirectHistory(
          jobId,
          `wholesale:${sku}:SHOPEE`,
          r.success ? "completed" : "failed",
          r.success ? undefined : r.error
        );
      }

      return res.json({
        success: result.success,
        message: `Wholesale: ${result.processed}/${result.uniqueItems} items`,
        data: {
          totalSkus: result.total,
          uniqueItems: result.uniqueItems,
          processed: result.processed,
          failed: result.failed,
          skipped: result.skipped,
          results: result.results,
          settingsUsed: settings,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Batch wholesale reset error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}

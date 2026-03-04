/**
 * Wholesale Batch Controller
 * RESPONSIBILITY: Handle batch wholesale operations
 *
 * Endpoints:
 * - POST /api/wholesale/shopee/batch-delete - Batch delete by itemIds
 * - POST /api/wholesale/shopee/batch-delete-skus - Batch delete by SKUs
 * - POST /api/wholesale/shopee/batch-update-skus - Batch update by SKUs
 */

import { Request, Response } from "express";
import {
  ShopeeWholesaleService,
  WholesaleTier,
} from "../services/wholesale/shopeeWholesaleService";
import { WholesaleSettingsService } from "../services/wholesale/wholesaleSettingsService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { JobHistoryManager } from "../services/jobHistoryManager";
import { getLogger } from "../utils/logger";
import { PrismaClient } from "@prisma/client";

const logger = getLogger("WholesaleBatchController");

export class WholesaleBatchController {
  private shopeeService: ShopeeWholesaleService;
  private settingsService: WholesaleSettingsService;
  private historyManager: JobHistoryManager;

  constructor(
    shopeeApiClient: ShopeeAPIClient,
    shopeeConfig: ShopeeConfigManager,
    prisma?: PrismaClient
  ) {
    this.shopeeService = new ShopeeWholesaleService(
      shopeeApiClient,
      shopeeConfig,
      prisma
    );
    this.settingsService = new WholesaleSettingsService(prisma);
    this.historyManager = new JobHistoryManager();
  }

  /**
   * POST /api/wholesale/shopee/batch-delete
   * Batch delete wholesale for multiple products
   */
  async batchDelete(req: Request, res: Response): Promise<Response> {
    try {
      const { itemIds } = req.body;

      if (!Array.isArray(itemIds) || itemIds.length === 0) {
        return res.status(400).json({
          success: false,
          error: "itemIds array is required",
        });
      }

      logger.info(`🗑️ Batch deleting wholesale for ${itemIds.length} items`);

      const results = await Promise.all(
        itemIds.map((id) => this.shopeeService.deleteWholesaleTiers(Number(id)))
      );

      const successful = results.filter((r) => r.success).length;
      const failed = results.filter((r) => !r.success).length;

      return res.json({
        success: failed === 0,
        message: `Deleted ${successful}/${itemIds.length} items`,
        data: { total: itemIds.length, successful, failed, results },
      });
    } catch (error: any) {
      logger.error(`❌ Batch delete error: ${error.message}`);
      return res.status(500).json({ success: false, error: error.message });
    }
  }

  /**
   * POST /api/wholesale/shopee/batch-delete-skus
   * Batch delete wholesale by SKUs (with automatic deduplication)
   */
  async batchDeleteBySkus(req: Request, res: Response): Promise<Response> {
    try {
      const { skus } = req.body;

      if (!Array.isArray(skus) || skus.length === 0) {
        return res.status(400).json({
          success: false,
          error: "skus array is required",
        });
      }

      logger.info(`🗑️ Batch delete wholesale for ${skus.length} SKUs`);

      const result = await this.shopeeService.batchDeleteBySkus(skus);

      logger.info(
        `📊 Result: ${result.uniqueItems} unique items from ${result.total} SKUs`
      );

      // Log to history per-SKU (like stock/price update)
      const batchId = Date.now();
      for (const r of result.results) {
        const sku =
          r.message?.match(/SKUs?: ([^)]+)/)?.[1] || `item_${r.itemId}`;
        const jobId = `delete_wholesale_${batchId}_${r.itemId}`;
        this.historyManager.addDirectHistory(
          jobId,
          `delete_wholesale:${sku}:SHOPEE`,
          r.success ? "completed" : "failed",
          r.success ? undefined : r.error
        );
      }

      return res.json({
        success: result.success,
        message: `Processed ${result.processed}/${result.uniqueItems} unique items`,
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
      logger.error(`❌ Batch delete by SKUs error: ${error.message}`);
      return res.status(500).json({ success: false, error: error.message });
    }
  }

  /**
   * POST /api/wholesale/shopee/batch-update-skus
   * Batch update wholesale by SKUs with auto-calculated tiers
   */
  async batchUpdateBySkus(req: Request, res: Response): Promise<Response> {
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

      // Validate items
      for (const item of items) {
        if (!item.sku || !item.price || item.price <= 0) {
          return res.status(400).json({
            success: false,
            error: `Invalid item: ${JSON.stringify(item)}`,
          });
        }
      }

      logger.info(`📦 Batch update wholesale for ${items.length} SKUs`);

      // Get tenant settings
      const settings = await this.settingsService.getSettings(tenantId);

      // Build SKU → Price map
      const skuPriceMap = new Map<string, number>();
      for (const item of items) {
        skuPriceMap.set(item.sku, item.price);
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

      const result = await this.shopeeService.batchUpdateBySkus(
        skuPriceMap,
        calculateTiers
      );

      logger.info(
        `📊 Result: ${result.processed}/${result.uniqueItems} updated`
      );

      return res.json({
        success: result.success,
        message: `Updated ${result.processed}/${result.uniqueItems} items`,
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
      logger.error(`❌ Batch update by SKUs error: ${error.message}`);
      return res.status(500).json({ success: false, error: error.message });
    }
  }
}

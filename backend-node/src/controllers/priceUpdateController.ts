import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductIdFetcher } from "../services/stock/productIdFetcher";
import { JobHistoryManager } from "../services/jobHistoryManager";
import { v4 as uuidv4 } from "uuid";
import { ShopeePriceUpdater } from "../services/price/shopeePriceUpdater";
import { LazadaPriceUpdater } from "../services/price/lazadaPriceUpdater";
import { TiktokPriceUpdater } from "../services/price/tiktokPriceUpdater";
import {
  PriceUpdateOrchestrator,
  PriceUpdateItem,
} from "../services/price/priceUpdateOrchestrator";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { getLogger } from "../utils/logger";

const logger = getLogger("PriceUpdateController");

/**
 * PriceUpdateController
 * RESPONSIBILITY: Handle HTTP requests for price updates
 * Delegates business logic to PriceUpdateOrchestrator
 *
 * Uses single "HARGA" column - same price for all platforms
 */
export class PriceUpdateController {
  private orchestrator: PriceUpdateOrchestrator | null = null;

  constructor(
    private prisma: PrismaClient,
    private shopeeApiClient: ShopeeAPIClient,
    private lazadaApiClient: LazadaAPIClient,
    private tiktokApiClient: TiktokAPIClient,
    private shopeeConfig: ShopeeConfigManager,
    private lazadaConfig: LazadaConfigManager,
    private tiktokConfig: TiktokConfigManager
  ) {}

  /**
   * Initialize orchestrator with all services
   */
  private getOrchestrator(): PriceUpdateOrchestrator {
    if (!this.orchestrator) {
      const productIdFetcher = new ProductIdFetcher(this.prisma);
      const shopeeUpdater = new ShopeePriceUpdater(
        this.shopeeApiClient,
        this.shopeeConfig
      );
      const lazadaUpdater = new LazadaPriceUpdater(
        this.lazadaApiClient,
        this.lazadaConfig
      );
      const tiktokUpdater = new TiktokPriceUpdater(
        this.tiktokApiClient,
        this.tiktokConfig
      );

      this.orchestrator = new PriceUpdateOrchestrator(
        productIdFetcher,
        shopeeUpdater,
        lazadaUpdater,
        tiktokUpdater
      );
    }
    return this.orchestrator;
  }

  /**
   * POST /api/inventory/update-price
   * Update price for a single SKU across platforms
   *
   * Body: { sku: string, price: number, platforms?: string[] }
   */
  async updateSinglePrice(req: Request, res: Response): Promise<Response> {
    try {
      const { sku, price, platforms } = req.body;

      if (!sku || typeof price !== "number") {
        return res.status(400).json({
          success: false,
          error: "Missing required fields: sku and price",
        });
      }

      if (price < 0) {
        return res.status(400).json({
          success: false,
          error: "Price must be a positive number",
        });
      }

      const orchestrator = this.getOrchestrator();
      const startTime = Date.now();
      const result = await orchestrator.updatePrice({
        sku,
        price,
        platforms: platforms as ("shopee" | "lazada" | "tiktok")[] | undefined,
      });
      const durationMs = Date.now() - startTime;

      // Log per-platform history for Direct Mode tracking
      try {
        const historyManager = new JobHistoryManager();
        const batchId = uuidv4().substring(0, 8);

        // Log each platform result separately
        if (result.platforms) {
          for (const [platform, platformResult] of Object.entries(
            result.platforms
          )) {
            if (!platformResult) continue;

            const jobId = `direct-price-${batchId}-${platform}`;
            const isSuccess = platformResult.success === true;
            const errorMsg = isSuccess
              ? undefined
              : platformResult.error || "Unknown error";

            historyManager.addDirectHistory(
              jobId,
              `price_update:${sku}:${platform.toUpperCase()}`,
              isSuccess ? "completed" : "failed",
              errorMsg,
              Math.round(durationMs / Object.keys(result.platforms).length)
            );
          }
        }
      } catch (historyError) {
        logger.warn(
          "[PriceUpdateController] Failed to log history:",
          historyError
        );
      }

      return res.json({
        success: result.success,
        data: result,
      });
    } catch (error: any) {
      logger.error("[PriceUpdateController] Single update error:", error);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/inventory/update-price-batch
   * Batch update price for multiple SKUs
   *
   * Body: { items: [{ sku: string, price: number, platforms?: string[] }] }
   */
  async updateBatchPrice(req: Request, res: Response): Promise<Response> {
    try {
      const { items } = req.body;

      if (!Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
          success: false,
          error: "Missing or invalid items array",
        });
      }

      // Validate items
      const validItems: PriceUpdateItem[] = [];
      const invalidItems: string[] = [];

      for (const item of items) {
        if (item.sku && typeof item.price === "number" && item.price >= 0) {
          validItems.push({
            sku: item.sku,
            price: item.price,
            platforms: item.platforms as
              | ("shopee" | "lazada" | "tiktok")[]
              | undefined,
          });
        } else {
          invalidItems.push(item.sku || "unknown");
        }
      }

      if (validItems.length === 0) {
        return res.status(400).json({
          success: false,
          error: "No valid items to process",
          invalidItems,
        });
      }

      const orchestrator = this.getOrchestrator();
      const startTime = Date.now();
      const result = await orchestrator.updateBatch(validItems);
      const durationMs = Date.now() - startTime;

      // Log per-platform history for Direct Mode tracking
      try {
        const historyManager = new JobHistoryManager();
        const batchId = uuidv4().substring(0, 8);

        // Log each SKU's platform results separately
        for (const itemResult of result.results) {
          if (!itemResult.platforms) continue;

          for (const [platform, platformResult] of Object.entries(
            itemResult.platforms
          )) {
            if (!platformResult) continue;

            const jobId = `direct-price-${batchId}-${itemResult.sku}-${platform}`;
            const isSuccess = platformResult.success === true;
            const errorMsg = isSuccess
              ? undefined
              : platformResult.error || "Unknown error";

            historyManager.addDirectHistory(
              jobId,
              `price_update:${itemResult.sku}:${platform.toUpperCase()}`,
              isSuccess ? "completed" : "failed",
              errorMsg,
              Math.round(durationMs / result.total)
            );
          }
        }
      } catch (historyError) {
        logger.warn(
          "[PriceUpdateController] Failed to log batch history:",
          historyError
        );
      }

      // Log results summary
      logger.info(
        `[Price Update] Batch complete: ${result.total} items (${result.successful} success, ${result.failed} failed)`
      );
      result.results.forEach((r) => {
        const platforms = Object.entries(r.platforms || {})
          .filter(([, p]: any) => p?.success)
          .map(([name]) => name.toUpperCase())
          .join(", ");
        const status = r.success ? "SUCCESS" : "FAILED";
        logger.info(
          `[Price Update] ${status} ${r.sku} -> ${platforms || "failed"}`
        );
      });

      return res.json({
        success: result.failed === 0,
        data: {
          total: result.total,
          successful: result.successful,
          failed: result.failed,
          skipped: result.skipped,
          results: result.results,
        },
        invalidItems: invalidItems.length > 0 ? invalidItems : undefined,
      });
    } catch (error: any) {
      logger.error(`[Price Update] Error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/inventory/lookup-platform-ids
   * Lookup platform product IDs for a SKU (reuse from stock controller)
   * This is shared functionality, already available in stock routes
   */
}

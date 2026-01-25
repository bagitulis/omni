import { Request, Response } from "express";
import { PrismaClient } from "@prisma/client";
import { ProductIdFetcher } from "../services/stock/productIdFetcher";
import { JobHistoryManager } from "../services/jobHistoryManager";
import { v4 as uuidv4 } from "uuid";
import { ShopeeStockUpdater } from "../services/stock/shopeeStockUpdater";
import { LazadaStockUpdater } from "../services/stock/lazadaStockUpdater";
import { TiktokStockUpdater } from "../services/stock/tiktokStockUpdater";
import {
  StockUpdateOrchestrator,
  StockUpdateItem,
} from "../services/stock/stockUpdateOrchestrator";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { getLogger } from "../utils/logger";

const logger = getLogger("StockUpdateController");

/**
 * StockUpdateController
 * RESPONSIBILITY: Handle HTTP requests for stock updates
 * Delegates business logic to StockUpdateOrchestrator
 */
export class StockUpdateController {
  private orchestrator: StockUpdateOrchestrator | null = null;

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
  private getOrchestrator(): StockUpdateOrchestrator {
    if (!this.orchestrator) {
      const productIdFetcher = new ProductIdFetcher(this.prisma);
      const shopeeUpdater = new ShopeeStockUpdater(
        this.shopeeApiClient,
        this.shopeeConfig
      );
      const lazadaUpdater = new LazadaStockUpdater(
        this.lazadaApiClient,
        this.lazadaConfig
      );
      const tiktokUpdater = new TiktokStockUpdater(
        this.tiktokApiClient,
        this.tiktokConfig
      );

      this.orchestrator = new StockUpdateOrchestrator(
        productIdFetcher,
        shopeeUpdater,
        lazadaUpdater,
        tiktokUpdater
      );
    }
    return this.orchestrator;
  }

  /**
   * POST /api/inventory/update-stock
   * Update stock for a single SKU across platforms
   *
   * Body: { sku: string, stock: number, platforms?: string[] }
   */
  async updateSingleStock(req: Request, res: Response): Promise<Response> {
    try {
      const { sku, stock, platforms } = req.body;

      if (!sku || typeof stock !== "number") {
        return res.status(400).json({
          success: false,
          error: "Missing required fields: sku and stock",
        });
      }

      const orchestrator = this.getOrchestrator();
      const startTime = Date.now();
      const result = await orchestrator.updateStock({
        sku,
        stock,
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

            const jobId = `direct-stock-${batchId}-${platform}`;
            const isSuccess = platformResult.success === true;
            const errorMsg = isSuccess
              ? undefined
              : platformResult.error || "Unknown error";

            historyManager.addDirectHistory(
              jobId,
              `stock_update:${sku}:${platform.toUpperCase()}`,
              isSuccess ? "completed" : "failed",
              errorMsg,
              Math.round(durationMs / Object.keys(result.platforms).length)
            );
          }
        }
      } catch (historyError) {
        logger.warn(
          "[StockUpdateController] Failed to log history:",
          historyError
        );
      }

      return res.json({
        success: result.success,
        data: result,
      });
    } catch (error: any) {
      logger.error("[StockUpdateController] Single update error:", error);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/inventory/update-stock-batch
   * Batch update stock for multiple SKUs
   *
   * Body: { items: [{ sku: string, stock: number, platforms?: string[] }] }
   */
  async updateBatchStock(req: Request, res: Response): Promise<Response> {
    try {
      const { items } = req.body;

      if (!Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
          success: false,
          error: "Missing or invalid items array",
        });
      }

      // Validate items
      const validItems: StockUpdateItem[] = [];
      const invalidItems: string[] = [];

      for (const item of items) {
        // Accept item if it has SKU and at least one stock value
        const hasStock =
          typeof item.stock === "number" ||
          typeof item.shopeeStock === "number" ||
          typeof item.tiktokStock === "number" ||
          typeof item.lazadaStock === "number";

        if (item.sku && hasStock) {
          validItems.push({
            sku: item.sku,
            stock: item.stock,
            shopeeStock: item.shopeeStock,
            tiktokStock: item.tiktokStock,
            lazadaStock: item.lazadaStock,
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
      const result = await orchestrator.updateBatch(validItems);

      // Log results summary
      logger.info(
        `[Stock Update] Batch complete: ${result.total} items (${result.successful} success, ${result.failed} failed)`
      );
      result.results.forEach((r) => {
        const platforms = Object.entries(r.platforms || {})
          .filter(([, p]: any) => p?.success)
          .map(([name]) => name.toUpperCase())
          .join(", ");
        const status = r.success ? "SUCCESS" : "FAILED";
        logger.info(
          `[Stock Update] ${status} ${r.sku} -> ${platforms || "failed"}`
        );
      });

      return res.json({
        success: result.failed === 0,
        data: {
          ...result,
          invalidItems: invalidItems.length > 0 ? invalidItems : undefined,
        },
      });
    } catch (error: any) {
      logger.error(`[Stock Update] Error: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/inventory/lookup-platform-ids
   * Lookup platform product IDs for a SKU (for debugging/verification)
   *
   * Body: { sku: string }
   */
  async lookupPlatformIds(req: Request, res: Response): Promise<Response> {
    try {
      const { sku } = req.body;

      if (!sku) {
        return res.status(400).json({
          success: false,
          error: "Missing required field: sku",
        });
      }

      const productIdFetcher = new ProductIdFetcher(this.prisma);
      const platformIds = await productIdFetcher.fetchBySku(sku);

      return res.json({
        success: true,
        data: platformIds,
      });
    } catch (error: any) {
      logger.error("[StockUpdateController] Lookup error:", error);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}

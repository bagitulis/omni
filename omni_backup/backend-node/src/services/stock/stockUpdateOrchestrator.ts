// import { PrismaClient } from "@prisma/client";
import { ProductIdFetcher } from "./productIdFetcher";
import {
  ShopeeStockUpdater,
  ShopeeStockUpdateResult,
} from "./shopeeStockUpdater";
import {
  LazadaStockUpdater,
  LazadaStockUpdateResult,
} from "./lazadaStockUpdater";
import {
  TiktokStockUpdater,
  TiktokStockUpdateResult,
} from "./tiktokStockUpdater";

/**
 * StockUpdateOrchestrator
 * RESPONSIBILITY: Coordinate stock updates across all platforms
 * Follows SRP by delegating platform-specific logic to individual updaters
 */
export interface StockUpdateItem {
  sku: string;
  stock?: number; // Legacy: single stock for all platforms
  shopeeStock?: number; // Platform-specific stock
  tiktokStock?: number;
  lazadaStock?: number;
  platforms?: ("shopee" | "lazada" | "tiktok")[];
}

export interface PlatformResult {
  shopee?: ShopeeStockUpdateResult;
  lazada?: LazadaStockUpdateResult;
  tiktok?: TiktokStockUpdateResult;
}

export interface StockUpdateResult {
  sku: string;
  success: boolean;
  platforms: PlatformResult;
  errors: string[];
}

export interface BatchStockUpdateResult {
  total: number;
  successful: number;
  failed: number;
  results: StockUpdateResult[];
}

export class StockUpdateOrchestrator {
  constructor(
    private productIdFetcher: ProductIdFetcher,
    private shopeeUpdater: ShopeeStockUpdater | null,
    private lazadaUpdater: LazadaStockUpdater | null,
    private tiktokUpdater: TiktokStockUpdater | null
  ) {}

  /**
   * Update stock for a single SKU across all platforms
   */
  async updateStock(item: StockUpdateItem): Promise<StockUpdateResult> {
    const { sku, stock, shopeeStock, tiktokStock, lazadaStock, platforms } =
      item;

    // Determine which platforms to update
    const targetPlatforms = platforms || ["shopee", "lazada", "tiktok"];

    // Fetch platform IDs
    const productIds = await this.productIdFetcher.fetchBySku(sku);

    const result: StockUpdateResult = {
      sku,
      success: false,
      platforms: {},
      errors: [],
    };

    // Update each platform with platform-specific stock or fallback to general stock
    const updatePromises: Promise<void>[] = [];

    // Shopee
    if (
      targetPlatforms.includes("shopee") &&
      this.shopeeUpdater &&
      productIds.shopee
    ) {
      const stockValue = shopeeStock !== undefined ? shopeeStock : stock;
      if (stockValue !== undefined) {
        updatePromises.push(
          this.shopeeUpdater
            .updateStock({
              itemId: productIds.shopee.itemId,
              modelId: productIds.shopee.modelId,
              stock: stockValue,
            })
            .then((res) => {
              result.platforms.shopee = res;
            })
            .catch((err) => {
              result.errors.push(`Shopee: ${err.message}`);
              result.platforms.shopee = {
                success: false,
                itemId: productIds.shopee!.itemId,
                modelId: productIds.shopee!.modelId,
                error: err.message,
              };
            })
        );
      }
    }

    // Lazada
    if (
      targetPlatforms.includes("lazada") &&
      this.lazadaUpdater &&
      productIds.lazada
    ) {
      const stockValue = lazadaStock !== undefined ? lazadaStock : stock;
      if (stockValue !== undefined) {
        updatePromises.push(
          this.lazadaUpdater
            .updateStock({
              itemId: productIds.lazada.itemId,
              skuId: productIds.lazada.skuId,
              sellerSku: productIds.lazada.sellerSku,
              stock: stockValue,
            })
            .then((res) => {
              result.platforms.lazada = res;
            })
            .catch((err) => {
              result.errors.push(`Lazada: ${err.message}`);
              result.platforms.lazada = {
                success: false,
                itemId: productIds.lazada!.itemId,
                skuId: productIds.lazada!.skuId,
                error: err.message,
              };
            })
        );
      }
    }

    // TikTok
    if (
      targetPlatforms.includes("tiktok") &&
      this.tiktokUpdater &&
      productIds.tiktok
    ) {
      const stockValue = tiktokStock !== undefined ? tiktokStock : stock;
      if (stockValue !== undefined) {
        updatePromises.push(
          this.tiktokUpdater
            .updateStock({
              productId: productIds.tiktok.productId,
              skuId: productIds.tiktok.skuId,
              stock: stockValue,
            })
            .then((res) => {
              result.platforms.tiktok = res;
            })
            .catch((err) => {
              result.errors.push(`TikTok: ${err.message}`);
              result.platforms.tiktok = {
                success: false,
                productId: productIds.tiktok!.productId,
                skuId: productIds.tiktok!.skuId,
                error: err.message,
              };
            })
        );
      }
    }

    // Wait for all updates
    await Promise.all(updatePromises);

    // Determine overall success
    const platformResults = Object.values(result.platforms);

    // If no platforms were updated (SKU not found in any platform), mark as skipped
    if (platformResults.length === 0) {
      result.success = true; // Not an error - just nothing to update
      result.errors.push("SKU not found in any platform - skipped");
    } else {
      result.success = platformResults.every((p) => p?.success === true);
    }

    return result;
  }

  /**
   * Batch update stock for multiple SKUs
   */
  async updateBatch(items: StockUpdateItem[]): Promise<BatchStockUpdateResult> {
    const results: StockUpdateResult[] = [];
    let successful = 0;
    let failed = 0;

    for (const item of items) {
      const result = await this.updateStock(item);
      results.push(result);

      if (result.success) {
        successful++;
      } else {
        failed++;
      }

      // Small delay between SKUs
      await this.delay(100);
    }

    return {
      total: items.length,
      successful,
      failed,
      results,
    };
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

import { ProductIdFetcher} from "../stock/productIdFetcher";
import { ShopeePriceUpdater, ShopeePriceUpdateResult } from "./shopeePriceUpdater";
import { LazadaPriceUpdater, LazadaPriceUpdateResult } from "./lazadaPriceUpdater";
import { TiktokPriceUpdater, TiktokPriceUpdateResult } from "./tiktokPriceUpdater";

/**
 * PriceUpdateOrchestrator
 * RESPONSIBILITY: Coordinate price updates across all platforms
 * Follows SRP by delegating platform-specific logic to individual updaters
 *
 * Uses single "HARGA" column - same price for all platforms
 */
export interface PriceUpdateItem {
  sku: string;
  price: number; // Single price for all platforms (from HARGA column)
  platforms?: ("shopee" | "lazada" | "tiktok")[]; // Platforms to update
}

export interface PlatformPriceResult {
  shopee?: ShopeePriceUpdateResult;
  lazada?: LazadaPriceUpdateResult;
  tiktok?: TiktokPriceUpdateResult;
}

export interface PriceUpdateResult {
  sku: string;
  success: boolean;
  platforms: PlatformPriceResult;
  errors: string[];
  skipped: string[]; // Platforms that were skipped
}

export interface BatchPriceUpdateResult {
  total: number;
  successful: number;
  failed: number;
  skipped: number;
  results: PriceUpdateResult[];
}

export class PriceUpdateOrchestrator {
  constructor(
    private productIdFetcher: ProductIdFetcher,
    private shopeeUpdater: ShopeePriceUpdater | null,
    private lazadaUpdater: LazadaPriceUpdater | null,
    private tiktokUpdater: TiktokPriceUpdater | null
  ) {}

  /**
   * Update price for a single SKU across all platforms
   * Uses single price value for all platforms (HARGA column)
   */
  async updatePrice(item: PriceUpdateItem): Promise<PriceUpdateResult> {
    const { sku, price, platforms } = item;

    // Determine which platforms to update (default: all)
    const targetPlatforms = platforms || ["shopee", "lazada", "tiktok"];

    // Fetch platform IDs
    const productIds = await this.productIdFetcher.fetchBySku(sku);

    const result: PriceUpdateResult = {
      sku,
      success: false,
      platforms: {},
      errors: [],
      skipped: [],
    };

    // Track which platforms to update
    const updatePromises: Promise<void>[] = [];

    // Shopee
    if (targetPlatforms.includes("shopee")) {
      if (this.shopeeUpdater && productIds.shopee) {
        updatePromises.push(
          this.shopeeUpdater
            .updatePrice({
              itemId: productIds.shopee.itemId,
              modelId: productIds.shopee.modelId,
              price: price,
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
      } else {
        result.skipped.push("shopee");
      }
    }

    // Lazada
    if (targetPlatforms.includes("lazada")) {
      if (this.lazadaUpdater && productIds.lazada) {
        updatePromises.push(
          this.lazadaUpdater
            .updatePrice({
              itemId: productIds.lazada.itemId,
              skuId: productIds.lazada.skuId,
              sellerSku: productIds.lazada.sellerSku,
              price: price,
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
      } else {
        result.skipped.push("lazada");
      }
    }

    // TikTok
    if (targetPlatforms.includes("tiktok")) {
      if (this.tiktokUpdater && productIds.tiktok) {
        updatePromises.push(
          this.tiktokUpdater
            .updatePrice({
              productId: productIds.tiktok.productId,
              skuId: productIds.tiktok.skuId,
              price: price,
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
      } else {
        result.skipped.push("tiktok");
      }
    }

    // Wait for all updates
    await Promise.all(updatePromises);

    // Determine overall success
    const platformResults = Object.values(result.platforms);
    result.success =
      platformResults.length > 0 &&
      platformResults.every((p) => p?.success === true);

    return result;
  }

  /**
   * Batch update price for multiple SKUs
   */
  async updateBatch(items: PriceUpdateItem[]): Promise<BatchPriceUpdateResult> {
    const results: PriceUpdateResult[] = [];
    let successful = 0;
    let failed = 0;
    let skipped = 0;

    for (const item of items) {
      const result = await this.updatePrice(item);
      results.push(result);

      if (result.success) {
        successful++;
      } else if (result.skipped.length === 3) {
        // All platforms skipped
        skipped++;
      } else {
        failed++;
      }

      // Small delay between items to avoid rate limiting
      await this.delay(100);
    }

    return {
      total: items.length,
      successful,
      failed,
      skipped,
      results,
    };
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

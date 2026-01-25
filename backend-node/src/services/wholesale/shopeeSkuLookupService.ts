/**
 * Shopee SKU Lookup Service
 * SRP: Handle SKU to ItemId lookups from database
 */
import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";
import { Logger } from "winston";

export interface SkuLookupResult {
  itemId: string;
  modelId: string | null;
  sellerSku: string | null;
  productName?: string;
}

export class ShopeeSkuLookupService {
  private logger: Logger;

  constructor(private prisma?: PrismaClient) {
    this.logger = getLogger("ShopeeSkuLookup");
  }

  /**
   * Lookup item_id from SKU (seller_sku or model_id)
   */
  async lookupItemIdBySku(sku: string): Promise<SkuLookupResult | null> {
    if (!this.prisma) {
      this.logger.warn("⚠️ Prisma not available for SKU lookup");
      return null;
    }

    try {
      // Try sellerSku first
      let result = await this.prisma.shopeeSku.findFirst({
        where: { sellerSku: sku },
        include: { product: true },
      });

      // Try modelId as BigInt
      if (!result) {
        try {
          const modelIdBig = BigInt(sku);
          result = await this.prisma.shopeeSku.findFirst({
            where: { modelId: modelIdBig },
            include: { product: true },
          });
        } catch {
          // Not a valid BigInt, skip
        }
      }

      if (!result) return null;

      return {
        itemId: result.itemId.toString(),
        modelId: result.modelId?.toString() || null,
        sellerSku: result.sellerSku,
        productName: result.product?.name || "Unknown",
      };
    } catch (error: any) {
      this.logger.error(`❌ SKU lookup error: ${error.message}`);
      return null;
    }
  }

  /**
   * Batch lookup and dedupe SKUs by itemId
   */
  async batchLookupAndDedupe(skus: string[]): Promise<{
    itemIdMap: Map<string, { skus: string[]; name: string }>;
    notFound: string[];
  }> {
    const itemIdMap = new Map<string, { skus: string[]; name: string }>();
    const notFound: string[] = [];

    for (const sku of skus) {
      const lookup = await this.lookupItemIdBySku(sku);
      if (lookup) {
        const existing = itemIdMap.get(lookup.itemId);
        if (existing) {
          existing.skus.push(sku);
        } else {
          itemIdMap.set(lookup.itemId, {
            skus: [sku],
            name: lookup.productName || "Unknown",
          });
        }
      } else {
        notFound.push(sku);
      }
    }

    this.logger.info(
      `📊 Deduped: ${skus.length} SKUs → ${itemIdMap.size} unique items`
    );

    return { itemIdMap, notFound };
  }

  /**
   * Batch lookup with price mapping
   */
  async batchLookupWithPrices(skuPriceMap: Map<string, number>): Promise<{
    itemIdMap: Map<
      string,
      { skus: string[]; name: string; basePrice: number; modelId?: string }
    >;
    notFound: string[];
  }> {
    const skus = Array.from(skuPriceMap.keys());
    const itemIdMap = new Map<
      string,
      { skus: string[]; name: string; basePrice: number; modelId?: string }
    >();
    const notFound: string[] = [];

    for (const sku of skus) {
      const lookup = await this.lookupItemIdBySku(sku);
      if (lookup) {
        const existing = itemIdMap.get(lookup.itemId);
        if (existing) {
          existing.skus.push(sku);
        } else {
          itemIdMap.set(lookup.itemId, {
            skus: [sku],
            name: lookup.productName || sku,
            basePrice: skuPriceMap.get(sku) || 0,
            modelId: lookup.modelId || undefined,
          });
        }
      } else {
        notFound.push(sku);
      }
    }

    this.logger.info(
      `📊 Deduped: ${skus.length} SKUs → ${itemIdMap.size} unique items`
    );

    return { itemIdMap, notFound };
  }
}

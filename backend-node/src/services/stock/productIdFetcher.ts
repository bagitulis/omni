import { PrismaClient } from "@prisma/client";

/**
 * ProductIdFetcher
 * RESPONSIBILITY: Fetch product/model/sku IDs from database by seller SKU
 * Required for platform API calls to update stock
 */
export interface PlatformProductIds {
  lazada?: {
    itemId: string;
    skuId: string;
    sellerSku: string;
  };
  shopee?: {
    itemId: string; // Converted from BigInt for JSON serialization
    modelId: string | null; // Converted from BigInt for JSON serialization (nullable for single products)
    sellerSku: string | null;
  };
  tiktok?: {
    productId: string;
    skuId: string;
    sellerSku: string | null;
  };
}

export class ProductIdFetcher {
  constructor(private prisma: PrismaClient) {}

  /**
   * Fetch all platform product IDs by seller SKU
   */
  async fetchBySku(sku: string): Promise<PlatformProductIds> {
    const [lazada, shopee, tiktok] = await Promise.all([
      this.fetchLazadaIds(sku),
      this.fetchShopeeIds(sku),
      this.fetchTiktokIds(sku),
    ]);

    return { lazada, shopee, tiktok };
  }

  /**
   * Fetch Lazada product IDs
   * Returns: itemId, skuId, sellerSku
   */
  private async fetchLazadaIds(sku: string) {
    try {
      const result = await this.prisma.lazadaSku.findFirst({
        where: {
          OR: [{ skuId: sku }, { sellerSku: sku }],
        },
        include: { product: true },
      });

      if (!result) return undefined;

      return {
        itemId: result.product.itemId,
        skuId: result.skuId,
        sellerSku: result.sellerSku || sku,
      };
    } catch (error) {
      console.warn(`⚠️ Error fetching Lazada IDs for ${sku}:`, error);
      return undefined;
    }
  }

  /**
   * Fetch Shopee product IDs
   * Returns: itemId (BigInt), modelId (BigInt), sellerSku
   */
  private async fetchShopeeIds(sku: string) {
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
        } catch (e) {
          // Not a valid BigInt, skip
        }
      }

      if (!result) return undefined;

      return {
        itemId: result.itemId.toString(),
        modelId: result.modelId?.toString() || null,
        sellerSku: result.sellerSku,
      };
    } catch (error) {
      console.warn(`⚠️ Error fetching Shopee IDs for ${sku}:`, error);
      return undefined;
    }
  }

  /**
   * Fetch TikTok product IDs
   * Returns: productId, skuId, sellerSku
   * Note: If multiple SKUs match, returns the latest one (highest id)
   */
  private async fetchTiktokIds(sku: string) {
    try {
      const result = await this.prisma.tiktokSku.findFirst({
        where: {
          OR: [{ skuId: sku }, { sellerSku: sku }],
        },
        include: { product: true },
        orderBy: { id: "desc" }, // Get the latest/newest entry
      });

      if (!result) return undefined;

      return {
        productId: result.product.productId,
        skuId: result.skuId,
        sellerSku: result.sellerSku,
      };
    } catch (error) {
      console.warn(`⚠️ Error fetching TikTok IDs for ${sku}:`, error);
      return undefined;
    }
  }

  /**
   * Batch fetch IDs for multiple SKUs
   */
  async fetchBatch(skus: string[]): Promise<Map<string, PlatformProductIds>> {
    const results = new Map<string, PlatformProductIds>();

    for (const sku of skus) {
      const ids = await this.fetchBySku(sku);
      results.set(sku, ids);
    }

    return results;
  }
}

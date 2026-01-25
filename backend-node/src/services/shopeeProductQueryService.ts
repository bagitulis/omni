/**
 * Shopee Product Query Service
 * Handles all read/query operations for Shopee products
 * Single Responsibility: Product queries only
 */

import { PrismaClient } from "@prisma/client";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";

export class ShopeeProductQueryService {
  private prisma: PrismaClient;
  private logger: Logger;

  constructor(prisma: PrismaClient, logger?: Logger) {
    this.prisma = prisma;
    this.logger = logger || getLogger("ShopeeProductQueryService");
  }

  async getMasterProductsFromDb(params: {
    offset?: number;
    limit?: number;
  }): Promise<any> {
    try {
      const { offset = 0, limit = 100 } = params;

      const products = await this.prisma.shopeeProduct.findMany({
        skip: offset,
        take: limit,
        include: { skus: true },
        orderBy: { updatedAt: "desc" },
      });

      const total = await this.prisma.shopeeProduct.count();

      return {
        success: true,
        products: products.map((p) => ({
          ...p,
          itemId: p.itemId.toString(),
          skus: p.skus.map((s) => ({
            ...s,
            itemId: s.itemId.toString(),
            modelId: s.modelId?.toString() || null,
          })),
        })),
        total,
      };
    } catch (error: any) {
      this.logger.error(
        `❌ Error getting Shopee products from DB: ${error.message}`
      );
      return { success: false, error: error.message };
    }
  }

  async getMasterProductsList(params: {
    limit?: number;
    offset?: number;
  }): Promise<any> {
    try {
      const { limit = 100, offset = 0 } = params;

      this.logger.debug(
        `🔄 Getting Shopee master products list | Offset: ${offset}, Limit: ${limit}`
      );

      const skus = await this.prisma.shopeeSku.findMany({
        skip: offset,
        take: limit,
        include: { product: true },
        orderBy: { updatedAt: "desc" },
      });

      const total = await this.prisma.shopeeSku.count();

      const masterProducts = skus.map((sku) => ({
        itemId: sku.itemId.toString(),
        modelId: sku.modelId?.toString() || null,
        itemName: sku.product?.name || "",
        modelName: sku.variantName || null,
        itemSku: sku.sellerSku || null,
        modelSku: sku.sellerSku || null,
        currentPrice: sku.price,
        originalPrice: sku.price,
        sellerStock: sku.quantity,
        shopeeStock: sku.quantity,
        createdAt: sku.createdAt,
        updatedAt: sku.updatedAt,
      }));

      return {
        success: true,
        data: masterProducts,
        count: masterProducts.length,
        total,
        limit,
        offset,
      };
    } catch (error: any) {
      this.logger.error(
        `❌ Error getting master products list: ${error.message}`
      );
      return { success: false, error: error.message, data: [] };
    }
  }

  async getMasterProductsFormatted(): Promise<any> {
    try {
      this.logger.debug(
        "🔄 Getting Shopee master products in formatted table structure..."
      );

      const skus = await this.prisma.shopeeSku.findMany({
        include: { product: true },
        orderBy: [{ itemId: "asc" }, { modelId: "asc" }],
      });

      const formattedProducts = skus.map((sku) => ({
        itemId: sku.itemId.toString(),
        modelId: sku.modelId?.toString() || null,
        modelSku: sku.sellerSku || null,
        itemName: sku.product?.name || "Unknown",
        modelName: sku.variantName || null,
        currentPrice: sku.price,
        originalPrice: sku.price,
        sellerStock: sku.quantity,
        shopeeStock: sku.quantity,
        updated: sku.updatedAt,
      }));

      this.logger.info(
        `✅ Retrieved ${formattedProducts.length} products in formatted structure`
      );

      return {
        success: true,
        products: formattedProducts,
        total: formattedProducts.length,
      };
    } catch (error: any) {
      this.logger.error(
        `❌ Error getting formatted master products: ${error.message}`
      );
      return { success: false, error: error.message, products: [], total: 0 };
    }
  }

  async getMasterProductsStats(): Promise<any> {
    try {
      this.logger.debug("🔄 Getting Shopee master product statistics...");

      const stats: Record<string, any> = {};
      stats.totalRecords = await this.prisma.shopeeSku.count();
      stats.uniqueItems = await this.prisma.shopeeProduct.count();

      const itemsWithModels = await this.prisma.shopeeProduct.findMany({
        include: { skus: true },
      });
      stats.itemsWithModels = itemsWithModels.filter(
        (p) => p.skus.length > 1
      ).length;

      const allSkus = await this.prisma.shopeeSku.findMany();
      if (allSkus.length > 0) {
        const prices = allSkus.map((s) => s.price);
        stats.minPrice = Math.min(...prices);
        stats.maxPrice = Math.max(...prices);
        stats.avgPrice =
          Math.round(
            (prices.reduce((a, b) => a + b, 0) / prices.length) * 100
          ) / 100;
      } else {
        stats.minPrice = 0;
        stats.maxPrice = 0;
        stats.avgPrice = 0;
      }

      const totalSellerStock = allSkus.reduce((sum, s) => sum + s.quantity, 0);
      const totalShopeeStock = allSkus.reduce((sum, s) => sum + s.quantity, 0);

      stats.totalSellerStock = totalSellerStock;
      stats.totalShopeeStock = totalShopeeStock;

      this.logger.debug("✅ Master product statistics retrieved");

      return { success: true, stats };
    } catch (error: any) {
      this.logger.error(
        `❌ Error getting master product stats: ${error.message}`
      );
      return { success: false, error: error.message, stats: {} };
    }
  }

  async getProductById(productId: string): Promise<any> {
    try {
      // Note: No tenantId filter needed - Prisma connection is already tenant-specific
      // Find by itemId only since DB is isolated per tenant
      const product = await this.prisma.shopeeProduct.findFirst({
        where: { itemId: BigInt(productId) },
        include: { skus: true },
      });

      if (!product) {
        return { success: false, error: "Product not found" };
      }

      return {
        success: true,
        data: {
          ...product,
          itemId: product.itemId.toString(),
          skus: product.skus.map((s: any) => ({
            ...s,
            itemId: s.itemId.toString(),
            modelId: s.modelId?.toString() || null,
          })),
        },
      };
    } catch (error: any) {
      this.logger.error(`❌ Error getting product: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  async searchProducts(query: string): Promise<any> {
    try {
      const products = await this.prisma.shopeeProduct.findMany({
        where: {
          OR: [
            { name: { contains: query } },
            { description: { contains: query } },
          ],
        },
        include: { skus: true },
      });

      return {
        success: true,
        data: products.map((p) => ({
          ...p,
          itemId: p.itemId.toString(),
          skus: p.skus.map((s) => ({
            ...s,
            itemId: s.itemId.toString(),
            modelId: s.modelId?.toString() || null,
          })),
        })),
        count: products.length,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error searching products: ${error.message}`);
      return { success: false, error: error.message, data: [], count: 0 };
    }
  }
}

/**
 * Lazada Product DB Service
 * Handles database operations for Lazada products
 * Single Responsibility: Lazada product database persistence
 */

import { PrismaClient } from "@prisma/client";

export class LazadaProductDBService {
  constructor(
    private prisma: PrismaClient,
    private tenantId: string = "default"
  ) {}

  private logger = {
    info: (msg: string) => console.log(`[LazadaProductDB] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[LazadaProductDB] ❌ ${msg}`),
  };

  /**
   * Save products to database
   */
  async saveProducts(
    products: any[],
    cleanBeforeSave: boolean = false
  ): Promise<void> {
    try {
      if (cleanBeforeSave) {
        this.logger.info(`🧹 Cleaning old Lazada products before save...`);
        const deleteSkus = await (this.prisma as any).lazadaSku.deleteMany({
          where: {
            product: { tenantId: this.tenantId },
          },
        });
        const deleteProducts = await (
          this.prisma as any
        ).lazadaProduct.deleteMany({
          where: { tenantId: this.tenantId },
        });
        this.logger.info(
          `✅ Deleted ${deleteProducts.count} products and ${deleteSkus.count} skus`
        );
      }

      let savedProducts = 0;
      for (const product of products) {
        try {
          let priceFromSku = 0;
          let quantityFromSku = 0;

          if (product.skus && product.skus.length > 0) {
            const firstSku = product.skus[0];
            priceFromSku = firstSku.price || firstSku.special_price || 0;
            quantityFromSku = firstSku.quantity || firstSku.Available || 0;
          }

          const mainProduct = await (this.prisma as any).lazadaProduct.upsert({
            where: {
              tenantId_itemId: {
                tenantId: this.tenantId,
                itemId: String(product.item_id),
              },
            },
            create: {
              tenantId: this.tenantId,
              itemId: String(product.item_id),
              name: product.attributes?.name || product.name || "Unknown",
              description:
                product.attributes?.description || product.description,
              status: product.status || "Active",
              price: priceFromSku || product.price || 0,
              quantity: quantityFromSku || product.quantity || 0,
              image: product.images?.[0],
              createdAt: new Date(),
              updatedAt: new Date(),
            },
            update: {
              name: product.attributes?.name || product.name || "Unknown",
              description:
                product.attributes?.description || product.description,
              status: product.status || "Active",
              price: priceFromSku || product.price || 0,
              quantity: quantityFromSku || product.quantity || 0,
              image: product.images?.[0],
              updatedAt: new Date(),
            },
          });

          if (product.skus && Array.isArray(product.skus)) {
            for (const sku of product.skus) {
              await this.saveSku(mainProduct.id, sku);
            }
          }
          savedProducts += 1;
        } catch (e: any) {
          this.logger.error(
            `❌ Error saving product ${product.item_id}: ${e.message}`
          );
        }
      }

      this.logger.info(
        `✅ Saved ${savedProducts}/${products.length} products to database`
      );
    } catch (error: any) {
      this.logger.error(`❌ Error saving products: ${error.message}`);
    }
  }

  /**
   * Save SKU to database
   */
  private async saveSku(productId: number, sku: any): Promise<void> {
    let variantName = "";
    let variantData: any = null;

    if (sku.saleProp) {
      try {
        variantData =
          typeof sku.saleProp === "string"
            ? JSON.parse(sku.saleProp)
            : sku.saleProp;
        const values = Object.values(variantData);
        variantName = values.length > 0 ? String(values[0]) : "";
      } catch (e) {
        variantName = String(sku.saleProp);
      }
    }

    if (!variantName && sku.Pilihan) {
      variantName = String(sku.Pilihan);
      if (!variantData) {
        variantData = { Pilihan: sku.Pilihan };
      }
    }

    if (!variantName && sku.Variation) {
      variantName = String(sku.Variation);
      if (!variantData) {
        variantData = { Variation: sku.Variation };
      }
    }

    await (this.prisma as any).lazadaSku.upsert({
      where: { skuId: String(sku.SkuId || sku.ShopSku) },
      create: {
        productId,
        skuId: String(sku.SkuId || sku.ShopSku),
        sellerSku: sku.SellerSku,
        price: sku.price || sku.special_price || 0,
        quantity: sku.quantity || sku.Available || 0,
        variantName: variantName || null,
        variantData: variantData ? JSON.stringify(variantData) : null,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      update: {
        productId,
        sellerSku: sku.SellerSku,
        price: sku.price || sku.special_price || 0,
        quantity: sku.quantity || sku.Available || 0,
        variantName: variantName || null,
        variantData: variantData ? JSON.stringify(variantData) : null,
        updatedAt: new Date(),
      },
    });
  }

  /**
   * Get master products from database
   */
  async getMasterProducts(offset: number, limit: number): Promise<any> {
    try {
      // Note: No tenantId filter needed - Prisma connection is already tenant-specific
      const [products, total] = await Promise.all([
        (this.prisma as any).lazadaProduct.findMany({
          skip: offset,
          take: limit,
          orderBy: { createdAt: "desc" },
          include: { skus: true },
        }),
        (this.prisma as any).lazadaProduct.count(),
      ]);

      return {
        success: true,
        products,
        count: products.length,
        total,
        offset,
        limit,
      };
    } catch (error: any) {
      this.logger.error(
        `❌ Error retrieving products from DB: ${error.message}`
      );
      return {
        success: false,
        error: error.message,
        count: 0,
        total: 0,
        data: [],
      };
    }
  }

  /**
   * Get product list (summary) from database
   */
  async getProductList(offset: number, limit: number): Promise<any> {
    try {
      // Note: No tenantId filter needed - Prisma connection is already tenant-specific
      const [products, total] = await Promise.all([
        (this.prisma as any).lazadaProduct.findMany({
          skip: offset,
          take: limit,
          select: {
            itemId: true,
            name: true,
            status: true,
            price: true,
            quantity: true,
            createdAt: true,
          },
          orderBy: { createdAt: "desc" },
        }),
        (this.prisma as any).lazadaProduct.count(),
      ]);

      return {
        success: true,
        count: products.length,
        total,
        offset,
        limit,
        data: products,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving product list: ${error.message}`);
      return {
        success: false,
        error: error.message,
        count: 0,
        total: 0,
        data: [],
      };
    }
  }

  /**
   * Get product detail by Item ID
   */
  async getProductDetail(itemId: string): Promise<any> {
    try {
      const product = await (this.prisma as any).lazadaProduct.findUnique({
        where: {
          tenantId_itemId: {
            tenantId: this.tenantId,
            itemId: String(itemId),
          },
        },
        include: { skus: true },
      });

      if (!product) {
        return {
          success: false,
          error: "Product not found",
          product: null,
        };
      }

      return {
        success: true,
        product,
      };
    } catch (error: any) {
      this.logger.error(`❌ Error retrieving product detail: ${error.message}`);
      return {
        success: false,
        error: error.message,
        product: null,
      };
    }
  }
}

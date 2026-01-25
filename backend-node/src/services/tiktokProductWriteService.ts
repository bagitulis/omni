/**
 * TikTok Product Write Service
 * SRP: Write operations for TikTok products (save, delete)
 */

import { PrismaClient } from "@prisma/client";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { tenantContext } from "../utils/tenantContext";

interface ProductData {
  id: string;
  title?: string;
  product_name?: string;
  description?: string;
  status?: string;
  price?: any;
  quantity?: number;
  cover_image_url?: string;
  images?: string[];
  skus?: any[];
}

export class TiktokProductWriteService {
  private prisma: PrismaClient;
  private logger: Logger;

  constructor(prisma: PrismaClient) {
    this.prisma = prisma;
    this.logger = getLogger("TiktokProductWriteService");
  }

  async deleteProduct(productId: string): Promise<any> {
    this.logger.info(`🔄 Deleting TikTok product from DB | ID: ${productId}`);

    const product = await (this.prisma as any).tiktokProduct.findFirst({
      where: { productId },
    });
    if (!product) throw new Error(`Product ${productId} not found`);

    const deleted = await (this.prisma as any).tiktokProduct.delete({
      where: { id: product.id },
    });

    this.logger.info(`✅ Deleted product ${productId} from DB`);
    return deleted;
  }

  async saveProducts(
    products: ProductData[],
    cleanBeforeSave = false
  ): Promise<void> {
    try {
      if (cleanBeforeSave) {
        await this.cleanOldProducts();
      }

      const productMap = await this.saveProductRecords(products);
      const totalSkusSaved = await this.saveSkuRecords(products, productMap);

      this.logger.info(
        `✅ Saved ${productMap.size} products and ${totalSkusSaved} SKUs`
      );
    } catch (error: any) {
      this.logger.error(`❌ Error saving products: ${error.message}`);
      throw error;
    }
  }

  private async cleanOldProducts(): Promise<void> {
    this.logger.info(`🧹 Cleaning old TikTok products before save...`);
    await (this.prisma as any).tiktokSku.deleteMany({});
    const deleteResult = await (this.prisma as any).tiktokProduct.deleteMany(
      {}
    );
    this.logger.info(`✅ Deleted ${deleteResult.count} old TikTok products`);
  }

  private async saveProductRecords(
    products: ProductData[]
  ): Promise<Map<string, any>> {
    const productMap = new Map<string, any>();
    const tenantId = tenantContext.getTenantId();

    for (const product of products) {
      const productId = product.id;
      if (!productId) {
        this.logger.warn(`⚠️ Skipping product without id`);
        continue;
      }

      const { price, quantity } = this.extractPriceAndQuantity(product);
      const imageUrl = this.extractImageUrl(product);

      const savedProduct = await (this.prisma as any).tiktokProduct.upsert({
        where: { tenantId_productId: { tenantId, productId } },
        create: {
          tenantId,
          productId,
          name: product.title || product.product_name || "Untitled",
          description: product.description || "",
          status: product.status || "ACTIVE",
          price,
          quantity,
          image: imageUrl,
          createdAt: new Date(),
          updatedAt: new Date(),
        },
        update: {
          name: product.title || product.product_name || "Untitled",
          description: product.description || "",
          status: product.status || "ACTIVE",
          price,
          quantity,
          image: imageUrl,
          updatedAt: new Date(),
        },
      });

      productMap.set(productId, savedProduct);
    }

    return productMap;
  }

  private async saveSkuRecords(
    products: ProductData[],
    productMap: Map<string, any>
  ): Promise<number> {
    let totalSkusSaved = 0;

    for (const product of products) {
      const productId = product.id;
      if (!productId || !productMap.has(productId)) continue;

      const savedProduct = productMap.get(productId);
      if (!product.skus || !Array.isArray(product.skus)) continue;

      for (const sku of product.skus) {
        try {
          const skuId = sku.id || sku.sku_id;
          if (!skuId) continue;

          const { variantName, variantData } = this.extractVariantInfo(sku);
          const skuPrice = this.extractSkuPrice(sku);
          const skuQuantity = this.extractSkuQuantity(sku);

          await (this.prisma as any).tiktokSku.upsert({
            where: { skuId: String(skuId) },
            create: {
              productId: savedProduct.id,
              skuId: String(skuId),
              sellerSku: sku.seller_sku || sku.sellerSku || "",
              variantName: variantName || null,
              variantData: variantData ? JSON.stringify(variantData) : null,
              price: skuPrice,
              quantity: skuQuantity,
              createdAt: new Date(),
              updatedAt: new Date(),
            },
            update: {
              productId: savedProduct.id,
              sellerSku: sku.seller_sku || sku.sellerSku || "",
              variantName: variantName || null,
              variantData: variantData ? JSON.stringify(variantData) : null,
              price: skuPrice,
              quantity: skuQuantity,
              updatedAt: new Date(),
            },
          });

          totalSkusSaved++;
        } catch (skuError: any) {
          this.logger.error(`❌ Error saving SKU: ${skuError.message}`);
        }
      }
    }

    return totalSkusSaved;
  }

  private extractPriceAndQuantity(product: ProductData): {
    price: number;
    quantity: number;
  } {
    let price = 0,
      quantity = 0;
    if (product.skus && product.skus.length > 0) {
      const mainSku = product.skus[0];
      if (mainSku.price) {
        const priceValue =
          typeof mainSku.price === "object"
            ? mainSku.price.tax_exclusive_price
            : mainSku.price;
        price = parseFloat(String(priceValue)) || 0;
      }
      if (mainSku.inventory && mainSku.inventory.length > 0) {
        quantity = mainSku.inventory[0].quantity || 0;
      }
    }
    return { price, quantity };
  }

  private extractImageUrl(product: ProductData): string | null {
    if (product.cover_image_url) return product.cover_image_url;
    if (product.images && product.images.length > 0) return product.images[0];
    return null;
  }

  private extractVariantInfo(sku: any): {
    variantName: string;
    variantData: any;
  } {
    let variantName = "",
      variantData: any = null;
    if (sku.variation_attributes && Array.isArray(sku.variation_attributes)) {
      variantName = sku.variation_attributes
        .map((attr: any) => `${attr.attribute_name}:${attr.attribute_value}`)
        .join(" | ");
      variantData = sku.variation_attributes;
    } else if (sku.variant_names && Array.isArray(sku.variant_names)) {
      variantName = sku.variant_names.join(" | ");
      variantData = sku.variant_names;
    }
    return { variantName, variantData };
  }

  private extractSkuPrice(sku: any): number {
    if (!sku.price) return 0;
    const priceValue =
      typeof sku.price === "object" ? sku.price.tax_exclusive_price : sku.price;
    return parseFloat(String(priceValue)) || 0;
  }

  private extractSkuQuantity(sku: any): number {
    if (
      sku.inventory &&
      Array.isArray(sku.inventory) &&
      sku.inventory.length > 0
    ) {
      return sku.inventory[0].quantity || 0;
    }
    return 0;
  }
}

/**
 * Shopee Product DB Service
 * Handles product save/write operations to database
 * Single Responsibility: Product save operations only
 */

import { PrismaClient } from "@prisma/client";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { tenantContext } from "../utils/tenantContext";

export class ShopeeProductDBService {
  private prisma: PrismaClient;
  private logger: Logger;

  constructor(prisma: PrismaClient, logger?: Logger) {
    this.prisma = prisma;
    this.logger = logger || getLogger("ShopeeProductDBService");
  }

  async saveProducts(
    products: any[],
    cleanOld: boolean = false
  ): Promise<void> {
    try {
      this.logger.info(
        `💾 Saving ${products.length} Shopee products to database...`
      );

      if (cleanOld) {
        this.logger.info(`🗑️  Deleting old Shopee products from database...`);
        await this.prisma.shopeeSku.deleteMany({});
        await this.prisma.shopeeProduct.deleteMany({});
        this.logger.info(`✅ Old data deleted successfully`);
      }

      for (const product of products) {
        const itemId = BigInt(product.item_id);
        let totalQuantity = 0;

        // Extract price from price_info or model_list
        let mainPrice = 0;
        if (product.price_info && product.price_info.length > 0) {
          mainPrice = product.price_info[0].original_price || 0;
        } else if (product.model_list && product.model_list.length > 0) {
          mainPrice =
            product.model_list[0].price_info?.[0]?.original_price || 0;
        }

        // Extract stock from stock_info or stock_info_v2
        let itemStock = 0;
        if (product.stock_info_v2?.summary_info) {
          itemStock =
            product.stock_info_v2.summary_info.total_available_stock || 0;
        } else if (product.stock_info && product.stock_info.length > 0) {
          itemStock = product.stock_info[0].normal_stock || 0;
        }

        // If item has models, calculate total stock from models
        if (product.model_list && product.model_list.length > 0) {
          let calculatedTotalStock = 0;
          for (const model of product.model_list) {
            if (model.stock_info_v2?.summary_info) {
              calculatedTotalStock +=
                model.stock_info_v2.summary_info.total_available_stock || 0;
            } else if (model.stock_info && model.stock_info.length > 0) {
              calculatedTotalStock += model.stock_info[0].normal_stock || 0;
            }
          }
          if (calculatedTotalStock > 0 || itemStock === 0) {
            itemStock = calculatedTotalStock;
          }
        }

        // Upsert product - use tenant from context for write operations
        const tenantId = tenantContext.getTenantId();
        const savedProduct = await this.prisma.shopeeProduct.upsert({
          where: { tenantId_itemId: { tenantId, itemId } },
          update: {
            name: product.item_name,
            description: product.description,
            status: product.item_status,
            image: product.image?.image_url_list?.[0] || null,
            price: mainPrice,
            quantity: itemStock,
            updatedAt: new Date(),
          },
          create: {
            tenantId,
            itemId,
            name: product.item_name,
            description: product.description,
            status: product.item_status,
            image: product.image?.image_url_list?.[0] || null,
            price: mainPrice,
            quantity: itemStock,
          },
        });

        // Handle SKUs (Models in Shopee terms)
        if (product.model_list && product.model_list.length > 0) {
          for (const model of product.model_list) {
            const modelId = BigInt(model.model_id);
            const price = model.price_info?.[0]?.original_price || 0;

            let quantity = 0;
            if (model.stock_info_v2?.summary_info) {
              quantity =
                model.stock_info_v2.summary_info.total_available_stock || 0;
            } else if (model.stock_info && model.stock_info.length > 0) {
              quantity = model.stock_info[0].normal_stock || 0;
            }

            totalQuantity += quantity;

            // Check if SKU exists for this modelId
            const existingSku = await this.prisma.shopeeSku.findFirst({
              where: { modelId },
            });

            if (existingSku) {
              // Update existing SKU
              await this.prisma.shopeeSku.update({
                where: { id: existingSku.id },
                data: {
                  sellerSku: model.model_sku,
                  price: price,
                  quantity: quantity,
                  variantName: model.tier_index
                    ?.map((idx: number, i: number) => {
                      const tier = product.tier_variation?.[i];
                      return tier?.option_list?.[idx]?.option;
                    })
                    .join(", "),
                  updatedAt: new Date(),
                },
              });
            } else {
              // Create new SKU
              await this.prisma.shopeeSku.create({
                data: {
                  productId: savedProduct.id,
                  itemId: itemId,
                  modelId: modelId,
                  sellerSku: model.model_sku,
                  price: price,
                  quantity: quantity,
                  variantName: model.tier_index
                    ?.map((idx: number, i: number) => {
                      const tier = product.tier_variation?.[i];
                      return tier?.option_list?.[idx]?.option;
                    })
                    .join(", "),
                },
              });
            }
          }
        } else {
          // Single product (no models)
          totalQuantity = itemStock;

          // For single products without models, check if SKU exists
          const existingSku = await this.prisma.shopeeSku.findFirst({
            where: {
              itemId: itemId,
              modelId: null,
            },
          });

          if (existingSku) {
            // Update existing SKU
            await this.prisma.shopeeSku.update({
              where: { id: existingSku.id },
              data: {
                sellerSku: product.item_sku,
                price: mainPrice,
                quantity: itemStock,
                updatedAt: new Date(),
              },
            });
          } else {
            // Create new SKU with null modelId
            await this.prisma.shopeeSku.create({
              data: {
                productId: savedProduct.id,
                itemId: itemId,
                modelId: null,
                sellerSku: product.item_sku,
                price: mainPrice,
                quantity: itemStock,
              },
            });
          }
        }

        // Update product total quantity and price
        await this.prisma.shopeeProduct.update({
          where: { id: savedProduct.id },
          data: {
            quantity: totalQuantity,
            price: mainPrice,
          },
        });
      }

      this.logger.info(
        `✅ Successfully saved ${products.length} Shopee products`
      );
    } catch (error: any) {
      this.logger.error(
        `❌ Error saving Shopee products to DB: ${error.message}`
      );
      throw error;
    }
  }
}

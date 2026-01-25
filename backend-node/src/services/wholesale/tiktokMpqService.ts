/**
 * TikTok MPQ (Minimum Purchase Quantity) Service
 * RESPONSIBILITY: Handle MPQ operations for TikTok products
 *
 * TikTok API:
 * - PUT /product/{version}/products/{product_id}
 * - Payload includes: minimum_order_quantity, skus with price
 *
 * IMPORTANT: TikTok does NOT support wholesale tiers, only MPQ + price
 */

import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";
import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";
import { Logger } from "winston";

export interface MpqUpdateResult {
  success: boolean;
  productId: string;
  message?: string;
  error?: string;
}

export interface TiktokSkuLookup {
  productId: string;
  skuId: string | null;
  sellerSku: string | null;
  productName?: string;
  currentPrice?: number;
}

export class TiktokMpqService {
  private logger: Logger;
  private apiVersion = "202309";

  constructor(
    private apiClient: TiktokAPIClient,
    _config: TiktokConfigManager, // Reserved for future use
    private prisma?: PrismaClient
  ) {
    this.logger = getLogger("TiktokMpqService");
  }

  /**
   * Lookup product_id from SKU
   * NOTE: result.productId is FK (int), we need result.product.productId (TikTok API ID string)
   */
  async lookupProductBySku(sku: string): Promise<TiktokSkuLookup | null> {
    if (!this.prisma) {
      this.logger.warn("⚠️ Prisma not available for SKU lookup");
      return null;
    }

    try {
      // Try sellerSku first (like productIdFetcher does)
      let result = await this.prisma.tiktokSku.findFirst({
        where: { sellerSku: sku },
        include: { product: true },
      });

      // If not found, try skuId
      if (!result) {
        result = await this.prisma.tiktokSku.findFirst({
          where: { skuId: sku },
          include: { product: true },
        });
      }

      if (!result || !result.product) {
        this.logger.warn(`⚠️ TikTok SKU not found in DB: ${sku}`);
        return null;
      }

      // IMPORTANT: Use result.product.productId (TikTok API ID), NOT result.productId (FK)
      this.logger.info(
        `✅ Found TikTok product: ${result.product.productId} for SKU: ${sku}`
      );

      return {
        productId: result.product.productId, // This is the TikTok API product ID!
        skuId: result.skuId?.toString() || null,
        sellerSku: result.sellerSku,
        productName: result.product?.name || "Unknown",
        currentPrice: result.price ? Number(result.price) : undefined,
      };
    } catch (error: any) {
      this.logger.error(`❌ SKU lookup error: ${error.message}`);
      return null;
    }
  }

  /**
   * Get raw product details from TikTok API
   */
  async getRawProductDetails(productId: string): Promise<any | null> {
    try {
      const endpoint = `/product/${this.apiVersion}/products/${productId}`;
      const response = await this.apiClient.request(endpoint, "GET");

      if (response && response.code === 0) {
        return response.data;
      }
      return null;
    } catch (error: any) {
      this.logger.error(`❌ Get product error: ${error.message}`);
      return null;
    }
  }

  /**
   * Update price for a TikTok product using dedicated price endpoint
   * TikTok API: POST /product/{version}/products/{product_id}/prices/update
   */
  async updatePrice(
    productId: string,
    skuId: string,
    newPrice: number
  ): Promise<{ success: boolean; error?: string }> {
    try {
      this.logger.info(
        `💰 Updating TikTok price: product=${productId}, sku=${skuId}, price=${newPrice}`
      );

      const endpoint = `/product/${this.apiVersion}/products/${productId}/prices/update`;
      const payload = {
        skus: [
          {
            id: skuId,
            price: {
              currency: "IDR",
              amount: String(Math.round(newPrice)),
            },
            external_list_prices: [],
          },
        ],
      };

      this.logger.info(`📤 TikTok price payload: ${JSON.stringify(payload)}`);

      const response = await this.apiClient.request(
        endpoint,
        "POST",
        {},
        payload
      );

      if (response && response.code === 0) {
        this.logger.info(`✅ TikTok price updated: ${newPrice}`);
        return { success: true };
      }

      const errorMsg = response?.message || "Unknown error";
      this.logger.error(`❌ TikTok price update failed: ${errorMsg}`);
      return { success: false, error: errorMsg };
    } catch (error: any) {
      this.logger.error(`❌ TikTok price update error: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  /**
   * Update MPQ (minimum_order_quantity) and price for a product
   * Step 1: Update MPQ via PUT /products/{id}
   * Step 2: Update price via POST /products/{id}/prices/update (separate endpoint)
   */
  async updateMpq(
    productId: string,
    mpq: number,
    newPrice?: number
  ): Promise<MpqUpdateResult> {
    try {
      this.logger.info(
        `📦 Updating MPQ for TikTok product: ${productId}, mpq=${mpq}, price=${
          newPrice || "N/A"
        }`
      );

      // Get current product data
      const productData = await this.getRawProductDetails(productId);
      if (!productData) {
        this.logger.error(`❌ TikTok product not found: ${productId}`);
        return {
          success: false,
          productId,
          error: `Product ${productId} not found`,
        };
      }

      // Get first SKU ID for price update later
      const firstSku = productData.skus?.[0];
      const skuId = firstSku?.id;

      // Build minimal update payload for MPQ only (don't include price in PUT)
      const payload = this.buildUpdatePayload(productData, mpq);

      const endpoint = `/product/${this.apiVersion}/products/${productId}`;
      const response = await this.apiClient.request(
        endpoint,
        "PUT",
        {},
        payload
      );

      if (response && response.code === 0) {
        this.logger.info(`✅ MPQ updated for TikTok product: ${productId}`);

        // Step 2: Update price via dedicated endpoint if provided
        if (newPrice && newPrice > 0 && skuId) {
          const priceResult = await this.updatePrice(
            productId,
            skuId,
            newPrice
          );
          if (!priceResult.success) {
            this.logger.warn(
              `⚠️ MPQ updated but price failed: ${priceResult.error}`
            );
            return {
              success: true, // MPQ succeeded
              productId,
              message: `MPQ set to ${mpq}, but price update failed: ${priceResult.error}`,
            };
          }
        }

        return {
          success: true,
          productId,
          message: `MPQ set to ${mpq}${newPrice ? `, price: ${newPrice}` : ""}`,
        };
      }

      const errorMsg =
        response?.message || JSON.stringify(response) || "Unknown error";
      this.logger.error(
        `❌ TikTok update MPQ failed [${productId}]: ${errorMsg}`
      );
      return { success: false, productId, error: errorMsg };
    } catch (error: any) {
      this.logger.error(
        `❌ TikTok update MPQ error [${productId}]: ${error.message}`
      );
      return { success: false, productId, error: error.message };
    }
  }

  /**
   * Build minimal payload for product update
   * IMPORTANT: Must include product_attributes - TikTok requires them
   */
  private buildUpdatePayload(productData: any, mpq: number): any {
    const payload: any = { minimum_order_quantity: mpq };

    // Add required fields
    if (productData.brand?.id) payload.brand_id = productData.brand.id;
    if (productData.category_chains?.length) {
      payload.category_id = productData.category_chains.slice(-1)[0]?.id;
    }
    if (productData.description) payload.description = productData.description;
    if (productData.main_images) payload.main_images = productData.main_images;
    if (productData.title) payload.title = productData.title;
    if (productData.skus) payload.skus = productData.skus;
    if (productData.package_weight)
      payload.package_weight = productData.package_weight;
    if (productData.video) payload.video = productData.video;

    // CRITICAL: Include product_attributes - contains required fields like "Contains Dangerous Goods?"
    if (productData.product_attributes) {
      payload.product_attributes = productData.product_attributes;
    }

    payload.is_cod_allowed = true;

    return payload;
  }

  /**
   * Batch update MPQ by SKUs
   */
  async batchUpdateMpq(
    skuPriceMap: Map<string, number>,
    mpq: number
  ): Promise<{
    success: boolean;
    total: number;
    uniqueProducts: number;
    processed: number;
    failed: number;
    results: MpqUpdateResult[];
    skipped: string[];
  }> {
    const skus = Array.from(skuPriceMap.keys());
    this.logger.info(`🔄 Batch MPQ update for ${skus.length} TikTok SKUs`);

    // Lookup and dedupe by productId
    const productMap = new Map<
      string,
      { name: string; skus: string[]; price: number }
    >();
    const notFound: string[] = [];

    for (const sku of skus) {
      const lookup = await this.lookupProductBySku(sku);
      if (lookup) {
        const existing = productMap.get(lookup.productId);
        if (existing) {
          existing.skus.push(sku);
        } else {
          productMap.set(lookup.productId, {
            name: lookup.productName || sku,
            skus: [sku],
            price: skuPriceMap.get(sku) || 0,
          });
        }
      } else {
        notFound.push(sku);
      }
    }

    this.logger.info(
      `📊 Deduped: ${skus.length} SKUs → ${productMap.size} unique products`
    );

    // Update each product
    const results: MpqUpdateResult[] = [];
    let processed = 0;
    let failed = 0;

    for (const [productId, data] of productMap) {
      const result = await this.updateMpq(productId, mpq, data.price);
      result.message = `${data.name} (SKUs: ${data.skus.join(", ")})`;
      results.push(result);

      if (result.success) processed++;
      else failed++;
    }

    return {
      success: failed === 0,
      total: skus.length,
      uniqueProducts: productMap.size,
      processed,
      failed,
      results,
      skipped: notFound,
    };
  }
}

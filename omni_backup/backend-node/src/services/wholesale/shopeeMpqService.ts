/**
 * Shopee MPQ (Minimum Purchase Quantity) Service
 * SRP: Handle MPQ operations for Shopee products
 */
import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { getLogger } from "../../utils/logger";
import { Logger } from "winston";
import { ShopeeSkuLookupService } from "./shopeeSkuLookupService";

export interface ShopeeMpqResult {
  success: boolean;
  itemId: number;
  message?: string;
  error?: string;
}

export class ShopeeMpqService {
  private logger: Logger;

  constructor(
    private apiClient: ShopeeAPIClient,
    private skuLookup: ShopeeSkuLookupService
  ) {
    this.logger = getLogger("ShopeeMpqService");
  }

  /**
   * Set MPQ (min_purchase_limit) for an item
   */
  async setMpq(itemId: number, mpq: number): Promise<ShopeeMpqResult> {
    try {
      this.logger.info(`🔢 Setting MPQ=${mpq} for item: ${itemId}`);

      const payload = {
        item_id: itemId,
        purchase_limit_info: {
          min_purchase_limit: mpq,
        },
      };

      this.logger.info(`📤 Shopee setMpq payload: ${JSON.stringify(payload)}`);

      const response = await this.apiClient.request(
        "/api/v2/product/update_item",
        "POST",
        {},
        payload
      );

      this.logger.info(
        `📥 Shopee setMpq response: ${JSON.stringify(response)}`
      );

      if (response.error || response.message) {
        const errMsg = response.message || response.error || "Unknown error";
        this.logger.error(`❌ Set MPQ failed [${itemId}]: ${errMsg}`);
        return { success: false, itemId, error: errMsg };
      }

      this.logger.info(`✅ MPQ set to ${mpq} for item: ${itemId}`);
      return { success: true, itemId, message: `MPQ set to ${mpq}` };
    } catch (error: any) {
      this.logger.error(`❌ Set MPQ error [${itemId}]: ${error.message}`);
      return { success: false, itemId, error: error.message };
    }
  }

  /**
   * Update price for an item/model
   */
  async updatePrice(
    itemId: number,
    modelId: number | null,
    newPrice: number
  ): Promise<ShopeeMpqResult> {
    try {
      this.logger.info(
        `💰 Updating price for item: ${itemId}, price=${newPrice}`
      );

      const payload = {
        item_id: itemId,
        price_list: [
          {
            model_id: modelId || 0,
            original_price: newPrice,
          },
        ],
      };

      const response = await this.apiClient.request(
        "/api/v2/product/update_price",
        "POST",
        {},
        payload
      );

      if (response.error || response.message) {
        const errMsg = response.message || response.error || "Unknown error";
        this.logger.error(`❌ Update price failed [${itemId}]: ${errMsg}`);
        return { success: false, itemId, error: errMsg };
      }

      return { success: true, itemId, message: `Price updated to ${newPrice}` };
    } catch (error: any) {
      this.logger.error(`❌ Update price error [${itemId}]: ${error.message}`);
      return { success: false, itemId, error: error.message };
    }
  }

  /**
   * Set MPQ mode: Delete wholesale → Update price → Set MPQ
   */
  async setMpqMode(
    itemId: number,
    mpq: number,
    newPrice: number,
    modelId?: number | null,
    deleteWholesaleFn?: (id: number) => Promise<any>
  ): Promise<ShopeeMpqResult> {
    try {
      this.logger.info(
        `🔢 Setting MPQ mode for item: ${itemId}, mpq=${mpq}, price=${newPrice}`
      );

      // Step 1: Delete wholesale if function provided
      if (deleteWholesaleFn) {
        const deleteResult = await deleteWholesaleFn(itemId);
        if (!deleteResult.success) {
          this.logger.warn(
            `⚠️ Delete wholesale warning: ${deleteResult.error}`
          );
        }
      }

      // Step 2: Update price if provided
      if (newPrice && newPrice > 0) {
        const priceResult = await this.updatePrice(
          itemId,
          modelId || null,
          newPrice
        );
        if (!priceResult.success) {
          this.logger.warn(`⚠️ Update price warning: ${priceResult.error}`);
        } else {
          this.logger.info(
            `✅ Price updated to ${newPrice} for item: ${itemId}`
          );
        }
      }

      // Step 3: Set MPQ
      return await this.setMpq(itemId, mpq);
    } catch (error: any) {
      this.logger.error(`❌ Set MPQ mode error: ${error.message}`);
      return { success: false, itemId, error: error.message };
    }
  }

  /**
   * Batch set MPQ by SKUs
   */
  async batchSetMpqBySkus(
    skuPriceMap: Map<string, number>,
    mpq: number,
    deleteWholesaleFn?: (id: number) => Promise<any>
  ): Promise<{
    success: boolean;
    total: number;
    uniqueItems: number;
    processed: number;
    failed: number;
    results: ShopeeMpqResult[];
    skipped: string[];
  }> {
    const skus = Array.from(skuPriceMap.keys());
    this.logger.info(`🔢 Batch MPQ for ${skus.length} SKUs, MPQ=${mpq}`);

    const { itemIdMap, notFound } =
      await this.skuLookup.batchLookupWithPrices(skuPriceMap);

    const results: ShopeeMpqResult[] = [];
    let processed = 0;
    let failed = 0;

    for (const [itemId, data] of itemIdMap) {
      const result = await this.setMpqMode(
        Number(itemId),
        mpq,
        data.basePrice,
        data.modelId ? Number(data.modelId) : null,
        deleteWholesaleFn
      );
      result.message = `${data.name} (${data.skus.join(", ")}) - ${
        result.message
      }`;
      results.push(result);

      if (result.success) processed++;
      else failed++;
    }

    return {
      success: failed === 0,
      total: skus.length,
      uniqueItems: itemIdMap.size,
      processed,
      failed,
      results,
      skipped: notFound,
    };
  }
}

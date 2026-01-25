import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("ShopeePriceUpdater");

/**
 * ShopeePriceUpdater
 * RESPONSIBILITY: Update price on Shopee platform
 * API: POST /api/v2/product/update_price
 *
 * Payload format:
 * {
 *   "item_id": number,
 *   "price_list": [{
 *     "model_id": number,
 *     "original_price": number
 *   }]
 * }
 */
export interface ShopeePriceUpdateRequest {
  itemId: string; // String (converted from BigInt for JSON serialization)
  modelId: string | null; // String (converted from BigInt for JSON serialization, nullable for single products)
  price: number;
}

export interface ShopeePriceUpdateResult {
  success: boolean;
  itemId: string;
  modelId: string | null;
  error?: string;
}

export class ShopeePriceUpdater {
  constructor(
    private apiClient: ShopeeAPIClient,
    _config: ShopeeConfigManager // Reserved for future use
  ) {}

  /**
   * Update price for a single Shopee product/model
   * API Request:
   * {
   *   "item_id": 1000,
   *   "price_list": [{
   *     "model_id": 0,
   *     "original_price": 10000
   *   }]
   * }
   */
  async updatePrice(
    request: ShopeePriceUpdateRequest
  ): Promise<ShopeePriceUpdateResult> {
    const { itemId, modelId, price } = request;

    try {
      const payload = {
        item_id: Number(itemId),
        price_list: [
          {
            model_id: Number(modelId),
            original_price: Math.max(0, price),
          },
        ],
      };

      const response = await this.apiClient.request(
        "/api/v2/product/update_price",
        "POST",
        {},
        payload
      );

      if (response.error) {
        logger.error(`Shopee price update failed: ${response.message}`);
        return {
          success: false,
          itemId: String(itemId),
          modelId: String(modelId),
          error: response.message || response.error,
        };
      }

      // Check for failures in response
      if (response.response?.failure_list?.length > 0) {
        const failure = response.response.failure_list[0];
        return {
          success: false,
          itemId: String(itemId),
          modelId: String(modelId),
          error: failure.failed_reason,
        };
      }

      return {
        success: true,
        itemId: String(itemId),
        modelId: String(modelId),
      };
    } catch (error: any) {
      logger.error(`Shopee price update error: ${error.message}`);
      return {
        success: false,
        itemId: String(itemId),
        modelId: String(modelId),
        error: error.message,
      };
    }
  }

  /**
   * Batch update price for multiple models
   */
  async updateBatch(
    requests: ShopeePriceUpdateRequest[]
  ): Promise<ShopeePriceUpdateResult[]> {
    const results: ShopeePriceUpdateResult[] = [];

    for (const request of requests) {
      const result = await this.updatePrice(request);
      results.push(result);

      // Small delay to avoid rate limiting
      await this.delay(200);
    }

    return results;
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

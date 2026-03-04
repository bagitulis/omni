import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("ShopeeStockUpdater");

/**
 * ShopeeStockUpdater
 * RESPONSIBILITY: Update stock on Shopee platform
 * API: POST /api/v2/product/update_stock
 */
export interface ShopeeStockUpdateRequest {
  itemId: string; // String (converted from BigInt for JSON serialization)
  modelId: string | null; // String (converted from BigInt for JSON serialization, nullable for single products)
  stock: number;
}

export interface ShopeeStockUpdateResult {
  success: boolean;
  itemId: string;
  modelId: string | null;
  error?: string;
}

export class ShopeeStockUpdater {
  constructor(
    private apiClient: ShopeeAPIClient,
    _config: ShopeeConfigManager // Reserved for future use
  ) {}

  /**
   * Update stock for a single Shopee product/model
   * API Request:
   * {
   *   "item_id": 1000,
   *   "stock_list": [{
   *     "model_id": 0,
   *     "seller_stock": [{ "location_id": "-", "stock": 0 }]
   *   }]
   * }
   */
  async updateStock(
    request: ShopeeStockUpdateRequest
  ): Promise<ShopeeStockUpdateResult> {
    const { itemId, modelId, stock } = request;

    try {
      const payload = {
        item_id: Number(itemId),
        stock_list: [
          {
            model_id: Number(modelId),
            seller_stock: [
              {
                stock: Math.max(0, stock),
              },
            ],
          },
        ],
      };

      const response = await this.apiClient.request(
        "/api/v2/product/update_stock",
        "POST",
        {},
        payload
      );

      if (response.error) {
        logger.error(`Shopee stock update failed: ${response.message}`);
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
      logger.error(`Shopee stock update error: ${error.message}`);
      return {
        success: false,
        itemId: String(itemId),
        modelId: String(modelId),
        error: error.message,
      };
    }
  }

  /**
   * Batch update stock for multiple items
   */
  async updateBatch(
    requests: ShopeeStockUpdateRequest[]
  ): Promise<ShopeeStockUpdateResult[]> {
    const results: ShopeeStockUpdateResult[] = [];

    for (const request of requests) {
      const result = await this.updateStock(request);
      results.push(result);

      // Small delay to avoid rate limiting
      await this.delay(100);
    }

    return results;
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

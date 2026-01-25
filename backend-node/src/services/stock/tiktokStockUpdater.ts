import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("TiktokStockUpdater");

/**
 * TiktokStockUpdater
 * RESPONSIBILITY: Update stock on TikTok platform
 * API: POST /product/202309/products/{product_id}/inventory/update
 *
 * Payload format:
 * {
 *   "skus": [{
 *     "id": "sku_id",
 *     "inventory": [{
 *       "warehouse_id": "warehouse_id",
 *       "quantity": number
 *     }]
 *   }]
 * }
 */
export interface TiktokStockUpdateRequest {
  productId: string;
  skuId: string;
  stock: number;
  warehouseId?: string;
}

export interface TiktokStockUpdateResult {
  success: boolean;
  productId: string;
  skuId: string;
  error?: string;
}

export class TiktokStockUpdater {
  private defaultWarehouseId: string | null = null;

  constructor(
    private apiClient: TiktokAPIClient,
    private config: TiktokConfigManager
  ) {}

  /**
   * Get warehouse ID - TikTok requires warehouse_id for inventory updates
   * First tries to get from config (database), then falls back to API
   */
  async getWarehouseId(): Promise<string | null> {
    if (this.defaultWarehouseId) {
      return this.defaultWarehouseId;
    }

    try {
      // First, try to get from config (database)
      const configWarehouseId = await this.config.getConfig("warehouse_id");
      if (configWarehouseId) {
        this.defaultWarehouseId = configWarehouseId;
        return this.defaultWarehouseId;
      }

      // Fallback: Try API endpoints
      const endpoints = [
        "/fulfillment/202309/warehouses",
        "/supply_chain/202309/warehouses",
        "/logistics/202309/warehouses",
      ];

      for (const endpoint of endpoints) {
        try {
          const response = await this.apiClient.request(endpoint, "GET", {});

          if (response.data?.warehouses?.length > 0) {
            this.defaultWarehouseId = response.data.warehouses[0].id;
            return this.defaultWarehouseId;
          }
        } catch (e) {
          // Try next endpoint
        }
      }

      // If no warehouses found, use default fallback
      logger.warn("No TikTok warehouses found in config or API");
      return null;
    } catch (error: any) {
      logger.error(`Failed to get TikTok warehouses: ${error.message}`);
      return null;
    }
  }

  /**
   * Update stock for a single TikTok product SKU
   * TikTok API requires product_id in path and skus array in body
   * API v202309: POST /product/202309/products/{product_id}/inventory/update
   *
   * Working format (from API Tool):
   * {"skus":[{"id":"sku_id","inventory":[{"quantity":number}]}]}
   * NO warehouse_id needed for basic stock update
   */
  async updateStock(
    request: TiktokStockUpdateRequest
  ): Promise<TiktokStockUpdateResult> {
    const { productId, skuId, stock } = request;

    try {
      // TikTok v202309 inventory update format
      // Simple format without warehouse_id
      const payload = {
        skus: [
          {
            id: skuId,
            inventory: [
              {
                quantity: Math.max(0, stock),
              },
            ],
          },
        ],
      };

      // IMPORTANT: request(endpoint, method, params, data)
      // params = query params (empty {}), data = POST body (payload)
      const response = await this.apiClient.request(
        `/product/202309/products/${productId}/inventory/update`,
        "POST",
        {}, // params (query string) - empty
        payload // data (POST body) - the actual payload
      );

      // TikTok returns code 0 for success
      if (response.code !== 0) {
        logger.error(`TikTok stock update failed: ${response.message || JSON.stringify(response)}`);
        return {
          success: false,
          productId,
          skuId,
          error: response.message || `Error code: ${response.code}`,
        };
      }

      return {
        success: true,
        productId,
        skuId,
      };
    } catch (error: any) {
      logger.error(`TikTok stock update error: ${error.response?.data?.message || error.message}`);
      return {
        success: false,
        productId,
        skuId,
        error: error.response?.data?.message || error.message,
      };
    }
  }

  /**
   * Batch update stock for multiple SKUs
   */
  async updateBatch(
    requests: TiktokStockUpdateRequest[]
  ): Promise<TiktokStockUpdateResult[]> {
    const results: TiktokStockUpdateResult[] = [];

    for (const request of requests) {
      const result = await this.updateStock(request);
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

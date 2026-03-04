import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { TiktokConfigManager } from "../../config/managers/tiktokConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("TiktokPriceUpdater");

/**
 * TiktokPriceUpdater
 * RESPONSIBILITY: Update price on TikTok platform
 * API: POST /product/202309/products/{product_id}/prices/update
 *
 * Payload format:
 * {
 *   "skus": [{
 *     "id": "sku_id",
 *     "price": {
 *       "amount": string,
 *       "currency": string
 *     }
 *   }]
 * }
 *
 * Note: TikTok price uses "amount" as string and requires currency code
 */
export interface TiktokPriceUpdateRequest {
  productId: string;
  skuId: string;
  price: number;
  currency?: string; // Default: IDR for Indonesia
}

export interface TiktokPriceUpdateResult {
  success: boolean;
  productId: string;
  skuId: string;
  error?: string;
}

export class TiktokPriceUpdater {
  private defaultCurrency: string = "IDR";

  constructor(
    private apiClient: TiktokAPIClient,
    private config: TiktokConfigManager
  ) {}

  /**
   * Get currency code from config or use default
   */
  async getCurrencyCode(): Promise<string> {
    try {
      const configCurrency = await this.config.getConfig("currency");
      if (configCurrency) {
        return configCurrency;
      }
    } catch (e) {
      // Use default
    }
    return this.defaultCurrency;
  }

  /**
   * Update price for a single TikTok product SKU
   * TikTok API requires product_id in path and skus array in body
   * API v202309: POST /product/202309/products/{product_id}/prices/update
   *
   * Payload format:
   * {"skus":[{"id":"sku_id","price":{"amount":"10000","currency":"IDR"}}]}
   */
  async updatePrice(
    request: TiktokPriceUpdateRequest
  ): Promise<TiktokPriceUpdateResult> {
    const { productId, skuId, price, currency } = request;

    try {
      const currencyCode = currency || await this.getCurrencyCode();

      // TikTok v202309 price update format
      // Price amount must be string and currency is required
      const payload = {
        skus: [
          {
            id: skuId,
            price: {
              amount: String(Math.max(0, price)),
              currency: currencyCode,
            },
          },
        ],
      };

      // IMPORTANT: request(endpoint, method, params, data)
      // params = query params (empty {}), data = POST body (payload)
      const response = await this.apiClient.request(
        `/product/202309/products/${productId}/prices/update`,
        "POST",
        {}, // params (query string) - empty
        payload // data (POST body) - the actual payload
      );

      // Check for errors in response
      if (response.code !== 0) {
        logger.error(`TikTok price update failed: ${response.message}`);
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
      logger.error(`TikTok price update error: ${error.message}`);
      return {
        success: false,
        productId,
        skuId,
        error: error.message,
      };
    }
  }

  /**
   * Batch update price for multiple SKUs
   */
  async updateBatch(
    requests: TiktokPriceUpdateRequest[]
  ): Promise<TiktokPriceUpdateResult[]> {
    const results: TiktokPriceUpdateResult[] = [];

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

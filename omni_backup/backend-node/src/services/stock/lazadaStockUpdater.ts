import { LazadaAPIClient } from "../../api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "../../config/managers/lazadaConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("LazadaStockUpdater");

/**
 * LazadaStockUpdater
 * RESPONSIBILITY: Update stock on Lazada platform
 * API: POST /product/price_quantity/update
 *
 * Payload format (XML-style in JSON):
 * <Request>
 *   <Product><Skus><Sku>
 *     <ItemId>...</ItemId>
 *     <SkuId>...</SkuId>
 *     <SellerSku>...</SellerSku>
 *     <Quantity>...</Quantity>
 *   </Sku></Skus></Product>
 * </Request>
 */
export interface LazadaStockUpdateRequest {
  itemId: string;
  skuId: string;
  sellerSku: string;
  stock: number;
}

export interface LazadaStockUpdateResult {
  success: boolean;
  itemId: string;
  skuId: string;
  error?: string;
}

export class LazadaStockUpdater {
  constructor(
    private apiClient: LazadaAPIClient,
    _config: LazadaConfigManager // Reserved for future use
  ) {}

  /**
   * Update stock for a single Lazada product SKU
   * Lazada uses XML-style payload parameter
   */
  async updateStock(
    request: LazadaStockUpdateRequest
  ): Promise<LazadaStockUpdateResult> {
    const { itemId, skuId, sellerSku, stock } = request;

    try {
      // Lazada expects XML payload as string parameter
      const xmlPayload = `<Request>
  <Product>
    <Skus>
      <Sku>
        <ItemId>${itemId}</ItemId>
        <SkuId>${skuId}</SkuId>
        <SellerSku>${sellerSku}</SellerSku>
        <Quantity>${Math.max(0, stock)}</Quantity>
      </Sku>
    </Skus>
  </Product>
</Request>`;

      const response = await this.apiClient.request(
        "/product/price_quantity/update",
        "POST",
        { payload: xmlPayload },
        {}
      );

      // Lazada returns code "0" for success
      if (response.code !== "0") {
        logger.error(`Lazada stock update failed: ${response.message}`);
        return {
          success: false,
          itemId,
          skuId,
          error: response.message || `Error code: ${response.code}`,
        };
      }

      return {
        success: true,
        itemId,
        skuId,
      };
    } catch (error: any) {
      logger.error(`Lazada stock update error: ${error.message}`);
      return {
        success: false,
        itemId,
        skuId,
        error: error.message,
      };
    }
  }

  /**
   * Batch update stock for multiple SKUs
   */
  async updateBatch(
    requests: LazadaStockUpdateRequest[]
  ): Promise<LazadaStockUpdateResult[]> {
    const results: LazadaStockUpdateResult[] = [];

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

import { LazadaAPIClient } from "../../api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "../../config/managers/lazadaConfigManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("LazadaPriceUpdater");

/**
 * LazadaPriceUpdater
 * RESPONSIBILITY: Update price on Lazada platform
 * API: POST /product/price_quantity/update
 *
 * Payload format (XML-style in JSON):
 * <Request>
 *   <Product><Skus><Sku>
 *     <ItemId>...</ItemId>
 *     <SkuId>...</SkuId>
 *     <SellerSku>...</SellerSku>
 *     <Price>...</Price>
 *     <SalePrice>...</SalePrice>
 *     <SaleStartDate>...</SaleStartDate>
 *     <SaleEndDate>...</SaleEndDate>
 *   </Sku></Skus></Product>
 * </Request>
 *
 * Note: Same endpoint as stock update, but with Price/SalePrice fields
 */
export interface LazadaPriceUpdateRequest {
  itemId: string;
  skuId: string;
  sellerSku: string;
  price: number;
  salePrice?: number;
  saleStartDate?: string; // ISO format or Lazada format
  saleEndDate?: string;
}

export interface LazadaPriceUpdateResult {
  success: boolean;
  itemId: string;
  skuId: string;
  error?: string;
}

export class LazadaPriceUpdater {
  constructor(
    private apiClient: LazadaAPIClient,
    _config: LazadaConfigManager // Reserved for future use
  ) {}

  /**
   * Update price for a single Lazada product SKU
   * Lazada uses XML-style payload parameter
   */
  async updatePrice(
    request: LazadaPriceUpdateRequest
  ): Promise<LazadaPriceUpdateResult> {
    const { itemId, skuId, sellerSku, price, salePrice, saleStartDate, saleEndDate } = request;

    try {
      // Build XML payload with optional sale price fields
      let xmlContent = `
        <ItemId>${itemId}</ItemId>
        <SkuId>${skuId}</SkuId>
        <SellerSku>${sellerSku}</SellerSku>
        <Price>${Math.max(0, price)}</Price>`;

      // Add sale price fields if provided
      if (salePrice !== undefined && salePrice > 0) {
        xmlContent += `
        <SalePrice>${salePrice}</SalePrice>`;
        
        if (saleStartDate) {
          xmlContent += `
        <SaleStartDate>${saleStartDate}</SaleStartDate>`;
        }
        if (saleEndDate) {
          xmlContent += `
        <SaleEndDate>${saleEndDate}</SaleEndDate>`;
        }
      }

      const xmlPayload = `<Request>
  <Product>
    <Skus>
      <Sku>${xmlContent}
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
        logger.error(`Lazada price update failed: ${response.message}`);
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
      logger.error(`Lazada price update error: ${error.message}`);
      return {
        success: false,
        itemId,
        skuId,
        error: error.message,
      };
    }
  }

  /**
   * Batch update price for multiple SKUs
   */
  async updateBatch(
    requests: LazadaPriceUpdateRequest[]
  ): Promise<LazadaPriceUpdateResult[]> {
    const results: LazadaPriceUpdateResult[] = [];

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

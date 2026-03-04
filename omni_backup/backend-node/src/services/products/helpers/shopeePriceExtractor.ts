/**
 * Extracts price information from Shopee product data
 * Handles both single-item and variant pricing
 */
export class ShopeePriceExtractor {
  static extractMainPrice(product: any): number {
    // Try price_info first
    if (product.price_info?.length > 0) {
      return product.price_info[0].original_price || 0;
    }

    // Fallback to first model price
    if (product.model_list?.length > 0) {
      return product.model_list[0].price_info?.[0]?.original_price || 0;
    }

    return 0;
  }

  static extractModelPrice(model: any): number {
    if (model.price_info?.length > 0) {
      return model.price_info[0].original_price || 0;
    }
    return 0;
  }
}

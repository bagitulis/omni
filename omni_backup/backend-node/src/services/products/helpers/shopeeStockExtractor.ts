/**
 * Extracts stock/quantity information from Shopee product data
 * Handles different API response formats (stock_info vs stock_info_v2)
 */
export class ShopeeStockExtractor {
  /**
   * Extract item-level stock
   */
  static extractItemStock(product: any): number {
    // Try stock_info_v2 first (newer format)
    if (product.stock_info_v2?.summary_info) {
      return product.stock_info_v2.summary_info.total_available_stock || 0;
    }

    // Fallback to stock_info (older format)
    if (product.stock_info?.length > 0) {
      return product.stock_info[0].normal_stock || 0;
    }

    return 0;
  }

  /**
   * Calculate total stock from models (variants)
   * Returns calculated stock or item stock as fallback
   */
  static calculateTotalStock(product: any, itemStock: number): number {
    if (!product.model_list?.length) {
      return itemStock;
    }

    let totalStock = 0;
    for (const model of product.model_list) {
      totalStock += this.extractModelStock(model);
    }

    // Use calculated stock if available, otherwise use item stock
    return totalStock > 0 || itemStock === 0 ? totalStock : itemStock;
  }

  /**
   * Extract model-level stock
   */
  static extractModelStock(model: any): number {
    // Try stock_info_v2 first
    if (model.stock_info_v2?.summary_info) {
      return model.stock_info_v2.summary_info.total_available_stock || 0;
    }

    // Fallback to stock_info
    if (model.stock_info?.length > 0) {
      return model.stock_info[0].normal_stock || 0;
    }

    return 0;
  }
}

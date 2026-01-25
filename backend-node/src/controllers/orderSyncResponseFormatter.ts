/**
 * Order Sync Response Formatter
 * Formats responses for order sync API endpoints
 * Single Responsibility: Response formatting and structure
 */

export class OrderSyncResponseFormatter {
  /**
   * Format sync by category response
   */
  static formatSyncResponse(
    category: string,
    days: number,
    results: Record<string, any>
  ) {
    return {
      success: true,
      category,
      days,
      results,
    };
  }

  /**
   * Format orders list response
   */
  static formatOrdersListResponse(category: string, items: any[]) {
    const uniqueOrders = new Set(items.map((item) => item.order_no)).size;
    return {
      success: true,
      category,
      count: uniqueOrders,
      items,
    };
  }

  /**
   * Format platform orders response
   */
  static formatPlatformOrdersResponse(
    platform: string,
    category: string,
    formattedItems: any[]
  ) {
    const uniqueOrders = new Set(formattedItems.map((item) => item.order_no))
      .size;
    return {
      success: true,
      platform,
      category,
      count: uniqueOrders,
      orders: formattedItems,
    };
  }

  /**
   * Format sync platform response
   */
  static formatSyncPlatformResponse(
    platform: string,
    category: string,
    days: number,
    orders: any[]
  ) {
    return {
      success: true,
      platform,
      category,
      days,
      count: orders.length,
      orders,
    };
  }

  /**
   * Format locked orders response
   */
  static formatLockedOrdersResponse(
    tenantId: string,
    days: number,
    lockedItems: any[],
    savedCount: number,
    totalQty: number
  ) {
    return {
      success: true,
      tenantId,
      days,
      count: lockedItems.length,
      saved: savedCount,
      total_qty: totalQty,
      items: lockedItems.map((item: any) => ({
        sku: item.sku,
        product_name: item.productName,
        variation_name: item.variationName || undefined,
        qty: item.qty,
        platforms: item.platforms || [],
      })),
    };
  }

  /**
   * Format saved locked orders (GET) response
   */
  static formatSavedLockedOrdersResponse(
    tenantId: string,
    days: number,
    lockedOrders: any[],
    totalQty: number
  ) {
    return {
      success: true,
      tenantId,
      days,
      count: lockedOrders.length,
      total_qty: totalQty,
      items: lockedOrders.map((order: any) => ({
        sku: order.sku,
        product_name: order.productName,
        variation_name: order.variationName || undefined,
        qty: order.qty,
        platforms: order.platforms || [],
      })),
    };
  }

  /**
   * Format order details response
   */
  static formatOrderDetailsResponse(platform: string, details: any[]) {
    return {
      success: true,
      platform,
      count: details.length,
      details,
    };
  }

  /**
   * Format error response
   */
  static formatErrorResponse(error: string) {
    return {
      success: false,
      error,
    };
  }
}

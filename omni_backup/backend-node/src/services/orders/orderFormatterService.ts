/**
 * Order Formatter Service
 * Transforms orders from database into format expected by frontend
 * Mirrors Python backend's get_formatted_orders() function
 */

export class OrderFormatterService {
  /**
   * Format orders for frontend - returns flattened items list
   * Mirrors Python: OrderDBHelper.get_formatted_orders(platform, category)
   */
  static formatOrdersForFrontend(platform: string, orders: any[]): any[] {
    try {
      const formatted: any[] = [];
      const platformLower = platform.toLowerCase();

      for (const order of orders) {
        const items = order.items || [];

        // For TikTok: deduplicate by product_id and count quantity
        if (platformLower === "tiktok") {
          const seenProducts: Record<string, any> = {};

          for (const item of items) {
            // TikTok uses: product_id, sku_id, seller_sku, product_name, sku_name
            const productId = item.productId || item.product_id || "";
            if (productId && !seenProducts[productId]) {
              // Count how many times this product appears across ALL items
              const qty = items.filter(
                (i: any) => (i.productId || i.product_id || "") === productId
              ).length;

              seenProducts[productId] = {
                platform: platform.toUpperCase(),
                order_no: order.orderSn || order.order_sn || "",
                sku: item.sellerSku || item.seller_sku || "-",
                product_name: item.productName || item.product_name || "-",
                variation_name: item.skuName || item.sku_name || "-",
                qty: qty,
              };
            }
          }

          formatted.push(...Object.values(seenProducts));
        } else {
          // Shopee/Lazada: iterate through items normally
          // NOTE: Shopee uses: itemName, modelName, itemSku, modelSku
          //       Lazada uses: productName, variationName, sellerSku
          for (const item of items) {
            formatted.push({
              platform: platform.toUpperCase(),
              order_no: order.orderSn || order.order_sn || "",
              // For SKU: try Shopee fields first, then Lazada fields
              sku:
                item.modelSku ||
                item.itemSku ||
                item.sellerSku ||
                item.seller_sku ||
                "-",
              // For product name: try Shopee fields first, then Lazada fields
              product_name:
                item.itemName || item.productName || item.product_name || "-",
              // For variation: try Shopee fields first, then Lazada fields
              variation_name:
                item.modelName ||
                item.variationName ||
                item.variation_name ||
                "-",
              qty: item.quantity || 0,
            });
          }
        }
      }

      return formatted;
    } catch (error) {
      console.error(`[OrderFormatter] Error formatting orders: ${error}`);
      return [];
    }
  }

  /**
   * Format orders from all platforms by category
   * Mirrors Python: OrderDBHelper.get_all_platforms_by_category(category)
   */
  static formatAllPlatformsForCategory(
    ordersMap: Record<string, any[]>
  ): any[] {
    try {
      const allItems: any[] = [];

      for (const [platform, orders] of Object.entries(ordersMap)) {
        const formattedItems = this.formatOrdersForFrontend(platform, orders);
        allItems.push(...formattedItems);
      }

      return allItems;
    } catch (error) {
      console.error(
        `[OrderFormatter] Error formatting all platforms: ${error}`
      );
      return [];
    }
  }
}

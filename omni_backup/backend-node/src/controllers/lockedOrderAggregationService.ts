/**
 * Locked Order Aggregation Service
 * Handles aggregation logic for locked orders (unprocess + processed)
 * Single Responsibility: Locked order data aggregation and transformation
 */

import { OrderFormatterService } from "../services/orders/orderFormatterService";

export class LockedOrderAggregationService {
  /**
   * Aggregate locked orders from unprocess and processed orders
   * Groups by SKU + Product Name + Variation and sums quantities
   */
  static aggregateLockedOrders(
    unprocessMap: Record<string, any[]>,
    processedMap: Record<string, any[]>
  ): any[] {
    // Format all orders using formatter service
    const unprocessFormatted =
      OrderFormatterService.formatAllPlatformsForCategory(unprocessMap);
    const processedFormatted =
      OrderFormatterService.formatAllPlatformsForCategory(processedMap);

    // Combine all formatted orders
    const allOrders = [...unprocessFormatted, ...processedFormatted];

    // Group by SKU + Product Name + Variation and aggregate qty + platforms
    const aggregated: Record<string, any> = {};

    for (const order of allOrders) {
      const sku = order.sku || "";
      const productName = order.product_name || "";
      const variationName = order.variation_name || "";
      const qty = order.qty || 1;
      const platform = order.platform || "unknown";

      const key = `${sku}|${productName}|${variationName}`;

      if (!aggregated[key]) {
        aggregated[key] = {
          sku,
          product_name: productName,
          variation_name: variationName,
          platforms: new Set(),
          total_qty: 0,
        };
      }

      aggregated[key].platforms.add(platform);
      aggregated[key].total_qty += qty;
    }

    // Convert to array and sort by qty descending
    return Object.values(aggregated)
      .map((item: any) => ({
        sku: item.sku,
        productName: item.product_name, // Will be converted to product_name in response formatter
        variationName: item.variation_name || undefined, // Will be converted to variation_name in response formatter
        qty: item.total_qty,
        platforms: Array.from(item.platforms), // Convert Set to array
      }))
      .sort((a: any, b: any) => b.qty - a.qty);
  }

  /**
   * Fetch orders by category from service and group by platform
   */
  static async fetchOrdersByCategory(
    orderSyncService: any,
    category: "unprocess" | "processed"
  ): Promise<Record<string, any[]>> {
    const map: Record<string, any[]> = {};

    const shopeeOrders = await orderSyncService.getOrdersByCategory(
      category,
      "shopee"
    );
    const lazadaOrders = await orderSyncService.getOrdersByCategory(
      category,
      "lazada"
    );
    const tiktokOrders = await orderSyncService.getOrdersByCategory(
      category,
      "tiktok"
    );

    map.shopee = shopeeOrders;
    map.lazada = lazadaOrders;
    map.tiktok = tiktokOrders;

    return map;
  }
}

/**
 * TikTok Order Manager
 * Handles TikTok-specific order operations
 * Mirrors Python backend's TiktokOrderManager
 */

import { BaseOrderManager } from "./baseOrderManager";
import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";

export class TiktokOrderManager extends BaseOrderManager {
  protected api: TiktokAPIClient;

  constructor(apiClient: TiktokAPIClient) {
    super(apiClient);
    this.api = apiClient;
  }

  /**
   * Get order list from TikTok by status
   * Mirrors Python: get_order_list()
   * TikTok uses POST method with body parameters
   */
  async getOrderList(status: string, days: number = 7): Promise<any[]> {
    try {
      this.logger.info(
        `📦 Fetching TikTok ${status} orders (last ${days} days)...`
      );

      const [timeFrom, timeTo] = this.getTimeRange(days);

      const allOrders: any[] = [];
      let nextPageToken = "";
      let pageCount = 0;

      // eslint-disable-next-line no-constant-condition
      while (true) {
        pageCount++;

        // TikTok requires POST with body parameters, not query params
        const requestBody: any = {
          create_time_ge: timeFrom, // Unix timestamp (seconds)
          create_time_lt: timeTo, // Unix timestamp (seconds)
          order_status: status, // Status filter
        };

        if (nextPageToken) {
          requestBody.next_page_token = nextPageToken;
        }

        // page_size goes in query params
        const queryParams = {
          page_size: "100",
        };

        try {
          const response = await this.api.request(
            "/order/202309/orders/search",
            "POST",
            queryParams,
            requestBody
          );

          if (!response || !response.data) {
            this.logger.debug(`No data in response at page ${pageCount}`);
            break;
          }

          const orders = response.data.orders || [];
          if (orders.length === 0) {
            this.logger.debug(`No orders returned at page ${pageCount}`);
            break;
          }

          // Transform TikTok orders: map 'id' to 'order_sn' if needed
          orders.forEach((order: any) => {
            if (!order.order_sn && order.id) {
              order.order_sn = order.id;
            }
          });

          allOrders.push(...orders);

          nextPageToken = response.data.next_page_token || "";
          const hasMore = !!nextPageToken && nextPageToken.length > 0;

          if (!hasMore) {
            this.logger.debug(
              `No more pages - ending pagination at page ${pageCount}`
            );
            break;
          }

          await this.sleep(this.RATE_LIMIT_DELAY);
        } catch (error) {
          this.logger.error(
            `Failed to fetch TikTok orders at page ${pageCount}: ${error}`
          );
          break;
        }
      }

      // TikTok might have duplicate orders, deduplicate by order_id
      const uniqueOrders = this.deduplicateOrders(allOrders);

      this.logger.info(
        `✅ Fetched ${uniqueOrders.length} unique TikTok ${status} orders`
      );

      return uniqueOrders;
    } catch (error) {
      this.logger.error(`Failed to fetch TikTok orders: ${error}`);
      throw error;
    }
  }

  /**
   * Get order details from TikTok (single order or batch)
   * Mirrors Python: get_order_details()
   *
   * TikTok API: GET /order/202309/orders with order IDs in query params
   * Returns: Complete order details including items, buyer info, shipping
   */
  async getOrderDetails(orderIds: string[]): Promise<any[]> {
    if (orderIds.length === 0) {
      return [];
    }

    try {
      this.logger.info(
        `✅ Processing order details for ${orderIds.length} orders`
      );

      const details: any[] = [];

      // TikTok can accept multiple IDs comma-separated in query params
      // Process in batches of 50 per API docs
      for (let i = 0; i < orderIds.length; i += this.BATCH_SIZE) {
        const batch = orderIds.slice(i, i + this.BATCH_SIZE);
        const idsString = batch.join(",");

        try {
          // Use GET /order/202309/orders with ids param
          const response = await this.api.request(
            "/order/202309/orders",
            "GET",
            {
              ids: idsString, // Comma-separated order IDs
            },
            undefined // No request body for GET
          );

          if (response && response.data && response.data.orders) {
            details.push(...response.data.orders);
            this.logger.debug(
              `✅ Retrieved ${response.data.orders.length} details from batch`
            );
          }

          if (i + this.BATCH_SIZE < orderIds.length) {
            await this.sleep(this.RATE_LIMIT_DELAY);
          }
        } catch (error) {
          this.logger.warn(
            `Failed to fetch details for batch starting at ${i}: ${error}`
          );
          // Continue fetching other batches instead of failing completely
        }
      }

      this.logger.info(
        `✅ Fetched complete details for ${details.length} TikTok orders`
      );

      return details;
    } catch (error) {
      this.logger.error(`Failed to fetch TikTok order details: ${error}`);
      throw error;
    }
  }

  /**
   * Deduplicate orders by order_sn (which is order.id)
   * TikTok API sometimes returns duplicates
   */
  private deduplicateOrders(orders: any[]): any[] {
    const seen = new Set<string>();
    const unique: any[] = [];

    for (const order of orders) {
      // Use order_sn or id (they should be the same after line 74 transformation)
      const orderId = order.order_sn || order.id || "";
      if (!seen.has(orderId)) {
        seen.add(orderId);
        unique.push(order);
      }
    }

    return unique;
  }

  /**
   * Format orders for export
   * Mirrors Python: format_items_for_export()
   */
  formatOrdersForExport(orders: any[]): Record<string, any>[] {
    const formatted: Record<string, any>[] = [];
    const processedItems = new Set<string>();

    for (const order of orders) {
      // Use order_sn or id for order identifier
      const orderId = order.order_sn || order.id || "";
      const status = order.status || "UNKNOWN";

      for (const item of order.line_items || []) {
        const skuId = item.sku_id || item.item_id || "";
        const sku = item.sku || "";
        const itemKey = `${orderId}_${skuId}_${sku}`;

        if (processedItems.has(itemKey)) {
          continue;
        }
        processedItems.add(itemKey);

        const quantity = item.quantity || 0;
        const price = item.price || item.original_price || "";

        formatted.push({
          order_id: orderId,
          sku,
          item_name: item.product_name || item.item_name || "",
          quantity,
          price,
          status,
        });
      }
    }

    return formatted;
  }

  /**
   * Get processed orders with tracking info for "Order Today" feature
   * TikTok already includes tracking info in the order search response
   */
  async getProcessedOrdersWithTracking(days: number = 7): Promise<any[]> {
    try {
      this.logger.info(
        `📦 Fetching TikTok AWAITING_COLLECTION with tracking...`
      );

      const orders = await this.getOrderList("AWAITING_COLLECTION", days);
      if (orders.length === 0) {
        return [];
      }

      const results: any[] = [];
      for (const order of orders) {
        const orderId = order.order_sn || order.id || "";
        const trackingNo = order.tracking_number || "";
        // Fallback chain: order-level shipping_provider_name -> shipping_provider -> first line_item's shipping_provider_name
        const lineItems = order.line_items || [];
        const carrier =
          order.shipping_provider_name ||
          order.shipping_provider ||
          (lineItems.length > 0 ? lineItems[0].shipping_provider_name : "") ||
          "";

        for (const item of lineItems) {
          // Per-item carrier fallback: item's shipping_provider_name -> order-level carrier
          const itemCarrier = item.shipping_provider_name || carrier;
          results.push({
            order_sn: orderId,
            tracking_number: item.tracking_number || trackingNo,
            shipping_carrier: itemCarrier,
            seller_sku: item.seller_sku || item.sku || "",
            product_name: item.product_name || "",
            variation_name: item.sku_name || "",
            quantity: item.quantity || 1,
          });
        }
      }

      this.logger.info(`✅ Fetched ${results.length} items with tracking`);
      return results;
    } catch (error) {
      this.logger.error(`Failed to fetch processed orders: ${error}`);
      throw error;
    }
  }
}

/**
 * Lazada Order Manager
 * Handles Lazada-specific order operations
 * Mirrors Python backend's LazadaOrderManager
 */

import { BaseOrderManager } from "./baseOrderManager";
import { LazadaAPIClient } from "../../api/clients/lazadaAPIClient";

export class LazadaOrderManager extends BaseOrderManager {
  protected api: LazadaAPIClient;

  constructor(apiClient: LazadaAPIClient) {
    super(apiClient);
    this.api = apiClient;
  }

  /**
   * Get order list from Lazada by status
   * Mirrors Python: get_order_list()
   * Note: Lazada API returns order_id, but we map it to order_sn for consistency
   */
  async getOrderList(status: string, days: number = 7): Promise<any[]> {
    try {
      this.logger.info(
        `📦 Fetching Lazada ${status} orders (last ${days} days)...`
      );

      // Get time range (not used in current implementation, but keeping for future)
      this.getTimeRange(days);

      // Lazada expects ISO 8601 format with timezone for created_after/created_before
      const endDate = new Date();
      const startDate = new Date(
        endDate.getTime() - days * 24 * 60 * 60 * 1000
      );

      // Format: YYYY-MM-DDTHH:MM:SS+07:00
      const createdAfter = startDate.toISOString().replace("Z", "+07:00");
      const createdBefore = endDate.toISOString().replace("Z", "+07:00");

      // Lazada endpoint is /orders/get with GET method
      const baseParams = {
        created_after: createdAfter,
        created_before: createdBefore,
        status: status,
        offset: 0,
        limit: 100,
      };

      const allOrders: any[] = [];
      let offset = 0;
      let hasMore = true;

      while (hasMore) {
        baseParams.offset = offset;

        try {
          const response = await this.api.request(
            "/orders/get",
            "GET",
            baseParams
          );

          if (!response || !response.data) {
            this.logger.debug(`No data in response at offset ${offset}`);
            break;
          }

          const orders = response.data.orders || [];
          if (orders.length === 0) {
            this.logger.debug(`No orders returned at offset ${offset}`);
            break;
          }

          // Map order_id → order_sn for consistency with database schema
          const mappedOrders = orders.map((order: any) => ({
            ...order,
            order_sn: String(order.order_id), // Map Lazada order_id to order_sn (convert to string for DB)
            order_status: status, // Add status for database save
            shop_id: "0", // Lazada doesn't return shop_id in list
            tracking_number: "", // Will be populated from /order/items/get
            shipping_carrier: "", // Will be populated from /order/items/get
          }));

          allOrders.push(...mappedOrders);
          offset += 100;

          // Check if there are more results
          if (orders.length < 100) {
            hasMore = false;
            this.logger.debug(`Reached end of results at offset ${offset}`);
          }

          await this.sleep(this.RATE_LIMIT_DELAY);
        } catch (error) {
          this.logger.error(
            `Failed to fetch Lazada orders at offset ${offset}: ${error}`
          );
          break;
        }
      }

      this.logger.info(
        `✅ Fetched ${allOrders.length} Lazada ${status} orders`
      );

      return allOrders;
    } catch (error) {
      this.logger.error(`Failed to fetch Lazada orders: ${error}`);
      throw error;
    }
  }

  /**
   * Get order items from Lazada using /order/items/get endpoint
   * This fetches detailed item information (seller_sku, product_name, etc)
   * Returns items with field mapping for database save
   */
  async getOrderItems(orderIds: string[]): Promise<any[]> {
    if (orderIds.length === 0) {
      return [];
    }

    try {
      this.logger.info(
        `✅ Processing order items for ${orderIds.length} orders`
      );

      const ordersWithItems: any[] = [];

      // Lazada's /order/items/get endpoint requires individual order IDs
      for (const orderId of orderIds) {
        try {
          const response = await this.api.request("/order/items/get", "GET", {
            order_id: orderId,
          });

          if (response && response.data && Array.isArray(response.data)) {
            // Response.data is array of items directly (NOT response.data.items)
            const mappedItems = response.data.map((item: any) => ({
              item_id: item.order_item_id || item.item_id || 0,
              sku_id: item.sku_id || "",
              seller_sku: item.sku || "", // Lazada uses 'sku' field
              product_name: item.name || "", // Lazada uses 'name' field
              variation_name: item.variation || "", // Lazada uses 'variation' field
              quantity: 1, // Not in Lazada item response - default to 1
              price: item.paid_price || item.item_price || 0,
            }));

            ordersWithItems.push({
              order_sn: orderId,
              order_id: orderId,
              item_list: mappedItems,
            });

            this.logger.debug(
              `✅ Fetched ${mappedItems.length} items for order ${orderId}`
            );
          } else {
            // Empty items for this order
            ordersWithItems.push({
              order_sn: orderId,
              order_id: orderId,
              item_list: [],
            });
          }

          await this.sleep(this.RATE_LIMIT_DELAY);
        } catch (error) {
          this.logger.warn(
            `Failed to fetch items for order ${orderId}: ${error}`
          );
          // Continue fetching other orders - still return order without items
          ordersWithItems.push({
            order_sn: orderId,
            order_id: orderId,
            item_list: [],
          });
        }
      }

      this.logger.info(
        `✅ Fetched items for ${ordersWithItems.length} Lazada orders`
      );

      return ordersWithItems;
    } catch (error) {
      this.logger.error(`Failed to fetch Lazada order items: ${error}`);
      throw error;
    }
  }

  /**
   * Format orders for export
   * Mirrors Python: format_items_for_export()
   */
  formatOrdersForExport(orders: any[]): Record<string, any>[] {
    const formatted: Record<string, any>[] = [];
    const processedItems = new Set<string>();

    for (const order of orders) {
      const orderId = order.order_id || order.order_number || "";
      const status = order.status || "UNKNOWN";

      for (const item of order.order_line_items || []) {
        const itemId = item.item_id || "";
        const sku = item.sku || "";
        const itemKey = `${orderId}_${itemId}_${sku}`;

        if (processedItems.has(itemKey)) {
          continue;
        }
        processedItems.add(itemKey);

        const quantity = item.quantity || 0;
        const price = item.price || item.purchase_price || "";

        formatted.push({
          order_id: orderId,
          sku,
          item_name: item.item_name || item.name || "",
          quantity,
          price,
          status,
        });
      }
    }

    return formatted;
  }

  /**
   * Get order details for specific order IDs
   * Required abstract method implementation
   */
  async getOrderDetails(_orderIds: string[]): Promise<any[]> {
    // Lazada doesn't have a separate order details endpoint
    // Details are already included in the order list
    return [];
  }

  /**
   * Get processed orders with tracking info for "Order Today" feature
   * Lazada returns tracking info from /order/items/get endpoint
   */
  async getProcessedOrdersWithTracking(days: number = 7): Promise<any[]> {
    try {
      this.logger.info(`📦 Fetching Lazada toship orders with tracking...`);

      // Get orders with toship status (processed)
      const orders = await this.getOrderList("toship", days);
      if (orders.length === 0) {
        return [];
      }

      const orderIds = orders.map((o) => o.order_sn);
      const results: any[] = [];

      // Fetch items with tracking info for each order
      for (const orderId of orderIds) {
        try {
          const response = await this.api.request("/order/items/get", "GET", {
            order_id: orderId,
          });

          if (response?.data && Array.isArray(response.data)) {
            const firstItem = response.data[0] || {};
            const trackingCode = firstItem.tracking_code || "";
            // Extract clean carrier name from shipment_provider
            // Lazada format: "Pickup: LEX ID, Delivery: LEX ID" -> extract just "LEX ID"
            let shippingProvider = firstItem.shipment_provider || "";
            if (shippingProvider.includes(":")) {
              // Try to extract from "Pickup: XXX" or "Delivery: XXX" format
              const match = shippingProvider.match(
                /(?:Pickup|Delivery):\s*([^,]+)/i
              );
              if (match && match[1]) {
                shippingProvider = match[1].trim();
              }
            }

            for (const item of response.data) {
              // Per-item tracking and carrier (some orders have different carriers per item)
              const itemTracking = item.tracking_code || trackingCode;
              let itemCarrier = item.shipment_provider || shippingProvider;
              // Clean up carrier name if it has Pickup/Delivery prefix
              if (itemCarrier.includes(":")) {
                const match = itemCarrier.match(
                  /(?:Pickup|Delivery):\s*([^,]+)/i
                );
                if (match && match[1]) {
                  itemCarrier = match[1].trim();
                }
              }
              results.push({
                order_sn: orderId,
                tracking_number: itemTracking,
                shipping_carrier: itemCarrier,
                seller_sku: item.sku || "",
                product_name: item.name || "",
                variation_name: item.variation || "",
                quantity: 1,
              });
            }
          }
          await this.sleep(this.RATE_LIMIT_DELAY);
        } catch (err) {
          this.logger.warn(`Failed to fetch items for order ${orderId}`);
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

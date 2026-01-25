/**
 * Shopee Order Manager
 * Handles Shopee-specific order operations
 * Mirrors Python backend's ShopeeOrderManager
 */

import { BaseOrderManager } from "./baseOrderManager";
import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { ShopeeWalletService } from "./shopeeWalletService";
import { ShopeeShippingFeeService } from "./shopeeShippingFeeService";

export class ShopeeOrderManager extends BaseOrderManager {
  protected api: ShopeeAPIClient;
  public wallet: ShopeeWalletService;
  public shippingFee: ShopeeShippingFeeService;

  constructor(apiClient: ShopeeAPIClient) {
    super(apiClient);
    this.api = apiClient;
    this.wallet = new ShopeeWalletService(apiClient);
    this.shippingFee = new ShopeeShippingFeeService(apiClient);
    this.shippingFee.setOrderManager(this);
  }

  /**
   * Get order list from Shopee by status
   * Mirrors Python: get_order_list()
   */
  async getOrderList(status: string, days: number = 7): Promise<any[]> {
    try {
      this.logger.info(
        `📦 Fetching Shopee ${status} orders (last ${days} days)...`
      );

      const [timeFrom, timeTo] = this.getTimeRange(days);

      const baseParams = {
        time_range_field: "create_time",
        time_from: timeFrom,
        time_to: timeTo,
        page_size: 100,
        order_status: status,
        response_optional_fields: "order_status",
      };

      // Fetch paginated order list
      const orders = await this.fetchPaginatedData(
        "/api/v2/order/get_order_list",
        baseParams,
        "order_list"
      );

      this.logger.info(`✅ Shopee ${status} orders fetched`);

      return orders;
    } catch (error) {
      this.logger.error(`Failed to fetch Shopee orders: ${error}`);
      throw error;
    }
  }

  /**
   * Get order details from Shopee with items
   * Mirrors Python: get_order_details()
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

      // Batch requests to avoid rate limiting
      for (let i = 0; i < orderIds.length; i += this.BATCH_SIZE) {
        const batch = orderIds.slice(i, i + this.BATCH_SIZE);

        const response = await this.api.request(
          "/api/v2/order/get_order_detail",
          "GET",
          {
            order_sn_list: batch.join(","),
            response_optional_fields: "order_status,tracking_number,item_list",
          }
        );

        if (response && response.response && response.response.order_list) {
          details.push(...response.response.order_list);
        }

        if (i + this.BATCH_SIZE < orderIds.length) {
          await this.sleep(this.RATE_LIMIT_DELAY);
        }
      }

      this.logger.info(
        `✅ Fetched details for ${details.length} Shopee orders with items`
      );

      return details;
    } catch (error) {
      this.logger.error(`Failed to fetch Shopee order details: ${error}`);
      throw error;
    }
  }

  /**
   * Get order items from Shopee by order IDs
   * Fetch item_list for orders
   */
  async getOrderItems(orderIds: string[]): Promise<any[]> {
    if (orderIds.length === 0) {
      return [];
    }

    try {
      this.logger.info(
        `📋 Fetching items for ${orderIds.length} Shopee orders...`
      );

      const allOrders: any[] = [];

      // Batch requests to avoid rate limiting
      for (let i = 0; i < orderIds.length; i += this.BATCH_SIZE) {
        const batch = orderIds.slice(i, i + this.BATCH_SIZE);

        const response = await this.api.request(
          "/api/v2/order/get_order_detail",
          "GET",
          {
            order_sn_list: batch.join(","),
            response_optional_fields: "item_list",
          }
        );

        if (response && response.response && response.response.order_list) {
          allOrders.push(...response.response.order_list);
        }

        if (i + this.BATCH_SIZE < orderIds.length) {
          await this.sleep(this.RATE_LIMIT_DELAY);
        }
      }

      this.logger.info(
        `✅ Fetched items for ${allOrders.length} Shopee orders`
      );

      return allOrders;
    } catch (error) {
      this.logger.error(`Failed to fetch Shopee order items: ${error}`);
      throw error;
    }
  }

  /**
   * Format orders for export
   * Mirrors Python: format_items_for_export()
   */
  formatOrdersForExport(orders: any[]): Record<string, any>[] {
    const formatted: Record<string, any>[] = [];

    for (const order of orders) {
      const orderId = order.order_sn || "";
      const status = order.order_status || "UNKNOWN";

      for (const item of order.item_list || []) {
        const sku = item.model_sku || item.item_sku || "";
        const modelName = item.model_name || "";
        const quantity = item.model_quantity_purchased || 0;
        const price = item.model_original_price || "";

        formatted.push({
          order_sn: orderId,
          sku,
          item_name: item.item_name || "",
          model_name: modelName,
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
   * Flow: searchPackageList(3) → getPackageDetail → getOrderDetails → Merge
   * Returns flattened items with all required fields
   */
  async getProcessedOrdersWithTracking(_days: number = 7): Promise<any[]> {
    try {
      this.logger.info(`📦 Fetching Shopee PROCESSED orders with tracking...`);

      // Step 1: Get package list with package_status=3 (Processed)
      let allPackages: Array<{
        order_sn: string;
        package_number: string;
      }> = [];
      let cursor = "";
      let hasMore = true;

      while (hasMore) {
        const result = await this.api.searchPackageList(3, cursor, 100);
        allPackages.push(...result.packages);
        cursor = result.nextCursor;
        hasMore = result.more;
        await this.sleep(this.RATE_LIMIT_DELAY);
      }

      if (allPackages.length === 0) {
        this.logger.info("No processed packages found");
        return [];
      }

      this.logger.info(`Found ${allPackages.length} processed packages`);

      // Step 2: Get package details (tracking, carrier, SKU, qty)
      interface PackageItem {
        item_id: number;
        model_id: number;
        item_sku: string;
        model_sku: string;
        model_quantity: number;
      }
      interface PackageInfo {
        tracking: string;
        carrier: string;
        packageNumber: string;
        items: PackageItem[];
      }
      const packageInfoMap = new Map<string, PackageInfo>();
      const packageNumbers = allPackages
        .map((p) => p.package_number)
        .filter((p) => p);

      for (let i = 0; i < packageNumbers.length; i += 50) {
        const batch = packageNumbers.slice(i, i + 50);
        const details = await this.api.getPackageDetail(batch);

        for (const pkg of details) {
          const trackingNo =
            pkg.tracking_number && pkg.tracking_number !== "-"
              ? pkg.tracking_number
              : pkg.package_number;
          packageInfoMap.set(pkg.order_sn, {
            tracking: trackingNo,
            carrier: pkg.shipping_carrier,
            packageNumber: pkg.package_number,
            items: pkg.items,
          });
        }
        await this.sleep(this.RATE_LIMIT_DELAY);
      }

      // Step 3: Get order details for item_name and model_name
      const orderIds = allPackages.map((p) => p.order_sn);
      const uniqueOrderIds = [...new Set(orderIds)];
      const orderDetails = await this.getOrderDetails(uniqueOrderIds);

      // Build item name lookup: key = "item_id|model_id"
      const itemNameMap = new Map<
        string,
        { item_name: string; model_name: string }
      >();
      for (const order of orderDetails) {
        for (const item of order.item_list || []) {
          const key = `${item.item_id}|${item.model_id || 0}`;
          itemNameMap.set(key, {
            item_name: item.item_name || "",
            model_name: item.model_name || "",
          });
        }
      }

      // Step 4: Build final result - merge all data per item
      const result: any[] = [];
      for (const order of orderDetails) {
        const pkgInfo = packageInfoMap.get(order.order_sn);
        if (!pkgInfo) continue;

        // Use items from package detail (has SKU & qty)
        for (const pkgItem of pkgInfo.items) {
          const nameKey = `${pkgItem.item_id}|${pkgItem.model_id || 0}`;
          const nameInfo = itemNameMap.get(nameKey) || {
            item_name: "",
            model_name: "",
          };

          result.push({
            order_sn: order.order_sn,
            tracking_number: pkgInfo.tracking,
            shipping_carrier: pkgInfo.carrier,
            item_list: [
              {
                item_id: pkgItem.item_id,
                model_id: pkgItem.model_id,
                item_sku: pkgItem.item_sku,
                model_sku: pkgItem.model_sku,
                model_quantity_purchased: pkgItem.model_quantity,
                item_name: nameInfo.item_name,
                model_name: nameInfo.model_name,
              },
            ],
          });
        }
      }

      this.logger.info(
        `✅ Fetched ${result.length} items from ${orderDetails.length} orders`
      );
      return result;
    } catch (error) {
      this.logger.error(
        `Failed to fetch processed orders with tracking: ${error}`
      );
      throw error;
    }
  }
}

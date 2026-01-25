/**
 * Order Sync Operations
 * SRP: Sync orders from API to database
 *
 * MULTI-TENANT: Receives tenantId from parent service
 */

import { OrderRepository } from "./orderRepository";
import { getPlatformCoordinationService } from "../platformCoordinationService";
import { getLogger } from "../../utils/logger";
import {
  PlatformType,
  OrderStatusCategory,
  ORDER_STATUS_MAPPINGS,
} from "./orderStatusMappings";

const logger = getLogger("OrderSyncOperations");

export class OrderSyncOperations {
  private orderManagers: Map<PlatformType, any>;
  private orderRepository: OrderRepository;
  private tenantId: string;

  /**
   * @param orderManagers - Map of platform order managers
   * @param orderRepository - Order repository instance
   * @param tenantId - Tenant ID for accessing platform coordination service
   */
  constructor(
    orderManagers: Map<PlatformType, any>,
    orderRepository: OrderRepository,
    tenantId: string
  ) {
    this.orderManagers = orderManagers;
    this.orderRepository = orderRepository;
    this.tenantId = tenantId;
  }

  private getOrderManager(platform: PlatformType): any {
    const manager = this.orderManagers.get(platform);
    if (!manager)
      throw new Error(`Order manager not found for platform: ${platform}`);
    return manager;
  }

  /**
   * Sync orders by category for all platforms
   */
  async syncByCategory(
    category: OrderStatusCategory,
    days: number = 7,
    platforms: PlatformType[] = ["shopee", "lazada", "tiktok"]
  ): Promise<Record<string, any>> {
    try {
      logger.info(
        `🔄 Syncing ${category} orders for ${platforms.join(
          ", "
        )} (last ${days} days) [tenant: ${this.tenantId}]`
      );

      // Get per-tenant platform coordination service
      const platformCoordinationService = getPlatformCoordinationService(
        this.tenantId
      );

      // Reload all configs from database
      logger.info("🔄 Reloading configs from database...");
      for (const platform of ["shopee", "lazada", "tiktok"] as PlatformType[]) {
        await platformCoordinationService.reloadPlatformConfig(platform);
      }

      const results: Record<string, any> = {};

      for (const platform of platforms) {
        try {
          const status = ORDER_STATUS_MAPPINGS[platform][category];
          const manager = this.getOrderManager(platform);

          logger.info(
            `  📦 Fetching ${platform} orders (status: ${status})...`
          );
          const orders: any[] = await manager.getOrderList(status, days);

          // Fetch order items if available
          if (orders.length > 0) {
            await this.fetchOrderItems(platform, manager, orders);
          }

          await this.saveOrdersToDatabase(platform, orders, category);
          results[platform] = { success: true, count: orders.length, orders };
          logger.info(`  ✅ ${platform}: Synced ${orders.length} orders`);
        } catch (error) {
          logger.error(`  ❌ ${platform}: Error syncing orders - ${error}`);
          results[platform] = {
            success: false,
            error: String(error),
            count: 0,
          };
        }
      }

      return results;
    } catch (error) {
      logger.error(`Failed to sync orders: ${error}`);
      throw error;
    }
  }

  /**
   * Sync orders for a specific platform
   */
  async syncPlatformOrders(
    platform: PlatformType,
    category: OrderStatusCategory,
    days: number = 7
  ): Promise<any[]> {
    try {
      logger.info(
        `🔄 Syncing ${platform} ${category} orders (last ${days} days)`
      );

      const status = ORDER_STATUS_MAPPINGS[platform][category];
      const manager = this.getOrderManager(platform);
      const orders: any[] = await manager.getOrderList(status, days);

      if (orders.length > 0) {
        await this.fetchOrderItems(platform, manager, orders);
      }

      await this.saveOrdersToDatabase(platform, orders, category);
      logger.info(`✅ Synced ${orders.length} orders from ${platform}`);

      return orders;
    } catch (error) {
      logger.error(`Failed to sync ${platform} orders: ${error}`);
      throw error;
    }
  }

  private async fetchOrderItems(
    platform: PlatformType,
    manager: any,
    orders: any[]
  ): Promise<void> {
    const orderIds = orders.map((o: any) => o.order_sn || o.order_id);

    if (
      (platform === "shopee" || platform === "lazada") &&
      "getOrderItems" in manager
    ) {
      logger.info(
        `📋 Fetching item details for ${orderIds.length} ${platform} orders...`
      );
      const ordersWithItems = await manager.getOrderItems(orderIds);

      let matchedCount = 0;
      let itemsTotalCount = 0;

      // Build a Map for O(1) lookup - fixes findIndex bug with Shopee orders
      const orderMap = new Map<string, any>();
      for (const order of orders) {
        const key = order.order_sn || order.order_id;
        if (key) {
          orderMap.set(String(key), order);
        }
      }

      for (const orderWithItems of ordersWithItems) {
        const key = String(orderWithItems.order_sn || orderWithItems.order_id);
        const matchedOrder = orderMap.get(key);

        if (matchedOrder) {
          matchedOrder.item_list = orderWithItems.item_list || [];
          matchedCount++;
          itemsTotalCount += (orderWithItems.item_list || []).length;
        }
      }

      logger.info(
        `📋 Matched ${matchedCount}/${ordersWithItems.length} orders with items. Total items: ${itemsTotalCount}`
      );
    }
  }

  private async saveOrdersToDatabase(
    platform: PlatformType,
    orders: any[],
    category: OrderStatusCategory
  ): Promise<void> {
    try {
      const status = ORDER_STATUS_MAPPINGS[platform][category];
      await this.orderRepository.clearOrdersByPlatformAndStatus(
        platform,
        status
      );

      if (orders.length === 0) {
        logger.info(
          `💾 Cleared stale ${platform} ${category} orders - no new orders`
        );
        return;
      }

      if (platform === "shopee")
        await this.orderRepository.saveShopeeOrders(orders);
      else if (platform === "lazada")
        await this.orderRepository.saveLazadaOrders(orders);
      else if (platform === "tiktok")
        await this.orderRepository.saveTiktokOrders(orders);

      logger.info(
        `💾 Saved ${orders.length} ${platform} ${category} orders to database`
      );
    } catch (error) {
      logger.error(`Failed to save orders to database: ${error}`);
    }
  }
}

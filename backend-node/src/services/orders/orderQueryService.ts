/**
 * Order Query Service
 * SRP: Read-only operations for fetching orders from database
 */

import { OrderRepository } from "./orderRepository";
import { getLogger } from "../../utils/logger";
import {
  PlatformType,
  OrderStatusCategory,
  ORDER_STATUS_MAPPINGS,
} from "./orderStatusMappings";

const logger = getLogger("OrderQueryService");

export class OrderQueryService {
  private orderRepository: OrderRepository;

  constructor(orderRepository: OrderRepository) {
    this.orderRepository = orderRepository;
  }

  /**
   * Get orders by category from database
   */
  async getOrdersByCategory(
    category: OrderStatusCategory,
    platform?: PlatformType
  ): Promise<any[]> {
    try {
      const platforms = platform
        ? [platform]
        : (["shopee", "lazada", "tiktok"] as PlatformType[]);
      const allOrders: any[] = [];

      for (const plat of platforms) {
        const status = ORDER_STATUS_MAPPINGS[plat][category];

        if (plat === "shopee") {
          const orders =
            await this.orderRepository.getShopeeOrdersByStatus(status);
          allOrders.push(
            ...orders.map((o: any) => ({ ...o, _platform: "shopee" }))
          );
        } else if (plat === "lazada") {
          const orders =
            await this.orderRepository.getLazadaOrdersByStatus(status);
          allOrders.push(
            ...orders.map((o: any) => ({ ...o, _platform: "lazada" }))
          );
        } else if (plat === "tiktok") {
          const orders =
            await this.orderRepository.getTiktokOrdersByStatus(status);
          allOrders.push(
            ...orders.map((o: any) => ({ ...o, _platform: "tiktok" }))
          );
        }
      }

      return allOrders;
    } catch (error) {
      logger.error(`Failed to get orders from database: ${error}`);
      return [];
    }
  }

  /**
   * Get latest synced orders for a platform
   */
  async getPlatformOrders(
    platform: PlatformType,
    category?: OrderStatusCategory
  ): Promise<any[]> {
    try {
      if (category) {
        return await this.getOrdersByCategory(category, platform);
      }

      // Get all orders for platform (all statuses)
      if (platform === "shopee") {
        const unpaid =
          await this.orderRepository.getShopeeOrdersByStatus("UNPAID");
        const processing =
          await this.orderRepository.getShopeeOrdersByStatus("READY_TO_SHIP");
        const processed =
          await this.orderRepository.getShopeeOrdersByStatus("PROCESSED");
        return [...unpaid, ...processing, ...processed];
      } else if (platform === "lazada") {
        const unpaid =
          await this.orderRepository.getLazadaOrdersByStatus("unpaid");
        const processing =
          await this.orderRepository.getLazadaOrdersByStatus("topack");
        const processed =
          await this.orderRepository.getLazadaOrdersByStatus("toship");
        return [...unpaid, ...processing, ...processed];
      } else if (platform === "tiktok") {
        const unpaid =
          await this.orderRepository.getTiktokOrdersByStatus("UNPAID");
        const processing =
          await this.orderRepository.getTiktokOrdersByStatus(
            "AWAITING_SHIPMENT"
          );
        const processed = await this.orderRepository.getTiktokOrdersByStatus(
          "AWAITING_COLLECTION"
        );
        return [...unpaid, ...processing, ...processed];
      }

      return [];
    } catch (error) {
      logger.error(`Failed to get ${platform} orders: ${error}`);
      return [];
    }
  }
}

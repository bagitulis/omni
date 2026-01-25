import { OrderData } from "../../types/order.types";
import { ShopeeOrderRepository } from "./shopeeOrderRepository";
import { LazadaOrderRepository } from "./lazadaOrderRepository";
import { TiktokOrderRepository } from "./tiktokOrderRepository";

/**
 * Order Repository
 * Facade for platform-specific order repositories
 * Single Responsibility: Coordinate platform-specific order operations
 */
export class OrderRepository {
  private shopeeRepo = new ShopeeOrderRepository();
  private lazadaRepo = new LazadaOrderRepository();
  private tiktokRepo = new TiktokOrderRepository();

  private logger = {
    info: (msg: string) => console.log(`[OrderRepository] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[OrderRepository] ❌ ${msg}`),
  };

  /**
   * Save Shopee orders to database with items
   */
  async saveShopeeOrders(orders: OrderData[]): Promise<void> {
    return this.shopeeRepo.saveOrders(orders);
  }

  /**
   * Save Lazada orders to database with items
   */
  async saveLazadaOrders(orders: OrderData[]): Promise<void> {
    return this.lazadaRepo.saveOrders(orders);
  }

  /**
   * Save TikTok orders to database with items
   */
  async saveTiktokOrders(orders: OrderData[]): Promise<void> {
    return this.tiktokRepo.saveOrders(orders);
  }

  /**
   * Get Shopee orders by status
   * Note: limit is optional - omit to fetch all records
   */
  async getShopeeOrdersByStatus(
    status: string,
    limit?: number
  ): Promise<any[]> {
    return this.shopeeRepo.getOrdersByStatus(status, limit);
  }

  /**
   * Get Lazada orders by status
   * Note: limit is optional - omit to fetch all records
   */
  async getLazadaOrdersByStatus(
    status: string,
    limit?: number
  ): Promise<any[]> {
    return this.lazadaRepo.getOrdersByStatus(status, limit);
  }

  /**
   * Get TikTok orders by status
   * Note: limit is optional - omit to fetch all records
   */
  async getTiktokOrdersByStatus(
    status: string,
    limit?: number
  ): Promise<any[]> {
    return this.tiktokRepo.getOrdersByStatus(status, limit);
  }

  /**
   * Clear orders by platform and status
   */
  async clearOrdersByPlatformAndStatus(
    platform: string,
    status: string
  ): Promise<void> {
    try {
      if (platform === "shopee") {
        await this.shopeeRepo.clearOrdersByStatus(status);
      } else if (platform === "lazada") {
        await this.lazadaRepo.clearOrdersByStatus(status);
      } else if (platform === "tiktok") {
        await this.tiktokRepo.clearOrdersByStatus(status);
      }
      this.logger.info(`✅ Orders cleared for ${platform}`);
    } catch (error: any) {
      this.logger.error(`Failed to clear orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get all orders count by platform and status
   */
  async getOrdersCount(platform: string, status?: string): Promise<number> {
    try {
      if (platform === "shopee") {
        return await this.shopeeRepo.getOrdersCount(status);
      } else if (platform === "lazada") {
        return await this.lazadaRepo.getOrdersCount(status);
      } else if (platform === "tiktok") {
        return await this.tiktokRepo.getOrdersCount(status);
      }
      return 0;
    } catch (error: any) {
      this.logger.error(`Failed to get count: ${error.message}`);
      throw error;
    }
  }
}



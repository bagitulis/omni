/**
 * Shopee Order Repository
 * Handles Shopee-specific order persistence operations
 * Single Responsibility: Shopee order database operations
 * NOTE: Webhook data is stored separately in WebhookOrderEvent table
 */

import { getPrisma } from "../prismaClient";
import { tenantContext } from "../../utils/tenantContext";
import { OrderData } from "../../types/order.types";
import { OrderDataConverter } from "./orderDataConverter";

export class ShopeeOrderRepository {
  private logger = {
    info: (msg: string) => console.log(`[ShopeeOrderRepo] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[ShopeeOrderRepo] ❌ ${msg}`),
  };

  private getTenantId(): string {
    return tenantContext.getTenantId();
  }

  /**
   * Save Shopee orders to database with items
   * Uses upsert for atomic operation - prevents race conditions
   */
  async saveOrders(orders: OrderData[]): Promise<void> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      if (orders.length === 0) return;

      // Debug counters
      let savedItemsCount = 0;
      let ordersWithItemsCount = 0;

      // Get distinct statuses from incoming orders for cleaning old data
      const statuses = [...new Set(orders.map((o) => o.order_status))];
      const orderSns = new Set(orders.map((o) => o.order_sn));

      // Delete orders with same status that are NOT in current batch (old orders)
      for (const status of statuses) {
        await prisma.shopeeOrder.deleteMany({
          where: {
            tenantId,
            orderStatus: status,
            orderSn: { notIn: Array.from(orderSns) },
          } as any,
        });
      }

      for (const order of orders) {
        // Upsert - atomic operation (no race condition)
        await prisma.shopeeOrder.upsert({
          where: { orderSn: order.order_sn },
          create: {
            tenantId,
            orderSn: order.order_sn,
            shopId: order.shop_id,
            orderStatus: order.order_status,
            orderTimestamp: order.order_timestamp,
          } as any,
          update: {
            tenantId,
            orderStatus: order.order_status,
            updatedAt: new Date(),
          },
        });

        // Save order items - delete existing and recreate
        const items = (order as any).item_list || [];
        if (items.length > 0) {
          await prisma.shopeeOrderItem.deleteMany({
            where: { orderSn: order.order_sn },
          });

          for (const item of items) {
            await prisma.shopeeOrderItem.create({
              data: {
                orderSn: order.order_sn,
                itemId: item.item_id || 0,
                modelId: item.model_id,
                itemName: item.item_name || "",
                modelName: item.model_name || "",
                itemSku: item.item_sku || "",
                modelSku: item.model_sku || "",
                quantity: item.model_quantity_purchased || 1,
                price: item.original_price || item.current_price || 0,
              },
            });
          }
          savedItemsCount += items.length;
          ordersWithItemsCount++;
        }
      }

      this.logger.info(
        `✅ Shopee orders saved (${ordersWithItemsCount} orders with ${savedItemsCount} items)`
      );
    } catch (error: any) {
      this.logger.error(`Failed to save Shopee orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get Shopee orders by status
   * Filters by current tenant from context
   * Note: limit is now optional - omit to fetch all records
   */
  async getOrdersByStatus(status: string, limit?: number): Promise<any[]> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();
      const queryOptions: any = {
        where: {
          tenantId,
          orderStatus: status,
        } as any,
        orderBy: { createdAt: "desc" },
        include: { items: true },
      };

      if (limit) {
        queryOptions.take = limit;
      }

      const orders = await prisma.shopeeOrder.findMany(queryOptions);
      return OrderDataConverter.convertOrdersToJSON(orders);
    } catch (error: any) {
      this.logger.error(`Failed to get Shopee orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Clear orders by status for current tenant
   */
  async clearOrdersByStatus(status: string): Promise<void> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();
      await prisma.shopeeOrder.deleteMany({
        where: { tenantId, orderStatus: status } as any,
      });
      this.logger.info(`✅ Orders cleared`);
    } catch (error: any) {
      this.logger.error(`Failed to clear orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get orders count by status for current tenant
   */
  async getOrdersCount(status?: string): Promise<number> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();
      return await prisma.shopeeOrder.count({
        where: { tenantId, ...(status && { orderStatus: status }) } as any,
      });
    } catch (error: any) {
      this.logger.error(`Failed to get count: ${error.message}`);
      throw error;
    }
  }
}

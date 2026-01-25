/**
 * Lazada Order Repository
 * Handles Lazada-specific order persistence operations
 * Single Responsibility: Lazada order database operations
 */

import { getPrisma } from "../prismaClient";
import { tenantContext } from "../../utils/tenantContext";
import { OrderData } from "../../types/order.types";
import { OrderDataConverter } from "./orderDataConverter";

export class LazadaOrderRepository {
  private logger = {
    info: (msg: string) => console.log(`[LazadaOrderRepo] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[LazadaOrderRepo] ❌ ${msg}`),
  };

  private getTenantId(): string {
    return tenantContext.getTenantId();
  }

  /**
   * Save Lazada orders to database with items
   * Uses upsert for atomic operation - prevents race conditions
   */
  async saveOrders(orders: OrderData[]): Promise<void> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      if (orders.length === 0) return;

      // Get distinct statuses from incoming orders for cleaning old data
      const statuses = [...new Set(orders.map((o) => o.order_status))];
      const orderSns = new Set(orders.map((o) => o.order_sn));

      // Delete orders with same status that are NOT in current batch (old orders)
      for (const status of statuses) {
        await prisma.lazadaOrder.deleteMany({
          where: {
            tenantId,
            orderStatus: status,
            orderSn: { notIn: Array.from(orderSns) },
          } as any,
        });
      }

      for (const order of orders) {
        // Upsert - atomic operation (no race condition)
        await prisma.lazadaOrder.upsert({
          where: { orderSn: order.order_sn },
          create: {
            tenantId,
            orderSn: order.order_sn,
            shopId: order.shop_id,
            orderStatus: order.order_status,
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
          await prisma.lazadaOrderItem.deleteMany({
            where: { orderSn: order.order_sn },
          });

          for (const item of items) {
            await prisma.lazadaOrderItem.create({
              data: {
                orderSn: order.order_sn,
                itemId: item.item_id || 0,
                skuId: item.sku_id || item.sku || "",
                sellerSku: item.seller_sku || item.sku || "",
                productName: item.product_name || item.name || "",
                variationName: item.variation_name || "",
                quantity: item.quantity || 1,
                price: item.price || item.original_price || 0,
              },
            });
          }
        }
      }

      this.logger.info(`✅ Lazada orders saved`);
    } catch (error: any) {
      this.logger.error(`Failed to save Lazada orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get Lazada orders by status
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
      
      const orders = await prisma.lazadaOrder.findMany(queryOptions);
      return OrderDataConverter.convertOrdersToJSON(orders);
    } catch (error: any) {
      this.logger.error(`Failed to get Lazada orders: ${error.message}`);
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
      await prisma.lazadaOrder.deleteMany({
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
      return await prisma.lazadaOrder.count({
        where: { tenantId, ...(status && { orderStatus: status }) } as any,
      });
    } catch (error: any) {
      this.logger.error(`Failed to get count: ${error.message}`);
      throw error;
    }
  }
}

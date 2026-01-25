/**
 * TikTok Order Repository
 * Handles TikTok-specific order persistence operations
 * Single Responsibility: TikTok order database operations
 */

import { getPrisma } from "../prismaClient";
import { tenantContext } from "../../utils/tenantContext";
import { OrderData } from "../../types/order.types";
import { OrderDataConverter } from "./orderDataConverter";

export class TiktokOrderRepository {
  private logger = {
    info: (msg: string) => console.log(`[TiktokOrderRepo] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[TiktokOrderRepo] ❌ ${msg}`),
    warn: (msg: string) => console.warn(`[TiktokOrderRepo] ⚠️  ${msg}`),
  };

  private getTenantId(): string {
    return tenantContext.getTenantId();
  }

  /**
   * Save TikTok orders to database with items
   * Uses upsert for atomic operation - prevents race conditions
   */
  async saveOrders(orders: OrderData[]): Promise<void> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      if (orders.length === 0) return;

      // Get distinct statuses from incoming orders for cleaning old data
      const statuses = [
        ...new Set(orders.map((o) => (o as any).status || o.order_status)),
      ];
      const orderSns = new Set(orders.map((o) => o.order_sn));

      // Delete orders with same status that are NOT in current batch (old orders)
      for (const status of statuses) {
        await prisma.tiktokOrder.deleteMany({
          where: {
            tenantId,
            orderStatus: status,
            orderSn: { notIn: Array.from(orderSns) },
          } as any,
        });
      }

      for (const order of orders) {
        // TikTok API returns 'status' not 'order_status'
        const orderStatus = (order as any).status || order.order_status;
        if (!orderStatus) {
          this.logger.warn(
            `Skipping TikTok order without status: ${order.order_sn}`
          );
          continue;
        }

        // Upsert - atomic operation (no race condition)
        await prisma.tiktokOrder.upsert({
          where: { orderSn: order.order_sn },
          create: {
            tenantId,
            orderSn: order.order_sn,
            shopId: order.shop_id,
            orderStatus: orderStatus,
          } as any,
          update: {
            tenantId,
            orderStatus: orderStatus,
            updatedAt: new Date(),
          },
        });

        // Save order items - delete existing and recreate
        const items = (order as any).line_items || [];
        if (items.length > 0) {
          await prisma.tiktokOrderItem.deleteMany({
            where: { orderSn: order.order_sn },
          });

          for (const item of items) {
            await prisma.tiktokOrderItem.create({
              data: {
                orderSn: order.order_sn,
                lineItemId: item.line_item_id || item.id || "",
                productId: parseInt(String(item.product_id)) || 0,
                skuId: item.sku_id || item.sku || "",
                sellerSku: item.seller_sku || item.sku || "",
                productName: item.product_name || item.name || "",
                variationName: item.variation_name || "",
                quantity: item.quantity || item.model_quantity_purchased || 1,
                price: item.price || item.model_original_price || 0,
              },
            });
          }
        }
      }

      this.logger.info(`✅ TikTok orders saved`);
    } catch (error: any) {
      this.logger.error(`Failed to save TikTok orders: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get TikTok orders by status
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
      
      const orders = await prisma.tiktokOrder.findMany(queryOptions);
      return OrderDataConverter.convertOrdersToJSON(orders);
    } catch (error: any) {
      this.logger.error(`Failed to get TikTok orders: ${error.message}`);
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
      await prisma.tiktokOrder.deleteMany({
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
      return await prisma.tiktokOrder.count({
        where: { tenantId, ...(status && { orderStatus: status }) } as any,
      });
    } catch (error: any) {
      this.logger.error(`Failed to get count: ${error.message}`);
      throw error;
    }
  }
}

/**
 * FBS (Fulfillment by Shopee) Webhook Event Repository
 * Handles: fbs_sellable_stock, fbs_br_invoice, fbs_br_block_shop/sku
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface FBSWebhookEvent extends BaseWebhookEvent {
  itemId?: string;
  skuId?: string;
  stockChange?: number;
  invoiceNumber?: string;
}

export class FBSWebhookRepository extends BaseWebhookRepository<FBSWebhookEvent> {
  protected tableName = "WebhookFBSEvent";

  async insertEvent(event: FBSWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookFBSEvent (
          tenantId, eventType, shopId, itemId, skuId, stockChange, invoiceNumber, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.itemId || null},
          ${event.skuId || null},
          ${event.stockChange || null},
          ${event.invoiceNumber || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted FBS event: ${event.eventType} item ${event.itemId}`
      );
    } catch (error: any) {
      this.log("error", `Failed to insert FBS event: ${error.message}`);
      throw error;
    }
  }

  async getEventsByItem(itemId: string): Promise<FBSWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<FBSWebhookEvent[]>`
        SELECT * FROM WebhookFBSEvent
        WHERE tenantId = ${tenantId} AND itemId = ${itemId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by item: ${error.message}`);
      return [];
    }
  }

  async getEventsBySku(skuId: string): Promise<FBSWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<FBSWebhookEvent[]>`
        SELECT * FROM WebhookFBSEvent
        WHERE tenantId = ${tenantId} AND skuId = ${skuId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by SKU: ${error.message}`);
      return [];
    }
  }
}

/**
 * Product Webhook Event Repository
 * Handles: reserved_stock_change, video_upload, brand_register, violation_item, etc.
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface ProductWebhookEvent extends BaseWebhookEvent {
  itemId?: string;
  variationId?: string;
  action?: string;
}

export class ProductWebhookRepository extends BaseWebhookRepository<ProductWebhookEvent> {
  protected tableName = "WebhookProductEvent";

  async insertEvent(event: ProductWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookProductEvent (
          tenantId, eventType, shopId, itemId, variationId, action, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.itemId || null},
          ${event.variationId || null},
          ${event.action || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted product event: ${event.eventType} for item ${event.itemId}`
      );
    } catch (error: any) {
      this.log("error", `Failed to insert product event: ${error.message}`);
      throw error;
    }
  }

  async getEventsByItem(itemId: string): Promise<ProductWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<ProductWebhookEvent[]>`
        SELECT * FROM WebhookProductEvent
        WHERE tenantId = ${tenantId} AND itemId = ${itemId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by item: ${error.message}`);
      return [];
    }
  }
}

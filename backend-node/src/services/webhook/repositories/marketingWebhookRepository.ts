/**
 * Marketing Webhook Event Repository
 * Handles: item_promotion_push, promotion_update_push
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface MarketingWebhookEvent extends BaseWebhookEvent {
  itemId?: string;
  promotionId?: string;
  promotionType?: string;
  action?: string;
}

export class MarketingWebhookRepository extends BaseWebhookRepository<MarketingWebhookEvent> {
  protected tableName = "WebhookMarketingEvent";

  async insertEvent(event: MarketingWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookMarketingEvent (
          tenantId, eventType, shopId, itemId, promotionId, promotionType, action, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.itemId || null},
          ${event.promotionId || null},
          ${event.promotionType || null},
          ${event.action || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted marketing event: ${event.eventType} promo ${event.promotionId}`
      );
    } catch (error: any) {
      this.log("error", `Failed to insert marketing event: ${error.message}`);
      throw error;
    }
  }

  async getEventsByPromotion(
    promotionId: string
  ): Promise<MarketingWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<MarketingWebhookEvent[]>`
        SELECT * FROM WebhookMarketingEvent
        WHERE tenantId = ${tenantId} AND promotionId = ${promotionId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by promotion: ${error.message}`);
      return [];
    }
  }
}

/**
 * Shopee System Webhook Event Repository
 * Handles: shop_authorization, auth_expiry, penalty_update, video_upload_result
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface ShopeeSystemWebhookEvent extends BaseWebhookEvent {
  action?: string;
  expiryTime?: string;
  penaltyPoints?: number;
}

export class ShopeeSystemWebhookRepository extends BaseWebhookRepository<ShopeeSystemWebhookEvent> {
  protected tableName = "WebhookShopeeEvent";

  async insertEvent(event: ShopeeSystemWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookShopeeEvent (
          tenantId, eventType, shopId, action, expiryTime, penaltyPoints, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.action || null},
          ${event.expiryTime || null},
          ${event.penaltyPoints || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted shopee system event: ${event.eventType} for shop ${event.shopId}`
      );
    } catch (error: any) {
      this.log(
        "error",
        `Failed to insert shopee system event: ${error.message}`
      );
      throw error;
    }
  }

  async getEventsByShop(shopId: string): Promise<ShopeeSystemWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<ShopeeSystemWebhookEvent[]>`
        SELECT * FROM WebhookShopeeEvent
        WHERE tenantId = ${tenantId} AND shopId = ${shopId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by shop: ${error.message}`);
      return [];
    }
  }
}

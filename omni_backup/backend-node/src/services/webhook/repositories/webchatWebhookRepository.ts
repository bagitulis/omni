/**
 * Webchat Webhook Event Repository
 * Handles: webchat_push
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface WebchatWebhookEvent extends BaseWebhookEvent {
  conversationId?: string;
  messageType?: string;
  senderId?: string;
}

export class WebchatWebhookRepository extends BaseWebhookRepository<WebchatWebhookEvent> {
  protected tableName = "WebhookWebchatEvent";

  async insertEvent(event: WebchatWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookWebchatEvent (
          tenantId, eventType, shopId, conversationId, messageType, senderId, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.conversationId || null},
          ${event.messageType || null},
          ${event.senderId || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted webchat event: ${event.eventType} conv ${event.conversationId}`
      );
    } catch (error: any) {
      this.log("error", `Failed to insert webchat event: ${error.message}`);
      throw error;
    }
  }

  async getEventsByConversation(
    conversationId: string
  ): Promise<WebchatWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<WebchatWebhookEvent[]>`
        SELECT * FROM WebhookWebchatEvent
        WHERE tenantId = ${tenantId} AND conversationId = ${conversationId}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log(
        "error",
        `Failed to get events by conversation: ${error.message}`
      );
      return [];
    }
  }
}

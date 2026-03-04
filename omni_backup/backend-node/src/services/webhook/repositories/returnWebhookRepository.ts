/**
 * Return Webhook Event Repository
 * Handles: return_updates_push
 */

import {
  BaseWebhookRepository,
  BaseWebhookEvent,
} from "./baseWebhookRepository";

export interface ReturnWebhookEvent extends BaseWebhookEvent {
  orderSn?: string;
  returnSn?: string;
  status?: string;
  reason?: string;
}

export class ReturnWebhookRepository extends BaseWebhookRepository<ReturnWebhookEvent> {
  protected tableName = "WebhookReturnEvent";

  async insertEvent(event: ReturnWebhookEvent): Promise<void> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const now = new Date().toISOString();

      await prisma.$executeRaw`
        INSERT INTO WebhookReturnEvent (
          tenantId, eventType, shopId, orderSn, returnSn, status, reason, payload, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.eventType},
          ${event.shopId || null},
          ${event.orderSn || null},
          ${event.returnSn || null},
          ${event.status || null},
          ${event.reason || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${now},
          ${now}
        )
      `;

      this.log(
        "info",
        `✅ Inserted return event: ${event.eventType} return ${event.returnSn}`
      );
    } catch (error: any) {
      this.log("error", `Failed to insert return event: ${error.message}`);
      throw error;
    }
  }

  async getEventsByOrder(orderSn: string): Promise<ReturnWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<ReturnWebhookEvent[]>`
        SELECT * FROM WebhookReturnEvent
        WHERE tenantId = ${tenantId} AND orderSn = ${orderSn}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by order: ${error.message}`);
      return [];
    }
  }

  async getEventsByReturn(returnSn: string): Promise<ReturnWebhookEvent[]> {
    try {
      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<ReturnWebhookEvent[]>`
        SELECT * FROM WebhookReturnEvent
        WHERE tenantId = ${tenantId} AND returnSn = ${returnSn}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get events by return: ${error.message}`);
      return [];
    }
  }
}

/**
 * Webhook Order Event Repository
 * Handles webhook event persistence (append-only, no conflicts with sync)
 * SRP: Webhook order event database operations only
 */

import { getPrisma } from "../prismaClient";
import { tenantContext } from "../../utils/tenantContext";

export interface WebhookOrderEventData {
  platform: string;
  eventType: string;
  orderSn: string;
  shopId?: string;
  oldStatus?: string;
  newStatus?: string;
  fulfillmentStatus?: string;
  packageNumber?: string;
  payload?: any;
  webhookLogId?: string;
}

export class WebhookOrderEventRepository {
  private logger = {
    info: (msg: string) => console.log(`[WebhookEventRepo] ℹ️  ${msg}`),
    error: (msg: string) => console.error(`[WebhookEventRepo] ❌ ${msg}`),
  };

  private getTenantId(): string {
    return tenantContext.getTenantId();
  }

  /**
   * Insert a new webhook order event (append-only)
   * Never updates or deletes - maintains full audit trail
   */
  async insertEvent(event: WebhookOrderEventData): Promise<void> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      await prisma.$executeRaw`
        INSERT INTO WebhookOrderEvent (
          tenantId, platform, eventType, orderSn, shopId,
          oldStatus, newStatus, fulfillmentStatus, packageNumber,
          payload, webhookLogId, processedAt, createdAt
        ) VALUES (
          ${tenantId},
          ${event.platform},
          ${event.eventType},
          ${event.orderSn},
          ${event.shopId || null},
          ${event.oldStatus || null},
          ${event.newStatus || null},
          ${event.fulfillmentStatus || null},
          ${event.packageNumber || null},
          ${event.payload ? JSON.stringify(event.payload) : null},
          ${event.webhookLogId || null},
          ${new Date().toISOString()},
          ${new Date().toISOString()}
        )
      `;

      this.logger.info(
        `✅ Inserted ${event.platform} ${event.eventType} for ${event.orderSn}`
      );
    } catch (error: any) {
      this.logger.error(`Failed to insert webhook event: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get latest status for an order from webhook events
   * Returns most recent event for the given orderSn
   */
  async getLatestEventForOrder(
    orderSn: string,
    platform: string
  ): Promise<any | null> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<any[]>`
        SELECT * FROM WebhookOrderEvent
        WHERE tenantId = ${tenantId}
          AND orderSn = ${orderSn}
          AND platform = ${platform}
        ORDER BY createdAt DESC
        LIMIT 1
      `;

      return result.length > 0 ? result[0] : null;
    } catch (error: any) {
      this.logger.error(`Failed to get latest event: ${error.message}`);
      return null;
    }
  }

  /**
   * Get all events for an order (full history)
   */
  async getEventsForOrder(orderSn: string, platform: string): Promise<any[]> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<any[]>`
        SELECT * FROM WebhookOrderEvent
        WHERE tenantId = ${tenantId}
          AND orderSn = ${orderSn}
          AND platform = ${platform}
        ORDER BY createdAt DESC
      `;

      return result;
    } catch (error: any) {
      this.logger.error(`Failed to get events: ${error.message}`);
      return [];
    }
  }

  /**
   * Get recent webhook events for dashboard/monitoring
   */
  async getRecentEvents(limit: number = 50): Promise<any[]> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<any[]>`
        SELECT * FROM WebhookOrderEvent
        WHERE tenantId = ${tenantId}
        ORDER BY createdAt DESC
        LIMIT ${limit}
      `;

      return result;
    } catch (error: any) {
      this.logger.error(`Failed to get recent events: ${error.message}`);
      return [];
    }
  }

  /**
   * Get event counts by platform and type (for monitoring)
   */
  async getEventCounts(): Promise<Record<string, number>> {
    try {
      const prisma = getPrisma();
      const tenantId = this.getTenantId();

      const result = await prisma.$queryRaw<any[]>`
        SELECT platform, eventType, COUNT(*) as count
        FROM WebhookOrderEvent
        WHERE tenantId = ${tenantId}
        GROUP BY platform, eventType
      `;

      const counts: Record<string, number> = {};
      for (const row of result) {
        const key = `${row.platform}_${row.eventType}`;
        counts[key] = Number(row.count);
      }

      return counts;
    } catch (error: any) {
      this.logger.error(`Failed to get event counts: ${error.message}`);
      return {};
    }
  }
}

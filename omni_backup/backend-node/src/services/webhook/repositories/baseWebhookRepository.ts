/**
 * Base Webhook Event Repository
 * Abstract class for all webhook category repositories
 */

import { getPrisma } from "../../prismaClient";
import { tenantContext } from "../../../utils/tenantContext";

export interface BaseWebhookEvent {
  tenantId?: string;
  eventType: string;
  shopId?: string;
  payload?: unknown;
}

// SECURITY: Whitelist of allowed table names to prevent SQL injection
const ALLOWED_TABLES = new Set([
  "ShopeeOrderWebhookEvent",
  "ShopeeFBSWebhookEvent",
  "ShopeeSystemWebhookEvent",
  "ShopeeMarketingWebhookEvent",
  "TiktokOrderWebhookEvent",
  "TiktokFBSWebhookEvent",
  "LazadaOrderWebhookEvent",
  "WebhookOrderEvent",
  "FBSWebhookEvent",
  "MarketingWebhookEvent",
]);

export abstract class BaseWebhookRepository<T extends BaseWebhookEvent> {
  protected abstract tableName: string;

  protected getTenantId(): string {
    return tenantContext.getTenantId();
  }

  protected getPrismaClient() {
    return getPrisma();
  }

  /**
   * SECURITY: Validate table name against whitelist to prevent SQL injection
   */
  protected validateTableName(): void {
    if (!ALLOWED_TABLES.has(this.tableName)) {
      throw new Error(`SECURITY_ERROR: Invalid table name: ${this.tableName}`);
    }
  }

  protected log(level: "info" | "error", msg: string): void {
    const prefix = `[${this.tableName}]`;
    if (level === "error") {
      console.error(`${prefix} ❌ ${msg}`);
    } else {
      console.log(`${prefix} ℹ️  ${msg}`);
    }
  }

  abstract insertEvent(event: T): Promise<void>;

  async getRecentEvents(limit: number = 50): Promise<T[]> {
    try {
      // SECURITY: Validate table name before using in query
      this.validateTableName();

      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const table = this.tableName;

      // SECURITY: Validate limit is a reasonable number
      const safeLimit = Math.min(Math.max(1, Math.floor(limit)), 1000);

      const result = await prisma.$queryRawUnsafe<T[]>(
        `SELECT * FROM ${table} WHERE tenantId = ? ORDER BY createdAt DESC LIMIT ?`,
        tenantId,
        safeLimit
      );

      return result;
    } catch (error: any) {
      this.log("error", `Failed to get recent events: ${error.message}`);
      return [];
    }
  }

  async getEventCounts(): Promise<Record<string, number>> {
    try {
      // SECURITY: Validate table name before using in query
      this.validateTableName();

      const prisma = this.getPrismaClient();
      const tenantId = this.getTenantId();
      const table = this.tableName;

      const result = await prisma.$queryRawUnsafe<any[]>(
        `SELECT eventType, COUNT(*) as count FROM ${table} WHERE tenantId = ? GROUP BY eventType`,
        tenantId
      );

      const counts: Record<string, number> = {};
      for (const row of result) {
        counts[row.eventType] = Number(row.count);
      }

      return counts;
    } catch (error: any) {
      this.log("error", `Failed to get event counts: ${error.message}`);
      return {};
    }
  }
}

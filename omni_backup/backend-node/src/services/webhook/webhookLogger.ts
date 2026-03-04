import { getPrisma } from "../prismaClient";
import { getDbManager } from "../dbConnectionManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("WebhookLogger");

type PlatformType = "shopee" | "tiktok" | "lazada";

/**
 * Get Prisma client for webhook logging
 * Uses first available tenant since webhooks don't have tenant context
 */
function getWebhookPrisma(): ReturnType<typeof getPrisma> {
  const dbManager = getDbManager();
  const tenants = dbManager.getAllTenants();
  if (tenants.length === 0) {
    throw new Error("No tenants configured");
  }
  // Use first tenant for webhook logs
  return getPrisma(tenants[0]);
}

export interface WebhookLogEntry {
  id: string;
  platform: string;
  eventType: string;
  status: string;
  createdAt: Date;
  processedAt?: Date | null;
  errorMsg?: string | null;
}

export interface PaginatedResult<T> {
  items: T[];
  total: number;
}

export interface WebhookStats {
  total: number;
  byPlatform: Record<string, number>;
  byStatus: Record<string, number>;
  last24h: number;
}

/**
 * Log webhook to database
 * Requires tenantId - throws error if missing
 */
export async function logWebhook(
  platform: PlatformType,
  eventType: string,
  payload: unknown,
  headers: Record<string, string>,
  tenantId?: string
): Promise<string> {
  // Validate tenantId is provided
  if (!tenantId) {
    logger.warn(`No tenantId provided for ${platform} webhook - cannot save`);
    throw new Error("Tenant ID is required to save webhook log");
  }

  // Validate tenant exists
  const dbManager = getDbManager();
  const tenants = dbManager.getAllTenants();
  if (!tenants.includes(tenantId)) {
    logger.warn(`Invalid tenantId '${tenantId}' for ${platform} webhook`);
    throw new Error(`Invalid tenant: ${tenantId}`);
  }

  const prisma = getPrisma(tenantId);

  const log = await prisma.webhookLog.create({
    data: {
      tenantId,
      platform,
      eventType,
      payload: JSON.stringify(payload),
      headers: JSON.stringify(headers),
      status: "received",
    },
  });

  logger.info(`Logged ${platform} webhook for tenant ${tenantId}: ${log.id}`);
  return log.id;
}

/**
 * Update webhook log status
 */
export async function updateWebhookStatus(
  logId: string,
  status: "processed" | "failed",
  errorMsg?: string
): Promise<void> {
  const prisma = getWebhookPrisma();

  await prisma.webhookLog.update({
    where: { id: logId },
    data: {
      status,
      processedAt: new Date(),
      errorMsg,
    },
  });
}

/**
 * Get webhook logs with pagination
 */
export async function getWebhookLogs(
  tenantId: string,
  options: { page: number; limit: number; status?: string }
): Promise<PaginatedResult<WebhookLogEntry>> {
  const prisma = getPrisma(tenantId);

  const where: Record<string, unknown> = {};
  if (options.status) {
    where.status = options.status;
  }

  const [items, total] = await Promise.all([
    prisma.webhookLog.findMany({
      where,
      orderBy: { createdAt: "desc" },
      skip: (options.page - 1) * options.limit,
      take: options.limit,
      select: {
        id: true,
        platform: true,
        eventType: true,
        status: true,
        createdAt: true,
        processedAt: true,
        errorMsg: true,
      },
    }),
    prisma.webhookLog.count({ where }),
  ]);

  return { items, total };
}

/**
 * Get webhook logs filtered by platform
 */
export async function getWebhookLogsByPlatform(
  tenantId: string,
  platform: string,
  options: { page: number; limit: number }
): Promise<PaginatedResult<WebhookLogEntry>> {
  const prisma = getPrisma(tenantId);

  const [items, total] = await Promise.all([
    prisma.webhookLog.findMany({
      where: { platform },
      orderBy: { createdAt: "desc" },
      skip: (options.page - 1) * options.limit,
      take: options.limit,
      select: {
        id: true,
        platform: true,
        eventType: true,
        status: true,
        createdAt: true,
        processedAt: true,
        errorMsg: true,
      },
    }),
    prisma.webhookLog.count({ where: { platform } }),
  ]);

  return { items, total };
}

/**
 * Get webhook statistics
 */
export async function getWebhookStats(tenantId: string): Promise<WebhookStats> {
  const prisma = getPrisma(tenantId);
  const last24h = new Date(Date.now() - 24 * 60 * 60 * 1000);

  const [total, byPlatformRaw, byStatusRaw, last24hCount] = await Promise.all([
    prisma.webhookLog.count(),
    prisma.webhookLog.groupBy({ by: ["platform"], _count: true }),
    prisma.webhookLog.groupBy({ by: ["status"], _count: true }),
    prisma.webhookLog.count({ where: { createdAt: { gte: last24h } } }),
  ]);

  const byPlatform: Record<string, number> = {};
  (byPlatformRaw as Array<{ platform: string; _count: number }>).forEach(
    (item) => {
      byPlatform[item.platform] = item._count;
    }
  );

  const byStatus: Record<string, number> = {};
  (byStatusRaw as Array<{ status: string; _count: number }>).forEach((item) => {
    byStatus[item.status] = item._count;
  });

  return { total, byPlatform, byStatus, last24h: last24hCount };
}

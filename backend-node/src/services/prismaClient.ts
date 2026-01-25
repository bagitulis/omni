import { PrismaClient } from "@prisma/client";
import { getDbManager } from "./dbConnectionManager";
import { tenantContext } from "../utils/tenantContext";

/**
 * Get or create Prisma client (singleton pattern per tenant)
 * Multi-tenant aware: Returns tenant-specific database connection
 * Uses DatabaseConnectionManager to manage connections
 *
 * @param tenantId - Optional explicit tenant ID. If not provided, uses tenantContext default
 */
export function getPrisma(tenantId?: string): PrismaClient {
  const dbManager = getDbManager();
  const finalTenantId = tenantId || tenantContext.getTenantId();
  return dbManager.getConnection(finalTenantId);
}

/**
 * Disconnect Prisma client for specific tenant
 */
export async function disconnectPrisma(tenantId?: string): Promise<void> {
  const dbManager = getDbManager();
  if (tenantId) {
    await dbManager.disconnect(tenantId);
  } else {
    await dbManager.disconnectAll();
  }
}

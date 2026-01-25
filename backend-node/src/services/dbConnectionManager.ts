import { PrismaClient } from "@prisma/client";
import path from "path";
import fs from "fs";
import { tenantContext } from "../utils/tenantContext";

interface TenantConfig {
  dbPath: string;
  shopName: string;
  isGlobal?: boolean;
}

interface TenantsConfig {
  [key: string]: TenantConfig;
}

class DatabaseConnectionManager {
  private connections: Map<string, PrismaClient> = new Map();
  private tenantsConfig: TenantsConfig;

  constructor() {
    // Use process.cwd() to get root directory (/app)
    const configPath = path.join(process.cwd(), "config/static/tenants.json");
    const configContent = fs.readFileSync(configPath, "utf-8");
    this.tenantsConfig = JSON.parse(configContent);
  }

  /**
   * Get Prisma client for specific tenant
   * Creates connection if doesn't exist (cached)
   */
  getConnection(tenantId: string): PrismaClient {
    if (!this.tenantsConfig[tenantId]) {
      throw new Error(`Tenant '${tenantId}' not found in configuration`);
    }

    // Return cached connection if exists
    if (this.connections.has(tenantId)) {
      return this.connections.get(tenantId)!;
    }

    // Create new connection
    const tenantConfig = this.tenantsConfig[tenantId];
    const dbUrl = `file:${path.resolve(tenantConfig.dbPath)}`;

    const prisma = new PrismaClient({
      datasources: {
        db: {
          url: dbUrl,
        },
      },
    });

    this.connections.set(tenantId, prisma);
    return prisma;
  }

  /**
   * Get default connection (from tenantContext config)
   */
  getDefaultConnection(): PrismaClient {
    return this.getConnection(tenantContext.getTenantId());
  }

  /**
   * Get tenant config
   */
  getTenantConfig(tenantId: string): TenantConfig {
    if (!this.tenantsConfig[tenantId]) {
      throw new Error(`Tenant '${tenantId}' not found`);
    }
    return this.tenantsConfig[tenantId];
  }

  /**
   * Get all available tenants (includes system for backward compatibility)
   * ⚠️ For job processing, use getRealTenants() instead!
   */
  getAllTenants(): string[] {
    return Object.keys(this.tenantsConfig);
  }

  /**
   * Get only REAL tenants (excludes 'system' and 'default')
   * Use this for job processing, webhooks, and operations that require real tenants
   */
  getRealTenants(): string[] {
    return Object.keys(this.tenantsConfig).filter(
      (id) =>
        id !== "system" &&
        id !== "default" &&
        !this.tenantsConfig[id]?.isGlobal,
    );
  }

  /**
   * Check if tenant exists
   */
  tenantExists(tenantId: string): boolean {
    return !!this.tenantsConfig[tenantId];
  }

  /**
   * Check if tenant is a real tenant (not system/global)
   */
  isRealTenant(tenantId: string): boolean {
    if (!this.tenantsConfig[tenantId]) return false;
    if (tenantId === "system" || tenantId === "default") return false;
    if (this.tenantsConfig[tenantId]?.isGlobal) return false;
    return true;
  }

  /**
   * Disconnect all connections
   */
  async disconnectAll(): Promise<void> {
    for (const prisma of this.connections.values()) {
      await prisma.$disconnect();
    }
    this.connections.clear();
  }

  /**
   * Disconnect specific tenant
   */
  async disconnect(tenantId: string): Promise<void> {
    const prisma = this.connections.get(tenantId);
    if (prisma) {
      await prisma.$disconnect();
      this.connections.delete(tenantId);
    }
  }

  /**
   * Get list of available tenants for UI selection
   * Excludes global tenants (like system) that don't have User tables
   */
  getAvailableTenants(): Array<{ id: string; shopName: string }> {
    return Object.entries(this.tenantsConfig)
      .filter(([, config]) => !config.isGlobal)
      .map(([id, config]) => ({
        id,
        shopName: config.shopName,
      }));
  }
}

// Singleton instance
let instance: DatabaseConnectionManager | null = null;

export function getDbManager(): DatabaseConnectionManager {
  if (!instance) {
    instance = new DatabaseConnectionManager();
  }
  return instance;
}

export default DatabaseConnectionManager;

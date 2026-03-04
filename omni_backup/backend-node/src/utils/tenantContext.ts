import { AsyncLocalStorage } from "async_hooks";
import fs from "fs";
import path from "path";

/**
 * TenantContext - Request-scoped tenant tracking (STRICT MODE)
 * Uses AsyncLocalStorage for per-request isolation
 *
 * ⚠️ NO DEFAULT TENANT - must be explicitly set per request!
 * If tenant is not set, throws error instead of using fallback.
 */

interface TenantStore {
  tenantId: string;
}

// Load valid tenants from config once at startup
let validTenants: Set<string> | null = null;

function loadValidTenants(): Set<string> {
  if (validTenants) return validTenants;

  try {
    const configPath = path.resolve("config/static/tenants.json");
    if (!fs.existsSync(configPath)) {
      console.error("[TenantContext] ❌ tenants.json not found!");
      validTenants = new Set();
      return validTenants;
    }

    const config = JSON.parse(fs.readFileSync(configPath, "utf-8"));
    validTenants = new Set(Object.keys(config));
    console.log(
      `[TenantContext] ✅ Loaded ${validTenants.size} valid tenant(s)`,
    );
    return validTenants;
  } catch (error) {
    console.error("[TenantContext] ❌ Error loading tenants.json:", error);
    validTenants = new Set();
    return validTenants;
  }
}

class TenantContextManager {
  private storage = new AsyncLocalStorage<TenantStore>();

  constructor() {
    // Load valid tenants at startup
    loadValidTenants();
  }

  /**
   * Run async code within tenant context
   * @param tenantId - Must be a valid tenant from tenants.json
   * @param fn - Function to run within context
   */
  async run<T>(tenantId: string, fn: () => Promise<T> | T): Promise<T> {
    // Validate tenant before running
    if (!this.isValidTenant(tenantId)) {
      throw new Error(`Invalid tenant ID: ${tenantId}`);
    }
    return this.storage.run({ tenantId }, () => fn());
  }

  /**
   * Get current tenant ID from context
   * ⚠️ THROWS ERROR if not set - no default fallback!
   */
  getTenantId(): string {
    const store = this.storage.getStore();
    if (!store?.tenantId) {
      throw new Error(
        "[TenantContext] ❌ No tenant ID in context! " +
          "Ensure request passes through tenant middleware or use tenantContext.run()",
      );
    }
    return store.tenantId;
  }

  /**
   * Get current tenant ID or null (safe version for optional checks)
   * Use this when you need to check if tenant is available without throwing
   */
  getTenantIdOrNull(): string | null {
    const store = this.storage.getStore();
    return store?.tenantId || null;
  }

  /**
   * Set tenant ID in current context for per-request isolation
   * Used by middleware to set tenant for each request
   * @param tenantId - Must be a valid tenant from tenants.json
   */
  setTenantId(tenantId: string): void {
    if (!this.isValidTenant(tenantId)) {
      throw new Error(`Invalid tenant ID: ${tenantId}`);
    }

    const store = this.storage.getStore();
    if (store) {
      store.tenantId = tenantId;
    } else {
      console.warn(
        `[TenantContext] ⚠️ setTenantId called without active context. ` +
          `Use tenantContext.run() instead for tenant: ${tenantId}`,
      );
    }
  }

  /**
   * Check if tenantId is available in current context
   */
  isAvailable(): boolean {
    const store = this.storage.getStore();
    return store?.tenantId !== undefined;
  }

  /**
   * Check if a tenant ID is valid (exists in tenants.json)
   */
  isValidTenant(tenantId: string): boolean {
    const tenants = loadValidTenants();
    return tenants.has(tenantId);
  }

  /**
   * Get list of all valid tenant IDs
   */
  getValidTenants(): string[] {
    return Array.from(loadValidTenants());
  }

  /**
   * Get list of tenant IDs that have job databases
   * EXCLUDES: 'system' (uses GlobalConfig only, no jobs database per AGENTS.MD)
   * EXCLUDES: Any tenant marked as isGlobal in tenants.json
   */
  getTenantsWithJobDatabases(): string[] {
    try {
      const configPath = path.resolve("config/static/tenants.json");
      if (!fs.existsSync(configPath)) {
        return [];
      }
      const config = JSON.parse(fs.readFileSync(configPath, "utf-8"));
      return Object.keys(config).filter((id) => {
        if (id === "system" || id === "default") return false;
        const tenant = config[id];
        if (tenant?.isGlobal) return false;
        return true;
      });
    } catch (error) {
      console.error("[TenantContext] Error loading tenants for jobs:", error);
      return [];
    }
  }
}

export const tenantContext = new TenantContextManager();

/**
 * Helper function to get tenant context with prisma client
 * Compatible with services expecting { tenantId, prisma } pattern
 * Uses AsyncLocalStorage for proper request isolation
 *
 * ⚠️ THROWS ERROR if tenant not set!
 */
export function getTenantContext(): {
  tenantId: string;
  prisma: import("@prisma/client").PrismaClient;
} {
  // Import here to avoid circular dependency
  const { getDbManager } = require("../services/dbConnectionManager");

  const tenantId = tenantContext.getTenantId(); // Throws if not set
  const prisma = getDbManager().getConnection(tenantId);

  return { tenantId, prisma };
}

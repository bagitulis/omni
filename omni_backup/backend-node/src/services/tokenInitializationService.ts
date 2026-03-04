import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { IConfigManager } from "../config/managers/baseConfigManager";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

type PlatformType = "shopee" | "lazada" | "tiktok";

const logger = getLogger("TokenInit");

/**
 * Token Initialization Service
 * Handles initialization of all platform config managers and API clients
 *
 * MULTI-TENANT: Each tenant has its own instance with separate config/clients
 *
 * Single Responsibility: Setup phase only
 * - Load config managers
 * - Create API clients
 * - Validate connections
 */
export class TokenInitializationService {
  private configManagers: Map<PlatformType, IConfigManager> = new Map();
  private apiClients: Map<PlatformType, any> = new Map();
  private initialized: boolean = false;
  private tenantId: string;

  /**
   * @param tenantId - Tenant ID for this service instance
   */
  constructor(tenantId: string) {
    this.tenantId = tenantId;
  }

  /**
   * Initialize all config managers and API clients for this tenant
   */
  async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }

    try {
      // Initialize Shopee with explicit tenant
      const shopeeConfig = new ShopeeConfigManager(this.tenantId);
      await shopeeConfig.loadConfig();
      this.configManagers.set("shopee", shopeeConfig);
      this.apiClients.set("shopee", new ShopeeAPIClient(shopeeConfig));

      // Initialize Lazada with explicit tenant
      const lazadaConfig = new LazadaConfigManager(this.tenantId);
      await lazadaConfig.loadConfig();
      this.configManagers.set("lazada", lazadaConfig);
      this.apiClients.set("lazada", new LazadaAPIClient(lazadaConfig));

      // Initialize TikTok with explicit tenant
      const tiktokConfig = new TiktokConfigManager(this.tenantId);
      await tiktokConfig.loadConfig();
      this.configManagers.set("tiktok", tiktokConfig);
      this.apiClients.set("tiktok", new TiktokAPIClient(tiktokConfig));

      this.initialized = true;
      logger.info(`Token services initialized for tenant: ${this.tenantId}`);
    } catch (error) {
      logger.error(
        `Failed to initialize token services for ${this.tenantId}: ${error}`
      );
      throw error;
    }
  }

  /**
   * Get config manager for a platform
   */
  getConfigManager(platform: PlatformType): IConfigManager {
    const config = this.configManagers.get(platform);
    if (!config) {
      throw new Error(`Config manager not found for platform: ${platform}`);
    }
    return config;
  }

  /**
   * Get API client for a platform
   */
  getAPIClient(platform: PlatformType): any {
    const client = this.apiClients.get(platform);
    if (!client) {
      throw new Error(`API client not found for platform: ${platform}`);
    }
    return client;
  }

  /**
   * Get all config managers
   */
  getConfigManagers(): Map<PlatformType, IConfigManager> {
    return new Map(this.configManagers);
  }

  /**
   * Get all API clients
   */
  getAPIClients(): Map<PlatformType, any> {
    return new Map(this.apiClients);
  }

  /**
   * Check if initialized
   */
  isInitialized(): boolean {
    return this.initialized;
  }

  /**
   * Get tenant ID for this instance
   */
  getTenantId(): string {
    return this.tenantId;
  }
}

/**
 * Per-tenant instance registry
 */
const tenantInstances: Map<string, TokenInitializationService> = new Map();

/**
 * Get TokenInitializationService for a specific tenant
 * @param tenantId - Optional tenant ID (defaults to current context)
 */
export function getTokenInitializationService(
  tenantId?: string
): TokenInitializationService {
  // Resolve tenant ID
  const resolvedTenantId =
    tenantId ?? tenantContext.getTenantIdOrNull() ?? undefined;

  if (!resolvedTenantId) {
    throw new Error(
      "[TokenInitializationService] No tenant ID provided and none in context"
    );
  }

  // Get or create instance for this tenant
  let instance = tenantInstances.get(resolvedTenantId);
  if (!instance) {
    instance = new TokenInitializationService(resolvedTenantId);
    tenantInstances.set(resolvedTenantId, instance);
    logger.info(
      `Created TokenInitializationService instance for tenant: ${resolvedTenantId}`
    );
  }

  return instance;
}

/**
 * Clear cached instance for a tenant (for testing)
 */
export function clearTokenInitializationService(tenantId: string): void {
  tenantInstances.delete(tenantId);
}

/**
 * Clear all cached instances (for testing)
 */
export function clearAllTokenInitializationServices(): void {
  tenantInstances.clear();
}

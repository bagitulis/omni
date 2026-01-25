import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

type PlatformType = "shopee" | "lazada" | "tiktok";

const logger = getLogger("PlatformCoord");

/**
 * Platform Coordination Service
 * Manages all platform API clients and config managers
 *
 * MULTI-TENANT: Each tenant has its own instance with separate clients/configs
 *
 * Single Responsibility: Platform client/config coordination
 * - Initialize platform clients
 * - Provide access to platform-specific resources
 * - Manage platform configurations
 */
export class PlatformCoordinationService {
  private platformClients: Map<PlatformType, any> = new Map();
  private configManagers: Map<PlatformType, any> = new Map();
  private tenantId: string;

  /**
   * @param tenantId - Tenant ID for this service instance
   */
  constructor(tenantId: string) {
    this.tenantId = tenantId;
  }

  /**
   * Initialize all platform clients for this tenant
   */
  async initializePlatforms(): Promise<void> {
    try {
      logger.info(
        `Initializing all platform clients for tenant: ${this.tenantId}...`
      );

      // Initialize Shopee with explicit tenant
      const shopeeConfig = new ShopeeConfigManager(this.tenantId);
      await shopeeConfig.loadConfig();
      this.configManagers.set("shopee", shopeeConfig);
      this.platformClients.set("shopee", new ShopeeAPIClient(shopeeConfig));

      // Initialize Lazada with explicit tenant
      const lazadaConfig = new LazadaConfigManager(this.tenantId);
      await lazadaConfig.loadConfig();
      this.configManagers.set("lazada", lazadaConfig);
      this.platformClients.set("lazada", new LazadaAPIClient(lazadaConfig));

      // Initialize TikTok with explicit tenant
      const tiktokConfig = new TiktokConfigManager(this.tenantId);
      await tiktokConfig.loadConfig();
      this.configManagers.set("tiktok", tiktokConfig);
      this.platformClients.set("tiktok", new TiktokAPIClient(tiktokConfig));

      logger.info(
        `All platform clients initialized for tenant: ${this.tenantId}`
      );
    } catch (error) {
      logger.error(
        `Failed to initialize platforms for ${this.tenantId}: ${error}`
      );
      throw error;
    }
  }

  /**
   * Get platform client
   */
  getClient(platform: PlatformType): any {
    const client = this.platformClients.get(platform);
    if (!client) {
      throw new Error(`Platform client not initialized: ${platform}`);
    }
    return client;
  }

  /**
   * Get platform config manager
   */
  getConfigManager(platform: PlatformType): any {
    const config = this.configManagers.get(platform);
    if (!config) {
      throw new Error(`Config manager not found: ${platform}`);
    }
    return config;
  }

  /**
   * Get all platform clients
   */
  getAllClients(): Map<PlatformType, any> {
    return new Map(this.platformClients);
  }

  /**
   * Get all config managers
   */
  getAllConfigManagers(): Map<PlatformType, any> {
    return new Map(this.configManagers);
  }

  /**
   * Check if platform is available
   */
  isPlatformAvailable(platform: PlatformType): boolean {
    return (
      this.platformClients.has(platform) && this.configManagers.has(platform)
    );
  }

  /**
   * Get available platforms
   */
  getAvailablePlatforms(): PlatformType[] {
    const platforms: PlatformType[] = ["shopee", "lazada", "tiktok"];
    return platforms.filter((p) => this.isPlatformAvailable(p));
  }

  /**
   * Reload configuration for a platform
   */
  async reloadPlatformConfig(platform: PlatformType): Promise<boolean> {
    try {
      logger.info(`Reloading config for ${platform}...`);
      const configManager = this.getConfigManager(platform);
      await configManager.loadConfig();

      // Reinitialize client with new config
      const ClientClass =
        platform === "shopee"
          ? ShopeeAPIClient
          : platform === "lazada"
            ? LazadaAPIClient
            : TiktokAPIClient;

      const newClient = new ClientClass(configManager);
      this.platformClients.set(platform, newClient);

      logger.info(`${platform} config reloaded`);
      return true;
    } catch (error) {
      logger.error(`Error reloading ${platform} config: ${error}`);
      return false;
    }
  }

  /**
   * Get the tenant ID for this service instance
   */
  getTenantId(): string {
    return this.tenantId;
  }
}

// Per-tenant instance registry
const tenantInstances: Map<string, PlatformCoordinationService> = new Map();

/**
 * Get PlatformCoordinationService for specific tenant
 *
 * @param tenantId - Tenant ID (optional, will use context if not provided)
 * @returns PlatformCoordinationService instance for the tenant
 * @throws Error if no tenant ID available
 */
export function getPlatformCoordinationService(
  tenantId?: string
): PlatformCoordinationService {
  const resolvedTenantId =
    tenantId ?? tenantContext.getTenantIdOrNull() ?? undefined;

  if (!resolvedTenantId) {
    throw new Error(
      "PlatformCoordinationService: No tenant ID provided and no tenant context available"
    );
  }

  let instance = tenantInstances.get(resolvedTenantId);
  if (!instance) {
    instance = new PlatformCoordinationService(resolvedTenantId);
    tenantInstances.set(resolvedTenantId, instance);
    logger.info(
      `Created new PlatformCoordinationService instance for tenant: ${resolvedTenantId}`
    );
  }

  return instance;
}

/**
 * Clear all tenant instances (for testing purposes)
 */
export function clearPlatformCoordinationInstances(): void {
  tenantInstances.clear();
}

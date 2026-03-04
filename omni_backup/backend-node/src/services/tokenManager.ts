import {
  getTokenInitializationService,
  TokenInitializationService,
} from "./tokenInitializationService";
import {
  getTokenRefreshService,
  TokenRefreshService,
} from "./tokenRefreshService";
import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";
import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

type PlatformType = "shopee" | "lazada" | "tiktok";

const logger = getLogger("TokenManager");

/**
 * Token Manager Service (Facade)
 * Central facade for managing all platform tokens
 *
 * MULTI-TENANT: Each tenant has its own instance with separate services
 *
 * Delegates to specialized services:
 * - TokenInitializationService: Setup phase
 * - TokenRefreshService: Runtime token management
 */
export class TokenManager {
  private initService: TokenInitializationService;
  private refreshService: TokenRefreshService;
  private tenantId: string;
  private initialized = false;

  /**
   * @param tenantId - Tenant ID for this manager instance
   */
  constructor(tenantId: string) {
    this.tenantId = tenantId;
    // Get per-tenant services
    this.initService = getTokenInitializationService(tenantId);
    this.refreshService = getTokenRefreshService(tenantId);
  }

  /**
   * Check if service is initialized
   */
  isInitialized(): boolean {
    return this.initialized;
  }

  /**
   * Delegate initialization to TokenInitializationService
   */
  async initialize(): Promise<void> {
    if (this.initialized) return;
    try {
      await this.initService.initialize();
      this.initialized = true;
      logger.info(`TokenManager initialized for tenant: ${this.tenantId}`);
    } catch (error) {
      logger.error(
        `Failed to initialize TokenManager for ${this.tenantId}: ${error}`
      );
      throw error;
    }
  }

  /**
   * Get API client for a platform
   */
  private getAPIClient(platform: PlatformType) {
    return this.initService.getAPIClient(platform);
  }

  /**
   * Delegate refresh to TokenRefreshService
   */
  async refreshToken(platform: PlatformType): Promise<boolean> {
    return this.refreshService.refreshToken(platform);
  }

  /**
   * Refresh all platform tokens
   * @param force - Force refresh even if not expired
   */
  async refreshAllTokens(force = false): Promise<Record<string, boolean>> {
    const results = await this.refreshService.refreshAllTokens(force);
    const record: Record<string, boolean> = {};
    results.forEach((value, key) => {
      record[key] = value;
    });
    return record;
  }

  /**
   * Delegate token status to TokenRefreshService
   */
  async getTokenStatus(platform: PlatformType): Promise<any> {
    const status = await this.refreshService.getTokenStatus(platform);
    return {
      platform: status.platform,
      isExpired: status.isExpired,
      expiresAt: status.expiresAt,
      refreshTokenExpiresAt: status.refreshTokenExpiresAt,
    };
  }

  /**
   * Get status for all platforms
   */
  async getAllTokenStatus(): Promise<Record<string, any>> {
    const statuses = await this.refreshService.getAllTokenStatus();
    const record: Record<string, any> = {};
    statuses.forEach((status) => {
      record[status.platform] = {
        platform: status.platform,
        isExpired: status.isExpired,
        expiresAt: status.expiresAt,
        refreshTokenExpiresAt: status.refreshTokenExpiresAt,
      };
    });
    return record;
  }

  /**
   * Check and auto-refresh tokens
   * Uses force=true to always refresh, ensuring tokens stay fresh
   */
  async checkAndRefreshExpiredTokens(): Promise<Record<string, boolean>> {
    return this.refreshAllTokens(true);
  }

  /**
   * Get Shopee API client
   */
  getShopeeClient(): ShopeeAPIClient | null {
    try {
      return this.getAPIClient("shopee");
    } catch {
      return null;
    }
  }

  /**
   * Get Lazada API client
   */
  getLazadaClient(): LazadaAPIClient | null {
    try {
      return this.getAPIClient("lazada");
    } catch {
      return null;
    }
  }

  /**
   * Get TikTok API client
   */
  getTiktokClient(): TiktokAPIClient | null {
    try {
      return this.getAPIClient("tiktok");
    } catch {
      return null;
    }
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
const tenantInstances: Map<string, TokenManager> = new Map();
const initializingPromises: Map<string, Promise<void>> = new Map();

/**
 * Get TokenManager for a specific tenant
 * Automatically initializes the manager if not already initialized
 * @param tenantId - Optional tenant ID (defaults to current context)
 */
export async function getTokenManager(
  tenantId?: string
): Promise<TokenManager> {
  // Resolve tenant ID
  const resolvedTenantId =
    tenantId ?? tenantContext.getTenantIdOrNull() ?? undefined;

  if (!resolvedTenantId) {
    throw new Error("[TokenManager] No tenant ID provided and none in context");
  }

  // Get or create instance for this tenant
  let instance = tenantInstances.get(resolvedTenantId);
  if (!instance) {
    instance = new TokenManager(resolvedTenantId);
    tenantInstances.set(resolvedTenantId, instance);
    logger.info(
      `Created TokenManager instance for tenant: ${resolvedTenantId}`
    );
  }

  // Check if already initialized
  if (!instance.isInitialized()) {
    // Check if initialization is already in progress
    let initPromise = initializingPromises.get(resolvedTenantId);
    if (!initPromise) {
      initPromise = instance.initialize();
      initializingPromises.set(resolvedTenantId, initPromise);
    }
    await initPromise;
    initializingPromises.delete(resolvedTenantId);
  }

  return instance;
}

/**
 * Clear cached instance for a tenant (for testing)
 */
export function clearTokenManager(tenantId: string): void {
  tenantInstances.delete(tenantId);
}

/**
 * Clear all cached instances (for testing)
 */
export function clearAllTokenManagers(): void {
  tenantInstances.clear();
}

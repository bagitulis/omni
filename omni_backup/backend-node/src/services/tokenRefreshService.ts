import {
  getTokenInitializationService,
  TokenInitializationService,
} from "./tokenInitializationService";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";

type PlatformType = "shopee" | "lazada" | "tiktok";

const logger = getLogger("TokenRefresh");

/**
 * Token Refresh Service
 * Handles token refresh, validation, and expiry management
 *
 * MULTI-TENANT: Each tenant has its own instance
 *
 * Single Responsibility: Runtime token management
 * - Check token expiry
 * - Refresh expired tokens
 * - Monitor token status
 */
export class TokenRefreshService {
  private initService: TokenInitializationService;
  private tenantId: string;

  /**
   * @param tenantId - Tenant ID for this service instance
   */
  constructor(tenantId: string) {
    this.tenantId = tenantId;
    // Get the per-tenant TokenInitializationService
    this.initService = getTokenInitializationService(tenantId);
  }

  /**
   * Check if platform token is expired
   */
  isTokenExpired(platform: PlatformType): boolean {
    try {
      const configManager = this.initService.getConfigManager(platform);

      if (!configManager.tokenExpiry) {
        return true;
      }

      const currentTime = Date.now();
      const isExpired = currentTime >= configManager.tokenExpiry;

      return isExpired;
    } catch (error) {
      logger.error(`Error checking ${platform} token expiry: ${error}`);
      return true; // Treat as expired if we can't check
    }
  }

  /**
   * Refresh platform token
   */
  async refreshToken(platform: PlatformType): Promise<boolean> {
    try {
      const apiClient = this.initService.getAPIClient(platform);
      const success = await apiClient.refreshAccessToken();

      if (success) {
        logger.info(`${platform} token refreshed`);
      } else {
        logger.error(`${platform} token refresh failed`);
      }

      return success;
    } catch (error) {
      logger.error(`Error refreshing ${platform} token: ${error}`);
      return false;
    }
  }

  /**
   * Check and refresh token if expired
   */
  async ensureTokenValid(platform: PlatformType): Promise<boolean> {
    try {
      if (this.isTokenExpired(platform)) {
        return await this.refreshToken(platform);
      }
      return true;
    } catch (error) {
      logger.error(`Error ensuring ${platform} token validity: ${error}`);
      return false;
    }
  }

  /**
   * Get token status for a platform
   */
  async getTokenStatus(platform: PlatformType): Promise<{
    platform: PlatformType;
    isExpired: boolean;
    expiresAt?: string;
    refreshTokenExpiresAt?: string;
  }> {
    try {
      const configManager = this.initService.getConfigManager(platform);

      // Reload config from database to get latest values
      await configManager.loadConfig();

      return {
        platform,
        isExpired: this.isTokenExpired(platform),
        expiresAt: configManager.tokenExpiry
          ? new Date(configManager.tokenExpiry).toISOString()
          : undefined,
        refreshTokenExpiresAt: configManager.refreshTokenExpiry
          ? new Date(configManager.refreshTokenExpiry).toISOString()
          : undefined,
      };
    } catch (error) {
      logger.error(`Error getting token status for ${platform}: ${error}`);
      return {
        platform,
        isExpired: true,
      };
    }
  }

  /**
   * Get token status for all platforms
   */
  async getAllTokenStatus(): Promise<
    Array<{
      platform: PlatformType;
      isExpired: boolean;
      expiresAt?: string;
      refreshTokenExpiresAt?: string;
    }>
  > {
    const platforms: PlatformType[] = ["shopee", "lazada", "tiktok"];
    const results = [];
    for (const platform of platforms) {
      results.push(await this.getTokenStatus(platform));
    }
    return results;
  }

  /**
   * Refresh all tokens
   * @param force - Force refresh even if not expired
   */
  async refreshAllTokens(force = false): Promise<Map<PlatformType, boolean>> {
    const results = new Map<PlatformType, boolean>();
    const platforms: PlatformType[] = ["shopee", "lazada", "tiktok"];

    for (const platform of platforms) {
      try {
        if (force) {
          // Force refresh regardless of expiry
          const success = await this.refreshToken(platform);
          results.set(platform, success);
        } else {
          // Only refresh if expired
          const success = await this.ensureTokenValid(platform);
          results.set(platform, success);
        }
      } catch (error) {
        logger.error(`Error refreshing ${platform} token: ${error}`);
        results.set(platform, false);
      }
    }

    return results;
  }

  /**
   * Refresh all expired tokens (legacy method)
   */
  async refreshAllExpiredTokens(): Promise<Map<PlatformType, boolean>> {
    return this.refreshAllTokens(false);
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
const tenantInstances: Map<string, TokenRefreshService> = new Map();

/**
 * Get TokenRefreshService for a specific tenant
 * @param tenantId - Optional tenant ID (defaults to current context)
 */
export function getTokenRefreshService(tenantId?: string): TokenRefreshService {
  // Resolve tenant ID
  const resolvedTenantId =
    tenantId ?? tenantContext.getTenantIdOrNull() ?? undefined;

  if (!resolvedTenantId) {
    throw new Error(
      "[TokenRefreshService] No tenant ID provided and none in context"
    );
  }

  // Get or create instance for this tenant
  let instance = tenantInstances.get(resolvedTenantId);
  if (!instance) {
    instance = new TokenRefreshService(resolvedTenantId);
    tenantInstances.set(resolvedTenantId, instance);
    logger.info(
      `Created TokenRefreshService instance for tenant: ${resolvedTenantId}`
    );
  }

  return instance;
}

/**
 * Clear cached instance for a tenant (for testing)
 */
export function clearTokenRefreshService(tenantId: string): void {
  tenantInstances.delete(tenantId);
}

/**
 * Clear all cached instances (for testing)
 */
export function clearAllTokenRefreshServices(): void {
  tenantInstances.clear();
}

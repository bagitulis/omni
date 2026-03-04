import crypto from "crypto";
import { getPrisma } from "./prismaClient";
import { getDbManager } from "./dbConnectionManager";
import { getLogger } from "../utils/logger";
import { getTokenManager } from "./tokenManager";
import { getTokenInitializationService } from "./tokenInitializationService";
import { platformAuthTokenSaver } from "./platformAuthTokenSaver";
import {
  generateShopeeAuthUrl,
  generateTiktokAuthUrl,
  generateLazadaAuthUrl,
} from "./oauth/oauthUrlGenerator";
import {
  exchangeShopeeCode,
  exchangeTiktokCode,
  exchangeLazadaCode,
  TokenExchangeResult,
} from "./oauth/tokenExchange";

const logger = getLogger("PlatformAuthService");

type PlatformType = "shopee" | "tiktok" | "lazada";

interface PlatformStatus {
  platform: string;
  connected: boolean;
  isExpired?: boolean;
  shopId?: string;
  expiresAt?: Date | string;
  lastActivity?: Date;
}

/**
 * Platform OAuth Service
 * SRP: Handle OAuth authorization flow for marketplace platforms
 * Token saving delegated to PlatformAuthTokenSaver
 */
export class PlatformAuthService {
  async generateAuthUrl(
    platform: PlatformType,
    tenantId: string,
    redirectUrl: string
  ): Promise<string> {
    const prisma = getPrisma(tenantId);
    const state = crypto.randomBytes(32).toString("hex");

    await prisma.oAuthState.create({
      data: {
        tenantId,
        platform,
        state,
        redirectUrl,
        expiresAt: new Date(Date.now() + 10 * 60 * 1000),
      },
    });

    const baseUrl = process.env.APP_BASE_URL || "https://yndigital.my.id";
    const callbackUrl = `${baseUrl}/api/platform-auth/${platform}/callback`;

    switch (platform) {
      case "shopee":
        return await generateShopeeAuthUrl(state, callbackUrl);
      case "tiktok":
        return await generateTiktokAuthUrl(state, callbackUrl);
      case "lazada":
        return await generateLazadaAuthUrl(state, callbackUrl);
      default:
        throw new Error(`Unsupported platform: ${platform}`);
    }
  }

  async exchangeCodeForTokens(
    platform: PlatformType,
    code: string,
    state: string | null,
    shopId?: string
  ): Promise<TokenExchangeResult> {
    let tenantId: string | null = null;

    if (state) {
      // Search for OAuth state across all tenants
      const foundState = await this.findOAuthStateAcrossTenants(state);
      if (!foundState) {
        logger.error(`OAuth state not found: ${state?.substring(0, 10)}...`);
        return {
          success: false,
          error:
            "SECURITY_ERROR: OAuth state not found. Please restart the authorization process.",
        };
      }
      const { prisma, oauthState } = foundState;

      if (new Date() > oauthState.expiresAt) {
        await prisma.oAuthState.delete({ where: { state } });
        await this.logOAuthEvent(
          oauthState.tenantId,
          platform,
          "callback_received",
          { shopId, error: "State expired" },
          "failed"
        );
        return {
          success: false,
          error:
            "SECURITY_ERROR: OAuth state expired. Please restart the authorization process.",
        };
      }
      tenantId = oauthState.tenantId;
      await prisma.oAuthState.delete({ where: { state } });
    } else if (shopId) {
      // Try to find tenant by shopId
      const foundTenant = await this.findTenantByShopId(platform, shopId);
      if (foundTenant) {
        tenantId = foundTenant;
        logger.info(`Found tenant ${tenantId} for shop_id ${shopId}`);
      }
    }

    // SECURITY: Reject if no tenant could be determined
    if (!tenantId) {
      logger.error(
        `SECURITY: No tenant found for OAuth callback. Platform: ${platform}, shopId: ${shopId}`
      );
      return {
        success: false,
        error:
          "SECURITY_ERROR: Unable to determine tenant. Please use proper OAuth flow with state parameter.",
      };
    }

    await this.logOAuthEvent(tenantId, platform, "callback_received", {
      shopId,
      code: code?.substring(0, 10) + "...",
      directCallback: !state,
    });

    try {
      let result: TokenExchangeResult;
      switch (platform) {
        case "shopee":
          result = await exchangeShopeeCode(code, shopId || "");
          break;
        case "tiktok":
          result = await exchangeTiktokCode(code);
          break;
        case "lazada":
          result = await exchangeLazadaCode(code);
          break;
        default:
          result = { success: false, error: "Unsupported platform" };
      }

      if (result.success && result.accessToken && result.refreshToken) {
        await platformAuthTokenSaver.saveTokens(
          platform,
          result,
          shopId,
          tenantId
        );
        await this.logOAuthEvent(tenantId, platform, "token_exchanged", {
          shopId,
          expiresAt: result.expiresAt,
          tokenSaved: true,
        });
        logger.info(`OAuth tokens saved for ${platform}`);
      } else {
        await this.logOAuthEvent(
          tenantId,
          platform,
          "token_exchanged",
          { shopId, error: result.error },
          "failed"
        );
      }

      return result;
    } catch (error: any) {
      logger.error(`Token exchange failed: ${error.message}`);
      await this.logOAuthEvent(
        tenantId,
        platform,
        "token_exchanged",
        { shopId, error: error.message },
        "failed"
      );
      return { success: false, error: error.message };
    }
  }

  private async findTenantByShopId(
    platform: PlatformType,
    shopId: string
  ): Promise<string | null> {
    const dbManager = getDbManager();
    const allTenants = dbManager.getAllTenants();
    const shopTenants = allTenants.filter((t) => {
      const config = dbManager.getTenantConfig(t);
      return !(config as any).isGlobal;
    });

    for (const tenant of shopTenants) {
      try {
        const prisma = getPrisma(tenant);
        const config = await prisma.platformConfig.findFirst({
          where: { platform, configKey: "shopId", configValue: shopId },
        });
        if (config) return tenant;
      } catch (error) {
        logger.debug(`Error checking tenant ${tenant}: ${error}`);
      }
    }
    return null;
  }

  /**
   * SECURITY: Search for OAuth state across all tenants
   * This prevents the need for hardcoded tenant fallbacks
   */
  private async findOAuthStateAcrossTenants(state: string): Promise<{
    prisma: ReturnType<typeof getPrisma>;
    oauthState: { tenantId: string; expiresAt: Date };
  } | null> {
    const dbManager = getDbManager();
    const allTenants = dbManager.getAllTenants();
    const shopTenants = allTenants.filter((t) => {
      const config = dbManager.getTenantConfig(t);
      return !(config as any).isGlobal;
    });

    for (const tenant of shopTenants) {
      try {
        const prisma = getPrisma(tenant);
        const oauthState = await prisma.oAuthState.findUnique({
          where: { state },
        });
        if (oauthState) {
          return { prisma, oauthState };
        }
      } catch (error) {
        logger.debug(
          `Error checking OAuth state in tenant ${tenant}: ${error}`
        );
      }
    }
    return null;
  }

  async getAllConnectionStatus(tenantId: string): Promise<PlatformStatus[]> {
    const tokenManager = await getTokenManager(tenantId);
    const initService = getTokenInitializationService();
    const platforms: PlatformType[] = ["shopee", "tiktok", "lazada"];
    const statuses: PlatformStatus[] = [];

    for (const platform of platforms) {
      try {
        const configManager = initService.getConfigManager(platform);
        await configManager.loadConfig();
        const hasToken = Boolean(configManager.accessToken);
        const status = await tokenManager.getTokenStatus(platform);
        statuses.push({
          platform,
          connected: hasToken,
          isExpired: status.isExpired,
          expiresAt: status.expiresAt,
        });
      } catch {
        statuses.push({ platform, connected: false, isExpired: true });
      }
    }
    return statuses;
  }

  async getPlatformStatus(
    platform: PlatformType,
    tenantId: string
  ): Promise<PlatformStatus> {
    const tokenManager = await getTokenManager(tenantId);
    const initService = getTokenInitializationService();

    try {
      const configManager = initService.getConfigManager(platform);
      await configManager.loadConfig();
      const hasToken = Boolean(configManager.accessToken);
      const status = await tokenManager.getTokenStatus(platform);
      return {
        platform,
        connected: hasToken,
        isExpired: status.isExpired,
        expiresAt: status.expiresAt,
      };
    } catch {
      return { platform, connected: false, isExpired: true };
    }
  }

  async disconnectPlatform(
    platform: PlatformType,
    tenantId: string
  ): Promise<void> {
    logger.info(`Disconnecting ${platform} for tenant ${tenantId}`);
  }

  private async logOAuthEvent(
    tenantId: string,
    platform: PlatformType,
    eventType:
      | "callback_received"
      | "token_exchanged"
      | "token_refreshed"
      | "error",
    metadata?: Record<string, any>,
    status: "received" | "success" | "failed" = "success"
  ): Promise<void> {
    try {
      const prisma = getPrisma(tenantId);
      await prisma.oAuthLog.create({
        data: {
          tenantId,
          platform,
          eventType,
          status,
          metadata: metadata ? JSON.stringify(metadata) : null,
          processedAt: new Date(),
        },
      });
    } catch (error: any) {
      logger.error(`Failed to log OAuth event: ${error.message}`);
    }
  }
}

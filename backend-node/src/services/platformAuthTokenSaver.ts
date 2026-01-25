/**
 * Platform Auth Token Saver
 * SRP: Save tokens to platform config managers
 */

import { getLogger } from "../utils/logger";
import { ShopeeConfigManager } from "../config/managers/shopeeConfigManager";
import { TiktokConfigManager } from "../config/managers/tiktokConfigManager";
import { LazadaConfigManager } from "../config/managers/lazadaConfigManager";
import { getTokenManager } from "./tokenManager";
import { TokenExchangeResult } from "./oauth/tokenExchange";

const logger = getLogger("PlatformAuthTokenSaver");

type PlatformType = "shopee" | "tiktok" | "lazada";

export class PlatformAuthTokenSaver {
  /**
   * Save tokens to platform config manager
   * @param platform - Platform type
   * @param result - Token exchange result
   * @param shopId - Shop ID (optional, for Shopee)
   * @param tenantId - Tenant ID (REQUIRED for multi-tenant security)
   */
  async saveTokens(
    platform: PlatformType,
    result: TokenExchangeResult,
    shopId?: string,
    tenantId?: string
  ): Promise<void> {
    if (!result.accessToken || !result.refreshToken || !result.expiresAt) {
      throw new Error("Missing token data");
    }

    if (!tenantId) {
      throw new Error(
        "Missing required tenantId - cannot save tokens without tenant context"
      );
    }

    const expiresIn = Math.floor(
      (result.expiresAt.getTime() - Date.now()) / 1000
    );

    logger.info(`[${tenantId}] Saving ${platform} tokens`);

    switch (platform) {
      case "shopee":
        await this.saveShopeeTokens(result, expiresIn, shopId, tenantId);
        break;
      case "tiktok":
        await this.saveTiktokTokens(result, expiresIn, tenantId);
        break;
      case "lazada":
        await this.saveLazadaTokens(result, expiresIn, tenantId);
        break;
    }

    await this.reinitializeTokenManager(tenantId);
  }

  private async saveShopeeTokens(
    result: TokenExchangeResult,
    expiresIn: number,
    shopId?: string,
    tenantId?: string
  ): Promise<void> {
    const shopeeConfig = new ShopeeConfigManager(tenantId);
    await shopeeConfig.loadConfig();
    if (shopId) {
      shopeeConfig.shopId = parseInt(shopId);
    }
    await shopeeConfig.updateTokenInfo(
      result.accessToken!,
      result.refreshToken!,
      expiresIn
    );
    await shopeeConfig.saveConfig();
    logger.info(`[${tenantId}] Shopee tokens saved successfully`);
  }

  private async saveTiktokTokens(
    result: TokenExchangeResult,
    expiresIn: number,
    tenantId?: string
  ): Promise<void> {
    const tiktokConfig = new TiktokConfigManager(tenantId);
    await tiktokConfig.loadConfig();
    await tiktokConfig.updateTokenInfo(
      result.accessToken!,
      result.refreshToken!,
      expiresIn
    );
    await tiktokConfig.saveConfig();
    logger.info(`[${tenantId}] TikTok tokens saved successfully`);
  }

  private async saveLazadaTokens(
    result: TokenExchangeResult,
    expiresIn: number,
    tenantId?: string
  ): Promise<void> {
    const lazadaConfig = new LazadaConfigManager(tenantId);
    await lazadaConfig.loadConfig();
    await lazadaConfig.updateTokenInfo(
      result.accessToken!,
      result.refreshToken!,
      expiresIn
    );
    await lazadaConfig.saveConfig();
    logger.info(`[${tenantId}] Lazada tokens saved successfully`);
  }

  private async reinitializeTokenManager(tenantId: string): Promise<void> {
    await getTokenManager(tenantId);
  }
}

export const platformAuthTokenSaver = new PlatformAuthTokenSaver();

import { BaseConfigManager } from "./baseConfigManager";
import { getGlobalConfigService } from "../../services/globalConfigService";
// import { maskSensitiveValue } from "../../utils/sanitizeLog";

/**
 * TikTok Configuration Manager
 * Handles TikTok-specific credentials and token management
 */
export class TiktokConfigManager extends BaseConfigManager {
  appKey: string = "";
  appSecret: string = "";
  openId: string = "";
  shopId: string = "";
  shopCipher: string = "";

  /**
   * @param tenantId - Optional explicit tenant ID (for jobs without request context)
   */
  constructor(tenantId?: string) {
    super("tiktok", tenantId);
  }

  /**
   * Load TikTok config from database
   */
  async loadConfig(): Promise<void> {
    await super.loadConfig();

    // 1. Get GLOBAL credentials from system.db
    const globalConfig = getGlobalConfigService();
    const { appKey, appSecret } = await globalConfig.getTiktokCredentials();

    // 2. Get TENANT specific config
    const openIdStr = await this.getConfig("openId");
    const shopIdStr = await this.getConfig("shopId");
    const shopCipherStr = await this.getConfig("shopCipher");

    // 3. Assign
    this.appKey = appKey || "";
    this.appSecret = appSecret || "";
    this.openId = openIdStr || "";
    this.shopId = shopIdStr || "";
    this.shopCipher = shopCipherStr || "";

    if (this.appKey && this.appSecret) {
      // Credentials validated
    } else {
      this.logger.warn("⚠️  TikTok Global Credentials missing in system.db");
    }
  }

  /**
   * Override setConfig to intercept Global Keys
   * Redirects appKey/appSecret to GlobalConfigService (system.db)
   */
  async setConfig(key: string, value: string): Promise<void> {
    if (key === "appKey" || key === "appSecret") {
      const globalConfig = getGlobalConfigService();

      const success = await globalConfig.setConfig("tiktok", key, value, true);

      if (!success) {
        throw new Error(`Failed to save global config: ${key}`);
      }

      if (key === "appKey") this.appKey = value;
      if (key === "appSecret") this.appSecret = value;

      this.logger.info(`✅ Updated Global Config (system.db): ${key}`);
      return;
    }

    await super.setConfig(key, value);
  }

  /**
   * Save TikTok config
   * NOTE: Only saves TENANT specific config.
   */
  async saveConfig(): Promise<void> {
    // Update in-memory config before saving
    if (this.openId) {
      this.config.set("openId", this.openId);
    }
    if (this.shopId) {
      this.config.set("shopId", this.shopId);
    }
    if (this.shopCipher) {
      this.config.set("shopCipher", this.shopCipher);
    }
    // Do NOT save appKey/appSecret to tenant DB

    await super.saveConfig();
  }

  /**
   * Update token info and save
   */
  async updateTokenInfo(
    accessToken: string,
    refreshToken: string,
    expiresIn: number
  ): Promise<void> {
    await super.updateTokenInfo(accessToken, refreshToken, expiresIn);
  }
}

import { BaseConfigManager } from "./baseConfigManager";
import { getGlobalConfigService } from "../../services/globalConfigService";

/**
 * Shopee Configuration Manager
 * Handles Shopee-specific credentials and token management
 */
export class ShopeeConfigManager extends BaseConfigManager {
  partnerId: number = 0;
  partnerKey: string = "";
  shopId: number = 0;

  /**
   * @param tenantId - Optional explicit tenant ID (for jobs without request context)
   */
  constructor(tenantId?: string) {
    super("shopee", tenantId);
  }

  /**
   * Load Shopee config from database
   */
  async loadConfig(): Promise<void> {
    await super.loadConfig();

    // 1. Get GLOBAL credentials from system.db (Partner ID, Key)
    const globalConfig = getGlobalConfigService();
    const { partnerId, partnerKey } = await globalConfig.getShopeeCredentials();

    // 2. Get TENANT specific config (Shop ID, Tokens) from loaded config
    const shopIdStr = await this.getConfig("shopId");

    // 3. Assign to properties
    this.partnerId = parseInt(partnerId, 10) || 0;
    this.partnerKey = partnerKey || "";
    this.shopId = parseInt(shopIdStr, 10) || 0;

    if (this.partnerId && this.partnerKey && this.shopId) {
      // Credentials validated
    } else {
      if (!this.partnerId || !this.partnerKey) {
        this.logger.warn(
          "⚠️  Shopee Global Credentials (partnerId/Key) missing in system.db",
        );
      }
      if (!this.shopId) {
        this.logger.warn(
          "⚠️  Shopee Tenant Config (shopId) missing in tenant db",
        );
      }
    }
  }

  /**
   * Override setConfig to intercept Global Keys
   * Redirects partnerId/partnerKey to GlobalConfigService (system.db)
   */
  async setConfig(key: string, value: string): Promise<void> {
    // Intercept Global Keys
    if (key === "partnerId" || key === "partnerKey") {
      const globalConfig = getGlobalConfigService();

      // Save to system.db (Encrypted)
      const success = await globalConfig.setConfig("shopee", key, value, true);

      if (!success) {
        throw new Error(`Failed to save global config: ${key}`);
      }

      // Update in-memory state immediately
      if (key === "partnerId") this.partnerId = parseInt(value, 10) || 0;
      if (key === "partnerKey") this.partnerKey = value;

      this.logger.info(`✅ Updated Global Config (system.db): ${key}`);
      return;
    }

    // Default: Save other keys to Tenant DB
    await super.setConfig(key, value);
  }

  /**
   * Save Shopee config
   * NOTE: Only saves TENANT specific config (ShopId, Tokens).
   * Global config (PartnerId/Key) is managed via GlobalConfigService elsewhere.
   */
  async saveConfig(): Promise<void> {
    // Update in-memory config before saving
    if (this.shopId) {
      this.config.set("shopId", String(this.shopId));
    }
    // Do NOT save partnerId/partnerKey to tenant DB

    await super.saveConfig();
  }

  /**
   * Update token info and save
   * @param refreshExpiresIn - Optional refresh token expiry in seconds (Shopee: 30 days)
   */
  async updateTokenInfo(
    accessToken: string,
    refreshToken: string,
    expiresIn: number,
    refreshExpiresIn?: number,
  ): Promise<void> {
    await super.updateTokenInfo(
      accessToken,
      refreshToken,
      expiresIn,
      refreshExpiresIn,
    );
  }
}

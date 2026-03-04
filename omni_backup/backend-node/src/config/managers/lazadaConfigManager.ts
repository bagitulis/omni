import { BaseConfigManager } from "./baseConfigManager";
import { getGlobalConfigService } from "../../services/globalConfigService";

/**
 * Lazada Configuration Manager
 * Handles Lazada-specific credentials and token management
 */
export class LazadaConfigManager extends BaseConfigManager {
  appKey: string = "";
  appSecret: string = "";
  country: string = "id"; // Default to Indonesia

  /**
   * @param tenantId - Optional explicit tenant ID (for jobs without request context)
   */
  constructor(tenantId?: string) {
    super("lazada", tenantId);
  }

  /**
   * Load Lazada config from database
   */
  async loadConfig(): Promise<void> {
    await super.loadConfig();

    // 1. Get GLOBAL credentials from system.db (App Key, Secret)
    const globalConfig = getGlobalConfigService();
    const { appKey, appSecret } = await globalConfig.getLazadaCredentials();

    // 2. Get TENANT specific config (Country, Tokens)
    const countryStr = await this.getConfig("country");

    // 3. Assign
    this.appKey = appKey || "";
    this.appSecret = appSecret || "";
    this.country = (countryStr || "id").toLowerCase();

    if (this.appKey && this.appSecret) {
      // Credentials validated
    } else {
      this.logger.warn("⚠️  Lazada Global Credentials missing in system.db");
    }
  }

  /**
   * Override setConfig to intercept Global Keys
   * Redirects appKey/appSecret to GlobalConfigService (system.db)
   */
  async setConfig(key: string, value: string): Promise<void> {
    if (key === "appKey" || key === "appSecret") {
      const globalConfig = getGlobalConfigService();

      // Save directly to system.db
      const success = await globalConfig.setConfig("lazada", key, value, true);

      if (!success) {
        throw new Error(`Failed to save global config: ${key}`);
      }

      // Update in-memory state
      if (key === "appKey") this.appKey = value;
      if (key === "appSecret") this.appSecret = value;

      this.logger.info(`✅ Updated Global Config (system.db): ${key}`);
      return;
    }

    await super.setConfig(key, value);
  }

  /**
   * Save Lazada config
   * NOTE: Only saves TENANT specific config.
   */
  async saveConfig(): Promise<void> {
    // Update in-memory config before saving
    if (this.country) {
      this.config.set("country", this.country);
    }
    // Do NOT save appKey/appSecret to tenant DB

    await super.saveConfig();
  }

  /**
   * Update token info and save
   * @param refreshExpiresIn - Optional refresh token expiry in seconds (Lazada: 30 days)
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

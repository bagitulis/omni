import { getTenantContext } from "../../utils/tenantContext";
import { EncryptionManager } from "../../utils/encryption";

/**
 * Base Configuration Manager
 * Mirrors Python backend's IConfigManager & BaseConfigManager
 * Handles loading/saving config from PlatformConfig table
 * Multi-tenant aware: reads/writes to tenant-specific database
 *
 * Supports two modes:
 * 1. Explicit tenantId (via constructor) - for job handlers, batch operations
 * 2. Context-based (via getTenantContext) - for request handlers
 */
export interface IConfigManager {
  platform: string;
  accessToken: string;
  refreshToken: string;
  tokenExpiry: number;
  refreshTokenExpiry: number;
  loadConfig(): Promise<void>;
  saveConfig(): Promise<void>;
  getConfig(key: string): Promise<string>;
  setConfig(key: string, value: string): Promise<void>;
  isTokenExpired(): boolean;
  updateTokenInfo(
    accessToken: string,
    refreshToken: string,
    expiresIn: number,
    refreshExpiresIn?: number
  ): Promise<void>;
}

export abstract class BaseConfigManager implements IConfigManager {
  platform: string;
  protected config: Map<string, string> = new Map();
  protected logger: any;
  protected explicitTenantId?: string; // Optional explicit tenant for job contexts

  // Common token fields
  accessToken: string = "";
  refreshToken: string = "";
  tokenExpiry: number = 0; // milliseconds
  refreshTokenExpiry: number = 0; // milliseconds

  /**
   * @param platform - Platform name (shopee, lazada, tiktok)
   * @param tenantId - Optional explicit tenant ID (for jobs without request context)
   */
  constructor(platform: string, tenantId?: string) {
    this.platform = platform;
    this.explicitTenantId = tenantId;
    this.logger = {
      info: (msg: string) => console.log(`[${this.platform}] ℹ️  ${msg}`),
      debug: (_msg: string) => {}, // Suppress debug logs
      error: (msg: string) => console.error(`[${this.platform}] ❌ ${msg}`),
      warn: (msg: string) => console.warn(`[${this.platform}] ⚠️  ${msg}`),
    };
  }

  /**
   * Get Prisma client for current tenant
   * Uses explicit tenantId if set, otherwise falls back to context
   */
  protected getPrismaClient(): import("@prisma/client").PrismaClient {
    if (this.explicitTenantId) {
      // Use explicit tenant ID (for job handlers)
      const { getDbManager } = require("../../services/dbConnectionManager");
      return getDbManager().getConnection(this.explicitTenantId);
    }

    // Fall back to request context
    const { prisma } = getTenantContext();
    return prisma;
  }

  /**
   * Load all config from database
   */
  async loadConfig(): Promise<void> {
    try {
      // Get tenant-specific database connection
      const prisma = this.getPrismaClient();

      // Fetch all config for this platform
      const configs = await prisma.platformConfig.findMany({
        where: { platform: this.platform },
      });

      // Clear and rebuild map
      this.config.clear();

      for (const cfg of configs) {
        let value = cfg.configValue;

        // Decrypt if needed
        if (cfg.isEncrypted) {
          try {
            value = EncryptionManager.decrypt(value);
            this.logger.debug(
              `✅ Decrypted ${cfg.configKey}: ${value.substring(0, 30)}...`
            );
          } catch (err) {
            this.logger.warn(
              `⚠️ Failed to decrypt ${cfg.configKey}, using plain value: ${err}`
            );
            // Continue with plain value if decryption fails
          }
        }

        this.config.set(cfg.configKey, value);
      }

      // Load common token fields (camelCase to match database keys)
      const accessTokenStr = this.config.get("accessToken") || "";
      const refreshTokenStr = this.config.get("refreshToken") || "";
      const tokenExpiryStr = this.config.get("tokenExpiry") || "0";
      const refreshTokenExpiryStr =
        this.config.get("refreshTokenExpiry") || "0";

      this.accessToken = accessTokenStr;
      this.refreshToken = refreshTokenStr;

      // Parse tokenExpiry - could be ISO string or millisecond timestamp
      if (tokenExpiryStr && !isNaN(Number(tokenExpiryStr))) {
        // Numeric timestamp
        this.tokenExpiry = parseInt(tokenExpiryStr, 10) || 0;
      } else if (tokenExpiryStr) {
        // ISO string format - convert to milliseconds
        this.tokenExpiry = new Date(tokenExpiryStr).getTime() || 0;
      } else {
        this.tokenExpiry = 0;
      }

      // Parse refreshTokenExpiry
      if (refreshTokenExpiryStr && !isNaN(Number(refreshTokenExpiryStr))) {
        // Numeric timestamp
        this.refreshTokenExpiry = parseInt(refreshTokenExpiryStr, 10) || 0;
      } else if (refreshTokenExpiryStr) {
        // ISO string format - convert to milliseconds
        this.refreshTokenExpiry =
          new Date(refreshTokenExpiryStr).getTime() || 0;
      } else {
        this.refreshTokenExpiry = 0;
      }
    } catch (error) {
      this.logger.error(`Failed to load config: ${error}`);
      throw error;
    }
  }

  /**
   * Get a config value
   */
  async getConfig(key: string): Promise<string> {
    return this.config.get(key) || "";
  }

  /**
   * Set a config value (in memory only, call saveConfig to persist)
   */
  async setConfig(key: string, value: string): Promise<void> {
    this.config.set(key, value);
  }

  /**
   * Save config to database
   */
  async saveConfig(): Promise<void> {
    try {
      // Get tenant-specific database connection
      const prisma = this.getPrismaClient();

      for (const [key, value] of this.config) {
        // Check if config already exists in database to get current isEncrypted flag
        const existing = await prisma.platformConfig.findUnique({
          where: {
            platform_configKey: {
              platform: this.platform,
              configKey: key,
            },
          },
        });

        // Use existing isEncrypted flag if available, otherwise default to specific fields
        const isEncrypted =
          existing?.isEncrypted ??
          [
            "accessToken",
            "refreshToken",
            "code",
            "appSecret",
            "partnerKey",
          ].includes(key);

        let storedValue = value;

        // Only encrypt if isEncrypted is true
        if (isEncrypted && value) {
          try {
            storedValue = EncryptionManager.encrypt(value);
          } catch (err) {
            this.logger.error(`Failed to encrypt ${key}: ${err}`);
            continue;
          }
        }

        // Upsert to database (preserve isEncrypted flag from database)
        await prisma.platformConfig.upsert({
          where: {
            platform_configKey: {
              platform: this.platform,
              configKey: key,
            },
          },
          update: {
            configValue: storedValue,
          },
          create: {
            platform: this.platform,
            configKey: key,
            configValue: storedValue,
            isEncrypted,
          },
        });
      }
    } catch (error) {
      this.logger.error(`Failed to save config: ${error}`);
      throw error;
    }
  }

  /**
   * Check if token is expired
   */
  isTokenExpired(): boolean {
    if (!this.tokenExpiry) return true;
    const now = Date.now();
    const isExpired = now >= this.tokenExpiry;

    if (isExpired) {
      const expiresIn = Math.floor((this.tokenExpiry - now) / 1000);
      this.logger.debug(`Token expired ${Math.abs(expiresIn)} seconds ago`);
    }

    return isExpired;
  }

  /**
   * Update token info and save to database
   */
  async updateTokenInfo(
    accessToken: string,
    refreshToken: string,
    expiresIn: number,
    refreshExpiresIn?: number
  ): Promise<void> {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;

    // expiresIn is in seconds, convert to milliseconds
    const expiryMs = Date.now() + expiresIn * 1000;
    this.tokenExpiry = expiryMs;

    // Update in-memory config (camelCase to match database keys)
    this.config.set("accessToken", accessToken);
    this.config.set("refreshToken", refreshToken);
    this.config.set("tokenExpiry", String(expiryMs));

    // Handle refresh token expiry if provided
    if (refreshExpiresIn) {
      const refreshExpiryMs = Date.now() + refreshExpiresIn * 1000;
      this.refreshTokenExpiry = refreshExpiryMs;
      this.config.set("refreshTokenExpiry", String(refreshExpiryMs));
    }

    // Save to database
    await this.saveConfig();
    this.logger.info(
      `✅ Token updated (expires in ${expiresIn}s at ${new Date(
        expiryMs
      ).toISOString()})`
    );
  }
}

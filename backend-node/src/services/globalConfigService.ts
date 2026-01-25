import Database from "better-sqlite3";
import path from "path";
import { EncryptionManager } from "../utils/encryption";
import { getLogger } from "../utils/logger";

const logger = getLogger("GlobalConfigService");

interface GlobalConfigRow {
  id: number;
  platform: string;
  configKey: string;
  configValue: string;
  isEncrypted: number;
  description: string | null;
}

/**
 * Service for accessing global configuration from system.db
 * Used for shared credentials like Partner ID/Key across all tenants
 */
class GlobalConfigService {
  private db: Database.Database | null = null;
  private dbPath: string;

  constructor() {
    this.dbPath = path.resolve(__dirname, "../../config/databases/system.db");
  }

  private getDb(): Database.Database {
    if (!this.db) {
      try {
        this.db = new Database(this.dbPath);
        // Use DELETE mode for better reliability in Docker environments
        this.db.pragma("journal_mode = DELETE");
        this.db.pragma("busy_timeout = 10000");
        this.db.pragma("synchronous = NORMAL");
      } catch (error) {
        logger.error(`Failed to open database: ${error}`);
        throw error;
      }
    }
    return this.db;
  }

  /**
   * Reconnect to database (useful after I/O errors)
   */
  private reconnect(): void {
    if (this.db) {
      try {
        this.db.close();
      } catch {
        // Ignore close errors
      }
      this.db = null;
    }
  }

  /**
   * Get a config value by platform and key
   * Returns decrypted value if encrypted
   * Has retry logic for I/O errors
   */
  async getConfig(platform: string, key: string): Promise<string> {
    const maxRetries = 2;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const db = this.getDb();
        const row = db
          .prepare(
            "SELECT configValue, isEncrypted FROM GlobalConfig WHERE platform = ? AND configKey = ?",
          )
          .get(platform, key) as GlobalConfigRow | undefined;

        if (!row || !row.configValue) {
          return "";
        }

        if (row.isEncrypted) {
          try {
            return EncryptionManager.decrypt(row.configValue);
          } catch {
            return row.configValue;
          }
        }

        return row.configValue;
      } catch (error: any) {
        const isIOError =
          error?.message?.includes("I/O") || error?.code === "SQLITE_IOERR";

        if (isIOError && attempt < maxRetries) {
          logger.warn(
            `I/O error getting ${platform}.${key}, retrying... (${
              attempt + 1
            }/${maxRetries})`,
          );
          this.reconnect();
          await new Promise((r) => setTimeout(r, 100)); // Small delay before retry
          continue;
        }

        logger.error(`Failed to get config ${platform}.${key}: ${error}`);
        return "";
      }
    }
    return "";
  }

  /**
   * Set a config value (encrypts if needed)
   */
  async setConfig(
    platform: string,
    key: string,
    value: string,
    encrypt: boolean = false,
  ): Promise<boolean> {
    try {
      const db = this.getDb();
      const finalValue = encrypt ? EncryptionManager.encrypt(value) : value;

      db.prepare(
        `INSERT INTO GlobalConfig (platform, configKey, configValue, isEncrypted, updatedAt)
         VALUES (?, ?, ?, ?, datetime('now'))
         ON CONFLICT(platform, configKey) 
         DO UPDATE SET configValue = ?, isEncrypted = ?, updatedAt = datetime('now')`,
      ).run(
        platform,
        key,
        finalValue,
        encrypt ? 1 : 0,
        finalValue,
        encrypt ? 1 : 0,
      );

      logger.info(`Updated global config: ${platform}.${key}`);
      return true;
    } catch (error) {
      logger.error(`Failed to set config ${platform}.${key}: ${error}`);
      return false;
    }
  }

  /**
   * Get Shopee credentials with env var fallback
   */
  async getShopeeCredentials(): Promise<{
    partnerId: string;
    partnerKey: string;
    pushPartnerKey: string;
  }> {
    const partnerId =
      (await this.getConfig("shopee", "partnerId")) ||
      process.env.SHOPEE_PARTNER_ID ||
      "";
    const partnerKey =
      (await this.getConfig("shopee", "partnerKey")) ||
      process.env.SHOPEE_PARTNER_KEY ||
      "";
    const pushPartnerKey =
      (await this.getConfig("shopee", "pushPartnerKey")) ||
      process.env.SHOPEE_PUSH_PARTNER_KEY ||
      "";

    return { partnerId, partnerKey, pushPartnerKey };
  }

  /**
   * Get TikTok credentials with env var fallback
   */
  async getTiktokCredentials(): Promise<{
    appKey: string;
    appSecret: string;
  }> {
    const appKey =
      (await this.getConfig("tiktok", "appKey")) ||
      process.env.TIKTOK_APP_KEY ||
      "";
    const appSecret =
      (await this.getConfig("tiktok", "appSecret")) ||
      process.env.TIKTOK_APP_SECRET ||
      "";

    return { appKey, appSecret };
  }

  /**
   * Get Lazada credentials with env var fallback
   */
  async getLazadaCredentials(): Promise<{
    appKey: string;
    appSecret: string;
  }> {
    const appKey =
      (await this.getConfig("lazada", "appKey")) ||
      process.env.LAZADA_APP_KEY ||
      "";
    const appSecret =
      (await this.getConfig("lazada", "appSecret")) ||
      process.env.LAZADA_APP_SECRET ||
      "";

    return { appKey, appSecret };
  }

  /**
   * Check if credentials are configured
   */
  async hasShopeeCredentials(): Promise<boolean> {
    const creds = await this.getShopeeCredentials();
    return Boolean(creds.partnerId && creds.partnerKey);
  }

  async hasTiktokCredentials(): Promise<boolean> {
    const creds = await this.getTiktokCredentials();
    return Boolean(creds.appKey && creds.appSecret);
  }

  async hasLazadaCredentials(): Promise<boolean> {
    const creds = await this.getLazadaCredentials();
    return Boolean(creds.appKey && creds.appSecret);
  }

  /**
   * Close database connection
   */
  close(): void {
    if (this.db) {
      this.db.close();
      this.db = null;
    }
  }
}

// Singleton instance
let instance: GlobalConfigService | null = null;

export function getGlobalConfigService(): GlobalConfigService {
  if (!instance) {
    instance = new GlobalConfigService();
  }
  return instance;
}

export { GlobalConfigService };

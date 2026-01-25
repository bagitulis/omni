/**
 * Service Account Rotation Manager
 * Manages multiple service accounts and auto-rotates based on quota/errors
 *
 * Single Responsibility: Service account rotation and lifecycle management
 */

import * as fs from "fs";
import * as path from "path";
import { google } from "googleapis";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { backendLogger } from "../utils/backendLogger";

export interface ServiceAccountConfig {
  keyPath: string;
  clientEmail: string;
  projectId: string;
  authClient?: any;
  usageCount: number;
  lastUsedAt: Date | null;
  isActive: boolean;
  quotaUsedPercent: number;
}

export class GoogleSheetsServiceAccountRotationManager {
  private accounts: Map<string, ServiceAccountConfig> = new Map();
  private currentAccountId: string | null = null;
  private logger: Logger;

  constructor() {
    this.logger = getLogger("ServiceAccountRotation");
  }

  /**
   * Load service accounts from configuration files
   */
  async loadServiceAccounts(keyPaths: string[]): Promise<boolean> {
    try {
      let loadedCount = 0;

      for (const keyPath of keyPaths) {
        const absolutePath = path.isAbsolute(keyPath)
          ? keyPath
          : path.join(process.cwd(), keyPath);

        if (!fs.existsSync(absolutePath)) {
          this.logger.warn(`Service account file not found: ${absolutePath}`);
          continue;
        }

        const keyContent = fs.readFileSync(absolutePath, "utf8");
        const keyData = JSON.parse(keyContent);

        const accountId = keyData.private_key_id;
        const config: ServiceAccountConfig = {
          keyPath: absolutePath,
          clientEmail: keyData.client_email,
          projectId: keyData.project_id,
          usageCount: 0,
          lastUsedAt: null,
          isActive: false,
          quotaUsedPercent: 0,
        };

        // Initialize OAuth2 client
        const auth = new google.auth.GoogleAuth({
          keyFile: absolutePath,
          scopes: [
            "https://www.googleapis.com/auth/spreadsheets",
            "https://www.googleapis.com/auth/drive",
          ],
        });

        config.authClient = await auth.getClient();
        this.accounts.set(accountId, config);
        loadedCount++;

        this.logger.info(`✅ Loaded service account: ${config.clientEmail}`);
      }

      if (loadedCount === 0) {
        this.logger.error("❌ No service accounts loaded");
        return false;
      }

      // Set first account as active
      if (this.accounts.size > 0) {
        const firstKey = this.accounts.keys().next().value as string;
        if (firstKey) {
          this.currentAccountId = firstKey;
          const firstAccount = this.accounts.get(firstKey);
          if (firstAccount) {
            firstAccount.isActive = true;
            this.logger.info(`🔄 Active account: ${firstKey}`);
          }
        }
      }

      return true;
    } catch (error) {
      this.logger.error(
        "Error loading service accounts",
        error instanceof Error ? error.message : String(error)
      );
      return false;
    }
  }

  /**
   * Get currently active service account
   */
  getActiveAccount(): ServiceAccountConfig | null {
    if (!this.currentAccountId) return null;
    return this.accounts.get(this.currentAccountId) || null;
  }

  /**
   * Get all service accounts
   */
  getAllAccounts(): ServiceAccountConfig[] {
    return Array.from(this.accounts.values());
  }

  /**
   * Rotate to next available service account
   */
  async rotateServiceAccount(
    reason: "limit" | "error" | "manual"
  ): Promise<boolean> {
    try {
      if (this.accounts.size < 2) {
        this.logger.warn("Cannot rotate: only 1 service account available");
        return false;
      }

      const currentId = this.currentAccountId;
      const currentAccount = this.accounts.get(currentId!);

      if (currentAccount) {
        currentAccount.isActive = false;
      }

      // Find next active account
      let nextId: string | null = null;
      for (const [id, account] of this.accounts) {
        if (id !== currentId && account.authClient) {
          nextId = id;
          break;
        }
      }

      if (!nextId) {
        this.logger.error("❌ No available service account to rotate to");
        return false;
      }

      this.currentAccountId = nextId;
      const nextAccount = this.accounts.get(nextId)!;
      nextAccount.isActive = true;
      nextAccount.lastUsedAt = new Date();
      nextAccount.usageCount++;

      backendLogger.info(
        "ServiceAccountRotation",
        `🔄 Rotated account (${reason}): ${nextAccount.clientEmail}`
      );

      return true;
    } catch (error) {
      this.logger.error(
        "Error rotating service account",
        error instanceof Error ? error.message : String(error)
      );
      return false;
    }
  }

  /**
   * Update usage quota for account
   */
  updateQuotaUsage(quotaUsedPercent: number): void {
    const activeAccount = this.getActiveAccount();
    if (activeAccount) {
      activeAccount.quotaUsedPercent = quotaUsedPercent;
    }
  }

  /**
   * Mark account usage
   */
  recordUsage(): void {
    const activeAccount = this.getActiveAccount();
    if (activeAccount) {
      activeAccount.usageCount++;
      activeAccount.lastUsedAt = new Date();
    }
  }

  /**
   * Get account statistics for monitoring
   */
  getAccountStats(): Record<string, any> {
    const stats: Record<string, any> = {};

    for (const [id, account] of this.accounts) {
      stats[id] = {
        email: account.clientEmail,
        usageCount: account.usageCount,
        lastUsedAt: account.lastUsedAt,
        isActive: account.isActive,
        quotaUsedPercent: account.quotaUsedPercent,
      };
    }

    return stats;
  }

  /**
   * Check if needs rotation based on quota
   */
  shouldRotate(quotaThreshold: number = 80): boolean {
    const activeAccount = this.getActiveAccount();
    if (!activeAccount) return false;
    return activeAccount.quotaUsedPercent >= quotaThreshold;
  }
}

// Singleton instance
let rotationManagerInstance: GoogleSheetsServiceAccountRotationManager | null =
  null;

export function getServiceAccountRotationManager(): GoogleSheetsServiceAccountRotationManager {
  if (!rotationManagerInstance) {
    rotationManagerInstance = new GoogleSheetsServiceAccountRotationManager();
  }
  return rotationManagerInstance;
}

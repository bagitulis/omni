import * as path from "path";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { backendLogger } from "../utils/backendLogger";
import {
  getServiceAccountRotationManager,
  GoogleSheetsServiceAccountRotationManager,
} from "./googleSheetsServiceAccountRotationManager";
import { getGoogleSheetsAutoDiscovery } from "./googleSheetsAutoDiscoveryService";
import { getQuotaManager } from "./quota";

/**
 * Google Sheets Auth Service (Refactored for Service Accounts)
 * Manages service account authentication and initialization
 * With auto-discovery of service account files
 *
 * Single Responsibility: Service account lifecycle management
 */

export interface AuthStatus {
  authenticated: boolean;
  settings: Record<string, string>;
  has_credentials: boolean;
}

class GoogleSheetsAuthService {
  private logger: Logger;
  private rotationManager: GoogleSheetsServiceAccountRotationManager;
  private autoDiscovery: any;
  private initializationPromise: Promise<boolean> | null = null;
  // Fallback paths for backward compatibility
  private fallbackServiceAccountPaths: string[] = [
    "config/static/google/bertigahemat-f1bd6932b229.json",
    "config/static/google/bertigamart-f887630c1a21.json",
  ];

  constructor() {
    this.logger = getLogger("GoogleSheetsAuth");
    this.rotationManager = getServiceAccountRotationManager();
    this.autoDiscovery = getGoogleSheetsAutoDiscovery();
    this.initializationPromise = this.initialize();
  }

  /**
   * Initialize service accounts with auto-discovery
   */
  private async initialize(): Promise<boolean> {
    try {
      // First, try auto-discovery
      let accountPaths = await this.autoDiscovery.autoDiscoverServiceAccounts();

      // If no accounts found, fallback to hardcoded paths
      if (accountPaths.length === 0) {
        this.logger.warn(
          "⚠️ No service accounts auto-discovered, using fallback paths"
        );
        accountPaths = this.fallbackServiceAccountPaths.map((p) =>
          path.isAbsolute(p) ? p : path.join(process.cwd(), p)
        );
      }

      const success =
        await this.rotationManager.loadServiceAccounts(accountPaths);

      if (success) {
        const accounts = this.rotationManager.getAllAccounts();
        backendLogger.success(
          "GoogleSheetsAuth",
          `✅ Service accounts initialized (${accounts.length} accounts loaded)`
        );

        // Initialize Quota Manager with discovered accounts
        const quotaManager = getQuotaManager();
        quotaManager.initialize();
        backendLogger.info(
          "GoogleSheetsAuth",
          `📊 Quota Manager initialized for ${accounts.length} account(s)`
        );

        // Watch for new service account files
        this.autoDiscovery.watchForNewAccounts(
          async (newAccounts: string[]) => {
            this.logger.info(`🆕 New service accounts detected, reloading...`);
            await this.rotationManager.loadServiceAccounts(newAccounts);
            // Re-initialize quota manager with new accounts
            quotaManager.reinitialize();
          }
        );

        return true;
      } else {
        backendLogger.error(
          "GoogleSheetsAuth",
          "❌ Failed to initialize service accounts"
        );
        return false;
      }
    } catch (error) {
      this.logger.error(
        "Initialization error",
        error instanceof Error ? error.message : String(error)
      );
      return false;
    }
  }

  /**
   * Wait for initialization to complete
   */
  async ensureInitialized(): Promise<boolean> {
    if (this.initializationPromise) {
      return await this.initializationPromise;
    }
    return false;
  }

  /**
   * Get current auth status
   */
  async getAuthStatus(): Promise<AuthStatus> {
    await this.ensureInitialized();

    const activeAccount = this.rotationManager.getActiveAccount();
    const allAccounts = this.rotationManager.getAllAccounts();

    return {
      authenticated: activeAccount !== null,
      settings: {
        active_account: activeAccount?.clientEmail || "none",
        account_count: allAccounts.length.toString(),
        service_account_mode: "enabled",
      },
      has_credentials: allAccounts.length > 0,
    };
  }

  /**
   * Get active service account auth client
   */
  getActiveAuthClient(): any {
    const activeAccount = this.rotationManager.getActiveAccount();
    return activeAccount?.authClient || null;
  }

  /**
   * Rotate to next service account
   */
  async rotateServiceAccount(
    reason: "limit" | "error" | "manual"
  ): Promise<boolean> {
    const success = await this.rotationManager.rotateServiceAccount(reason);
    if (success) {
      backendLogger.info(
        "GoogleSheetsAuth",
        `🔄 Service account rotated: ${reason}`
      );
    }
    return success;
  }

  /**
   * Check if needs rotation based on quota
   */
  shouldRotate(quotaThreshold?: number): boolean {
    return this.rotationManager.shouldRotate(quotaThreshold);
  }

  /**
   * Record usage in rotation manager
   */
  recordUsage(): void {
    this.rotationManager.recordUsage();
  }

  /**
   * Update quota usage
   */
  updateQuotaUsage(quotaUsedPercent: number): void {
    this.rotationManager.updateQuotaUsage(quotaUsedPercent);
  }

  /**
   * Get account statistics
   */
  getAccountStats(): Record<string, any> {
    return this.rotationManager.getAccountStats();
  }

  /**
   * Get rotation manager (for advanced usage)
   */
  getRotationManager(): GoogleSheetsServiceAccountRotationManager {
    return this.rotationManager;
  }

  /**
   * Check if authenticated
   */
  isAuthenticated(): boolean {
    const activeAccount = this.rotationManager.getActiveAccount();
    return activeAccount !== null && activeAccount.authClient !== undefined;
  }
}

// Singleton instance
let googleSheetsAuthServiceInstance: GoogleSheetsAuthService | null = null;

export function getGoogleSheetsAuthService(): GoogleSheetsAuthService {
  if (!googleSheetsAuthServiceInstance) {
    googleSheetsAuthServiceInstance = new GoogleSheetsAuthService();
  }
  return googleSheetsAuthServiceInstance;
}

export { GoogleSheetsAuthService };

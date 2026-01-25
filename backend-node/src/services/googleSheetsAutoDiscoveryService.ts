/**
 * Google Sheets Auto Discovery Service
 * Automatically detects and loads service account JSON files
 *
 * Single Responsibility: Auto-discover and register service accounts
 */

import * as fs from "fs";
import * as path from "path";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
// import { getServiceAccountRotationManager } from "./googleSheetsServiceAccountRotationManager";

const GOOGLE_CONFIG_DIR = path.join(process.cwd(), "config/static/google");

export class GoogleSheetsAutoDiscoveryService {
  private logger: Logger;
  private discoveredAccounts: Set<string> = new Set();

  constructor() {
    this.logger = getLogger("GoogleSheetsAutoDiscovery");
  }

  /**
   * Auto-discover all service account JSON files in the google config directory
   */
  async autoDiscoverServiceAccounts(): Promise<string[]> {
    try {
      // Ensure directory exists
      if (!fs.existsSync(GOOGLE_CONFIG_DIR)) {
        this.logger.warn(
          `Google config directory not found: ${GOOGLE_CONFIG_DIR}`
        );
        return [];
      }

      const files = fs.readdirSync(GOOGLE_CONFIG_DIR);
      const accountFiles: string[] = [];

      for (const file of files) {
        // Only process JSON files that look like service accounts
        if (!file.endsWith(".json")) continue;
        if (file === "registry.json" || file === "google_credentials.json")
          continue;

        const filePath = path.join(GOOGLE_CONFIG_DIR, file);

        try {
          const content = fs.readFileSync(filePath, "utf8");
          const data = JSON.parse(content);

          // Validate it's a service account file
          if (
            data.type === "service_account" &&
            data.project_id &&
            data.private_key_id
          ) {
            accountFiles.push(filePath);
            this.discoveredAccounts.add(file);
            this.logger.info(
              `✅ Discovered service account: ${file} (${data.client_email})`
            );
          }
        } catch (error) {
          this.logger.debug(`Skipped non-service-account file: ${file}`);
        }
      }

      if (accountFiles.length === 0) {
        this.logger.warn("⚠️ No service account files discovered");
        return [];
      }

      this.logger.info(
        `🔍 Auto-discovered ${accountFiles.length} service accounts`
      );
      return accountFiles;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error discovering service accounts: ${errorMsg}`);
      return [];
    }
  }

  /**
   * Get list of discovered service accounts
   */
  getDiscoveredAccounts(): string[] {
    return Array.from(this.discoveredAccounts);
  }

  /**
   * Watch for new service account files (for real-time discovery)
   */
  watchForNewAccounts(callback: (newAccounts: string[]) => void): void {
    try {
      if (!fs.existsSync(GOOGLE_CONFIG_DIR)) {
        this.logger.warn("Cannot watch directory - does not exist");
        return;
      }

      fs.watch(GOOGLE_CONFIG_DIR, (_eventType, filename) => {
        if (!filename || !filename.endsWith(".json")) return;

        const filePath = path.join(GOOGLE_CONFIG_DIR, filename);

        // Check if file exists and is a service account
        if (fs.existsSync(filePath)) {
          try {
            const content = fs.readFileSync(filePath, "utf8");
            const data = JSON.parse(content);

            if (
              data.type === "service_account" &&
              !this.discoveredAccounts.has(filename)
            ) {
              this.discoveredAccounts.add(filename);
              this.logger.info(`🆕 New service account detected: ${filename}`);
              callback([filePath]);
            }
          } catch {
            // Not a valid service account file
          }
        }
      });

      this.logger.info(
        `👀 Watching for new service accounts in ${GOOGLE_CONFIG_DIR}`
      );
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.warn(`Could not watch for new accounts: ${errorMsg}`);
    }
  }

  /**
   * Get all service account info
   */
  async getServiceAccountsInfo(): Promise<
    Array<{ filename: string; email: string; projectId: string }>
  > {
    const info: Array<{ filename: string; email: string; projectId: string }> =
      [];

    for (const filename of this.discoveredAccounts) {
      try {
        const filePath = path.join(GOOGLE_CONFIG_DIR, filename);
        const content = fs.readFileSync(filePath, "utf8");
        const data = JSON.parse(content);

        info.push({
          filename,
          email: data.client_email,
          projectId: data.project_id,
        });
      } catch (error) {
        this.logger.debug(`Error reading ${filename}: ${error}`);
      }
    }

    return info;
  }
}

let discoveryInstance: GoogleSheetsAutoDiscoveryService | null = null;

export function getGoogleSheetsAutoDiscovery(): GoogleSheetsAutoDiscoveryService {
  if (!discoveryInstance) {
    discoveryInstance = new GoogleSheetsAutoDiscoveryService();
  }
  return discoveryInstance;
}

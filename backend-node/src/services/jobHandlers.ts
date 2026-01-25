/**
 * Job Handlers
 * SRP: Initialize and register job handlers
 * Stock/Price handlers delegated to jobHandlersStock
 */

import { getJobExecutor } from "./JobExecutor";
import { getLogger } from "../utils/logger";
import { getTokenManager } from "./tokenManager";
import { getInventoryService } from "./inventoryService";
import { registerStockHandlers } from "./jobHandlersStock";

const logger = getLogger("JobHandlers");

/**
 * Initialize all job handlers
 */
export function initializeJobHandlers(): void {
  const executor = getJobExecutor();

  // Sync inventory handler
  executor.registerHandler("sync_inventory", async (data) => {
    const { source = "local", target = "cloud" } = data || {};

    logger.info(`🔄 Starting inventory sync: ${source} → ${target}...`);

    try {
      const inventoryService = getInventoryService();
      const settings = await inventoryService.getSettings();

      if (!settings.spreadsheetId || !settings.sheetName) {
        throw new Error(
          "Inventory settings not configured. Please configure Google Sheets integration first."
        );
      }

      let result;

      if (target === "cloud") {
        logger.info("📤 Syncing from local database to Google Sheets...");
        result = await inventoryService.syncToSheets(
          settings.spreadsheetId,
          settings.sheetName
        );
      } else if (target === "local") {
        logger.info("📥 Syncing from Google Sheets to local database...");
        result = await inventoryService.syncFromSheets(
          settings.spreadsheetId,
          settings.sheetName
        );
      } else {
        throw new Error(
          `Invalid sync target: ${target}. Use 'local' or 'cloud'.`
        );
      }

      if (result.status === "success") {
        logger.info(
          `✅ Inventory sync completed: ${result.synced_records}/${result.total_records} records synced`
        );
      } else {
        throw new Error(result.message || "Inventory sync failed");
      }
    } catch (error) {
      logger.error(`❌ Inventory sync failed: ${error}`);
      throw error;
    }
  });

  // Sync FROM Google Sheets handler
  executor.registerHandler("sync_from_sheets", async (_data) => {
    logger.info(`📥 Starting scheduled sync from Google Sheets...`);

    try {
      const inventoryService = getInventoryService();
      const settings = await inventoryService.getSettings();

      if (!settings.spreadsheetId || !settings.sheetName) {
        throw new Error(
          "Inventory settings not configured. Please configure Google Sheets integration first."
        );
      }

      logger.info("📥 Syncing from Google Sheets to local database...");
      const result = await inventoryService.syncFromSheets(
        settings.spreadsheetId,
        settings.sheetName
      );

      if (result.status === "success") {
        logger.info(
          `✅ Sync from sheets completed: ${result.synced_records}/${result.total_records} records synced`
        );
      } else {
        throw new Error(result.message || "Sync from sheets failed");
      }
    } catch (error) {
      logger.error(`❌ Sync from sheets failed: ${error}`);
      throw error;
    }
  });

  // Auto-update tokens handler
  executor.registerHandler("auto_update_token", async (_data) => {
    logger.info(`🔑 Starting auto-update token refresh...`);

    try {
      const tokenManager = await getTokenManager();

      const refreshResults = await tokenManager.checkAndRefreshExpiredTokens();

      let successCount = 0;
      let failCount = 0;

      for (const [platform, success] of Object.entries(refreshResults)) {
        if (success) {
          successCount++;
          logger.info(`✅ ${platform} token refreshed`);
        } else {
          failCount++;
          logger.error(`❌ ${platform} token refresh failed`);
        }
      }

      logger.info(
        `✅ Auto-update token completed: ${successCount} succeeded, ${failCount} failed`
      );
    } catch (error) {
      logger.error(`❌ Auto-update token failed: ${error}`);
      throw error;
    }
  });

  // Locked Today handler
  executor.registerHandler("locked_today", async (data) => {
    const { days = 7 } = data || {};

    logger.info(`🔒 Fetching locked today items (last ${days} days)...`);

    try {
      const response = await fetch(
        "http://localhost:3001/api/orders/locked-today",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ days }),
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch locked items: ${response.statusText}`);
      }

      const result = (await response.json()) as {
        success: boolean;
        count: number;
        total_qty: number;
        items: any[];
        error?: string;
      };

      if (result.success) {
        logger.info(
          `✅ Locked today completed: ${result.count} unique SKUs, ${result.total_qty} total quantity`
        );
      } else {
        throw new Error(result.error || "Failed to fetch locked items");
      }
    } catch (error) {
      logger.error(`❌ Locked today failed: ${error}`);
      throw error;
    }
  });

  // Register stock/price handlers
  registerStockHandlers();

  logger.info("✅ All job handlers initialized");
}

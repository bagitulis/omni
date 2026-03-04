/**
 * Stock/Price Job Handlers
 * SRP: Handle stock and price update jobs
 */

import { getJobExecutor } from "./JobExecutor";
import { getLogger } from "../utils/logger";
import { getStockUpdateOrchestrator } from "./stock";
import { getPriceUpdateOrchestrator } from "./price";
import { JobHistoryManager } from "./jobHistoryManager";
import { v4 as uuidv4 } from "uuid";

const logger = getLogger("JobHandlersStock");

export function registerStockHandlers(): void {
  const executor = getJobExecutor();

  // Stock Update handler
  executor.registerHandler("stock_update", async (data) => {
    const { sku, stock, tenantId } = data || {};

    if (!tenantId) {
      throw new Error(
        "Missing required field: tenantId - job must specify tenant context"
      );
    }

    if (!sku) {
      throw new Error("Missing required field: sku");
    }

    logger.info(
      `📦 [Queue] [${tenantId}] Updating stock for SKU: ${sku} → ${stock}`
    );
    const startTime = Date.now();

    try {
      const orchestrator = await getStockUpdateOrchestrator(tenantId);
      const result = await orchestrator.updateStock({ sku, stock });
      const durationMs = Date.now() - startTime;

      const platformSummary: string[] = [];
      const failedPlatforms: string[] = [];
      const historyManager = new JobHistoryManager();
      const batchId = uuidv4().substring(0, 8);

      if (result.platforms) {
        for (const [platform, platformResult] of Object.entries(
          result.platforms
        )) {
          if (!platformResult) continue;

          const jobId = `queue-stock-${batchId}-${platform}`;
          const platformDuration = Math.round(
            durationMs / Object.keys(result.platforms).length
          );

          if (platformResult.success) {
            platformSummary.push(`${platform.toUpperCase()}:✅`);
            logger.info(
              `✅ [Queue] ${sku} → ${platform.toUpperCase()} SUCCESS`
            );
            try {
              historyManager.addDirectHistory(
                jobId,
                `stock_update:${sku}:${platform.toUpperCase()}`,
                "completed",
                undefined,
                platformDuration
              );
            } catch {
              /* ignore */
            }
          } else {
            const errMsg = platformResult.error || "Unknown error";
            platformSummary.push(`${platform.toUpperCase()}:❌`);
            failedPlatforms.push(`${platform.toUpperCase()}: ${errMsg}`);
            logger.error(
              `❌ [Queue] ${sku} → ${platform.toUpperCase()} FAILED: ${errMsg}`
            );
            try {
              historyManager.addDirectHistory(
                jobId,
                `stock_update:${sku}:${platform.toUpperCase()}`,
                "failed",
                errMsg,
                platformDuration
              );
            } catch {
              /* ignore */
            }
          }
        }
      }

      logger.info(`📊 [Queue] ${sku}: ${platformSummary.join(" | ")}`);

      if (platformSummary.length === 0) {
        logger.warn(
          `⏭️ [Queue] ${sku}: SKU not found in any platform - skipped`
        );
        return;
      }

      if (!result.success) {
        throw new Error(`Stock update failed - ${failedPlatforms.join("; ")}`);
      }
    } catch (error) {
      logger.error(`❌ [Queue] Stock update failed for ${sku}: ${error}`);
      throw error;
    }
  });

  // Price Update handler
  executor.registerHandler("price_update", async (data) => {
    const { sku, price, tenantId } = data || {};

    if (!tenantId) {
      throw new Error(
        "Missing required field: tenantId - job must specify tenant context"
      );
    }

    if (!sku) {
      throw new Error("Missing required field: sku");
    }

    logger.info(
      `💰 [Queue] [${tenantId}] Updating price for SKU: ${sku} → ${price}`
    );
    const startTime = Date.now();

    try {
      const orchestrator = await getPriceUpdateOrchestrator(tenantId);
      const result = await orchestrator.updatePrice({ sku, price });
      const durationMs = Date.now() - startTime;

      const platformSummary: string[] = [];
      const failedPlatforms: string[] = [];
      const historyManager = new JobHistoryManager();
      const batchId = uuidv4().substring(0, 8);

      if (result.platforms) {
        for (const [platform, platformResult] of Object.entries(
          result.platforms
        )) {
          if (!platformResult) continue;

          const jobId = `queue-price-${batchId}-${platform}`;
          const platformDuration = Math.round(
            durationMs / Object.keys(result.platforms).length
          );

          if (platformResult.success) {
            platformSummary.push(`${platform.toUpperCase()}:✅`);
            logger.info(
              `✅ [Queue] ${sku} → ${platform.toUpperCase()} PRICE SUCCESS`
            );
            try {
              historyManager.addDirectHistory(
                jobId,
                `price_update:${sku}:${platform.toUpperCase()}`,
                "completed",
                undefined,
                platformDuration
              );
            } catch {
              /* ignore */
            }
          } else {
            const errMsg = platformResult.error || "Unknown error";
            platformSummary.push(`${platform.toUpperCase()}:❌`);
            failedPlatforms.push(`${platform.toUpperCase()}: ${errMsg}`);
            logger.error(
              `❌ [Queue] ${sku} → ${platform.toUpperCase()} PRICE FAILED: ${errMsg}`
            );
            try {
              historyManager.addDirectHistory(
                jobId,
                `price_update:${sku}:${platform.toUpperCase()}`,
                "failed",
                errMsg,
                platformDuration
              );
            } catch {
              /* ignore */
            }
          }
        }
      }

      logger.info(`📊 [Queue] ${sku} PRICE: ${platformSummary.join(" | ")}`);

      if (!result.success) {
        throw new Error(`Price update failed - ${failedPlatforms.join("; ")}`);
      }
    } catch (error) {
      logger.error(`❌ [Queue] Price update failed for ${sku}: ${error}`);
      throw error;
    }
  });

  logger.info("✅ Stock/Price handlers registered");
}

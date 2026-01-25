/**
 * System Integration - Tenant Aware
 * Single Responsibility: Initialize tenant-aware monitoring systems
 * Max 100 lines
 */

import { queueProcessor } from "./services/queueProcessor";
import { metricsService } from "./services/metricsService";
import { alertSystem } from "./services/alertSystem";
import { getLogger } from "./utils/logger";

const logger = getLogger("SystemIntegration");

/**
 * Initialize all monitoring systems
 */
export function initializeMonitoring(): void {
  logger.info("🚀 Initializing tenant-aware monitoring systems...");

  // Configure alert thresholds
  alertSystem.configure({
    slowRouteThreshold: 3000, // 3 seconds
    errorRateThreshold: 10, // 10%
    queueBacklogThreshold: 50,
    alertCooldown: 60000, // 1 minute
  });

  // Start queue processor
  queueProcessor.start();

  // Setup event listeners for queue
  queueProcessor.on("job:enqueued", (job) => {
    metricsService.recordQueueEnqueued(job.tenantId);
  });

  queueProcessor.on("job:completed", ({ job }) => {
    const processingTime = Date.now() - job.createdAt.getTime();
    metricsService.recordQueueCompleted(job.tenantId, processingTime);
  });

  queueProcessor.on("job:failed", ({ job }) => {
    metricsService.recordQueueFailed(job.tenantId);
  });

  queueProcessor.on("job:retry", ({ job }) => {
    metricsService.recordQueueRetried(job.tenantId);
  });

  queueProcessor.on("job:cancelled", (job) => {
    metricsService.recordQueueCancelled(job.tenantId);
  });

  // Setup alert checking
  queueProcessor.on("job:completed", ({ job }) => {
    const stats = queueProcessor.getTenantStats(job.tenantId);
    alertSystem.checkQueueBacklog(job.tenantId, stats.queued);
  });

  logger.info("✅ Tenant-aware monitoring systems initialized");
  logger.info(`   - Queue processor: Active`);
  logger.info(`   - Metrics tracking: Active (per tenant)`);
  logger.info(`   - Alert system: Active (per tenant)`);
}

/**
 * Get monitoring status
 */
export function getMonitoringStatus() {
  const queueStats = queueProcessor.getStats();
  const metricsSummary = metricsService.getSummary();
  const alertStats = alertSystem.getAggregateStats();

  return {
    queue: {
      isRunning: queueStats.isRunning,
      totalQueued: queueStats.totalQueued,
      processing: queueStats.processing,
      tenants: queueStats.tenantCount,
    },
    metrics: {
      totalTenants: metricsSummary.totalTenants,
      cacheHitRate: metricsSummary.cacheHitRate,
      queueSuccessRate: metricsSummary.queueSuccessRate,
    },
    alerts: {
      totalAlerts: alertStats.totalAlerts,
      totalTenants: alertStats.totalTenants,
      critical: alertStats.bySeverity.critical,
    },
  };
}

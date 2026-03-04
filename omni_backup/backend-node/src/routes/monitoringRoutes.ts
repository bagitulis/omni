/**
 * Monitoring Routes - Tenant Aware
 * Single Responsibility: Expose monitoring endpoints with tenant context
 * Max 300 lines
 */

import { Router, Request, Response } from "express";
import { metricsService } from "../services/metricsService";
import { alertSystem } from "../services/alertSystem";
import { queueProcessor } from "../services/queueProcessor";
import { authMiddleware } from "../middleware/authMiddleware";
import { getLogger } from "../utils/logger";
import { sendSuccess, sendUnauthorized, handleError } from "../utils/apiResponse";

const logger = getLogger("MonitoringRoutes");
const router = Router();

// Apply auth middleware to all routes
router.use(authMiddleware);

/**
 * Helper: Extract tenant ID from request
 */
function getTenantId(req: Request): string | null {
  return req.tenantId || null;
}

/**
 * GET /api/monitoring/metrics
 * Get tenant-specific metrics
 */
router.get("/metrics", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    const tenantMetrics = metricsService.getTenantMetricsData(tenantId);
    const queueStats = queueProcessor.getTenantStats(tenantId);

    return sendSuccess(res, {
      tenantId,
      metrics: tenantMetrics,
      queue: queueStats,
    });
  } catch (error: any) {
    logger.error(`Failed to get metrics: ${error.message}`);
    return handleError(res, error, "Failed to get metrics");
  }
});

/**
 * GET /api/monitoring/cache
 * Get tenant-specific cache metrics
 */
router.get("/cache", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    const cacheMetrics = metricsService.getCacheMetrics(tenantId);

    return sendSuccess(res, {
      tenantId,
      cache: cacheMetrics,
    });
  } catch (error: any) {
    logger.error(`Failed to get cache metrics: ${error.message}`);
    return handleError(res, error, "Failed to get cache metrics");
  }
});

/**
 * GET /api/monitoring/queue
 * Get tenant-specific queue metrics
 */
router.get("/queue", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    const queueMetrics = metricsService.getQueueMetrics(tenantId);
    const queueStats = queueProcessor.getTenantStats(tenantId);

    return sendSuccess(res, {
      tenantId,
      queue: {
        metrics: queueMetrics,
        stats: queueStats,
      },
    });
  } catch (error: any) {
    logger.error(`Failed to get queue metrics: ${error.message}`);
    return handleError(res, error, "Failed to get queue metrics");
  }
});

/**
 * GET /api/monitoring/alerts
 * Get tenant-specific alerts
 */
router.get("/alerts", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    const limit = req.query.limit
      ? parseInt(req.query.limit as string)
      : undefined;
    const alerts = alertSystem.getTenantAlerts(tenantId, { limit });
    const stats = alertSystem.getTenantStats(tenantId);

    return sendSuccess(res, {
      tenantId,
      alerts,
      stats,
    });
  } catch (error: any) {
    logger.error(`Failed to get alerts: ${error.message}`);
    return handleError(res, error, "Failed to get alerts");
  }
});

/**
 * GET /api/monitoring/summary
 * Get tenant-specific summary
 */
router.get("/summary", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    const tenantMetrics = metricsService.getTenantMetricsData(tenantId);
    const queueStats = queueProcessor.getTenantStats(tenantId);
    const alertStats = alertSystem.getTenantStats(tenantId);

    return sendSuccess(res, {
      tenantId,
      metrics: tenantMetrics,
      queue: queueStats,
      alerts: alertStats,
    });
  } catch (error: any) {
    logger.error(`Failed to get summary: ${error.message}`);
    return handleError(res, error, "Failed to get summary");
  }
});

/**
 * GET /api/monitoring/aggregate
 * Get aggregate metrics (admin view - all tenants)
 */
router.get("/aggregate", (_req: Request, res: Response) => {
  try {
    const aggregateMetrics = metricsService.getAggregateMetrics();
    const queueStats = queueProcessor.getStats();
    const alertStats = alertSystem.getAggregateStats();

    return sendSuccess(res, {
      metrics: aggregateMetrics,
      queue: queueStats,
      alerts: alertStats,
    });
  } catch (error: any) {
    logger.error(`Failed to get aggregate: ${error.message}`);
    return handleError(res, error, "Failed to get aggregate metrics");
  }
});

/**
 * POST /api/monitoring/clear
 * Clear tenant-specific metrics
 */
router.post("/clear", (req: Request, res: Response) => {
  try {
    const tenantId = getTenantId(req);
    if (!tenantId) {
      return sendUnauthorized(res);
    }

    metricsService.resetTenantMetrics(tenantId);
    alertSystem.clearTenantAlerts(tenantId);
    queueProcessor.clearTenantQueue(tenantId);

    logger.info(`🧹 [${tenantId}] Metrics cleared`);
    return sendSuccess(res, { message: "Metrics cleared" });
  } catch (error: any) {
    logger.error(`Failed to clear metrics: ${error.message}`);
    return handleError(res, error, "Failed to clear metrics");
  }
});

/**
 * POST /api/monitoring/configure-alerts
 * Configure alert thresholds
 */
router.post("/configure-alerts", (req: Request, res: Response) => {
  try {
    const config = req.body;
    alertSystem.configure(config);

    logger.info("⚙️ Alert configuration updated");
    return sendSuccess(res, { message: "Alert configuration updated", config });
  } catch (error: any) {
    logger.error(`Failed to configure alerts: ${error.message}`);
    return handleError(res, error, "Failed to configure alerts");
  }
});

export default router;

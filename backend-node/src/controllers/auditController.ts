import { Response } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import AuditService from "../services/auditService";
import { getLogger } from "../utils/logger";

const logger = getLogger("AuditController");

/**
 * Audit Controller - Handle audit log queries
 *
 * Endpoints:
 *   GET /api/audit/logs/user - Get audit logs for current user
 *   GET /api/audit/logs/tenant - Get audit logs for tenant
 */
export class AuditController {
  /**
   * GET /api/audit/logs/user
   * Get audit logs for current user
   */
  async getUserAuditLogs(req: AuthRequest, res: Response): Promise<void> {
    try {
      const userId = req.userId || "";
      const limit = parseInt(req.query.limit as string) || 50;

      logger.info(`📋 Fetching audit logs for user: ${userId}`);

      const logs = await AuditService.getUserAuditLogs(userId, limit);

      res.json({
        success: true,
        data: {
          logs,
          total: logs.length,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Failed to fetch user audit logs: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * GET /api/audit/logs/tenant
   * Get audit logs for entire tenant (admin only)
   */
  async getTenantAuditLogs(req: AuthRequest, res: Response): Promise<void> {
    try {
      // ⛔ NO DEFAULT TENANT per AGENTS.MD - must have valid tenantId
      if (!req.tenantId) {
        res.status(401).json({
          success: false,
          error: "Missing tenantId - authentication required",
        });
        return;
      }
      const tenantId = req.tenantId;
      const limit = parseInt(req.query.limit as string) || 100;

      logger.info(`📋 Fetching audit logs for tenant: ${tenantId}`);

      const logs = await AuditService.getTenantAuditLogs(tenantId, limit);

      res.json({
        success: true,
        data: {
          tenantId,
          logs,
          total: logs.length,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Failed to fetch tenant audit logs: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/audit/cleanup
   * Delete audit logs older than specified days (admin only)
   *
   * Body:
   * {
   *   "daysOld": 90
   * }
   */
  async cleanupAuditLogs(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { daysOld = 90 } = req.body;

      logger.warn(`🧹 Cleaning up audit logs older than ${daysOld} days`);

      const deletedCount = await AuditService.cleanupOldLogs(daysOld);

      logger.info(`✅ Cleaned up ${deletedCount} audit logs`);
      res.json({
        success: true,
        message: `Deleted ${deletedCount} audit log files`,
      });
    } catch (error: any) {
      logger.error(`❌ Failed to cleanup audit logs: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
}

export function getAuditController(): AuditController {
  return new AuditController();
}

/**
 * Auto Function Action Routes
 * SRP: Handle auto-function actions (enable/disable/cancel)
 */

import { Router } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { getAutoFunctionScheduler } from "../services/AutoFunctionScheduler";
import { getJobManager } from "../services/JobManager";
import { getLogger } from "../utils/logger";
import { getTenantJobDatabase } from "../utils/jobDb";
import { tenantContext } from "../utils/tenantContext";

const logger = getLogger("AutoFunctionActionRoutes");
export const autoFunctionActionRouter = Router();

/**
 * POST /enable - Enable auto-function
 */
autoFunctionActionRouter.post("/:name/enable", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const name = req.params.name;
    const scheduler = getAutoFunctionScheduler();
    const success = scheduler.enableFunction(tenantId, name);

    res.json({
      success: success,
      data: { enabled: success },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to enable auto-function: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /disable - Disable auto-function
 */
autoFunctionActionRouter.post("/:name/disable", (req: AuthRequest, res) => {
  try {
    const tenantId = req.tenantId || tenantContext.getTenantId();
    const name = req.params.name;
    const scheduler = getAutoFunctionScheduler();
    const success = scheduler.disableFunction(tenantId, name);

    res.json({
      success: success,
      data: { disabled: success },
    });
  } catch (error: any) {
    logger.error(`❌ Failed to disable auto-function: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /cancel-scheduled - Cancel scheduled execution
 */
autoFunctionActionRouter.post(
  "/:id/cancel-scheduled",
  (req: AuthRequest, res) => {
    try {
      const tenantId = req.tenantId || tenantContext.getTenantId();
      const id = req.params.id;
      const scheduler = getAutoFunctionScheduler();
      const jobManager = getJobManager();

      const db = getTenantJobDatabase(tenantId);
      const config: any = db
        .prepare("SELECT id, name FROM auto_functions_config WHERE id = ?")
        .get(id);

      if (!config) {
        return res.status(404).json({
          success: false,
          error: "Auto-function not found",
        });
      }

      const success = scheduler.cancelScheduledExecution(tenantId, config.name);

      if (success) {
        const cancelJobId = `cancelled-${config.name}-${Date.now()}`;
        const now = new Date().toISOString();

        try {
          const insertStmt = db.prepare(`
          INSERT INTO jobs (id, type, status, priority, data, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, ?, ?)
        `);

          insertStmt.run(
            cancelJobId,
            config.name,
            "cancelled",
            "normal",
            JSON.stringify({ reason: "manually_cancelled" }),
            now,
            now
          );

          jobManager.addJobHistory(
            cancelJobId,
            "cancelled",
            `Scheduled execution for ${config.name} was manually cancelled`,
            0
          );

          logger.info(`✅ Logged cancellation for ${config.name} to history`);
        } catch (dbError: any) {
          logger.warn(
            `⚠️ Failed to log cancellation to history: ${dbError.message}`
          );
        }
      }

      return res.json({
        success: success,
        data: {
          cancelled: success,
          functionName: config.name,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Failed to cancel scheduled execution: ${error.message}`);
      return res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }
);

export default autoFunctionActionRouter;

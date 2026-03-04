/**
 * Job Queue Routes
 * Handles job queue operations: enqueue, status, history, cancellation
 * Single Responsibility: Job queue management only
 */

import { Router } from "express";
import { getJobManager } from "../services/JobManager";
import { getLogger } from "../utils/logger";
import { authMiddleware, requireAuth } from "../middleware/authMiddleware";
import { sendSuccess, sendBadRequest, handleError } from "../utils/apiResponse";

const logger = getLogger("JobQueueRoutes");
export const jobQueueRouter = Router();

// Apply authentication middleware to all routes
jobQueueRouter.use(authMiddleware);
jobQueueRouter.use(requireAuth);

/**
 * POST /api/jobs/enqueue - Enqueue a new job
 */
jobQueueRouter.post("/enqueue", (req, res) => {
  try {
    const { type, data, priority } = req.body;

    if (!type || !data) {
      return sendBadRequest(res, "Missing required fields: type, data");
    }

    const jobManager = getJobManager();
    const jobId = jobManager.enqueueJob(type, data, priority || "normal");

    return sendSuccess(res, {
      jobId: jobId,
      status: "pending",
    });
  } catch (error: any) {
    logger.error(`❌ Failed to enqueue job: ${error.message}`);
    return handleError(res, error, "Failed to enqueue job");
  }
});

/**
 * GET /api/jobs/status - Get current running job status
 */
jobQueueRouter.get("/status", (_req, res) => {
  try {
    const jobManager = getJobManager();
    const currentJob = jobManager.getCurrentRunningJob();
    const stats = jobManager.getStatistics();

    return sendSuccess(res, {
      currentJob: currentJob || null,
      stats: stats || {},
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get job status: ${error.message}`);
    return handleError(res, error, "Failed to get job status");
  }
});

/**
 * GET /api/jobs/queue - Get pending queue
 */
jobQueueRouter.get("/queue", (req, res) => {
  try {
    const limit = parseInt(req.query.limit as string) || 100;
    const jobManager = getJobManager();
    const queue = jobManager.getPendingQueue(limit);

    return sendSuccess(res, {
      queue: queue,
      total: queue.length,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get queue: ${error.message}`);
    return handleError(res, error, "Failed to get queue");
  }
});

/**
 * GET /api/jobs/history - Get job history
 */
jobQueueRouter.get("/history", (req, res) => {
  try {
    const jobId = req.query.jobId as string | undefined;
    const limit = parseInt(req.query.limit as string) || 100;
    const jobManager = getJobManager();
    const history = jobManager.getJobHistory(jobId, limit);

    return sendSuccess(res, {
      history: history,
      total: history.length,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get history: ${error.message}`);
    return handleError(res, error, "Failed to get history");
  }
});

/**
 * GET /api/jobs/history-paginated - Get paginated job history with filters
 */
jobQueueRouter.get("/history-paginated", (req, res) => {
  try {
    const jobManager = getJobManager();
    const result = jobManager.getJobHistoryPaginated({
      page: parseInt(req.query.page as string) || 1,
      pageSize: parseInt(req.query.pageSize as string) || 20,
      status: req.query.status as string,
      jobType: req.query.jobType as string,
      search: req.query.search as string,
      sortBy: req.query.sortBy as string,
      sortOrder: (req.query.sortOrder as "asc" | "desc") || "desc",
    });

    return sendSuccess(res, result);
  } catch (error: any) {
    logger.error(`❌ Failed to get paginated history: ${error.message}`);
    return handleError(res, error, "Failed to get paginated history");
  }
});

/**
 * GET /api/jobs/history-job-types - Get distinct job types for filter
 */
jobQueueRouter.get("/history-job-types", (_req, res) => {
  try {
    const jobManager = getJobManager();
    const jobTypes = jobManager.getDistinctJobTypes();

    return sendSuccess(res, { jobTypes });
  } catch (error: any) {
    logger.error(`❌ Failed to get job types: ${error.message}`);
    return handleError(res, error, "Failed to get job types");
  }
});

/**
 * GET /api/jobs/status/:jobId - Get specific job status and history
 */
jobQueueRouter.get("/status/:jobId", (req, res) => {
  try {
    const jobId = req.params.jobId;
    const jobManager = getJobManager();
    const currentJob = jobManager.getCurrentRunningJob();
    const history = jobManager.getJobHistory(jobId, 10);

    // Determine job status
    let status = "unknown";
    if (currentJob && currentJob.id === jobId) {
      status = "running";
    } else if (history.length > 0) {
      status = history[0].status || "unknown";
    }

    return sendSuccess(res, {
      jobId,
      status,
      currentJob: currentJob?.id === jobId ? currentJob : null,
      history: history,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get job status: ${error.message}`);
    return handleError(res, error, "Failed to get job status");
  }
});

/**
 * GET /api/jobs/monitor - Get full monitor status (current + queue + stats)
 */
jobQueueRouter.get("/monitor", (_req, res) => {
  try {
    const jobManager = getJobManager();
    const queueStatus = jobManager.getQueueStatus();
    const recentHistory = jobManager.getJobHistory(undefined, 10);

    return sendSuccess(res, {
      currentJob: queueStatus.currentJob || null,
      pendingQueue: queueStatus.pendingQueue,
      totalPending: queueStatus.totalPending,
      totalCompleted: queueStatus.totalCompleted,
      recentHistory: recentHistory,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to get monitor status: ${error.message}`);
    return handleError(res, error, "Failed to get monitor status");
  }
});

/**
 * POST /api/jobs/cancel/:jobId - Cancel a job
 */
jobQueueRouter.post("/cancel/:jobId", (req, res) => {
  try {
    const jobId = req.params.jobId;
    const jobManager = getJobManager();
    const success = jobManager.cancelJob(jobId, false);

    return sendSuccess(res, { cancelled: success });
  } catch (error: any) {
    logger.error(`❌ Failed to cancel job: ${error.message}`);
    return handleError(res, error, "Failed to cancel job");
  }
});

/**
 * POST /api/jobs/force-cancel/:jobId - Force cancel a job (even if running)
 */
jobQueueRouter.post("/force-cancel/:jobId", (req, res) => {
  try {
    const jobId = req.params.jobId;
    const jobManager = getJobManager();
    const success = jobManager.cancelJob(jobId, true);

    if (success) {
      logger.info(`✅ Job force cancelled: ${jobId}`);
    }

    return sendSuccess(res, { forceCancelled: success });
  } catch (error: any) {
    logger.error(`❌ Failed to force cancel job: ${error.message}`);
    return handleError(res, error, "Failed to force cancel job");
  }
});

/**
 * GET /api/jobs/check-timeout - Check for stuck/timed out jobs
 */
jobQueueRouter.get("/check-timeout", (req, res) => {
  try {
    const timeoutMinutes = parseInt(req.query.timeout as string) || 5;
    const jobManager = getJobManager();
    const timedOutJobs = jobManager.checkAndTimeoutStuckJobs(timeoutMinutes);

    return sendSuccess(res, {
      timedOutJobs,
      count: timedOutJobs.length,
      timeoutMinutes,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to check timeout: ${error.message}`);
    return handleError(res, error, "Failed to check timeout");
  }
});

/**
 * DELETE /api/jobs/history - Clear all job history
 */
jobQueueRouter.delete("/history", (_req, res) => {
  try {
    const jobManager = getJobManager();
    const deleted = jobManager.clearJobHistory();

    return sendSuccess(res, {
      deleted: deleted,
      message: `${deleted} history records deleted`,
    });
  } catch (error: any) {
    logger.error(`❌ Failed to clear history: ${error.message}`);
    return handleError(res, error, "Failed to clear history");
  }
});

export default jobQueueRouter;

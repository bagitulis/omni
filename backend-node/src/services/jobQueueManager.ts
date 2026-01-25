import { v4 as uuidv4 } from "uuid";
import { getTenantJobDatabase, withRetry } from "../utils/jobDb";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import {
  IJob,
  JobStatus,
  JobPriority,
  IJobQueueStatus,
} from "../types/job.types";

const logger = getLogger("JobQueueManager");

/**
 * Job Queue Manager - Tenant-aware job queueing
 * Uses TenantContext to automatically get correct database per tenant
 *
 * Benefits:
 *   ✅ No concurrent write contention (separate DB per tenant)
 *   ✅ Automatic tenant isolation via context
 *   ✅ No method signature changes needed
 */
export class JobQueueManager {
  enqueueJob(
    type: string,
    data: Record<string, any>,
    priority: JobPriority = "normal"
  ): string {
    const jobId = uuidv4();
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(`
        INSERT INTO jobs (id, type, status, priority, data)
        VALUES (?, ?, ?, ?, ?)
      `);

      stmt.run(jobId, type, "pending", priority, JSON.stringify(data));
      logger.info(`📥 Job enqueued: ${jobId} (type: ${type})`);
      return jobId;
    } catch (error) {
      logger.error(`❌ Failed to enqueue job: ${error}`);
      throw error;
    }
  }

  getJobStatus(jobId: string): IJob | null {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare("SELECT * FROM jobs WHERE id = ?");
      const row = stmt.get(jobId) as any;

      if (!row) return null;
      return this.mapRowToJob(row);
    } catch (error) {
      logger.error(`❌ Failed to get job status: ${error}`);
      throw error;
    }
  }

  getCurrentRunningJob(): IJob | null {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    return withRetry(
      () => {
        const stmt = db.prepare(
          "SELECT * FROM jobs WHERE status = ? ORDER BY started_at DESC LIMIT 1"
        );
        const row = stmt.get("running") as any;
        return row ? this.mapRowToJob(row) : null;
      },
      { operationName: "getCurrentRunningJob", maxRetries: 3 }
    );
  }

  getPendingQueue(limit: number = 100): IJob[] {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    return withRetry(
      () => {
        const stmt = db.prepare(`
          SELECT * FROM jobs 
          WHERE status = 'pending'
          ORDER BY 
            CASE priority
              WHEN 'high' THEN 0
              WHEN 'normal' THEN 1
              WHEN 'low' THEN 2
            END ASC,
            created_at ASC
          LIMIT ?
        `);

        const rows = stmt.all(limit) as any[];
        return rows.map((row) => this.mapRowToJob(row));
      },
      { operationName: "getPendingQueue", maxRetries: 3 }
    );
  }

  getQueueStatus(): IJobQueueStatus {
    const currentJob = this.getCurrentRunningJob();
    const pendingQueue = this.getPendingQueue(1000);

    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);
    const completedResult = withRetry(
      () => {
        const completedStmt = db.prepare(
          "SELECT COUNT(*) as count FROM jobs WHERE status = ?"
        );
        return completedStmt.get("completed") as any;
      },
      { operationName: "getQueueStatus-completed", maxRetries: 3 }
    );

    return {
      currentJob: currentJob || undefined,
      pendingQueue,
      totalPending: pendingQueue.length,
      totalCompleted: completedResult?.count || 0,
    };
  }

  updateJobStatus(
    jobId: string,
    status: JobStatus,
    errorMessage?: string
  ): void {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const now = new Date().toISOString();
      const updates: string[] = ["status = ?"];
      const values: any[] = [status];

      if (status === "running") {
        updates.push("started_at = ?");
        values.push(now);
      } else if (status === "completed" || status === "failed") {
        updates.push("completed_at = ?");
        values.push(now);
      }

      if (errorMessage) {
        updates.push("error_message = ?");
        values.push(errorMessage);
      }

      updates.push("updated_at = ?");
      values.push(now);
      values.push(jobId);

      const stmt = db.prepare(`
        UPDATE jobs SET ${updates.join(", ")} WHERE id = ?
      `);

      stmt.run(...values);
      logger.info(`✅ Job ${jobId} status updated to ${status}`);
    } catch (error) {
      logger.error(`❌ Failed to update job status: ${error}`);
      throw error;
    }
  }

  cancelJob(jobId: string, force: boolean = false): boolean {
    try {
      const job = this.getJobStatus(jobId);
      if (!job) {
        logger.warn(`⚠️ Job ${jobId} not found`);
        return false;
      }

      if (job.status === "pending") {
        this.updateJobStatus(jobId, "cancelled");
        logger.info(`✅ Job cancelled: ${jobId}`);
        return true;
      }

      if (job.status === "running") {
        if (!force) {
          logger.warn(
            `⚠️ Cannot cancel running job ${jobId}. Use force=true to force cancel.`
          );
          return false;
        }
        this.updateJobStatus(jobId, "cancelled");
        logger.info(`✅ Job force cancelled: ${jobId}`);
        return true;
      }

      if (
        job.status === "completed" ||
        job.status === "failed" ||
        job.status === "cancelled"
      ) {
        logger.warn(`⚠️ Cannot cancel ${job.status} job ${jobId}`);
        return false;
      }

      return false;
    } catch (error) {
      logger.error(`❌ Failed to cancel job: ${error}`);
      return false;
    }
  }

  checkAndTimeoutStuckJobs(timeoutMinutes: number = 5): string[] {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);
    const timedOutJobs: string[] = [];

    const runningJobs = withRetry(
      () => {
        const stmt = db.prepare(`
          SELECT * FROM jobs 
          WHERE status = 'running' AND started_at IS NOT NULL
        `);

        return stmt.all() as any[];
      },
      { operationName: "checkStuckJobs", maxRetries: 3 }
    );

    const now = Date.now();
    const timeoutMs = timeoutMinutes * 60 * 1000;

    for (const job of runningJobs) {
      const startTime = new Date(job.started_at).getTime();
      const elapsedMs = now - startTime;

      if (elapsedMs > timeoutMs) {
        const errorMsg = `Job timed out after ${timeoutMinutes} minutes`;
        this.updateJobStatus(job.id, "failed", errorMsg);
        timedOutJobs.push(job.id);

        logger.warn(
          `⏱️ Job timed out and marked as failed: ${
            job.id
          } (running for ${Math.round(elapsedMs / 1000)}s)`
        );
      }
    }

    return timedOutJobs;
  }

  getRunningJobDuration(jobId: string): number | null {
    const job = this.getJobStatus(jobId);
    if (!job || job.status !== "running" || !job.startedAt) {
      return null;
    }

    return Date.now() - new Date(job.startedAt).getTime();
  }

  clearOldJobs(olderThanDays: number = 7): number {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(`
        DELETE FROM jobs 
        WHERE (status = 'completed' OR status = 'failed')
        AND created_at < datetime('now', '-' || ? || ' days')
      `);

      const result = stmt.run(olderThanDays);
      logger.info(`🗑️ Cleared ${result.changes} old jobs`);
      return result.changes;
    } catch (error) {
      logger.error(`❌ Failed to clear old jobs: ${error}`);
      return 0;
    }
  }

  getStatistics() {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const totalStmt = db.prepare("SELECT COUNT(*) as count FROM jobs");
      const pendingStmt = db.prepare(
        "SELECT COUNT(*) as count FROM jobs WHERE status = 'pending'"
      );
      const runningStmt = db.prepare(
        "SELECT COUNT(*) as count FROM jobs WHERE status = 'running'"
      );
      const completedStmt = db.prepare(
        "SELECT COUNT(*) as count FROM jobs WHERE status = 'completed'"
      );
      const failedStmt = db.prepare(
        "SELECT COUNT(*) as count FROM jobs WHERE status = 'failed'"
      );

      return {
        total: (totalStmt.get() as any).count,
        pending: (pendingStmt.get() as any).count,
        running: (runningStmt.get() as any).count,
        completed: (completedStmt.get() as any).count,
        failed: (failedStmt.get() as any).count,
      };
    } catch (error) {
      logger.error(`❌ Failed to get statistics: ${error}`);
      return null;
    }
  }

  private mapRowToJob(row: any): IJob {
    return {
      id: row.id,
      type: row.type,
      status: row.status,
      priority: row.priority,
      data: JSON.parse(row.data),
      errorMessage: row.error_message,
      startedAt: row.started_at ? new Date(row.started_at) : undefined,
      completedAt: row.completed_at ? new Date(row.completed_at) : undefined,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at),
    };
  }
}

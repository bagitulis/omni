/**
 * Tenant-Aware Queue Processor Service
 * Single Responsibility: Process queued jobs with tenant isolation
 * DRY: Centralized queue management per tenant
 * Max 300 lines per file
 */

import { EventEmitter } from "events";
import { getLogger } from "../utils/logger";

const logger = getLogger("QueueProcessor");

export interface QueueJob {
  id: string;
  tenantId: string;
  type: string;
  category: string;
  priority: "low" | "medium" | "high";
  data: any;
  retries: number;
  maxRetries: number;
  createdAt: Date;
  scheduledAt?: Date;
  handler: () => Promise<any>;
}

export interface QueueConfig {
  maxConcurrent: number;
  retryDelay: number;
  maxRetryDelay: number;
}

export class TenantQueueProcessor extends EventEmitter {
  private queues: Map<string, QueueJob[]> = new Map();
  private processing: Map<string, QueueJob> = new Map();
  private config: QueueConfig;
  private isRunning = false;

  constructor(config: Partial<QueueConfig> = {}) {
    super();
    this.config = {
      maxConcurrent: config.maxConcurrent || 5,
      retryDelay: config.retryDelay || 1000,
      maxRetryDelay: config.maxRetryDelay || 30000,
    };
  }

  private getTenantQueue(tenantId: string): QueueJob[] {
    if (!this.queues.has(tenantId)) {
      this.queues.set(tenantId, []);
    }
    return this.queues.get(tenantId)!;
  }

  enqueue(
    job: Omit<QueueJob, "id" | "createdAt" | "retries"> & { tenantId: string }
  ): string {
    const queueJob: QueueJob = {
      ...job,
      id: this.generateJobId(),
      createdAt: new Date(),
      retries: 0,
    };

    const tenantQueue = this.getTenantQueue(job.tenantId);
    const priorityValues = { high: 3, medium: 2, low: 1 };
    const insertIndex = tenantQueue.findIndex(
      (j) => priorityValues[j.priority] < priorityValues[queueJob.priority]
    );

    if (insertIndex === -1) {
      tenantQueue.push(queueJob);
    } else {
      tenantQueue.splice(insertIndex, 0, queueJob);
    }

    logger.info(
      `📥 [${job.tenantId}] Job enqueued: ${job.type} [${job.priority}]`
    );
    this.emit("job:enqueued", queueJob);

    if (!this.isRunning) {
      this.start();
    }

    return queueJob.id;
  }

  start(): void {
    if (this.isRunning) return;
    this.isRunning = true;
    logger.info("▶️ Queue processor started");
    this.processNext();
  }

  stop(): void {
    this.isRunning = false;
    logger.info("⏸️ Queue processor stopped");
  }

  private async processNext(): Promise<void> {
    if (!this.isRunning) return;

    if (this.processing.size >= this.config.maxConcurrent) {
      setTimeout(() => this.processNext(), 100);
      return;
    }

    let job: QueueJob | null = null;
    for (const [_tenantId, tenantQueue] of this.queues.entries()) {
      const index = tenantQueue.findIndex(
        (j) => !j.scheduledAt || j.scheduledAt <= new Date()
      );

      if (index > -1) {
        job = tenantQueue.splice(index, 1)[0];
        break;
      }
    }

    if (!job) {
      setTimeout(() => this.processNext(), 500);
      return;
    }

    this.processing.set(job.id, job);
    this.emit("job:processing", job);

    try {
      logger.info(`⚙️ [${job.tenantId}] Processing: ${job.type}`);
      const result = await job.handler();
      this.processing.delete(job.id);
      this.emit("job:completed", { job, result });
      logger.info(`✅ [${job.tenantId}] Completed: ${job.type}`);
    } catch (error: any) {
      logger.error(`❌ [${job.tenantId}] Failed: ${job.type}`);
      await this.handleJobFailure(job, error);
    }

    setImmediate(() => this.processNext());
  }

  private async handleJobFailure(job: QueueJob, error: Error): Promise<void> {
    this.processing.delete(job.id);
    job.retries++;

    if (job.retries < job.maxRetries) {
      const delay = Math.min(
        this.config.retryDelay * Math.pow(2, job.retries),
        this.config.maxRetryDelay
      );

      job.scheduledAt = new Date(Date.now() + delay);
      const tenantQueue = this.getTenantQueue(job.tenantId);
      tenantQueue.unshift(job);

      logger.warn(`🔄 [${job.tenantId}] Retry in ${delay}ms`);
      this.emit("job:retry", { job, delay });
    } else {
      logger.error(`💀 [${job.tenantId}] Failed permanently`);
      this.emit("job:failed", { job, error });
    }
  }

  getStats() {
    let totalQueued = 0;
    const tenantStats: Record<string, any> = {};

    for (const [tenantId, tenantQueue] of this.queues.entries()) {
      totalQueued += tenantQueue.length;
      tenantStats[tenantId] = {
        queued: tenantQueue.length,
        byPriority: {
          high: tenantQueue.filter((j) => j.priority === "high").length,
          medium: tenantQueue.filter((j) => j.priority === "medium").length,
          low: tenantQueue.filter((j) => j.priority === "low").length,
        },
      };
    }

    return {
      totalQueued,
      processing: this.processing.size,
      isRunning: this.isRunning,
      tenantCount: this.queues.size,
      byTenant: tenantStats,
    };
  }

  getTenantStats(tenantId: string) {
    const tenantQueue = this.queues.get(tenantId) || [];
    return {
      tenantId,
      queued: tenantQueue.length,
      byPriority: {
        high: tenantQueue.filter((j) => j.priority === "high").length,
        medium: tenantQueue.filter((j) => j.priority === "medium").length,
        low: tenantQueue.filter((j) => j.priority === "low").length,
      },
    };
  }

  cancelJob(jobId: string): boolean {
    for (const [tenantId, tenantQueue] of this.queues.entries()) {
      const index = tenantQueue.findIndex((j) => j.id === jobId);
      if (index > -1) {
        const job = tenantQueue.splice(index, 1)[0];
        this.emit("job:cancelled", job);
        logger.info(`🚫 [${tenantId}] Job cancelled`);
        return true;
      }
    }
    return false;
  }

  clearTenantQueue(tenantId: string): number {
    const tenantQueue = this.queues.get(tenantId);
    const count = tenantQueue?.length || 0;
    this.queues.set(tenantId, []);
    logger.info(`🗑️ [${tenantId}] Queue cleared: ${count} jobs`);
    return count;
  }

  clearAllQueues(): number {
    let totalCount = 0;
    for (const [_, tenantQueue] of this.queues.entries()) {
      totalCount += tenantQueue.length;
    }
    this.queues.clear();
    logger.info(`🗑️ All queues cleared: ${totalCount} jobs`);
    return totalCount;
  }

  private generateJobId(): string {
    return `job_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

// Singleton instance with tenant isolation
export const queueProcessor = new TenantQueueProcessor({
  maxConcurrent: 5,
  retryDelay: 1000,
  maxRetryDelay: 30000,
});

// Backward compatibility
export const tenantQueueProcessor = queueProcessor;

import { getJobManager } from "./JobManager";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import { getAutoFunctionScheduler } from "./AutoFunctionScheduler";
import { JobHandler, IJob, IJobExecutionResult } from "../types/job.types";
import fs from "fs";
import path from "path";

const logger = getLogger("JobExecutor");

/**
 * Get list of all tenant IDs that support job databases
 * EXCLUDES: 'system' tenant (uses GlobalConfig only, no jobs database)
 * EXCLUDES: 'default' (not a real tenant)
 */
function getAllTenantIds(): string[] {
  try {
    const configPath = path.resolve("config/static/tenants.json");
    if (!fs.existsSync(configPath)) {
      return [];
    }
    const config = JSON.parse(fs.readFileSync(configPath, "utf-8"));
    // Filter out system tenant and any global tenants - they don't have jobs databases
    return Object.keys(config).filter((id) => {
      if (id === "system" || id === "default") return false;
      // Also exclude tenants marked as global
      const tenant = config[id];
      if (tenant?.isGlobal) return false;
      return true;
    });
  } catch (error) {
    logger.error(`Failed to load tenants: ${error}`);
    return [];
  }
}

export class JobExecutor {
  private static instance: JobExecutor;
  private isRunning: boolean = false;
  private jobHandlers: Map<string, JobHandler> = new Map();
  private retryAttempts: number = 3;
  private retryDelayMs: number = 5000;
  private jobTimeoutMinutes: number = 5; // Default 5 minutes timeout
  private timeoutCheckIntervalMs: number = 30000; // Check every 30 seconds
  private timeoutCheckTimeoutId: NodeJS.Timeout | null = null;

  // Circuit breaker for database errors
  private consecutiveErrors: number = 0;
  private circuitBreakerBackoffMs: number = 10000; // Start with 10s backoff
  private maxBackoffMs: number = 60000; // Max 1 minute backoff
  private lastErrorLog: number = 0;
  private errorLogIntervalMs: number = 30000; // Log errors max every 30s

  private constructor() {}

  public static getInstance(): JobExecutor {
    if (!JobExecutor.instance) {
      JobExecutor.instance = new JobExecutor();
    }
    return JobExecutor.instance;
  }

  /**
   * Register a job handler for a specific job type
   */
  public registerHandler(jobType: string, handler: JobHandler): void {
    this.jobHandlers.set(jobType, handler);
    logger.info(`✅ Job handler registered: ${jobType}`);
  }

  /**
   * Start the job executor (polling loop)
   */
  public start(): void {
    if (this.isRunning) {
      logger.warn("⚠️ Job executor already running");
      return;
    }

    this.isRunning = true;
    logger.info("🚀 Job executor started");

    // Process jobs every 500ms
    this.processQueue();

    // Start timeout checker
    this.startTimeoutChecker();
  }

  /**
   * Stop the job executor
   */
  public stop(): void {
    this.isRunning = false;
    if (this.timeoutCheckTimeoutId) {
      clearTimeout(this.timeoutCheckTimeoutId);
      this.timeoutCheckTimeoutId = null;
    }
    logger.info("🛑 Job executor stopped");
  }

  /**
   * Check if executor is running
   */
  public isExecutorRunning(): boolean {
    return this.isRunning;
  }

  /**
   * Process queue continuously - checks ALL tenant databases
   * Implements circuit breaker pattern for database errors
   */
  private async processQueue(): Promise<void> {
    while (this.isRunning) {
      try {
        let jobFound = false;
        const tenants = getAllTenantIds();

        // Iterate through all tenants to find pending jobs
        for (const tenantId of tenants) {
          if (!this.isRunning) break;

          // Run queue check within tenant context
          const result = await tenantContext.run(tenantId, async () => {
            const jobManager = getJobManager();

            // Skip if there's already a running job for this tenant
            const currentJob = jobManager.getCurrentRunningJob();
            if (currentJob) {
              return { hasRunningJob: true, job: null };
            }

            // Get next pending job for this tenant
            const pendingJobs = jobManager.getPendingQueue(1);
            if (pendingJobs.length > 0) {
              return { hasRunningJob: false, job: pendingJobs[0] };
            }

            return { hasRunningJob: false, job: null };
          });

          // If this tenant has a running job, skip to next
          if (result.hasRunningJob) {
            continue;
          }

          // If we found a pending job, execute it
          if (result.job) {
            jobFound = true;
            // Execute job within tenant context
            await tenantContext.run(tenantId, async () => {
              await this.executeJob(result.job!);
            });
            break; // Process one job per cycle
          }
        }

        // Reset circuit breaker on success
        this.consecutiveErrors = 0;

        // If no jobs found in any tenant, wait before checking again
        if (!jobFound) {
          await this.sleep(500);
        }
      } catch (error) {
        // Circuit breaker: increase backoff on consecutive errors
        this.consecutiveErrors++;

        // Only log errors periodically to avoid spam
        const now = Date.now();
        if (now - this.lastErrorLog >= this.errorLogIntervalMs) {
          logger.error(
            `❌ Job queue processing error (${this.consecutiveErrors}x): ${error}`,
          );
          this.lastErrorLog = now;
        }

        // Calculate backoff with exponential increase
        const backoff = Math.min(
          this.circuitBreakerBackoffMs *
            Math.pow(2, Math.min(this.consecutiveErrors - 1, 4)),
          this.maxBackoffMs,
        );

        await this.sleep(backoff);
      }
    }
  }

  /**
   * Execute a single job with retry logic
   * Assumes tenant context is already set by caller (processQueue)
   */
  private async executeJob(job: IJob): Promise<IJobExecutionResult> {
    const jobManager = getJobManager();
    const startTime = Date.now();
    const tenantId = tenantContext.getTenantId();

    try {
      // Check if handler exists
      const handler = this.jobHandlers.get(job.type);
      if (!handler) {
        const duration = Date.now() - startTime;
        const error = `No handler registered for job type: ${job.type}`;
        logger.error(`❌ ${error}`);
        jobManager.updateJobStatus(job.id, "failed", error);
        jobManager.addJobHistory(job.id, "failed", error, duration);
        return { jobId: job.id, success: false, error, durationMs: duration };
      }

      // Mark job as running
      jobManager.updateJobStatus(job.id, "running");
      logger.info(
        `⏳ Executing job: ${job.id} (type: ${job.type}) for tenant: ${tenantId}`,
      );

      // Execute with retry
      let lastError: string | undefined;
      for (let attempt = 1; attempt <= this.retryAttempts; attempt++) {
        try {
          await handler(job.data);

          const duration = Date.now() - startTime;
          jobManager.updateJobStatus(job.id, "completed");
          jobManager.addJobHistory(job.id, "completed", undefined, duration);

          // If this is an auto-function job, update last_executed in scheduler
          if (job.data?.functionName) {
            const scheduler = getAutoFunctionScheduler();
            const tenantId = job.data?.tenantId || tenantContext.getTenantId();
            scheduler.updateLastExecutedAfterCompletion(
              tenantId,
              job.data.functionName,
            );
          }

          logger.info(`✅ Job completed: ${job.id} (duration: ${duration}ms)`);

          return { jobId: job.id, success: true, durationMs: duration };
        } catch (error: any) {
          lastError = error.message || String(error);

          if (attempt < this.retryAttempts) {
            logger.warn(
              `⚠️ Job ${job.id} attempt ${attempt}/${this.retryAttempts} failed: ${lastError}. Retrying in ${this.retryDelayMs}ms...`,
            );
            await this.sleep(this.retryDelayMs);
          } else {
            logger.error(
              `❌ Job ${job.id} failed after ${this.retryAttempts} attempts: ${lastError}`,
            );
          }
        }
      }

      // All retries failed
      const duration = Date.now() - startTime;
      jobManager.updateJobStatus(job.id, "failed", lastError);
      jobManager.addJobHistory(job.id, "failed", lastError, duration);

      // Still update last_executed even if job failed (to avoid infinite retries)
      if (job.data?.functionName) {
        const scheduler = getAutoFunctionScheduler();
        const tenantId = job.data?.tenantId || tenantContext.getTenantId();
        scheduler.updateLastExecutedAfterCompletion(
          tenantId,
          job.data.functionName,
        );
      }

      return {
        jobId: job.id,
        success: false,
        error: lastError,
        durationMs: duration,
      };
    } catch (error: any) {
      const duration = Date.now() - startTime;
      const errorMsg = error.message || String(error);
      logger.error(`❌ Critical error executing job ${job.id}: ${errorMsg}`);

      jobManager.updateJobStatus(job.id, "failed", errorMsg);
      jobManager.addJobHistory(job.id, "failed", errorMsg, duration);

      return {
        jobId: job.id,
        success: false,
        error: errorMsg,
        durationMs: duration,
      };
    }
  }

  /**
   * Set retry configuration
   */
  public setRetryConfig(attempts: number, delayMs: number): void {
    this.retryAttempts = Math.max(1, attempts);
    this.retryDelayMs = Math.max(100, delayMs);
    logger.info(
      `⚙️ Retry config: attempts=${this.retryAttempts}, delay=${this.retryDelayMs}ms`,
    );
  }

  /**
   * Get registered handlers
   */
  public getRegisteredHandlers(): string[] {
    return Array.from(this.jobHandlers.keys());
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * Start timeout checker for stuck jobs - checks ALL tenant databases
   */
  private startTimeoutChecker(): void {
    let consecutiveTimeoutErrors = 0;
    let lastTimeoutErrorLog = 0;

    const checkTimeout = async () => {
      if (!this.isRunning) return;

      try {
        const tenants = getAllTenantIds();

        for (const tenantId of tenants) {
          await tenantContext.run(tenantId, async () => {
            const jobManager = getJobManager();
            const timedOutJobs = jobManager.checkAndTimeoutStuckJobs(
              this.jobTimeoutMinutes,
            );

            if (timedOutJobs.length > 0) {
              logger.warn(
                `⏱️ [${tenantId}] Detected ${
                  timedOutJobs.length
                } timed out job(s): ${timedOutJobs.join(", ")}`,
              );
            }
          });
        }

        // Reset error counter on success
        consecutiveTimeoutErrors = 0;
      } catch (error) {
        consecutiveTimeoutErrors++;

        // Only log errors periodically to avoid spam
        const now = Date.now();
        if (now - lastTimeoutErrorLog >= this.errorLogIntervalMs) {
          logger.error(
            `❌ Error checking for stuck jobs (${consecutiveTimeoutErrors}x): ${error}`,
          );
          lastTimeoutErrorLog = now;
        }
      }

      // Schedule next check
      this.timeoutCheckTimeoutId = setTimeout(
        checkTimeout,
        this.timeoutCheckIntervalMs,
      );
    };

    checkTimeout();
  }

  /**
   * Set job timeout configuration
   * @param timeoutMinutes - Timeout in minutes
   */
  public setJobTimeout(timeoutMinutes: number): void {
    this.jobTimeoutMinutes = Math.max(1, timeoutMinutes);
    logger.info(`⏱️ Job timeout configured: ${this.jobTimeoutMinutes} minutes`);
  }

  /**
   * Get current job timeout in minutes
   */
  public getJobTimeout(): number {
    return this.jobTimeoutMinutes;
  }
}

export function getJobExecutor(): JobExecutor {
  return JobExecutor.getInstance();
}

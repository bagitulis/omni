import { JobQueueManager } from "./jobQueueManager";
import { JobHistoryManager } from "./jobHistoryManager";
import {
  IJob,
  IJobHistory,
  JobStatus,
  JobPriority,
  IJobQueueStatus,
} from "../types/job.types";

/**
 * Job Manager - Facade/Orchestrator
 * Delegates to JobQueueManager and JobHistoryManager
 */
export class JobManager {
  private static instance: JobManager;
  private queueManager: JobQueueManager;
  private historyManager: JobHistoryManager;

  private constructor() {
    this.queueManager = new JobQueueManager();
    this.historyManager = new JobHistoryManager();
  }

  public static getInstance(): JobManager {
    if (!JobManager.instance) {
      JobManager.instance = new JobManager();
    }
    return JobManager.instance;
  }

  // Queue operations
  public enqueueJob(
    type: string,
    data: Record<string, any>,
    priority: JobPriority = "normal"
  ): string {
    return this.queueManager.enqueueJob(type, data, priority);
  }

  public getJobStatus(jobId: string): IJob | null {
    return this.queueManager.getJobStatus(jobId);
  }

  public getCurrentRunningJob(): IJob | null {
    return this.queueManager.getCurrentRunningJob();
  }

  public getPendingQueue(limit: number = 100): IJob[] {
    return this.queueManager.getPendingQueue(limit);
  }

  public getQueueStatus(): IJobQueueStatus {
    return this.queueManager.getQueueStatus();
  }

  public updateJobStatus(
    jobId: string,
    status: JobStatus,
    errorMessage?: string
  ): void {
    this.queueManager.updateJobStatus(jobId, status, errorMessage);
  }

  public cancelJob(jobId: string, force: boolean = false): boolean {
    return this.queueManager.cancelJob(jobId, force);
  }

  public checkAndTimeoutStuckJobs(timeoutMinutes: number = 5): string[] {
    return this.queueManager.checkAndTimeoutStuckJobs(timeoutMinutes);
  }

  public getRunningJobDuration(jobId: string): number | null {
    return this.queueManager.getRunningJobDuration(jobId);
  }

  public clearOldJobs(olderThanDays: number = 7): number {
    return this.queueManager.clearOldJobs(olderThanDays);
  }

  public getStatistics() {
    return this.queueManager.getStatistics();
  }

  // History operations
  public addJobHistory(
    jobId: string,
    status: JobStatus,
    errorMessage?: string,
    durationMs?: number,
    jobType?: string
  ): void {
    this.historyManager.addJobHistory(
      jobId,
      status,
      errorMessage,
      durationMs,
      jobType
    );
  }

  public getJobHistory(jobId?: string, limit: number = 100): IJobHistory[] {
    return this.historyManager.getJobHistory(jobId, limit);
  }

  public getJobHistoryPaginated(options: {
    page?: number;
    pageSize?: number;
    status?: string;
    jobType?: string;
    search?: string;
    sortBy?: string;
    sortOrder?: "asc" | "desc";
  }) {
    return this.historyManager.getJobHistoryPaginated(options);
  }

  public getDistinctJobTypes(): string[] {
    return this.historyManager.getDistinctJobTypes();
  }

  public clearJobHistory(): number {
    return this.historyManager.clearJobHistory();
  }
}

export function getJobManager(): JobManager {
  return JobManager.getInstance();
}

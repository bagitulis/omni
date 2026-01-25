export interface IJob {
  id: string;
  type: string;
  status: JobStatus;
  priority: JobPriority;
  data: Record<string, any>;
  errorMessage?: string;
  startedAt?: Date;
  completedAt?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface IJobHistory {
  id: number;
  jobId: string;
  jobType?: string;
  status: JobStatus;
  errorMessage?: string;
  durationMs?: number;
  startedAt?: Date;
  completedAt?: Date;
  createdAt: Date;
}

export interface IAutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  intervalMinutes: number;
  startTime?: string; // HH:mm format
  endTime?: string; // HH:mm format
  lastExecuted?: Date;
  nextScheduledExecution?: Date; // When the next execution is scheduled
  createdAt: Date;
  updatedAt: Date;
}

export type JobStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled";
export type JobPriority = "low" | "normal" | "high";

export interface IJobQueueStatus {
  currentJob?: IJob;
  pendingQueue: IJob[];
  totalPending: number;
  totalCompleted: number;
}

export interface IJobExecutionResult {
  jobId: string;
  success: boolean;
  error?: string;
  durationMs: number;
}

export type JobHandler = (data: Record<string, any>) => Promise<void>;

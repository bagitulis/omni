/**
 * Queue Tab Utilities
 * Shared formatting and calculation functions for queue displays
 */

export interface Job {
  id: string;
  type: string;
  status: string;
  priority: string;
  data: Record<string, any>;
  startedAt?: Date;
  createdAt: Date;
}

export interface AutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  intervalMinutes: number;
  startTime?: string;
  endTime?: string;
  lastExecuted?: string | Date;
  nextScheduledExecution?: string | Date;
  createdAt?: string | Date;
  updatedAt?: string | Date;
}

/**
 * Format time in Jakarta timezone (HH:mm:ss)
 */
export function formatTime(date: Date | string | undefined): string {
  if (!date) return "-";

  const d = typeof date === "string" ? new Date(date) : date;
  return new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Jakarta",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  }).format(d);
}

/**
 * Format datetime in Jakarta timezone (MMM DD, YYYY HH:mm:ss)
 */
export function formatDateTime(date: Date | string | undefined): string {
  if (!date) return "-";

  const d = typeof date === "string" ? new Date(date) : date;
  return new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Jakarta",
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  }).format(d);
}

/**
 * Calculate countdown to next trigger
 */
export function getTimeUntilTrigger(config: AutoFunctionConfig): string {
  if (!config.nextScheduledExecution) return "";

  const nextTime = typeof config.nextScheduledExecution === "string"
    ? new Date(config.nextScheduledExecution).getTime()
    : (config.nextScheduledExecution as Date).getTime();

  const diffMs = nextTime - Date.now();

  if (diffMs < 0) {
    return "(should trigger now)";
  }

  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);

  if (diffHours > 0) {
    return `(in ${diffHours}h ${diffMins % 60}m)`;
  } else if (diffMins > 0) {
    return `(in ${diffMins}m ${diffSecs % 60}s)`;
  } else {
    return `(in ${diffSecs}s)`;
  }
}

/**
 * Check if schedule is overdue
 */
export function isScheduleOverdue(config: AutoFunctionConfig): boolean {
  if (!config.nextScheduledExecution) return false;

  const nextTime = typeof config.nextScheduledExecution === "string"
    ? new Date(config.nextScheduledExecution).getTime()
    : (config.nextScheduledExecution as Date).getTime();

  return nextTime <= Date.now();
}

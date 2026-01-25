/**
 * Config Table Utilities
 * Formatting functions for auto-function configuration display
 */

export interface AutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  intervalMinutes: number;
  startTime?: string;
  endTime?: string;
  lastExecuted?: string | Date;
  nextScheduledExecution?: string | Date;
}

/**
 * Format date to Jakarta timezone
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
 * Calculate time remaining until next trigger
 */
export function getTimeUntilTrigger(config: AutoFunctionConfig): string {
  if (!config.nextScheduledExecution) return "";
  
  const nextTime = typeof config.nextScheduledExecution === "string"
    ? new Date(config.nextScheduledExecution).getTime()
    : (config.nextScheduledExecution as Date).getTime();
  
  const diffMs = nextTime - Date.now();
  if (diffMs < 0) return "(should trigger now)";
  
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  
  if (diffHours > 0) return `(in ${diffHours}h ${diffMins % 60}m)`;
  if (diffMins > 0) return `(in ${diffMins}m ${diffSecs % 60}s)`;
  return `(in ${diffSecs}s)`;
}

/**
 * Get CSS class for next trigger urgency indicator
 */
export function getNextTriggerClass(config: AutoFunctionConfig): string {
  if (!config.nextScheduledExecution) return "";
  
  const nextTime = typeof config.nextScheduledExecution === "string"
    ? new Date(config.nextScheduledExecution).getTime()
    : (config.nextScheduledExecution as Date).getTime();
  
  const diffMs = nextTime - Date.now();
  if (diffMs < 0) return "trigger-overdue";
  if (diffMs < 60000) return "trigger-soon";
  return "";
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

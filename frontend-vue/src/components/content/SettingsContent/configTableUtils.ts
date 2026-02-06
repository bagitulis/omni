/**
 * Config Table Utilities
 * Formatting functions for auto-function configuration display
 * JSON uses snake_case as per AGENTS.md standard
 */

export interface AutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  interval_minutes: number;
  start_time?: string;
  end_time?: string;
  last_executed?: string | Date;
  next_scheduled_execution?: string | Date;
  created_at?: string | Date;
  updated_at?: string | Date;
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
  if (!config.next_scheduled_execution) return "";

  const nextTime =
    typeof config.next_scheduled_execution === "string"
      ? new Date(config.next_scheduled_execution).getTime()
      : (config.next_scheduled_execution as Date).getTime();

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
  if (!config.next_scheduled_execution) return "";

  const nextTime =
    typeof config.next_scheduled_execution === "string"
      ? new Date(config.next_scheduled_execution).getTime()
      : (config.next_scheduled_execution as Date).getTime();

  const diffMs = nextTime - Date.now();
  if (diffMs < 0) return "trigger-overdue";
  if (diffMs < 60000) return "trigger-soon";
  return "";
}

/**
 * Check if schedule is overdue
 */
export function isScheduleOverdue(config: AutoFunctionConfig): boolean {
  if (!config.next_scheduled_execution) return false;
  const nextTime =
    typeof config.next_scheduled_execution === "string"
      ? new Date(config.next_scheduled_execution).getTime()
      : (config.next_scheduled_execution as Date).getTime();
  return nextTime <= Date.now();
}

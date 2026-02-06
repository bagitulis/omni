/**
 * Current Job Tab Utilities
 * Formatting and calculation functions for job display
 * API types use snake_case to match backend JSON response
 */

const JOB_TIMEOUT_MINUTES = 5;

export interface Job {
  id: string;
  type: string;
  status: string;
  priority: string;
  data: Record<string, any>;
  started_at?: Date;
  created_at: Date;
}

/**
 * Format datetime in Jakarta timezone
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
 * Calculate elapsed time duration
 */
export function calculateDuration(
  started_at: Date | string | undefined,
): string {
  if (!started_at) return "-";

  const start =
    typeof started_at === "string" ? new Date(started_at) : started_at;
  const duration = Date.now() - start.getTime();
  const seconds = Math.floor(duration / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

/**
 * Get job duration in minutes (with decimals)
 */
export function getJobDurationMinutes(
  started_at: Date | string | undefined,
): number {
  if (!started_at) return 0;

  const startTime =
    typeof started_at === "string"
      ? new Date(started_at).getTime()
      : (started_at as Date).getTime();

  return Math.round(((Date.now() - startTime) / (1000 * 60)) * 10) / 10;
}

/**
 * Calculate job progress percentage (0-100)
 */
export function getJobProgressPercent(
  started_at: Date | string | undefined,
): number {
  if (!started_at) return 0;

  const startTime =
    typeof started_at === "string"
      ? new Date(started_at).getTime()
      : (started_at as Date).getTime();

  const elapsedMinutes = (Date.now() - startTime) / (1000 * 60);
  const percent = (elapsedMinutes / JOB_TIMEOUT_MINUTES) * 100;

  return Math.min(percent, 100);
}

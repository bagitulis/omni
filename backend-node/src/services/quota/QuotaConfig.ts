/**
 * QuotaConfig - Configuration for Google Sheets API Quota Management
 *
 * SINGLE RESPONSIBILITY: Define quota limits and configuration constants
 *
 * Google Sheets API Limits:
 * - 60 read requests per minute per user (service account)
 * - 60 write requests per minute per user (service account)
 * - Shared across all tenants using the same service account
 */

export interface QuotaLimits {
  requestsPerMinute: number;
  requestsPerSecond: number;
  burstLimit: number;
  windowSizeMs: number;
}

export interface BackoffConfig {
  initialDelayMs: number;
  maxDelayMs: number;
  multiplier: number;
  maxRetries: number;
}

export interface QuotaManagerConfig {
  limits: QuotaLimits;
  backoff: BackoffConfig;
  rotationThresholdPercent: number;
  enablePersistence: boolean;
  enableAutoRotation: boolean;
}

/**
 * Default quota configuration based on Google Sheets API limits
 */
export const DEFAULT_QUOTA_CONFIG: QuotaManagerConfig = {
  limits: {
    requestsPerMinute: 60,
    requestsPerSecond: 1, // Conservative: 1 req/sec = 60/min
    burstLimit: 10, // Allow small bursts
    windowSizeMs: 60000, // 1 minute sliding window
  },
  backoff: {
    initialDelayMs: 1000, // Start with 1 second
    maxDelayMs: 60000, // Max 60 seconds
    multiplier: 2, // Double each retry
    maxRetries: 5, // Max 5 retries before giving up
  },
  rotationThresholdPercent: 80, // Rotate when 80% quota used
  enablePersistence: true, // Save state to database
  enableAutoRotation: true, // Auto-rotate on quota hit
};

/**
 * Get dynamic quota limits based on number of service accounts
 * More accounts = more total capacity
 */
export function getDynamicLimits(accountCount: number): QuotaLimits {
  const baseLimit = DEFAULT_QUOTA_CONFIG.limits.requestsPerMinute;

  return {
    requestsPerMinute: baseLimit, // Per account limit stays same
    requestsPerSecond: Math.max(1, Math.floor(baseLimit / 60)),
    burstLimit: Math.min(15, accountCount * 5), // Scale burst with accounts
    windowSizeMs: 60000,
  };
}

/**
 * Calculate safe request rate for given account count
 * Returns requests per second that's safe across all accounts
 */
export function getSafeRequestRate(accountCount: number): number {
  if (accountCount <= 0) return 0;

  // Total capacity = accounts * 60 requests/min
  // Safe rate = 80% of capacity / 60 seconds
  const totalCapacity = accountCount * 60;
  const safeCapacity = totalCapacity * 0.8;

  return Math.floor(safeCapacity / 60); // requests per second
}

export type RequestPriority = "high" | "normal" | "low";

export const PRIORITY_WEIGHTS: Record<RequestPriority, number> = {
  high: 3,
  normal: 2,
  low: 1,
};

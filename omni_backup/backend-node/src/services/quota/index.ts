/**
 * Quota Management Module Exports
 *
 * Provides comprehensive quota management for Google Sheets API
 * Handles rate limiting, backoff, and service account rotation
 */

// Configuration
export {
  QuotaLimits,
  BackoffConfig,
  QuotaManagerConfig,
  DEFAULT_QUOTA_CONFIG,
  getDynamicLimits,
  getSafeRequestRate,
  RequestPriority,
  PRIORITY_WEIGHTS,
} from "./QuotaConfig";

// Usage Tracking
export {
  UsageRecord,
  UsageSnapshot,
  UsageTracker,
  getUsageTracker,
} from "./UsageTracker";

// Service Account Selection
export {
  SelectionStrategy,
  SelectionResult,
  ServiceAccountSelector,
  getServiceAccountSelector,
} from "./ServiceAccountSelector";

// Backoff Strategy
export {
  BackoffState,
  BackoffStrategy,
  createBackoff,
} from "./BackoffStrategy";

// Request Throttling
export {
  QueuedRequest,
  ThrottlerConfig,
  RequestThrottler,
  getRequestThrottler,
} from "./RequestThrottler";

// Main Quota Manager
export {
  ExecutionResult,
  QuotaStatus,
  GoogleSheetsQuotaManager,
  getQuotaManager,
} from "./GoogleSheetsQuotaManager";

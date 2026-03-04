/**
 * GoogleSheetsQuotaManager - Main orchestrator for quota management
 *
 * SINGLE RESPONSIBILITY: Coordinate all quota-related components
 *
 * This is the main entry point that combines:
 * - Usage tracking
 * - Service account selection
 * - Rate limiting
 * - Backoff/retry logic
 */

import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { backendLogger } from "../../utils/backendLogger";
import {
  QuotaManagerConfig,
  DEFAULT_QUOTA_CONFIG,
  getSafeRequestRate,
  RequestPriority,
} from "./QuotaConfig";
import {
  ServiceAccountSelector,
  getServiceAccountSelector,
} from "./ServiceAccountSelector";
import { RequestThrottler, getRequestThrottler } from "./RequestThrottler";
import { createBackoff } from "./BackoffStrategy";
import { GoogleSheetsErrorHandler } from "../../utils/googleSheetsErrorHandler";

export interface ExecutionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
  accountUsed?: string;
  retriesUsed: number;
  totalTimeMs: number;
}

export interface QuotaStatus {
  healthy: boolean;
  totalAccounts: number;
  availableAccounts: number;
  totalCapacity: number;
  usedCapacity: number;
  availableCapacity: number;
  usagePercent: number;
  queueSize: number;
  accounts: any[];
}

export class GoogleSheetsQuotaManager {
  private logger: Logger;
  private config: QuotaManagerConfig;
  private selector: ServiceAccountSelector;
  private throttler: RequestThrottler;
  private errorHandler: GoogleSheetsErrorHandler;
  private initialized: boolean = false;

  constructor(config?: Partial<QuotaManagerConfig>) {
    this.logger = getLogger("QuotaManager");
    this.config = { ...DEFAULT_QUOTA_CONFIG, ...config };
    this.selector = getServiceAccountSelector();
    this.throttler = getRequestThrottler();
    this.errorHandler = new GoogleSheetsErrorHandler(this.logger);
  }

  /**
   * Initialize the quota manager
   * Call this after service accounts are loaded
   */
  initialize(): void {
    if (this.initialized) return;
    this.doInitialize();
  }

  /**
   * Reinitialize when service accounts change
   * Called when new accounts are discovered
   */
  reinitialize(): void {
    this.doInitialize();
  }

  /**
   * Internal initialization logic
   */
  private doInitialize(): void {
    // Initialize selector with current accounts
    this.selector.initialize();

    // Update throttler rate based on account count
    const accountCount = this.selector.getAccountCount();
    const safeRate = getSafeRequestRate(accountCount);
    this.throttler.setRateLimit(Math.max(1, safeRate));

    this.initialized = true;

    backendLogger.info(
      "QuotaManager",
      `✅ Initialized with ${accountCount} account(s), rate limit: ${safeRate} req/s`
    );
  }

  /**
   * Execute a Google Sheets API operation with full quota management
   *
   * This method:
   * 1. Selects best available service account
   * 2. Queues request if rate limited
   * 3. Handles errors with backoff/retry
   * 4. Auto-rotates on quota errors
   */
  async execute<T>(
    tenantId: string,
    operation: (authClient: any) => Promise<T>,
    options: {
      priority?: RequestPriority;
      operationName?: string;
      skipQueue?: boolean;
    } = {}
  ): Promise<ExecutionResult<T>> {
    const startTime = Date.now();
    const {
      priority = "normal",
      operationName = "sheets-operation",
      skipQueue = false,
    } = options;

    if (!this.initialized) {
      this.initialize();
    }

    // Create backoff strategy for this execution
    const backoff = createBackoff(this.config.backoff);
    let retriesUsed = 0;
    let lastError: any = null;

    while (backoff.canRetry()) {
      try {
        // Select best service account
        const selection = await this.selector.selectAccount();

        if (!selection) {
          // No account available, wait for quota reset
          const waitTime = this.selector.getTimeUntilAvailable();

          if (waitTime > 0 && waitTime < 60000) {
            this.logger.warn(
              `⏳ All accounts at capacity, waiting ${waitTime}ms for reset`
            );
            await this.sleep(waitTime + 1000); // +1s buffer
            continue;
          }

          throw new Error(
            "All service accounts exhausted, no capacity available"
          );
        }

        // Execute operation
        const executeOperation = async (): Promise<T> => {
          const result = await operation(selection.authClient);

          // Record successful usage
          this.selector.recordSuccess(selection.accountId);

          return result;
        };

        // Either queue or execute directly
        const result = skipQueue
          ? await executeOperation()
          : await this.throttler.enqueue(tenantId, executeOperation, priority);

        // Success!
        return {
          success: true,
          data: result,
          accountUsed: selection.accountEmail,
          retriesUsed,
          totalTimeMs: Date.now() - startTime,
        };
      } catch (error: any) {
        lastError = error;
        retriesUsed++;

        // Check error type
        if (
          this.errorHandler.isQuotaError(error) ||
          this.errorHandler.isRateLimitError(error)
        ) {
          backendLogger.warning(
            "QuotaManager",
            `⚠️ Quota/Rate limit hit for ${operationName}, backing off...`
          );

          // Pause throttler briefly
          this.throttler.pause(5000);

          // Wait with quota-specific backoff
          const shouldRetry = await backoff.waitForQuota();
          if (!shouldRetry) break;
        } else if (this.errorHandler.isNetworkError(error)) {
          // Network error - use normal backoff
          const shouldRetry = await backoff.wait();
          if (!shouldRetry) break;
        } else {
          // Unknown error - don't retry
          break;
        }
      }
    }

    // All retries exhausted
    return {
      success: false,
      error: lastError?.message || "Operation failed after retries",
      retriesUsed,
      totalTimeMs: Date.now() - startTime,
    };
  }

  /**
   * Execute operation without queue (immediate execution)
   * Use for critical/time-sensitive operations
   */
  async executeImmediate<T>(
    tenantId: string,
    operation: (authClient: any) => Promise<T>,
    operationName?: string
  ): Promise<ExecutionResult<T>> {
    return this.execute(tenantId, operation, {
      priority: "high",
      operationName,
      skipQueue: true,
    });
  }

  /**
   * Get current quota status
   */
  getStatus(): QuotaStatus {
    // Ensure selector is initialized with latest accounts
    this.selector.initialize();

    const selectorStats = this.selector.getStats();
    const throttlerStats = this.throttler.getStats();

    return {
      healthy: this.selector.hasAvailableAccount(),
      totalAccounts: selectorStats.accountCount,
      availableAccounts:
        selectorStats.accounts?.filter((a: any) => a.available).length || 0,
      totalCapacity: selectorStats.totalCapacityPerMinute,
      usedCapacity: selectorStats.usedCapacity,
      availableCapacity: selectorStats.availableCapacity,
      usagePercent: selectorStats.usagePercent,
      queueSize: throttlerStats.queueSize,
      accounts: selectorStats.accounts || [],
    };
  }

  /**
   * Get detailed statistics
   */
  getDetailedStats(): Record<string, any> {
    // Ensure selector is initialized with latest accounts
    this.selector.initialize();

    return {
      quota: this.getStatus(),
      selector: this.selector.getStats(),
      throttler: this.throttler.getStats(),
      config: {
        rotationThreshold: this.config.rotationThresholdPercent,
        maxRetries: this.config.backoff.maxRetries,
      },
    };
  }

  /**
   * Pause all operations (e.g., during maintenance)
   */
  pause(durationMs: number): void {
    this.throttler.pause(durationMs);
    backendLogger.info("QuotaManager", `⏸️ Paused for ${durationMs}ms`);
  }

  /**
   * Resume operations
   */
  resume(): void {
    this.throttler.resume();
    backendLogger.info("QuotaManager", "▶️ Resumed");
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

// Singleton instance (shared across all tenants)
let quotaManagerInstance: GoogleSheetsQuotaManager | null = null;

export function getQuotaManager(): GoogleSheetsQuotaManager {
  if (!quotaManagerInstance) {
    quotaManagerInstance = new GoogleSheetsQuotaManager();
  }
  return quotaManagerInstance;
}

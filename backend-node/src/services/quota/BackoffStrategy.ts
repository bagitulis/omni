/**
 * BackoffStrategy - Exponential backoff for retry logic
 *
 * SINGLE RESPONSIBILITY: Calculate and execute retry delays
 *
 * Algorithm: delay = min(initialDelay * (multiplier ^ attempt), maxDelay)
 * With jitter to prevent thundering herd
 */

import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { BackoffConfig, DEFAULT_QUOTA_CONFIG } from "./QuotaConfig";

export interface BackoffState {
  attempt: number;
  lastDelay: number;
  totalWaitTime: number;
  startedAt: Date;
}

export class BackoffStrategy {
  private logger: Logger;
  private config: BackoffConfig;
  private state: BackoffState;

  constructor(config?: Partial<BackoffConfig>) {
    this.logger = getLogger("BackoffStrategy");
    this.config = { ...DEFAULT_QUOTA_CONFIG.backoff, ...config };
    this.state = this.createInitialState();
  }

  /**
   * Create initial backoff state
   */
  private createInitialState(): BackoffState {
    return {
      attempt: 0,
      lastDelay: 0,
      totalWaitTime: 0,
      startedAt: new Date(),
    };
  }

  /**
   * Calculate next delay with exponential backoff + jitter
   */
  calculateDelay(): number {
    const baseDelay =
      this.config.initialDelayMs *
      Math.pow(this.config.multiplier, this.state.attempt);

    // Add jitter (±20%) to prevent thundering herd
    const jitter = baseDelay * 0.2 * (Math.random() - 0.5);
    const delay = Math.min(baseDelay + jitter, this.config.maxDelayMs);

    return Math.floor(delay);
  }

  /**
   * Execute backoff wait
   * Returns true if should retry, false if max retries reached
   */
  async wait(): Promise<boolean> {
    if (this.state.attempt >= this.config.maxRetries) {
      this.logger.warn(
        `❌ Max retries (${this.config.maxRetries}) reached after ${this.state.totalWaitTime}ms total wait`
      );
      return false;
    }

    const delay = this.calculateDelay();
    this.state.attempt++;
    this.state.lastDelay = delay;
    this.state.totalWaitTime += delay;

    this.logger.info(
      `⏳ Backoff: waiting ${delay}ms (attempt ${this.state.attempt}/${this.config.maxRetries})`
    );

    await this.sleep(delay);
    return true;
  }

  /**
   * Execute backoff for quota error (longer initial delay)
   */
  async waitForQuota(suggestedDelayMs?: number): Promise<boolean> {
    // For quota errors, start with suggested delay or longer initial delay
    const minQuotaDelay = suggestedDelayMs || 5000; // 5 seconds minimum for quota

    if (this.state.attempt >= this.config.maxRetries) {
      return false;
    }

    const baseDelay = Math.max(
      minQuotaDelay,
      this.config.initialDelayMs *
        Math.pow(this.config.multiplier, this.state.attempt)
    );

    const delay = Math.min(baseDelay, this.config.maxDelayMs);

    this.state.attempt++;
    this.state.lastDelay = delay;
    this.state.totalWaitTime += delay;

    this.logger.info(
      `⏳ Quota backoff: waiting ${delay}ms (attempt ${this.state.attempt}/${this.config.maxRetries})`
    );

    await this.sleep(delay);
    return true;
  }

  /**
   * Reset backoff state (call after successful operation)
   */
  reset(): void {
    this.state = this.createInitialState();
  }

  /**
   * Check if can retry
   */
  canRetry(): boolean {
    return this.state.attempt < this.config.maxRetries;
  }

  /**
   * Get current state
   */
  getState(): BackoffState {
    return { ...this.state };
  }

  /**
   * Get remaining retries
   */
  getRemainingRetries(): number {
    return Math.max(0, this.config.maxRetries - this.state.attempt);
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * Create a new backoff strategy instance
 * Each operation should have its own backoff state
 */
export function createBackoff(
  config?: Partial<BackoffConfig>
): BackoffStrategy {
  return new BackoffStrategy(config);
}

/**
 * RequestThrottler - Rate limiting and request queue management
 *
 * SINGLE RESPONSIBILITY: Control request flow to stay within quota limits
 *
 * Features:
 * - Token bucket algorithm for smooth rate limiting
 * - Fair queue for multiple tenants
 * - Priority support (high/normal/low)
 */

import { Logger } from "winston";
import { getLogger } from "../../utils/logger";
import { RequestPriority, PRIORITY_WEIGHTS } from "./QuotaConfig";

export interface QueuedRequest<T> {
  id: string;
  tenantId: string;
  priority: RequestPriority;
  operation: () => Promise<T>;
  resolve: (value: T) => void;
  reject: (error: any) => void;
  enqueuedAt: Date;
  timeoutMs?: number;
}

export interface ThrottlerConfig {
  maxRequestsPerSecond: number;
  maxQueueSize: number;
  defaultTimeoutMs: number;
}

const DEFAULT_THROTTLER_CONFIG: ThrottlerConfig = {
  maxRequestsPerSecond: 2, // Conservative default
  maxQueueSize: 100,
  defaultTimeoutMs: 30000, // 30 seconds
};

export class RequestThrottler {
  private logger: Logger;
  private config: ThrottlerConfig;
  private queue: QueuedRequest<any>[] = [];
  private isProcessing: boolean = false;
  private lastRequestTime: number = 0;
  private requestCounter: number = 0;
  private paused: boolean = false;
  private pauseUntil: number = 0;

  constructor(config?: Partial<ThrottlerConfig>) {
    this.logger = getLogger("RequestThrottler");
    this.config = { ...DEFAULT_THROTTLER_CONFIG, ...config };
  }

  /**
   * Enqueue a request for rate-limited execution
   */
  enqueue<T>(
    tenantId: string,
    operation: () => Promise<T>,
    priority: RequestPriority = "normal",
    timeoutMs?: number
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      if (this.queue.length >= this.config.maxQueueSize) {
        reject(new Error("Request queue is full"));
        return;
      }

      const request: QueuedRequest<T> = {
        id: `req_${++this.requestCounter}`,
        tenantId,
        priority,
        operation,
        resolve,
        reject,
        enqueuedAt: new Date(),
        timeoutMs: timeoutMs || this.config.defaultTimeoutMs,
      };

      // Insert based on priority
      this.insertByPriority(request);

      this.logger.debug(
        `📥 Queued request ${request.id} for ${tenantId} (${priority}) - Queue size: ${this.queue.length}`
      );

      // Start processing if not already running
      this.processQueue();
    });
  }

  /**
   * Insert request into queue maintaining priority order
   */
  private insertByPriority<T>(request: QueuedRequest<T>): void {
    const weight = PRIORITY_WEIGHTS[request.priority];
    let insertIndex = this.queue.length;

    // Find position to insert (maintain FIFO within same priority)
    for (let i = 0; i < this.queue.length; i++) {
      const existingWeight = PRIORITY_WEIGHTS[this.queue[i].priority];
      if (weight > existingWeight) {
        insertIndex = i;
        break;
      }
    }

    this.queue.splice(insertIndex, 0, request);
  }

  /**
   * Process queued requests with rate limiting
   */
  private async processQueue(): Promise<void> {
    if (this.isProcessing) return;
    this.isProcessing = true;

    while (this.queue.length > 0) {
      // Check if paused
      if (this.paused && Date.now() < this.pauseUntil) {
        const waitTime = this.pauseUntil - Date.now();
        this.logger.info(`⏸️ Throttler paused, waiting ${waitTime}ms`);
        await this.sleep(waitTime);
        this.paused = false;
      }

      // Rate limiting: ensure minimum interval between requests
      const minInterval = 1000 / this.config.maxRequestsPerSecond;
      const timeSinceLastRequest = Date.now() - this.lastRequestTime;

      if (timeSinceLastRequest < minInterval) {
        const waitTime = minInterval - timeSinceLastRequest;
        await this.sleep(waitTime);
      }

      const request = this.queue.shift();
      if (!request) break;

      // Check timeout
      const waitedTime = Date.now() - request.enqueuedAt.getTime();
      if (request.timeoutMs && waitedTime > request.timeoutMs) {
        request.reject(
          new Error(`Request timeout after ${waitedTime}ms in queue`)
        );
        continue;
      }

      // Execute request
      this.lastRequestTime = Date.now();

      try {
        const result = await request.operation();
        request.resolve(result);
      } catch (error) {
        request.reject(error);
      }
    }

    this.isProcessing = false;
  }

  /**
   * Pause processing for specified duration
   * Used when quota is hit and need to wait for reset
   */
  pause(durationMs: number): void {
    this.paused = true;
    this.pauseUntil = Date.now() + durationMs;
    this.logger.warn(`⏸️ Throttler paused for ${durationMs}ms`);
  }

  /**
   * Resume processing immediately
   */
  resume(): void {
    this.paused = false;
    this.pauseUntil = 0;
    this.logger.info("▶️ Throttler resumed");
    this.processQueue();
  }

  /**
   * Update rate limit dynamically
   */
  setRateLimit(requestsPerSecond: number): void {
    this.config.maxRequestsPerSecond = requestsPerSecond;
    this.logger.info(`🔧 Rate limit updated to ${requestsPerSecond} req/s`);
  }

  /**
   * Get queue statistics
   */
  getStats(): Record<string, any> {
    const byPriority = {
      high: this.queue.filter((r) => r.priority === "high").length,
      normal: this.queue.filter((r) => r.priority === "normal").length,
      low: this.queue.filter((r) => r.priority === "low").length,
    };

    const byTenant: Record<string, number> = {};
    for (const request of this.queue) {
      byTenant[request.tenantId] = (byTenant[request.tenantId] || 0) + 1;
    }

    return {
      queueSize: this.queue.length,
      maxQueueSize: this.config.maxQueueSize,
      isProcessing: this.isProcessing,
      isPaused: this.paused,
      pausedUntil: this.paused ? new Date(this.pauseUntil) : null,
      rateLimit: `${this.config.maxRequestsPerSecond} req/s`,
      byPriority,
      byTenant,
    };
  }

  /**
   * Clear all pending requests
   */
  clear(reason: string = "Queue cleared"): void {
    const count = this.queue.length;

    for (const request of this.queue) {
      request.reject(new Error(reason));
    }

    this.queue = [];
    this.logger.info(`🗑️ Cleared ${count} pending requests: ${reason}`);
  }

  /**
   * Get current queue size
   */
  getQueueSize(): number {
    return this.queue.length;
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

// Singleton instance (shared across all tenants)
let throttlerInstance: RequestThrottler | null = null;

export function getRequestThrottler(): RequestThrottler {
  if (!throttlerInstance) {
    throttlerInstance = new RequestThrottler();
  }
  return throttlerInstance;
}

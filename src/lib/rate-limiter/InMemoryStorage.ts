/**
 * In-Memory Bucket Storage
 *
 * A simple in-memory storage implementation for single-instance deployments.
 * Uses a Map for O(1) access and supports atomic updates via synchronous operations.
 *
 * Note: For distributed systems, use RedisStorage instead.
 */

import type { BucketStorage, BucketState, RateLimitTier } from "./types";

/**
 * In-Memory Storage Implementation
 *
 * Thread-safety is achieved through JavaScript's single-threaded event loop:
 * - All operations complete synchronously within a single tick
 * - No interleaving of operations is possible
 * - Wrapped in Promises for interface compatibility
 */
export class InMemoryStorage implements BucketStorage {
  private readonly buckets: Map<string, BucketState> = new Map();
  private readonly ttlMs: number;
  private cleanupInterval: ReturnType<typeof setInterval> | null = null;

  /**
   * Creates a new InMemoryStorage instance
   *
   * @param ttlMs - Time-to-live for bucket entries in milliseconds (default: 5 minutes)
   * @param cleanupIntervalMs - Interval for cleanup of expired entries (default: 1 minute)
   *
   * @example
   * ```typescript
   * const storage = new InMemoryStorage();
   * const limiter = new TokenBucketRateLimiter({ storage });
   * ```
   */
  constructor(
    ttlMs: number = 5 * 60 * 1000,
    cleanupIntervalMs: number = 60 * 1000,
  ) {
    this.ttlMs = ttlMs;

    // Start cleanup interval to prevent memory leaks
    if (cleanupIntervalMs > 0) {
      this.cleanupInterval = setInterval(() => {
        this.cleanup();
      }, cleanupIntervalMs);

      // Don't block process exit
      if (this.cleanupInterval.unref) {
        this.cleanupInterval.unref();
      }
    }
  }

  /**
   * Generate a unique key for user + tier combination
   */
  private getKey(userId: string, tier: RateLimitTier): string {
    return `${userId}:${tier}`;
  }

  /**
   * Clean up expired entries to prevent memory leaks
   */
  private cleanup(): void {
    const now = Date.now();
    const expiredKeys: string[] = [];

    for (const [key, state] of this.buckets) {
      // Consider expired if not accessed within TTL
      if (now - state.lastRefillTime > this.ttlMs) {
        expiredKeys.push(key);
      }
    }

    for (const key of expiredKeys) {
      this.buckets.delete(key);
    }
  }

  /**
   * Stop the cleanup interval (for graceful shutdown)
   */
  destroy(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = null;
    }
  }

  /**
   * Get the current bucket state for a user
   */
  async get(userId: string, tier: RateLimitTier): Promise<BucketState | null> {
    const key = this.getKey(userId, tier);
    return this.buckets.get(key) ?? null;
  }

  /**
   * Set the bucket state for a user
   */
  async set(userId: string, state: BucketState): Promise<void> {
    const key = this.getKey(userId, state.tier);
    this.buckets.set(key, { ...state });
  }

  /**
   * Atomically get and update bucket state
   *
   * In-memory implementation is inherently atomic due to JS single-threading.
   * The updater function is called synchronously with the current state,
   * and the result is stored immediately.
   */
  async atomicUpdate(
    userId: string,
    tier: RateLimitTier,
    updater: (current: BucketState | null) => BucketState,
  ): Promise<BucketState> {
    const key = this.getKey(userId, tier);
    const current = this.buckets.get(key) ?? null;

    // Apply the update
    const updated = updater(current);

    // Store the result
    this.buckets.set(key, { ...updated });

    return updated;
  }

  /**
   * Delete a bucket
   */
  async delete(userId: string): Promise<void> {
    // Delete all tier entries for this user
    for (const tier of ["free", "pro", "enterprise"] as RateLimitTier[]) {
      const key = this.getKey(userId, tier);
      this.buckets.delete(key);
    }
  }

  /**
   * Clear all buckets
   */
  async clear(): Promise<void> {
    this.buckets.clear();
  }

  /**
   * Get the number of buckets (for monitoring/testing)
   */
  size(): number {
    return this.buckets.size;
  }
}

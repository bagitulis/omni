/**
 * Token Bucket Rate Limiter Types
 *
 * Type definitions for the multi-tier, distributed-ready rate limiter.
 */

/**
 * Rate limit tiers with predefined limits
 */
export type RateLimitTier = "free" | "pro" | "enterprise";

/**
 * Result of a rate limit check or consume operation
 */
export interface RateLimitResult {
  /** Whether the request is allowed */
  allowed: boolean;
  /** Number of remaining tokens */
  remaining: number;
  /** When the bucket will be fully refilled */
  resetAt: Date;
  /** Seconds until next token is available (only if not allowed) */
  retryAfter?: number;
}

/**
 * Configuration for a rate limit tier
 */
export interface TierConfig {
  /** Maximum tokens in the bucket (burst capacity) */
  maxTokens: number;
  /** Tokens refilled per minute */
  refillRatePerMinute: number;
}

/**
 * All tier configurations
 */
export type TierConfigs = Record<RateLimitTier, TierConfig>;

/**
 * Internal bucket state for persistence
 */
export interface BucketState {
  /** Current number of tokens (may be fractional) */
  tokens: number;
  /** Timestamp of last refill in milliseconds */
  lastRefillTime: number;
  /** The tier this bucket belongs to */
  tier: RateLimitTier;
}

/**
 * Rate limiter interface for check and consume operations
 */
export interface RateLimiter {
  /**
   * Check if a request would be allowed without consuming tokens
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @returns Promise resolving to the rate limit result
   */
  check(userId: string, tier: RateLimitTier): Promise<RateLimitResult>;

  /**
   * Consume tokens for a request
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @param tokens - Number of tokens to consume (default: 1)
   * @returns Promise resolving to the rate limit result
   */
  consume(
    userId: string,
    tier: RateLimitTier,
    tokens?: number,
  ): Promise<RateLimitResult>;
}

/**
 * Storage interface for bucket state persistence
 * Implementations can be in-memory, Redis, or any other backing store
 */
export interface BucketStorage {
  /**
   * Get the current bucket state for a user
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @returns Promise resolving to bucket state or null if not found
   */
  get(userId: string, tier: RateLimitTier): Promise<BucketState | null>;

  /**
   * Set the bucket state for a user
   * @param userId - Unique identifier for the user
   * @param state - The bucket state to persist
   * @returns Promise resolving when state is saved
   */
  set(userId: string, state: BucketState): Promise<void>;

  /**
   * Atomically get and update bucket state (for thread safety)
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @param updater - Function to update the bucket state
   * @returns Promise resolving to the updated bucket state
   */
  atomicUpdate(
    userId: string,
    tier: RateLimitTier,
    updater: (current: BucketState | null) => BucketState,
  ): Promise<BucketState>;

  /**
   * Delete a bucket (for cleanup/testing)
   * @param userId - Unique identifier for the user
   */
  delete(userId: string): Promise<void>;

  /**
   * Clear all buckets (for testing)
   */
  clear(): Promise<void>;
}

/**
 * Configuration options for the rate limiter
 */
export interface RateLimiterConfig {
  /** Storage backend for bucket state */
  storage: BucketStorage;
  /** Optional custom tier configurations (overrides defaults) */
  tierConfigs?: Partial<TierConfigs>;
}

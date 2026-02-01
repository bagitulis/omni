/**
 * Token Bucket Rate Limiter
 *
 * A multi-tier, distributed-ready rate limiter using the token bucket algorithm.
 * Supports smooth refill (tokens are added continuously based on elapsed time).
 *
 * Features:
 * - Multiple rate limit tiers (free, pro, enterprise)
 * - Smooth token refill (not burst refill)
 * - Thread-safe via atomic storage operations
 * - Pluggable storage backend (in-memory or Redis)
 * - Returns remaining tokens and reset time
 */

import type {
  RateLimiter,
  RateLimitResult,
  RateLimitTier,
  RateLimiterConfig,
  TierConfig,
  TierConfigs,
  BucketState,
  BucketStorage,
} from "./types";

/**
 * Default tier configurations
 * - free: 10 requests per minute
 * - pro: 100 requests per minute
 * - enterprise: 1000 requests per minute
 */
export const DEFAULT_TIER_CONFIGS: TierConfigs = {
  free: {
    maxTokens: 10,
    refillRatePerMinute: 10,
  },
  pro: {
    maxTokens: 100,
    refillRatePerMinute: 100,
  },
  enterprise: {
    maxTokens: 1000,
    refillRatePerMinute: 1000,
  },
};

/**
 * Token Bucket Rate Limiter Implementation
 *
 * This implementation uses a token bucket algorithm with smooth refill:
 * - Each user has a bucket that starts full
 * - Tokens are consumed on each request
 * - Tokens refill continuously based on elapsed time
 * - The bucket cannot exceed its maximum capacity
 */
export class TokenBucketRateLimiter implements RateLimiter {
  private readonly storage: BucketStorage;
  private readonly tierConfigs: TierConfigs;

  /**
   * Creates a new TokenBucketRateLimiter
   *
   * @param config - Configuration including storage backend and optional tier overrides
   *
   * @example
   * ```typescript
   * const limiter = new TokenBucketRateLimiter({
   *   storage: new InMemoryStorage(),
   * });
   *
   * const result = await limiter.consume('user-123', 'pro');
   * if (result.allowed) {
   *   // Process request
   * } else {
   *   // Return 429 with retryAfter
   * }
   * ```
   */
  constructor(config: RateLimiterConfig) {
    this.storage = config.storage;
    this.tierConfigs = {
      ...DEFAULT_TIER_CONFIGS,
      ...config.tierConfigs,
    };
  }

  /**
   * Get the tier configuration
   */
  private getTierConfig(tier: RateLimitTier): TierConfig {
    const config = this.tierConfigs[tier];
    if (!config) {
      throw new Error(`Unknown rate limit tier: ${tier}`);
    }
    return config;
  }

  /**
   * Calculate the current token count after smooth refill
   *
   * @param state - Current bucket state
   * @param config - Tier configuration
   * @param now - Current timestamp
   * @returns Updated token count (capped at maxTokens)
   */
  private calculateRefill(
    state: BucketState,
    config: TierConfig,
    now: number,
  ): number {
    const elapsedMs = now - state.lastRefillTime;
    const elapsedMinutes = elapsedMs / 60000;

    // Smooth refill: tokens = elapsed_time * refill_rate
    const tokensToAdd = elapsedMinutes * config.refillRatePerMinute;

    // Cap at maxTokens
    return Math.min(config.maxTokens, state.tokens + tokensToAdd);
  }

  /**
   * Calculate when the bucket will be fully refilled
   *
   * @param currentTokens - Current token count
   * @param config - Tier configuration
   * @returns Date when bucket will be full
   */
  private calculateResetAt(currentTokens: number, config: TierConfig): Date {
    if (currentTokens >= config.maxTokens) {
      return new Date();
    }

    const tokensNeeded = config.maxTokens - currentTokens;
    const minutesNeeded = tokensNeeded / config.refillRatePerMinute;
    const msNeeded = minutesNeeded * 60000;

    return new Date(Date.now() + msNeeded);
  }

  /**
   * Calculate seconds until next token is available
   *
   * @param currentTokens - Current token count
   * @param config - Tier configuration
   * @param tokensNeeded - Number of tokens needed
   * @returns Seconds until tokens are available
   */
  private calculateRetryAfter(
    currentTokens: number,
    config: TierConfig,
    tokensNeeded: number,
  ): number {
    if (currentTokens >= tokensNeeded) {
      return 0;
    }

    const deficit = tokensNeeded - currentTokens;
    const minutesNeeded = deficit / config.refillRatePerMinute;
    const secondsNeeded = minutesNeeded * 60;

    // Round up to next whole second
    return Math.ceil(secondsNeeded);
  }

  /**
   * Create a new bucket state initialized to full capacity
   *
   * @param tier - The rate limit tier
   * @param now - Current timestamp (for consistency with refill calculations)
   */
  private createNewBucket(tier: RateLimitTier, now: number): BucketState {
    const config = this.getTierConfig(tier);
    return {
      tokens: config.maxTokens,
      lastRefillTime: now,
      tier,
    };
  }

  /**
   * Check if a request would be allowed without consuming tokens
   *
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @returns Promise resolving to the rate limit result
   */
  async check(userId: string, tier: RateLimitTier): Promise<RateLimitResult> {
    if (!userId || typeof userId !== "string") {
      throw new Error("userId must be a non-empty string");
    }

    const config = this.getTierConfig(tier);
    const now = Date.now();

    // Get current state (don't modify)
    let state = await this.storage.get(userId, tier);

    if (!state) {
      // New user, would start with full bucket
      return {
        allowed: true,
        remaining: config.maxTokens,
        resetAt: new Date(now),
      };
    }

    // Calculate current tokens after refill
    const currentTokens = this.calculateRefill(state, config, now);

    return {
      allowed: currentTokens >= 1,
      remaining: Math.floor(currentTokens),
      resetAt: this.calculateResetAt(currentTokens, config),
      ...(currentTokens < 1 && {
        retryAfter: this.calculateRetryAfter(currentTokens, config, 1),
      }),
    };
  }

  /**
   * Consume tokens for a request
   *
   * @param userId - Unique identifier for the user
   * @param tier - The user's rate limit tier
   * @param tokens - Number of tokens to consume (default: 1)
   * @returns Promise resolving to the rate limit result
   */
  async consume(
    userId: string,
    tier: RateLimitTier,
    tokens: number = 1,
  ): Promise<RateLimitResult> {
    if (!userId || typeof userId !== "string") {
      throw new Error("userId must be a non-empty string");
    }

    if (tokens <= 0) {
      throw new Error("tokens must be a positive number");
    }

    const config = this.getTierConfig(tier);

    // Check if request can never be satisfied
    if (tokens > config.maxTokens) {
      return {
        allowed: false,
        remaining: 0,
        resetAt: new Date(Date.now() + 60000), // 1 minute from now
        retryAfter: Infinity,
      };
    }

    const now = Date.now();

    // Track whether consumption was allowed
    let wasAllowed = false;

    // Atomically update bucket state
    const updatedState = await this.storage.atomicUpdate(
      userId,
      tier,
      (current) => {
        // Initialize new bucket if doesn't exist
        if (!current) {
          current = this.createNewBucket(tier, now);
        }

        // Calculate current tokens after smooth refill
        const refilled = this.calculateRefill(current, config, now);

        // Check if we have enough tokens
        if (refilled >= tokens) {
          // Consume tokens
          wasAllowed = true;
          return {
            tokens: refilled - tokens,
            lastRefillTime: now,
            tier,
          };
        }

        // Not enough tokens, just update refill time without consuming
        wasAllowed = false;
        return {
          tokens: refilled,
          lastRefillTime: now,
          tier,
        };
      },
    );

    const remaining = Math.floor(Math.max(0, updatedState.tokens));

    return {
      allowed: wasAllowed,
      remaining,
      resetAt: this.calculateResetAt(updatedState.tokens, config),
      ...(!wasAllowed && {
        retryAfter: this.calculateRetryAfter(
          updatedState.tokens,
          config,
          tokens,
        ),
      }),
    };
  }
}

export type {
  RateLimiter,
  RateLimitResult,
  RateLimitTier,
  RateLimiterConfig,
  TierConfig,
  TierConfigs,
  BucketState,
  BucketStorage,
};

/**
 * Rate Limiter Implementation using Token Bucket Algorithm
 *
 * The token bucket algorithm maintains a bucket of tokens that is refilled
 * at a constant rate. Each request consumes tokens from the bucket.
 * If there are not enough tokens, the request is rejected.
 *
 * Key features:
 * - Allows burst traffic up to maxTokens
 * - Sustained rate is limited to refillRate tokens/second
 * - Snapshot-based state updates for predictable behavior
 */

/**
 * Configuration options for the RateLimiter.
 */
export interface RateLimiterConfig {
  /** Maximum number of tokens the bucket can hold. */
  maxTokens: number;
  /** Number of tokens added per second. */
  refillRate: number;
}

/**
 * Internal state representation.
 */
interface BucketState {
  tokens: number;
  lastRefillTime: number;
}

/**
 * Token Bucket Rate Limiter.
 */
export class RateLimiter {
  private readonly maxTokens: number;
  private readonly refillRate: number;
  private state: BucketState;

  /**
   * Creates a new RateLimiter instance.
   *
   * @param config - Configuration options
   * @throws {Error} If maxTokens or refillRate is not positive
   */
  constructor(config: RateLimiterConfig) {
    if (config.maxTokens <= 0) {
      throw new Error("maxTokens must be a positive number");
    }
    if (config.refillRate <= 0) {
      throw new Error("refillRate must be a positive number");
    }

    this.maxTokens = config.maxTokens;
    this.refillRate = config.refillRate;
    this.state = {
      tokens: config.maxTokens,
      lastRefillTime: this.getCurrentTime(),
    };
  }

  /**
   * Attempts to consume the specified number of tokens.
   *
   * @param tokens - Number of tokens to consume (default: 1)
   * @returns true if tokens were consumed successfully, false if rate limited
   */
  tryConsume(tokens: number = 1): boolean {
    if (tokens <= 0) {
      throw new Error("tokens must be a positive number");
    }

    const currentTime = this.getCurrentTime();
    const snapshot = this.captureState();
    const elapsedSeconds = (currentTime - snapshot.lastRefillTime) / 1000;
    const tokensToAdd = elapsedSeconds * this.refillRate;
    const newTokenCount = Math.min(
      this.maxTokens,
      snapshot.tokens + tokensToAdd,
    );

    if (newTokenCount < tokens) {
      this.updateState({
        tokens: newTokenCount,
        lastRefillTime: currentTime,
      });
      return false;
    }

    this.updateState({
      tokens: newTokenCount - tokens,
      lastRefillTime: currentTime,
    });

    return true;
  }

  /**
   * Gets the current number of available tokens.
   *
   * @returns Current number of available tokens (may be fractional)
   */
  getAvailableTokens(): number {
    const currentTime = this.getCurrentTime();
    const snapshot = this.captureState();
    const elapsedSeconds = (currentTime - snapshot.lastRefillTime) / 1000;
    const tokensToAdd = elapsedSeconds * this.refillRate;

    return Math.min(this.maxTokens, snapshot.tokens + tokensToAdd);
  }

  /**
   * Resets the rate limiter to full capacity.
   */
  reset(): void {
    this.updateState({
      tokens: this.maxTokens,
      lastRefillTime: this.getCurrentTime(),
    });
  }

  /**
   * Gets the maximum tokens (bucket capacity).
   */
  getMaxTokens(): number {
    return this.maxTokens;
  }

  /**
   * Gets the refill rate.
   */
  getRefillRate(): number {
    return this.refillRate;
  }

  /**
   * Calculates the time until the specified number of tokens is available.
   *
   * @param tokens - Number of tokens needed
   * @returns Milliseconds until tokens are available, 0 if already available
   */
  getTimeUntilTokens(tokens: number): number {
    const available = this.getAvailableTokens();
    if (available >= tokens) {
      return 0;
    }

    const tokensNeeded = tokens - available;
    const secondsNeeded = tokensNeeded / this.refillRate;
    return Math.ceil(secondsNeeded * 1000);
  }

  /**
   * Captures the current state as an immutable snapshot.
   */
  private captureState(): Readonly<BucketState> {
    return Object.freeze({ ...this.state });
  }

  /**
   * Atomically updates the state.
   */
  private updateState(newState: BucketState): void {
    this.state = { ...newState };
  }

  /**
   * Gets the current time in milliseconds.
   */
  private getCurrentTime(): number {
    return Date.now();
  }
}

/**
 * Creates a rate limiter for API calls.
 *
 * @param requestsPerSecond - Maximum sustained requests per second
 * @param burstMultiplier - Burst allowance multiplier (default: 2)
 * @returns Configured RateLimiter instance
 */
export function createApiRateLimiter(
  requestsPerSecond: number,
  burstMultiplier: number = 2,
): RateLimiter {
  return new RateLimiter({
    maxTokens: requestsPerSecond * burstMultiplier,
    refillRate: requestsPerSecond,
  });
}

/**
 * Creates a rate limiter with retry-after support.
 *
 * @returns limiter instance and getRetryAfter helper (in seconds)
 */
export function createRateLimiterWithRetry(config: RateLimiterConfig): {
  limiter: RateLimiter;
  getRetryAfter: (tokens?: number) => number;
} {
  const limiter = new RateLimiter(config);

  return {
    limiter,
    getRetryAfter: (tokens: number = 1) => {
      const waitMs = limiter.getTimeUntilTokens(tokens);
      return Math.ceil(waitMs / 1000);
    },
  };
}

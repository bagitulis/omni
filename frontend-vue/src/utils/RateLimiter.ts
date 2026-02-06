/**
 * Rate Limiter Implementation using Token Bucket Algorithm
 *
 * The token bucket algorithm works by maintaining a "bucket" of tokens that gets
 * refilled at a constant rate. Each request consumes tokens from the bucket.
 * If there aren't enough tokens, the request is rejected.
 *
 * Key features:
 * - Allows burst traffic up to maxTokens
 * - Sustained rate limited to refillRate tokens/second
 * - Thread-safe through atomic operations and state snapshots
 *
 * @module RateLimiter
 */

/**
 * Configuration options for the RateLimiter
 */
export interface RateLimiterConfig {
  /** Maximum number of tokens the bucket can hold */
  maxTokens: number;
  /** Number of tokens added per second */
  refillRate: number;
}

/**
 * Internal state representation for atomic operations
 */
interface BucketState {
  tokens: number;
  lastRefillTime: number;
}

/**
 * Token Bucket Rate Limiter
 *
 * Implements the token bucket algorithm for rate limiting.
 * This implementation is conceptually thread-safe by using
 * snapshot-based state updates with optimistic concurrency control.
 *
 * @example
 * ```typescript
 * // Create a rate limiter allowing 10 requests per second with burst of 20
 * const limiter = new RateLimiter({ maxTokens: 20, refillRate: 10 });
 *
 * // Try to consume a token for a request
 * if (limiter.tryConsume(1)) {
 *   // Request allowed - proceed
 *   handleRequest();
 * } else {
 *   // Rate limited - reject or queue
 *   return { error: 'Too many requests' };
 * }
 * ```
 */
export class RateLimiter {
  private readonly maxTokens: number;
  private readonly refillRate: number;
  private state: BucketState;

  /**
   * Creates a new RateLimiter instance
   *
   * @param config - Configuration options
   * @throws {Error} If maxTokens or refillRate is not positive
   *
   * @example
   * ```typescript
   * const limiter = new RateLimiter({
   *   maxTokens: 100,    // Allow burst of 100 requests
   *   refillRate: 10     // Sustain 10 requests per second
   * });
   * ```
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
   * Attempts to consume the specified number of tokens
   *
   * This method is conceptually thread-safe. It uses a snapshot-based
   * approach where the state is read, computed, and atomically updated.
   * In JavaScript's single-threaded event loop, this prevents race conditions
   * between async operations that yield to the event loop.
   *
   * @param tokens - Number of tokens to consume (default: 1)
   * @returns true if tokens were consumed successfully, false if rate limited
   *
   * @example
   * ```typescript
   * // Single token consumption
   * if (limiter.tryConsume()) {
   *   console.log('Request allowed');
   * }
   *
   * // Batch consumption (e.g., for heavy operations)
   * if (limiter.tryConsume(5)) {
   *   console.log('Heavy operation allowed');
   * }
   * ```
   */
  tryConsume(tokens: number = 1): boolean {
    if (tokens <= 0) {
      throw new Error("tokens must be a positive number");
    }

    // Atomic read: capture current state snapshot
    const currentTime = this.getCurrentTime();
    const snapshot = this.captureState();

    // Calculate refill
    const elapsedSeconds = (currentTime - snapshot.lastRefillTime) / 1000;
    const tokensToAdd = elapsedSeconds * this.refillRate;
    const newTokenCount = Math.min(
      this.maxTokens,
      snapshot.tokens + tokensToAdd,
    );

    // Check if we have enough tokens
    if (newTokenCount < tokens) {
      // Not enough tokens - update lastRefillTime but don't consume
      // This ensures accurate refill calculations for future attempts
      this.updateState({
        tokens: newTokenCount,
        lastRefillTime: currentTime,
      });
      return false;
    }

    // Atomic write: consume tokens and update state
    this.updateState({
      tokens: newTokenCount - tokens,
      lastRefillTime: currentTime,
    });

    return true;
  }

  /**
   * Gets the current number of available tokens
   *
   * This method recalculates the token count based on elapsed time
   * since the last operation, providing an accurate real-time value.
   *
   * @returns Current number of available tokens (may be fractional)
   *
   * @example
   * ```typescript
   * const available = limiter.getAvailableTokens();
   * console.log(`${available} tokens available`);
   *
   * // Check if operation is likely to succeed
   * if (limiter.getAvailableTokens() >= 5) {
   *   limiter.tryConsume(5);
   * }
   * ```
   */
  getAvailableTokens(): number {
    const currentTime = this.getCurrentTime();
    const snapshot = this.captureState();

    const elapsedSeconds = (currentTime - snapshot.lastRefillTime) / 1000;
    const tokensToAdd = elapsedSeconds * this.refillRate;

    return Math.min(this.maxTokens, snapshot.tokens + tokensToAdd);
  }

  /**
   * Resets the rate limiter to full capacity
   *
   * Useful for:
   * - Clearing rate limits after user authentication changes
   * - Testing scenarios
   * - Administrative override of rate limits
   *
   * @example
   * ```typescript
   * // After user upgrades to premium tier
   * limiter.reset();
   * ```
   */
  reset(): void {
    this.updateState({
      tokens: this.maxTokens,
      lastRefillTime: this.getCurrentTime(),
    });
  }

  /**
   * Gets the maximum tokens (bucket capacity)
   *
   * @returns The maximum number of tokens the bucket can hold
   */
  getMaxTokens(): number {
    return this.maxTokens;
  }

  /**
   * Gets the refill rate
   *
   * @returns The number of tokens added per second
   */
  getRefillRate(): number {
    return this.refillRate;
  }

  /**
   * Calculates the time until the specified number of tokens will be available
   *
   * @param tokens - Number of tokens needed
   * @returns Milliseconds until tokens are available, 0 if already available
   *
   * @example
   * ```typescript
   * const waitTime = limiter.getTimeUntilTokens(5);
   * if (waitTime > 0) {
   *   console.log(`Please wait ${waitTime}ms before retrying`);
   * }
   * ```
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
   * Captures the current state as an immutable snapshot
   * This is the core of our thread-safety mechanism
   *
   * @private
   */
  private captureState(): Readonly<BucketState> {
    // Return a frozen copy to prevent mutation
    return Object.freeze({ ...this.state });
  }

  /**
   * Atomically updates the state
   * In JavaScript, this assignment is atomic
   *
   * @private
   */
  private updateState(newState: BucketState): void {
    // Single atomic assignment
    this.state = { ...newState };
  }

  /**
   * Gets the current time in milliseconds
   * Extracted for easier testing/mocking
   *
   * @private
   */
  private getCurrentTime(): number {
    return Date.now();
  }
}

/**
 * Creates a rate limiter for API calls
 *
 * @param requestsPerSecond - Maximum sustained requests per second
 * @param burstMultiplier - How many times the sustained rate to allow for bursts (default: 2)
 * @returns Configured RateLimiter instance
 *
 * @example
 * ```typescript
 * // Allow 10 req/s sustained, 20 req/s burst
 * const apiLimiter = createApiRateLimiter(10);
 *
 * // Allow 100 req/s sustained, 500 req/s burst
 * const highVolumeLimiter = createApiRateLimiter(100, 5);
 * ```
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
 * Creates a rate limiter with retry-after support
 *
 * Wraps a RateLimiter to provide Retry-After header values
 * for HTTP 429 responses.
 *
 * @example
 * ```typescript
 * const { limiter, getRetryAfter } = createRateLimiterWithRetry({
 *   maxTokens: 100,
 *   refillRate: 10
 * });
 *
 * if (!limiter.tryConsume()) {
 *   res.setHeader('Retry-After', getRetryAfter(1));
 *   res.status(429).send('Too Many Requests');
 * }
 * ```
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
      // Return seconds (rounded up) for Retry-After header
      return Math.ceil(waitMs / 1000);
    },
  };
}

// ============================================================================
// EXAMPLE USAGE
// ============================================================================

/*
// Basic usage
const limiter = new RateLimiter({ maxTokens: 10, refillRate: 2 });

// API endpoint protection
async function handleApiRequest(req: Request, res: Response) {
  if (!limiter.tryConsume()) {
    const retryAfter = limiter.getTimeUntilTokens(1);
    res.setHeader('Retry-After', Math.ceil(retryAfter / 1000));
    return res.status(429).json({
      error: 'Too Many Requests',
      retryAfterMs: retryAfter
    });
  }

  // Process the request
  return res.json({ success: true });
}

// Per-user rate limiting with a Map
const userLimiters = new Map<string, RateLimiter>();

function getUserLimiter(userId: string): RateLimiter {
  if (!userLimiters.has(userId)) {
    userLimiters.set(userId, createApiRateLimiter(10));
  }
  return userLimiters.get(userId)!;
}

// Weighted rate limiting (some operations cost more)
const weightedLimiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });

function processOperation(type: 'light' | 'heavy') {
  const cost = type === 'light' ? 1 : 10;

  if (weightedLimiter.tryConsume(cost)) {
    console.log(`Processing ${type} operation`);
  } else {
    console.log(`Rate limited. Wait ${weightedLimiter.getTimeUntilTokens(cost)}ms`);
  }
}
*/

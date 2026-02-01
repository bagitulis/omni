/**
 * Token Bucket Rate Limiter
 *
 * Implements the token bucket algorithm for rate limiting.
 * Tokens are added at a constant rate up to a maximum capacity.
 * Requests consume tokens; if insufficient tokens, request is rejected.
 *
 * @example
 * ```typescript
 * const limiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });
 *
 * if (limiter.tryConsume(5)) {
 *   // Request allowed
 * } else {
 *   // Rate limited
 * }
 * ```
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
 * Snapshot of the rate limiter's current state
 */
export interface RateLimiterState {
  /** Current available tokens */
  availableTokens: number;
  /** Maximum bucket capacity */
  maxTokens: number;
  /** Tokens added per second */
  refillRate: number;
  /** Timestamp of last refill calculation */
  lastRefillTime: number;
}

/**
 * Token Bucket Rate Limiter Implementation
 *
 * This class implements the token bucket algorithm, which is widely used
 * for rate limiting in distributed systems. The algorithm allows for
 * burst traffic while maintaining an average rate limit.
 *
 * Thread-safety considerations:
 * - JavaScript is single-threaded, but async operations can interleave
 * - All state mutations are atomic (single assignment operations)
 * - The refill calculation uses timestamps to handle concurrent access correctly
 * - No locks needed due to JS event loop, but design supports conceptual concurrency
 */
export class RateLimiter {
  private readonly maxTokens: number;
  private readonly refillRate: number;
  private tokens: number;
  private lastRefillTime: number;

  /**
   * Creates a new RateLimiter instance
   *
   * @param config - Configuration for the rate limiter
   * @throws {Error} If maxTokens or refillRate is not positive
   *
   * @example
   * ```typescript
   * // Allow 100 requests with 10 tokens refilled per second
   * const limiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });
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
    this.tokens = config.maxTokens; // Start with full bucket
    this.lastRefillTime = Date.now();
  }

  /**
   * Refills the token bucket based on elapsed time
   *
   * This method calculates how many tokens should be added based on
   * the time elapsed since the last refill. The bucket cannot exceed
   * maxTokens capacity.
   *
   * @private
   */
  private refill(): void {
    const now = Date.now();
    const elapsedMs = now - this.lastRefillTime;
    const elapsedSeconds = elapsedMs / 1000;

    // Calculate tokens to add based on elapsed time
    const tokensToAdd = elapsedSeconds * this.refillRate;

    // Update tokens, capped at maxTokens
    // Using Math.min ensures we never exceed capacity
    this.tokens = Math.min(this.maxTokens, this.tokens + tokensToAdd);

    // Update last refill time
    // This is the key to handling "concurrent" calls correctly:
    // Each call sees accurate elapsed time since last calculation
    this.lastRefillTime = now;
  }

  /**
   * Attempts to consume the specified number of tokens
   *
   * This method first refills the bucket based on elapsed time,
   * then attempts to consume the requested tokens. If sufficient
   * tokens are available, they are consumed and true is returned.
   * Otherwise, no tokens are consumed and false is returned.
   *
   * @param tokens - Number of tokens to consume (must be positive)
   * @returns true if tokens were consumed, false if insufficient tokens
   * @throws {Error} If tokens is not a positive number
   *
   * @example
   * ```typescript
   * const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
   *
   * // Consume 5 tokens
   * if (limiter.tryConsume(5)) {
   *   console.log('Request allowed');
   * } else {
   *   console.log('Rate limited - try again later');
   * }
   * ```
   */
  tryConsume(tokens: number): boolean {
    if (tokens <= 0) {
      throw new Error("tokens must be a positive number");
    }

    if (tokens > this.maxTokens) {
      // Request can never be satisfied - always reject
      return false;
    }

    // Refill bucket based on elapsed time
    this.refill();

    // Check if we have enough tokens
    if (this.tokens >= tokens) {
      // Atomic consumption - single assignment prevents race conditions
      this.tokens -= tokens;
      return true;
    }

    // Insufficient tokens - reject without consuming
    return false;
  }

  /**
   * Gets the current number of available tokens
   *
   * This method refills the bucket first to provide an accurate count.
   * Note that the returned value may be stale by the time it's used
   * due to other concurrent operations.
   *
   * @returns Current number of available tokens (may include fractional tokens)
   *
   * @example
   * ```typescript
   * const limiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });
   * limiter.tryConsume(30);
   * console.log(limiter.getAvailableTokens()); // ~70 (may vary based on timing)
   * ```
   */
  getAvailableTokens(): number {
    this.refill();
    return this.tokens;
  }

  /**
   * Resets the rate limiter to full capacity
   *
   * This method restores the bucket to its maximum token capacity
   * and resets the refill timer. Useful for testing or when rate
   * limiting conditions need to be cleared.
   *
   * @example
   * ```typescript
   * const limiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });
   * limiter.tryConsume(100); // Empty the bucket
   * limiter.reset(); // Back to full capacity
   * console.log(limiter.getAvailableTokens()); // 100
   * ```
   */
  reset(): void {
    this.tokens = this.maxTokens;
    this.lastRefillTime = Date.now();
  }

  /**
   * Gets a snapshot of the current rate limiter state
   *
   * Useful for debugging, monitoring, or serialization.
   *
   * @returns Current state of the rate limiter
   */
  getState(): RateLimiterState {
    this.refill();
    return {
      availableTokens: this.tokens,
      maxTokens: this.maxTokens,
      refillRate: this.refillRate,
      lastRefillTime: this.lastRefillTime,
    };
  }

  /**
   * Calculates when enough tokens will be available
   *
   * @param tokens - Number of tokens needed
   * @returns Milliseconds until tokens are available, or 0 if available now
   *
   * @example
   * ```typescript
   * const limiter = new RateLimiter({ maxTokens: 100, refillRate: 10 });
   * limiter.tryConsume(100);
   * const waitTime = limiter.getWaitTime(50);
   * console.log(`Wait ${waitTime}ms before retrying`);
   * ```
   */
  getWaitTime(tokens: number): number {
    if (tokens <= 0 || tokens > this.maxTokens) {
      return tokens > this.maxTokens ? Infinity : 0;
    }

    this.refill();

    if (this.tokens >= tokens) {
      return 0;
    }

    const tokensNeeded = tokens - this.tokens;
    const secondsNeeded = tokensNeeded / this.refillRate;
    return Math.ceil(secondsNeeded * 1000);
  }
}

// ============================================================================
// EXAMPLE USAGE
// ============================================================================

/**
 * Example: API Rate Limiting
 *
 * Demonstrates how to use the RateLimiter for API request throttling.
 */
async function exampleApiRateLimiting(): Promise<void> {
  console.log("=== API Rate Limiting Example ===\n");

  // Create a limiter: 10 requests max, refill 2 per second
  const apiLimiter = new RateLimiter({
    maxTokens: 10,
    refillRate: 2,
  });

  // Simulate API requests
  async function makeApiRequest(requestId: number): Promise<void> {
    if (apiLimiter.tryConsume(1)) {
      console.log(
        `Request ${requestId}: ✅ Allowed (${apiLimiter.getAvailableTokens().toFixed(1)} tokens left)`,
      );
    } else {
      const waitTime = apiLimiter.getWaitTime(1);
      console.log(`Request ${requestId}: ❌ Rate limited (wait ${waitTime}ms)`);
    }
  }

  // Make 15 rapid requests (should allow ~10, reject ~5)
  for (let i = 1; i <= 15; i++) {
    await makeApiRequest(i);
  }

  console.log("\nWaiting 2 seconds for token refill...\n");
  await new Promise((resolve) => setTimeout(resolve, 2000));

  // Try again after waiting
  for (let i = 16; i <= 20; i++) {
    await makeApiRequest(i);
  }
}

/**
 * Example: Burst Traffic with Sustained Rate
 *
 * Demonstrates allowing burst traffic while maintaining average rate.
 */
function exampleBurstTraffic(): void {
  console.log("\n=== Burst Traffic Example ===\n");

  // Allow bursts of 100, but sustain only 10/second average
  const burstLimiter = new RateLimiter({
    maxTokens: 100,
    refillRate: 10,
  });

  // Burst: consume 50 tokens at once
  const burstSuccess = burstLimiter.tryConsume(50);
  console.log(`Burst of 50: ${burstSuccess ? "✅ Allowed" : "❌ Rejected"}`);
  console.log(
    `Remaining: ${burstLimiter.getAvailableTokens().toFixed(1)} tokens`,
  );

  // Another burst
  const burst2Success = burstLimiter.tryConsume(50);
  console.log(`Burst of 50: ${burst2Success ? "✅ Allowed" : "❌ Rejected"}`);
  console.log(
    `Remaining: ${burstLimiter.getAvailableTokens().toFixed(1)} tokens`,
  );

  // Third burst should fail
  const burst3Success = burstLimiter.tryConsume(50);
  console.log(`Burst of 50: ${burst3Success ? "✅ Allowed" : "❌ Rejected"}`);
  console.log(`Wait time for 50 tokens: ${burstLimiter.getWaitTime(50)}ms`);
}

/**
 * Example: Different Cost Operations
 *
 * Demonstrates varying token costs for different operations.
 */
function exampleDifferentCosts(): void {
  console.log("\n=== Different Operation Costs Example ===\n");

  const operationLimiter = new RateLimiter({
    maxTokens: 100,
    refillRate: 20,
  });

  // Define operation costs
  const operationCosts = {
    read: 1,
    write: 5,
    delete: 10,
    bulkOperation: 25,
  };

  function performOperation(op: keyof typeof operationCosts): void {
    const cost = operationCosts[op];
    if (operationLimiter.tryConsume(cost)) {
      console.log(`${op} (cost: ${cost}): ✅ Allowed`);
    } else {
      console.log(`${op} (cost: ${cost}): ❌ Rate limited`);
    }
  }

  performOperation("read");
  performOperation("write");
  performOperation("bulkOperation");
  performOperation("delete");
  performOperation("bulkOperation");
  performOperation("bulkOperation");

  console.log(
    `\nFinal state: ${operationLimiter.getAvailableTokens().toFixed(1)} tokens available`,
  );
}

// Run examples if this file is executed directly
if (require.main === module) {
  (async () => {
    await exampleApiRateLimiting();
    exampleBurstTraffic();
    exampleDifferentCosts();
  })();
}

export default RateLimiter;

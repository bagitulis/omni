/**
 * Token Bucket Rate Limiter Tests
 *
 * Comprehensive unit tests demonstrating rate limiting behavior.
 * Tests cover:
 * - Basic rate limiting (allow/reject)
 * - Multiple tiers (free, pro, enterprise)
 * - Smooth refill behavior
 * - Concurrent request handling
 * - Error handling
 * - Storage operations
 */

import {
  TokenBucketRateLimiter,
  DEFAULT_TIER_CONFIGS,
} from "./TokenBucketRateLimiter";
import { InMemoryStorage } from "./InMemoryStorage";
import type {
  RateLimitResult,
  BucketStorage,
  BucketState,
  RateLimitTier,
} from "./types";

/**
 * Simple test runner for demonstration
 * In production, use Jest, Vitest, or similar
 */
class TestRunner {
  private passed = 0;
  private failed = 0;
  private tests: Array<{ name: string; fn: () => Promise<void> }> = [];

  describe(name: string, fn: () => void): void {
    console.log(`\n📦 ${name}`);
    fn();
  }

  it(name: string, fn: () => Promise<void>): void {
    this.tests.push({ name, fn });
  }

  async run(): Promise<void> {
    for (const test of this.tests) {
      try {
        await test.fn();
        console.log(`  ✅ ${test.name}`);
        this.passed++;
      } catch (error) {
        console.log(`  ❌ ${test.name}`);
        console.log(`     Error: ${(error as Error).message}`);
        this.failed++;
      }
    }

    console.log(`\n📊 Results: ${this.passed} passed, ${this.failed} failed\n`);
  }
}

function expect<T>(actual: T) {
  return {
    toBe(expected: T): void {
      if (actual !== expected) {
        throw new Error(`Expected ${expected}, got ${actual}`);
      }
    },
    toBeGreaterThan(expected: number): void {
      if ((actual as number) <= expected) {
        throw new Error(`Expected ${actual} to be greater than ${expected}`);
      }
    },
    toBeGreaterThanOrEqual(expected: number): void {
      if ((actual as number) < expected) {
        throw new Error(`Expected ${actual} to be >= ${expected}`);
      }
    },
    toBeLessThan(expected: number): void {
      if ((actual as number) >= expected) {
        throw new Error(`Expected ${actual} to be less than ${expected}`);
      }
    },
    toBeLessThanOrEqual(expected: number): void {
      if ((actual as number) > expected) {
        throw new Error(`Expected ${actual} to be <= ${expected}`);
      }
    },
    toBeTruthy(): void {
      if (!actual) {
        throw new Error(`Expected truthy, got ${actual}`);
      }
    },
    toBeFalsy(): void {
      if (actual) {
        throw new Error(`Expected falsy, got ${actual}`);
      }
    },
    toBeInstanceOf(expected: new (...args: unknown[]) => unknown): void {
      if (!(actual instanceof expected)) {
        throw new Error(`Expected instance of ${expected.name}`);
      }
    },
    toBeDefined(): void {
      if (actual === undefined) {
        throw new Error("Expected defined, got undefined");
      }
    },
    toBeUndefined(): void {
      if (actual !== undefined) {
        throw new Error(`Expected undefined, got ${actual}`);
      }
    },
  };
}

async function expectAsync(fn: () => Promise<unknown>) {
  return {
    async toReject(): Promise<void> {
      try {
        await fn();
        throw new Error("Expected function to throw");
      } catch {
        // Expected
      }
    },
  };
}

// Helper to simulate time passing
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ============================================================================
// TESTS
// ============================================================================

async function runTests(): Promise<void> {
  const runner = new TestRunner();

  // ---------------------------------------------------------------------------
  // Basic Rate Limiting Tests
  // ---------------------------------------------------------------------------
  runner.describe("Basic Rate Limiting", () => {
    runner.it("should allow requests within limit", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      const result = await limiter.consume("user-1", "free");

      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(9); // 10 - 1
      expect(result.resetAt).toBeInstanceOf(Date);
    });

    runner.it("should reject requests exceeding limit", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // Consume all 10 free tier tokens
      for (let i = 0; i < 10; i++) {
        await limiter.consume("user-1", "free");
      }

      // 11th request should be rejected
      const result = await limiter.consume("user-1", "free");

      expect(result.allowed).toBe(false);
      expect(result.remaining).toBe(0);
      expect(result.retryAfter).toBeDefined();
      expect(result.retryAfter).toBeGreaterThan(0);
    });

    runner.it("should allow checking without consuming", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // Consume some tokens
      await limiter.consume("user-1", "free");
      await limiter.consume("user-1", "free");

      // Check should not consume
      const checkResult = await limiter.check("user-1", "free");
      expect(checkResult.allowed).toBe(true);
      expect(checkResult.remaining).toBe(8);

      // Consume should still have 8 available
      const consumeResult = await limiter.consume("user-1", "free");
      expect(consumeResult.remaining).toBe(7);
    });
  });

  // ---------------------------------------------------------------------------
  // Tier Configuration Tests
  // ---------------------------------------------------------------------------
  runner.describe("Tier Configurations", () => {
    runner.it("should apply free tier limits (10/min)", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      const result = await limiter.check("user-1", "free");
      expect(result.remaining).toBe(10);
    });

    runner.it("should apply pro tier limits (100/min)", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      const result = await limiter.check("user-1", "pro");
      expect(result.remaining).toBe(100);
    });

    runner.it("should apply enterprise tier limits (1000/min)", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      const result = await limiter.check("user-1", "enterprise");
      expect(result.remaining).toBe(1000);
    });

    runner.it("should allow custom tier configurations", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({
        storage,
        tierConfigs: {
          free: { maxTokens: 5, refillRatePerMinute: 5 },
        },
      });

      const result = await limiter.check("user-1", "free");
      expect(result.remaining).toBe(5);
    });
  });

  // ---------------------------------------------------------------------------
  // Smooth Refill Tests
  // ---------------------------------------------------------------------------
  runner.describe("Smooth Refill Behavior", () => {
    runner.it("should refill tokens based on elapsed time", async () => {
      const storage = new InMemoryStorage();
      // Custom config: 10 tokens, 600 tokens/minute = 10 tokens/second
      const limiter = new TokenBucketRateLimiter({
        storage,
        tierConfigs: {
          free: { maxTokens: 10, refillRatePerMinute: 600 },
        },
      });

      // Consume all tokens
      for (let i = 0; i < 10; i++) {
        await limiter.consume("user-1", "free");
      }

      let result = await limiter.check("user-1", "free");
      expect(result.remaining).toBe(0);

      // Wait 500ms (should refill ~5 tokens at 10/sec)
      await sleep(500);

      result = await limiter.check("user-1", "free");
      expect(result.remaining).toBeGreaterThanOrEqual(4); // Allow some timing variance
      expect(result.remaining).toBeLessThanOrEqual(6);
    });

    runner.it("should not exceed max tokens during refill", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({
        storage,
        tierConfigs: {
          free: { maxTokens: 10, refillRatePerMinute: 6000 }, // Very fast refill
        },
      });

      // Consume some tokens
      await limiter.consume("user-1", "free");

      // Wait for refill
      await sleep(100);

      const result = await limiter.check("user-1", "free");
      expect(result.remaining).toBeLessThanOrEqual(10);
    });
  });

  // ---------------------------------------------------------------------------
  // Multi-token Consumption Tests
  // ---------------------------------------------------------------------------
  runner.describe("Multi-token Consumption", () => {
    runner.it("should allow consuming multiple tokens at once", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      const result = await limiter.consume("user-1", "free", 5);

      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(5); // 10 - 5
    });

    runner.it(
      "should reject if not enough tokens for multi-consume",
      async () => {
        const storage = new InMemoryStorage();
        const limiter = new TokenBucketRateLimiter({ storage });

        // Consume 8 tokens
        await limiter.consume("user-1", "free", 8);

        // Try to consume 5 more (only 2 left)
        const result = await limiter.consume("user-1", "free", 5);

        expect(result.allowed).toBe(false);
        expect(result.remaining).toBe(2);
        expect(result.retryAfter).toBeDefined();
      },
    );

    runner.it("should reject requests exceeding max bucket size", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // Try to consume more than max (10 for free tier)
      const result = await limiter.consume("user-1", "free", 15);

      expect(result.allowed).toBe(false);
    });
  });

  // ---------------------------------------------------------------------------
  // User Isolation Tests
  // ---------------------------------------------------------------------------
  runner.describe("User Isolation", () => {
    runner.it("should maintain separate limits per user", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // User 1 consumes tokens
      await limiter.consume("user-1", "free", 8);

      // User 2 should have full bucket
      const result = await limiter.check("user-2", "free");
      expect(result.remaining).toBe(10);
    });

    runner.it(
      "should maintain separate limits per tier for same user",
      async () => {
        const storage = new InMemoryStorage();
        const limiter = new TokenBucketRateLimiter({ storage });

        // Consume from free tier
        await limiter.consume("user-1", "free", 10);

        // Pro tier should be unaffected
        const result = await limiter.check("user-1", "pro");
        expect(result.remaining).toBe(100);
      },
    );
  });

  // ---------------------------------------------------------------------------
  // Concurrent Request Tests
  // ---------------------------------------------------------------------------
  runner.describe("Concurrent Requests", () => {
    runner.it("should handle concurrent requests correctly", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // Fire 15 concurrent requests
      const requests = Array.from({ length: 15 }, () =>
        limiter.consume("user-1", "free"),
      );

      const results = await Promise.all(requests);

      // Exactly 10 should be allowed (free tier limit)
      const allowed = results.filter((r) => r.allowed).length;
      const rejected = results.filter((r) => !r.allowed).length;

      expect(allowed).toBe(10);
      expect(rejected).toBe(5);
    });

    runner.it(
      "should handle concurrent requests for multiple users",
      async () => {
        const storage = new InMemoryStorage();
        const limiter = new TokenBucketRateLimiter({ storage });

        // 5 users, 15 requests each
        const allRequests: Promise<RateLimitResult>[] = [];
        for (let user = 1; user <= 5; user++) {
          for (let i = 0; i < 15; i++) {
            allRequests.push(limiter.consume(`user-${user}`, "free"));
          }
        }

        const results = await Promise.all(allRequests);

        // Each user should have exactly 10 allowed, 5 rejected
        const byUser = new Map<string, RateLimitResult[]>();
        for (let user = 1; user <= 5; user++) {
          const startIdx = (user - 1) * 15;
          byUser.set(`user-${user}`, results.slice(startIdx, startIdx + 15));
        }

        for (const [, userResults] of byUser) {
          const allowed = userResults.filter((r) => r.allowed).length;
          expect(allowed).toBe(10);
        }
      },
    );
  });

  // ---------------------------------------------------------------------------
  // Error Handling Tests
  // ---------------------------------------------------------------------------
  runner.describe("Error Handling", () => {
    runner.it("should reject empty userId", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      let threw = false;
      try {
        await limiter.consume("", "free");
      } catch {
        threw = true;
      }
      expect(threw).toBe(true);
    });

    runner.it("should reject non-positive token count", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      let threw = false;
      try {
        await limiter.consume("user-1", "free", 0);
      } catch {
        threw = true;
      }
      expect(threw).toBe(true);
    });

    runner.it("should reject negative token count", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      let threw = false;
      try {
        await limiter.consume("user-1", "free", -5);
      } catch {
        threw = true;
      }
      expect(threw).toBe(true);
    });
  });

  // ---------------------------------------------------------------------------
  // Storage Tests
  // ---------------------------------------------------------------------------
  runner.describe("InMemoryStorage", () => {
    runner.it("should store and retrieve bucket state", async () => {
      const storage = new InMemoryStorage();

      const state: BucketState = {
        tokens: 5,
        lastRefillTime: Date.now(),
        tier: "free",
      };

      await storage.set("user-1", state);
      const retrieved = await storage.get("user-1", "free");

      expect(retrieved).toBeDefined();
      expect(retrieved?.tokens).toBe(5);
      expect(retrieved?.tier).toBe("free");
    });

    runner.it("should return null for non-existent bucket", async () => {
      const storage = new InMemoryStorage();

      const result = await storage.get("non-existent", "free");
      expect(result).toBe(null);
    });

    runner.it("should delete bucket", async () => {
      const storage = new InMemoryStorage();

      await storage.set("user-1", {
        tokens: 5,
        lastRefillTime: Date.now(),
        tier: "free",
      });

      await storage.delete("user-1");
      const result = await storage.get("user-1", "free");

      expect(result).toBe(null);
    });

    runner.it("should clear all buckets", async () => {
      const storage = new InMemoryStorage();

      await storage.set("user-1", {
        tokens: 5,
        lastRefillTime: Date.now(),
        tier: "free",
      });
      await storage.set("user-2", {
        tokens: 10,
        lastRefillTime: Date.now(),
        tier: "pro",
      });

      await storage.clear();

      expect(await storage.get("user-1", "free")).toBe(null);
      expect(await storage.get("user-2", "pro")).toBe(null);
    });

    runner.it("should perform atomic updates", async () => {
      const storage = new InMemoryStorage();

      // Initial state
      await storage.set("user-1", {
        tokens: 10,
        lastRefillTime: Date.now(),
        tier: "free",
      });

      // Atomic update
      const result = await storage.atomicUpdate("user-1", "free", (current) => {
        return {
          tokens: (current?.tokens ?? 0) - 3,
          lastRefillTime: Date.now(),
          tier: "free",
        };
      });

      expect(result.tokens).toBe(7);

      const retrieved = await storage.get("user-1", "free");
      expect(retrieved?.tokens).toBe(7);
    });
  });

  // ---------------------------------------------------------------------------
  // Reset Time and Retry After Tests
  // ---------------------------------------------------------------------------
  runner.describe("Reset Time and Retry After", () => {
    runner.it("should return correct resetAt time", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({ storage });

      // Consume all tokens
      for (let i = 0; i < 10; i++) {
        await limiter.consume("user-1", "free");
      }

      const result = await limiter.check("user-1", "free");

      // Reset time should be in the future
      expect(result.resetAt.getTime()).toBeGreaterThan(Date.now());

      // Should take 1 minute to fully refill at 10 tokens/min
      const expectedResetTime = Date.now() + 60000; // 60 seconds
      const actualResetTime = result.resetAt.getTime();

      // Allow 5 second tolerance
      expect(actualResetTime).toBeGreaterThan(expectedResetTime - 5000);
      expect(actualResetTime).toBeLessThan(expectedResetTime + 5000);
    });

    runner.it("should return correct retryAfter seconds", async () => {
      const storage = new InMemoryStorage();
      const limiter = new TokenBucketRateLimiter({
        storage,
        tierConfigs: {
          free: { maxTokens: 10, refillRatePerMinute: 60 }, // 1 token per second
        },
      });

      // Consume all tokens
      for (let i = 0; i < 10; i++) {
        await limiter.consume("user-1", "free");
      }

      const result = await limiter.consume("user-1", "free");

      expect(result.allowed).toBe(false);
      expect(result.retryAfter).toBeDefined();
      // Should be ~1 second for next token at 1 token/sec
      expect(result.retryAfter).toBeGreaterThanOrEqual(1);
      expect(result.retryAfter).toBeLessThanOrEqual(2);
    });
  });

  // ---------------------------------------------------------------------------
  // Run all tests
  // ---------------------------------------------------------------------------
  await runner.run();
}

// Run tests if executed directly
if (require.main === module) {
  runTests().catch(console.error);
}

export { runTests };

/**
 * Token Bucket Rate Limiter Examples
 *
 * Demonstrates practical usage of the rate limiter in various scenarios.
 */

import { TokenBucketRateLimiter, InMemoryStorage } from "./index";
import type { RateLimitResult, RateLimitTier } from "./types";

/**
 * Example 1: Basic API Rate Limiting
 *
 * Demonstrates how to use the rate limiter for API request throttling.
 */
async function basicApiRateLimiting(): Promise<void> {
  console.log("=== Example 1: Basic API Rate Limiting ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  // Simulate user making requests
  const userId = "user-123";
  const tier: RateLimitTier = "free"; // 10 requests per minute

  for (let i = 1; i <= 12; i++) {
    const result = await limiter.consume(userId, tier);

    if (result.allowed) {
      console.log(`Request ${i}: ✅ Allowed (${result.remaining} tokens left)`);
    } else {
      console.log(
        `Request ${i}: ❌ Rate limited! Retry after ${result.retryAfter} seconds`,
      );
      console.log(`   Reset at: ${result.resetAt.toISOString()}`);
    }
  }
}

/**
 * Example 2: Multi-tier Rate Limiting
 *
 * Demonstrates different rate limits for different user tiers.
 */
async function multiTierRateLimiting(): Promise<void> {
  console.log("\n=== Example 2: Multi-tier Rate Limiting ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  const users = [
    { id: "free-user", tier: "free" as const },
    { id: "pro-user", tier: "pro" as const },
    { id: "enterprise-user", tier: "enterprise" as const },
  ];

  for (const user of users) {
    // Check initial state
    const initial = await limiter.check(user.id, user.tier);
    console.log(
      `${user.tier.toUpperCase()} user - Initial tokens: ${initial.remaining}`,
    );

    // Simulate 50 rapid requests
    let allowed = 0;
    for (let i = 0; i < 50; i++) {
      const result = await limiter.consume(user.id, user.tier);
      if (result.allowed) allowed++;
    }

    console.log(`  → ${allowed}/50 requests allowed\n`);
  }
}

/**
 * Example 3: Check Before Consume Pattern
 *
 * Demonstrates checking availability before performing expensive operations.
 */
async function checkBeforeConsume(): Promise<void> {
  console.log("\n=== Example 3: Check Before Consume Pattern ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  const userId = "user-456";
  const tier: RateLimitTier = "pro";

  // Check if we can make the request first
  const checkResult = await limiter.check(userId, tier);
  console.log(`Pre-check: ${checkResult.remaining} tokens available`);

  if (checkResult.allowed) {
    console.log("→ Proceeding with expensive operation...");

    // Only consume if we actually do the work
    const consumeResult = await limiter.consume(userId, tier);
    console.log(`→ Consumed 1 token, ${consumeResult.remaining} remaining`);
  } else {
    console.log("→ Skipping expensive operation (rate limited)");
  }
}

/**
 * Example 4: Variable Cost Operations
 *
 * Demonstrates using different token costs for different operations.
 */
async function variableCostOperations(): Promise<void> {
  console.log("\n=== Example 4: Variable Cost Operations ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  const userId = "user-789";
  const tier: RateLimitTier = "pro"; // 100 tokens per minute

  // Define operation costs
  const operations = [
    { name: "Read single record", cost: 1 },
    { name: "Write single record", cost: 5 },
    { name: "Bulk read (100 records)", cost: 10 },
    { name: "Bulk write (100 records)", cost: 50 },
    { name: "Export all data", cost: 100 },
  ];

  for (const op of operations) {
    const result = await limiter.consume(userId, tier, op.cost);

    if (result.allowed) {
      console.log(
        `${op.name} (cost: ${op.cost}): ✅ Allowed, ${result.remaining} tokens left`,
      );
    } else {
      console.log(
        `${op.name} (cost: ${op.cost}): ❌ Rate limited, need ${result.retryAfter}s`,
      );
    }
  }
}

/**
 * Example 5: Express.js Middleware Pattern
 *
 * Demonstrates how to integrate with Express.js as middleware.
 */
async function expressMiddlewarePattern(): Promise<void> {
  console.log("\n=== Example 5: Express.js Middleware Pattern ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  // Simulated Express middleware
  async function rateLimitMiddleware(
    userId: string,
    tier: RateLimitTier,
  ): Promise<{
    status: number;
    headers: Record<string, string>;
    body?: string;
  }> {
    const result = await limiter.consume(userId, tier);

    // Build rate limit headers (common convention)
    const headers: Record<string, string> = {
      "X-RateLimit-Limit": String(
        tier === "free" ? 10 : tier === "pro" ? 100 : 1000,
      ),
      "X-RateLimit-Remaining": String(result.remaining),
      "X-RateLimit-Reset": String(Math.floor(result.resetAt.getTime() / 1000)),
    };

    if (!result.allowed) {
      headers["Retry-After"] = String(result.retryAfter);
      return {
        status: 429,
        headers,
        body: JSON.stringify({
          error: "Too Many Requests",
          retryAfter: result.retryAfter,
        }),
      };
    }

    return { status: 200, headers };
  }

  // Simulate requests
  console.log("Simulating Express middleware responses:\n");

  for (let i = 1; i <= 12; i++) {
    const response = await rateLimitMiddleware("api-user", "free");
    console.log(`Request ${i}: Status ${response.status}`);
    console.log(
      `  Headers: X-RateLimit-Remaining: ${response.headers["X-RateLimit-Remaining"]}`,
    );
    if (response.status === 429) {
      console.log(`  Body: ${response.body}`);
    }
  }
}

/**
 * Example 6: Custom Tier Configuration
 *
 * Demonstrates using custom rate limit configurations.
 */
async function customTierConfiguration(): Promise<void> {
  console.log("\n=== Example 6: Custom Tier Configuration ===\n");

  const storage = new InMemoryStorage();

  // Custom configuration for a gaming app
  const limiter = new TokenBucketRateLimiter({
    storage,
    tierConfigs: {
      free: { maxTokens: 5, refillRatePerMinute: 5 }, // 5 actions/min
      pro: { maxTokens: 30, refillRatePerMinute: 30 }, // 30 actions/min
      enterprise: { maxTokens: 120, refillRatePerMinute: 120 }, // 120 actions/min
    },
  });

  const tiers: RateLimitTier[] = ["free", "pro", "enterprise"];

  for (const tier of tiers) {
    const result = await limiter.check(`player-${tier}`, tier);
    console.log(
      `${tier.toUpperCase()} tier: ${result.remaining} actions available`,
    );
  }
}

/**
 * Example 7: Concurrent Request Handling
 *
 * Demonstrates that the limiter handles concurrent requests correctly.
 */
async function concurrentRequestHandling(): Promise<void> {
  console.log("\n=== Example 7: Concurrent Request Handling ===\n");

  const storage = new InMemoryStorage();
  const limiter = new TokenBucketRateLimiter({ storage });

  const userId = "concurrent-user";
  const tier: RateLimitTier = "free"; // 10 tokens

  // Fire 20 concurrent requests
  console.log("Firing 20 concurrent requests (limit: 10)...\n");

  const requests = Array.from({ length: 20 }, (_, i) =>
    limiter.consume(userId, tier).then((result) => ({
      index: i + 1,
      ...result,
    })),
  );

  const results = await Promise.all(requests);

  const allowed = results.filter((r) => r.allowed);
  const rejected = results.filter((r) => !r.allowed);

  console.log(
    `Results: ${allowed.length} allowed, ${rejected.length} rejected`,
  );
  console.log(`Allowed requests: ${allowed.map((r) => r.index).join(", ")}`);
  console.log(`Rejected requests: ${rejected.map((r) => r.index).join(", ")}`);
}

// Run all examples
async function main(): Promise<void> {
  await basicApiRateLimiting();
  await multiTierRateLimiting();
  await checkBeforeConsume();
  await variableCostOperations();
  await expressMiddlewarePattern();
  await customTierConfiguration();
  await concurrentRequestHandling();

  console.log("\n✨ All examples completed!\n");
}

// Run if executed directly
if (require.main === module) {
  main().catch(console.error);
}

export { main as runExamples };

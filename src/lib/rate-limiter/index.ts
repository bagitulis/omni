/**
 * Rate Limiter Module
 *
 * A multi-tier, distributed-ready token bucket rate limiter.
 *
 * @example Basic Usage
 * ```typescript
 * import { TokenBucketRateLimiter, InMemoryStorage } from './rate-limiter';
 *
 * const limiter = new TokenBucketRateLimiter({
 *   storage: new InMemoryStorage(),
 * });
 *
 * // Check if request is allowed
 * const result = await limiter.consume('user-123', 'pro');
 *
 * if (result.allowed) {
 *   // Process request
 *   console.log(`Remaining: ${result.remaining}`);
 * } else {
 *   // Rate limited
 *   console.log(`Retry after ${result.retryAfter} seconds`);
 * }
 * ```
 *
 * @example Distributed Usage with Redis
 * ```typescript
 * import Redis from 'ioredis';
 * import { TokenBucketRateLimiter, RedisStorage } from './rate-limiter';
 *
 * const redis = new Redis('redis://localhost:6379');
 * const limiter = new TokenBucketRateLimiter({
 *   storage: new RedisStorage({ client: redis }),
 * });
 * ```
 */

export {
  TokenBucketRateLimiter,
  DEFAULT_TIER_CONFIGS,
} from "./TokenBucketRateLimiter";
export { InMemoryStorage } from "./InMemoryStorage";
export { RedisStorage } from "./RedisStorage";
export type { RedisClient, RedisStorageConfig } from "./RedisStorage";

export type {
  RateLimiter,
  RateLimitResult,
  RateLimitTier,
  RateLimiterConfig,
  TierConfig,
  TierConfigs,
  BucketState,
  BucketStorage,
} from "./types";

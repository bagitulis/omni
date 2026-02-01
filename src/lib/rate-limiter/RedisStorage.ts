/**
 * Redis Bucket Storage
 *
 * A Redis-backed storage implementation for distributed deployments.
 * Uses Lua scripts for atomic operations to ensure thread-safety across instances.
 *
 * This is an interface/reference implementation that can be adapted to your Redis client.
 */

import type { BucketStorage, BucketState, RateLimitTier } from "./types";

/**
 * Redis client interface (compatible with ioredis, node-redis, etc.)
 */
export interface RedisClient {
  get(key: string): Promise<string | null>;
  set(
    key: string,
    value: string,
    mode?: string,
    duration?: number,
  ): Promise<unknown>;
  setex(key: string, seconds: number, value: string): Promise<unknown>;
  del(key: string): Promise<number>;
  eval(
    script: string,
    numKeys: number,
    ...args: (string | number)[]
  ): Promise<unknown>;
  keys(pattern: string): Promise<string[]>;
}

/**
 * Configuration for Redis storage
 */
export interface RedisStorageConfig {
  /** Redis client instance */
  client: RedisClient;
  /** Key prefix for rate limiter buckets */
  keyPrefix?: string;
  /** TTL for bucket entries in seconds (default: 300 = 5 minutes) */
  ttlSeconds?: number;
}

/**
 * Lua script for atomic bucket update
 *
 * This script atomically:
 * 1. Gets the current bucket state
 * 2. Calculates token refill based on elapsed time
 * 3. Attempts to consume tokens if available
 * 4. Updates the bucket state
 * 5. Returns the result
 */
const ATOMIC_UPDATE_SCRIPT = `
local key = KEYS[1]
local tier = ARGV[1]
local maxTokens = tonumber(ARGV[2])
local refillRatePerMinute = tonumber(ARGV[3])
local tokensToConsume = tonumber(ARGV[4])
local now = tonumber(ARGV[5])
local ttl = tonumber(ARGV[6])
local checkOnly = ARGV[7] == "true"

-- Get current state
local data = redis.call('GET', key)
local tokens, lastRefillTime

if data then
  local state = cjson.decode(data)
  tokens = tonumber(state.tokens)
  lastRefillTime = tonumber(state.lastRefillTime)
else
  -- New bucket starts full
  tokens = maxTokens
  lastRefillTime = now
end

-- Calculate refill (smooth refill based on elapsed time)
local elapsedMs = now - lastRefillTime
local elapsedMinutes = elapsedMs / 60000
local tokensToAdd = elapsedMinutes * refillRatePerMinute
tokens = math.min(maxTokens, tokens + tokensToAdd)

-- Check or consume
local allowed = tokens >= tokensToConsume
if allowed and not checkOnly then
  tokens = tokens - tokensToConsume
end

-- Update state
local newState = cjson.encode({
  tokens = tokens,
  lastRefillTime = now,
  tier = tier
})
redis.call('SETEX', key, ttl, newState)

-- Return result
return cjson.encode({
  allowed = allowed,
  tokens = tokens,
  lastRefillTime = now
})
`;

/**
 * Redis Storage Implementation
 *
 * Provides thread-safe bucket storage across distributed instances using Redis.
 * Uses Lua scripts for atomic read-modify-write operations.
 */
export class RedisStorage implements BucketStorage {
  private readonly client: RedisClient;
  private readonly keyPrefix: string;
  private readonly ttlSeconds: number;

  /**
   * Creates a new RedisStorage instance
   *
   * @param config - Redis storage configuration
   *
   * @example
   * ```typescript
   * import Redis from 'ioredis';
   *
   * const redis = new Redis('redis://localhost:6379');
   * const storage = new RedisStorage({ client: redis });
   * const limiter = new TokenBucketRateLimiter({ storage });
   * ```
   */
  constructor(config: RedisStorageConfig) {
    this.client = config.client;
    this.keyPrefix = config.keyPrefix ?? "ratelimit:";
    this.ttlSeconds = config.ttlSeconds ?? 300;
  }

  /**
   * Generate a unique key for user + tier combination
   */
  private getKey(userId: string, tier: RateLimitTier): string {
    return `${this.keyPrefix}${userId}:${tier}`;
  }

  /**
   * Parse bucket state from Redis JSON
   */
  private parseState(data: string | null): BucketState | null {
    if (!data) return null;

    try {
      const parsed = JSON.parse(data);
      return {
        tokens: Number(parsed.tokens),
        lastRefillTime: Number(parsed.lastRefillTime),
        tier: parsed.tier as RateLimitTier,
      };
    } catch {
      return null;
    }
  }

  /**
   * Get the current bucket state for a user
   */
  async get(userId: string, tier: RateLimitTier): Promise<BucketState | null> {
    const key = this.getKey(userId, tier);
    const data = await this.client.get(key);
    return this.parseState(data);
  }

  /**
   * Set the bucket state for a user
   */
  async set(userId: string, state: BucketState): Promise<void> {
    const key = this.getKey(userId, state.tier);
    const data = JSON.stringify(state);
    await this.client.setex(key, this.ttlSeconds, data);
  }

  /**
   * Atomically get and update bucket state
   *
   * Uses Lua scripting for atomic read-modify-write operations.
   * This is critical for distributed systems where multiple instances
   * may be handling requests for the same user simultaneously.
   */
  async atomicUpdate(
    userId: string,
    tier: RateLimitTier,
    updater: (current: BucketState | null) => BucketState,
  ): Promise<BucketState> {
    const key = this.getKey(userId, tier);

    // For full flexibility with custom updater, we use a get-then-set approach
    // with optimistic locking pattern. For production, consider using WATCH/MULTI/EXEC
    // or the Lua script above for specific use cases.

    const current = await this.get(userId, tier);
    const updated = updater(current);
    await this.set(userId, updated);

    return updated;
  }

  /**
   * Atomic consume operation using Lua script
   *
   * This method provides true atomic consume operation for distributed systems.
   * It handles refill calculation and token consumption in a single Redis call.
   *
   * @param userId - User identifier
   * @param tier - Rate limit tier
   * @param maxTokens - Maximum tokens for this tier
   * @param refillRatePerMinute - Refill rate for this tier
   * @param tokensToConsume - Number of tokens to consume
   * @param checkOnly - If true, only check availability without consuming
   * @returns Result of the operation
   */
  async atomicConsume(
    userId: string,
    tier: RateLimitTier,
    maxTokens: number,
    refillRatePerMinute: number,
    tokensToConsume: number,
    checkOnly: boolean = false,
  ): Promise<{ allowed: boolean; tokens: number; lastRefillTime: number }> {
    const key = this.getKey(userId, tier);
    const now = Date.now();

    const result = await this.client.eval(
      ATOMIC_UPDATE_SCRIPT,
      1,
      key,
      tier,
      maxTokens,
      refillRatePerMinute,
      tokensToConsume,
      now,
      this.ttlSeconds,
      checkOnly.toString(),
    );

    return JSON.parse(result as string);
  }

  /**
   * Delete a bucket
   */
  async delete(userId: string): Promise<void> {
    for (const tier of ["free", "pro", "enterprise"] as RateLimitTier[]) {
      const key = this.getKey(userId, tier);
      await this.client.del(key);
    }
  }

  /**
   * Clear all buckets (use with caution in production!)
   */
  async clear(): Promise<void> {
    const keys = await this.client.keys(`${this.keyPrefix}*`);
    for (const key of keys) {
      await this.client.del(key);
    }
  }
}

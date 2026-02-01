/**
 * Distributed Lock Manager - Safe Implementation
 *
 * A Redis-based distributed lock manager designed with safety as the primary concern.
 * This implementation follows the principles outlined in Martin Kleppmann's critique
 * of Redlock, prioritizing safety over liveness.
 *
 * ## Safety Guarantees
 *
 * 1. **Fencing Tokens**: Every lock acquisition generates a monotonically increasing
 *    fencing token. Downstream systems MUST validate these tokens before accepting writes.
 *    This is the ONLY way to prevent stale lock holders from causing data corruption.
 *
 * 2. **Clock Drift Compensation**: Lock validity is reduced by a configurable drift factor
 *    to account for clock differences between nodes.
 *
 * 3. **Atomic Operations**: All check-then-modify operations use Lua scripts to ensure
 *    atomicity within Redis.
 *
 * 4. **Conservative Validity**: We always err on the side of releasing locks early rather
 *    than risk a stale lock holder continuing to operate.
 *
 * ## Known Limitations (READ CAREFULLY)
 *
 * 1. **Single Redis Instance**: This implementation uses a single Redis instance.
 *    It does NOT implement the full Redlock algorithm with multiple independent nodes.
 *    For truly critical systems, consider using a consensus system like Zookeeper.
 *
 * 2. **GC Pauses**: If the lock holder experiences a long GC pause, the lock may expire
 *    on Redis while the holder still thinks it has the lock. The fencing token is your
 *    only protection against this scenario.
 *
 * 3. **Network Partitions**: If the lock holder is partitioned from Redis, it cannot
 *    extend the lock. The lock will expire and another node can acquire it. Again,
 *    fencing tokens are critical.
 *
 * 4. **Redis Persistence**: If Redis crashes and restarts with AOF/RDB, locks may be
 *    restored. This can lead to two holders thinking they have the same lock.
 *    Consider using Redis Cluster with WAIT for stronger guarantees.
 *
 * @module DistributedLockManager
 */

import { randomBytes } from "crypto";
import { EventEmitter } from "events";
import type {
  Lock,
  LockResult,
  LockManagerConfig,
  AcquireLockOptions,
  ExtendLockOptions,
  RedisClientInterface,
  Logger,
  LockManagerEvents,
  FencingToken,
} from "./types";
import {
  LockError,
  LockNotHeldError,
  LockExpiredError,
  ClockDriftError,
  LockFailureReason,
} from "./types";

/**
 * Default configuration values
 */
const DEFAULT_CONFIG: Omit<LockManagerConfig, "nodeId"> = {
  defaultTtlMs: 30000, // 30 seconds
  clockDriftFactor: 0.01, // 1% clock drift allowance
  maxRetries: 3,
  retryDelayMs: 200,
  maxRetryDelayMs: 3000,
  operationTimeoutMs: 5000,
  enableAutoHeartbeat: false,
  heartbeatIntervalMs: 10000,
};

/**
 * Lua script for safe lock release.
 * Only releases if the value matches (proves ownership).
 */
const RELEASE_SCRIPT = `
  if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
  else
    return 0
  end
`;

/**
 * Lua script for safe lock extension.
 * Only extends if the value matches (proves ownership).
 */
const EXTEND_SCRIPT = `
  if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
  else
    return 0
  end
`;

/**
 * Lua script to atomically get fencing token and increment.
 * Returns the new (post-increment) value.
 */
const GET_FENCING_TOKEN_SCRIPT = `
  return redis.call("INCR", KEYS[1])
`;

/**
 * Key prefix for lock resources
 */
const LOCK_KEY_PREFIX = "dlock:";

/**
 * Key for fencing token counter
 */
const FENCING_TOKEN_KEY = "dlock:fencing:counter";

/**
 * Typed event emitter for lock manager events
 */
export interface DistributedLockManager {
  on<K extends keyof LockManagerEvents>(
    event: K,
    listener: LockManagerEvents[K],
  ): this;
  off<K extends keyof LockManagerEvents>(
    event: K,
    listener: LockManagerEvents[K],
  ): this;
  emit<K extends keyof LockManagerEvents>(
    event: K,
    ...args: Parameters<LockManagerEvents[K]>
  ): boolean;
}

/**
 * Distributed Lock Manager
 *
 * Provides safe distributed locking with fencing tokens for Redis-based coordination.
 */
export class DistributedLockManager extends EventEmitter {
  private readonly config: LockManagerConfig;
  private readonly redis: RedisClientInterface;
  private readonly logger: Logger;
  private readonly activeLocks: Map<string, Lock> = new Map();
  private readonly heartbeatIntervals: Map<string, NodeJS.Timeout> = new Map();
  private isShuttingDown = false;
  private clockDriftMs = 0;
  private lastClockCheck = 0;

  constructor(
    redis: RedisClientInterface,
    config: Partial<LockManagerConfig> & { nodeId: string },
    logger?: Logger,
  ) {
    super();
    this.redis = redis;
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.logger = logger ?? this.createDefaultLogger();

    // Validate configuration
    this.validateConfig();

    this.logger.info("DistributedLockManager initialized", {
      nodeId: this.config.nodeId,
      defaultTtlMs: this.config.defaultTtlMs,
      clockDriftFactor: this.config.clockDriftFactor,
    });
  }

  /**
   * Acquire a distributed lock on a resource.
   *
   * @param resource - The resource name to lock
   * @param options - Optional lock configuration
   * @returns LockResult indicating success or failure
   *
   * @example
   * ```typescript
   * const result = await lockManager.acquire("user:123:profile");
   * if (result.success) {
   *   try {
   *     // Use result.lock.fencingToken for downstream validation
   *     await performCriticalOperation(result.lock.fencingToken);
   *   } finally {
   *     await lockManager.release(result.lock);
   *   }
   * }
   * ```
   */
  async acquire(
    resource: string,
    options: AcquireLockOptions = {},
  ): Promise<LockResult> {
    if (this.isShuttingDown) {
      return {
        success: false,
        reason: LockFailureReason.REDIS_ERROR,
      };
    }

    const ttlMs = options.ttlMs ?? this.config.defaultTtlMs;
    const waitForLock = options.waitForLock ?? false;
    const waitTimeoutMs = options.waitTimeoutMs ?? ttlMs * 2;

    // Check clock drift before acquiring
    await this.checkClockDrift();

    const startTime = Date.now();
    let attempt = 0;

    while (true) {
      attempt++;
      const result = await this.tryAcquire(resource, ttlMs);

      if (result.success) {
        const lock = result.lock;
        this.activeLocks.set(resource, lock);

        // Set up automatic heartbeat if enabled
        if (this.config.enableAutoHeartbeat) {
          this.startHeartbeat(lock);
        }

        this.emit("lockAcquired", lock);
        this.logger.info("Lock acquired", {
          resource,
          holderId: lock.holderId,
          fencingToken: lock.fencingToken.value.toString(),
          validUntil: new Date(lock.validUntil).toISOString(),
        });

        return result;
      }

      // Don't retry on certain errors
      if (
        result.reason === LockFailureReason.CLOCK_DRIFT_EXCEEDED ||
        result.reason === LockFailureReason.REDIS_ERROR
      ) {
        return result;
      }

      // Check if we should retry
      if (!waitForLock || attempt >= this.config.maxRetries) {
        this.logger.debug("Lock acquisition failed, not retrying", {
          resource,
          reason: result.reason,
          attempt,
          waitForLock,
        });
        return result;
      }

      // Check timeout
      const elapsed = Date.now() - startTime;
      if (elapsed >= waitTimeoutMs) {
        return {
          success: false,
          reason: LockFailureReason.TIMEOUT,
        };
      }

      // Calculate retry delay with exponential backoff and jitter
      const delay = this.calculateRetryDelay(attempt);
      this.logger.debug("Lock held by another, retrying", {
        resource,
        attempt,
        delayMs: delay,
      });

      await this.sleep(delay);
    }
  }

  /**
   * Extend the TTL of an existing lock.
   *
   * WARNING: This does NOT make the lock safe against GC pauses or network partitions.
   * The fencing token remains unchanged - downstream systems should continue to
   * validate with the same token.
   *
   * @param lock - The lock to extend
   * @param options - Extension options
   * @returns Updated lock or throws if extension fails
   */
  async extend(lock: Lock, options: ExtendLockOptions): Promise<Lock> {
    // First, verify we still think we hold this lock
    const activeLock = this.activeLocks.get(lock.resource);
    if (!activeLock || activeLock.token !== lock.token) {
      throw new LockNotHeldError(lock.resource, lock.holderId);
    }

    // Check if lock has already expired (locally)
    if (Date.now() >= lock.validUntil) {
      this.activeLocks.delete(lock.resource);
      this.stopHeartbeat(lock.resource);
      throw new LockExpiredError(lock.resource, lock.validUntil);
    }

    const key = this.getLockKey(lock.resource);
    const startTime = Date.now();

    try {
      const result = (await this.redis.eval(
        EXTEND_SCRIPT,
        [key],
        [lock.token, options.ttlMs.toString()],
      )) as number;

      if (result !== 1) {
        // Lock was released or taken by another holder
        this.activeLocks.delete(lock.resource);
        this.stopHeartbeat(lock.resource);
        throw new LockNotHeldError(lock.resource, lock.holderId);
      }

      // Calculate new validity, accounting for drift and operation time
      const operationTime = Date.now() - startTime;
      const driftAllowance = Math.floor(
        options.ttlMs * this.config.clockDriftFactor,
      );
      const validUntil =
        Date.now() + options.ttlMs - driftAllowance - operationTime;

      const extendedLock: Lock = {
        ...lock,
        ttlMs: options.ttlMs,
        validUntil,
      };

      this.activeLocks.set(lock.resource, extendedLock);
      this.emit("lockExtended", extendedLock);

      this.logger.debug("Lock extended", {
        resource: lock.resource,
        newTtlMs: options.ttlMs,
        validUntil: new Date(validUntil).toISOString(),
      });

      return extendedLock;
    } catch (error) {
      if (
        error instanceof LockNotHeldError ||
        error instanceof LockExpiredError
      ) {
        throw error;
      }
      this.logger.error("Failed to extend lock", {
        resource: lock.resource,
        error: (error as Error).message,
      });
      throw new LockError(
        `Failed to extend lock: ${(error as Error).message}`,
        "EXTEND_FAILED",
        lock.resource,
        error as Error,
      );
    }
  }

  /**
   * Release a lock.
   *
   * This is a graceful release that only succeeds if we still hold the lock.
   * It's safe to call even if the lock has already expired - it will simply
   * do nothing in that case.
   *
   * @param lock - The lock to release
   * @returns true if released, false if already released/expired
   */
  async release(lock: Lock): Promise<boolean> {
    const activeLock = this.activeLocks.get(lock.resource);
    if (!activeLock || activeLock.token !== lock.token) {
      this.logger.debug("Lock not in active locks during release", {
        resource: lock.resource,
        hadActiveLock: !!activeLock,
      });
      return false;
    }

    // Always clean up local state first
    this.activeLocks.delete(lock.resource);
    this.stopHeartbeat(lock.resource);

    const key = this.getLockKey(lock.resource);

    try {
      const result = (await this.redis.eval(
        RELEASE_SCRIPT,
        [key],
        [lock.token],
      )) as number;

      const released = result === 1;

      if (released) {
        this.emit("lockReleased", lock.resource, lock.holderId);
        this.logger.info("Lock released", {
          resource: lock.resource,
          holderId: lock.holderId,
        });
      } else {
        this.logger.warn("Lock release returned 0 - lock may have expired", {
          resource: lock.resource,
          expectedToken: lock.token.substring(0, 8) + "...",
        });
      }

      return released;
    } catch (error) {
      this.logger.error("Error during lock release", {
        resource: lock.resource,
        error: (error as Error).message,
      });
      // Still return false - we've cleaned up locally
      return false;
    }
  }

  /**
   * Check if a lock is currently held by this manager.
   * NOTE: This is a local check only and may be stale.
   */
  isHeld(resource: string): boolean {
    const lock = this.activeLocks.get(resource);
    if (!lock) return false;

    // Check if expired locally
    if (Date.now() >= lock.validUntil) {
      this.activeLocks.delete(resource);
      this.stopHeartbeat(resource);
      this.emit("lockExpired", resource, lock.holderId);
      return false;
    }

    return true;
  }

  /**
   * Get the current lock for a resource, if held by this manager.
   */
  getLock(resource: string): Lock | undefined {
    if (!this.isHeld(resource)) {
      return undefined;
    }
    return this.activeLocks.get(resource);
  }

  /**
   * Validate that a lock is still valid.
   * Returns the remaining validity time in milliseconds, or -1 if invalid.
   *
   * IMPORTANT: This only validates local state. The lock may have been
   * taken by another holder if TTL expired on Redis.
   */
  validateLock(lock: Lock): number {
    const activeLock = this.activeLocks.get(lock.resource);
    if (!activeLock || activeLock.token !== lock.token) {
      return -1;
    }

    const remaining = lock.validUntil - Date.now();
    if (remaining <= 0) {
      this.activeLocks.delete(lock.resource);
      this.stopHeartbeat(lock.resource);
      this.emit("lockExpired", lock.resource, lock.holderId);
      return -1;
    }

    return remaining;
  }

  /**
   * Execute a function while holding a lock.
   * Automatically acquires and releases the lock.
   *
   * @param resource - Resource to lock
   * @param fn - Function to execute while holding lock
   * @param options - Lock options
   *
   * @example
   * ```typescript
   * const result = await lockManager.withLock(
   *   "user:123:balance",
   *   async (lock) => {
   *     const balance = await getBalance(123);
   *     // Pass fencing token to database for validation
   *     await updateBalance(123, balance + 100, lock.fencingToken);
   *     return balance + 100;
   *   },
   *   { ttlMs: 10000 }
   * );
   * ```
   */
  async withLock<T>(
    resource: string,
    fn: (lock: Lock) => Promise<T>,
    options: AcquireLockOptions = {},
  ): Promise<T> {
    const result = await this.acquire(resource, options);

    if (!result.success) {
      throw new LockError(
        `Failed to acquire lock: ${result.reason}`,
        result.reason,
        resource,
      );
    }

    try {
      return await fn(result.lock);
    } finally {
      await this.release(result.lock);
    }
  }

  /**
   * Gracefully shutdown the lock manager.
   * Releases all held locks and stops heartbeats.
   */
  async shutdown(): Promise<void> {
    this.isShuttingDown = true;
    this.logger.info("Shutting down lock manager", {
      activeLocks: this.activeLocks.size,
    });

    // Stop all heartbeats
    for (const resource of this.heartbeatIntervals.keys()) {
      this.stopHeartbeat(resource);
    }

    // Release all locks
    const releases = Array.from(this.activeLocks.values()).map((lock) =>
      this.release(lock).catch((err) => {
        this.logger.warn("Failed to release lock during shutdown", {
          resource: lock.resource,
          error: (err as Error).message,
        });
      }),
    );

    await Promise.all(releases);

    this.logger.info("Lock manager shutdown complete");
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private async tryAcquire(
    resource: string,
    ttlMs: number,
  ): Promise<LockResult> {
    const key = this.getLockKey(resource);
    const token = this.generateToken();
    const holderId = this.generateHolderId();
    const startTime = Date.now();

    try {
      // Attempt to set the lock
      const acquired = await this.redis.setNX(key, token, ttlMs);

      if (!acquired) {
        return {
          success: false,
          reason: LockFailureReason.ALREADY_HELD,
          retryAfterMs: this.calculateRetryDelay(1),
        };
      }

      // Get fencing token atomically
      const fencingValue = (await this.redis.eval(
        GET_FENCING_TOKEN_SCRIPT,
        [FENCING_TOKEN_KEY],
        [],
      )) as number;

      const acquireTime = Date.now();
      const operationTime = acquireTime - startTime;

      // Calculate validity, accounting for drift and operation time
      const driftAllowance = Math.floor(ttlMs * this.config.clockDriftFactor);
      const validUntil = acquireTime + ttlMs - driftAllowance - operationTime;

      // Check if we have meaningful validity remaining
      const minValidityMs = 100; // Must have at least 100ms of validity
      if (validUntil - Date.now() < minValidityMs) {
        // Lock acquired but too little time remaining - release and fail
        await this.redis.eval(RELEASE_SCRIPT, [key], [token]).catch(() => {}); // Best effort release
        return {
          success: false,
          reason: LockFailureReason.CLOCK_DRIFT_EXCEEDED,
        };
      }

      const fencingToken: FencingToken = {
        value: BigInt(fencingValue),
        issuedAt: new Date(acquireTime).toISOString(),
        resource,
        holderId,
      };

      const lock: Lock = {
        resource,
        holderId,
        token,
        fencingToken,
        acquiredAt: acquireTime,
        validUntil,
        ttlMs,
      };

      return { success: true, lock };
    } catch (error) {
      this.logger.error("Redis error during lock acquisition", {
        resource,
        error: (error as Error).message,
      });
      return {
        success: false,
        reason: LockFailureReason.REDIS_ERROR,
      };
    }
  }

  private startHeartbeat(lock: Lock): void {
    if (this.heartbeatIntervals.has(lock.resource)) {
      return;
    }

    const interval = setInterval(async () => {
      try {
        const currentLock = this.activeLocks.get(lock.resource);
        if (!currentLock || currentLock.token !== lock.token) {
          this.stopHeartbeat(lock.resource);
          return;
        }

        // Extend with original TTL
        await this.extend(currentLock, { ttlMs: lock.ttlMs });
      } catch (error) {
        this.logger.warn("Heartbeat failed", {
          resource: lock.resource,
          error: (error as Error).message,
        });
        this.emit("heartbeatFailed", lock, error as Error);
        this.stopHeartbeat(lock.resource);
      }
    }, this.config.heartbeatIntervalMs);

    this.heartbeatIntervals.set(lock.resource, interval);
  }

  private stopHeartbeat(resource: string): void {
    const interval = this.heartbeatIntervals.get(resource);
    if (interval) {
      clearInterval(interval);
      this.heartbeatIntervals.delete(resource);
    }
  }

  private async checkClockDrift(): Promise<void> {
    const now = Date.now();

    // Only check every 60 seconds
    if (now - this.lastClockCheck < 60000) {
      return;
    }

    try {
      const beforeCall = Date.now();
      const [seconds, microseconds] = await this.redis.time();
      const afterCall = Date.now();

      const redisTimeMs = seconds * 1000 + Math.floor(microseconds / 1000);
      const roundTripTime = afterCall - beforeCall;
      const localTimeAtRedisResponse = beforeCall + roundTripTime / 2;

      this.clockDriftMs = Math.abs(redisTimeMs - localTimeAtRedisResponse);
      this.lastClockCheck = now;

      // Warn if drift is significant
      const maxAcceptableDrift =
        this.config.defaultTtlMs * this.config.clockDriftFactor * 2;
      if (this.clockDriftMs > maxAcceptableDrift) {
        this.emit("clockDriftDetected", this.clockDriftMs);
        this.logger.warn("Significant clock drift detected", {
          driftMs: this.clockDriftMs,
          maxAcceptableDrift,
        });
      }
    } catch (error) {
      this.logger.debug("Failed to check clock drift", {
        error: (error as Error).message,
      });
    }
  }

  private getLockKey(resource: string): string {
    return `${LOCK_KEY_PREFIX}${resource}`;
  }

  private generateToken(): string {
    return randomBytes(16).toString("hex");
  }

  private generateHolderId(): string {
    return `${this.config.nodeId}:${Date.now()}:${randomBytes(4).toString("hex")}`;
  }

  private calculateRetryDelay(attempt: number): number {
    // Exponential backoff: baseDelay * 2^(attempt-1)
    const exponentialDelay =
      this.config.retryDelayMs * Math.pow(2, attempt - 1);
    const cappedDelay = Math.min(exponentialDelay, this.config.maxRetryDelayMs);

    // Add jitter (0-50% of delay)
    const jitter = Math.random() * 0.5 * cappedDelay;
    return Math.floor(cappedDelay + jitter);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  private validateConfig(): void {
    if (this.config.defaultTtlMs < 1000) {
      throw new Error("defaultTtlMs must be at least 1000ms");
    }

    if (
      this.config.clockDriftFactor < 0 ||
      this.config.clockDriftFactor > 0.5
    ) {
      throw new Error("clockDriftFactor must be between 0 and 0.5");
    }

    if (this.config.enableAutoHeartbeat) {
      if (this.config.heartbeatIntervalMs >= this.config.defaultTtlMs / 2) {
        throw new Error(
          "heartbeatIntervalMs must be less than half of defaultTtlMs",
        );
      }
    }

    if (!this.config.nodeId) {
      throw new Error("nodeId is required");
    }
  }

  private createDefaultLogger(): Logger {
    return {
      debug: () => {},
      info: (message, context) =>
        console.log(`[DLM] ${message}`, JSON.stringify(context)),
      warn: (message, context) =>
        console.warn(`[DLM] ${message}`, JSON.stringify(context)),
      error: (message, context) =>
        console.error(`[DLM] ${message}`, JSON.stringify(context)),
    };
  }
}

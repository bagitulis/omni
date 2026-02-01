/**
 * Distributed Lock Manager - Type Definitions
 *
 * This module provides type-safe interfaces for a distributed locking system
 * designed for Redis-based coordination across multiple nodes/processes.
 */

/**
 * Fencing token - a monotonically increasing value that prevents
 * stale lock holders from performing operations. Each lock acquisition
 * generates a new fencing token that MUST be validated by downstream
 * systems before accepting writes.
 */
export interface FencingToken {
  /** The numeric token value - monotonically increasing */
  readonly value: bigint;
  /** ISO timestamp when this token was issued */
  readonly issuedAt: string;
  /** The lock resource this token is valid for */
  readonly resource: string;
  /** Unique identifier of the lock holder */
  readonly holderId: string;
}

/**
 * Represents an acquired lock with all necessary metadata
 * to safely use and release it.
 */
export interface Lock {
  /** The resource name this lock protects */
  readonly resource: string;
  /** Unique identifier for this lock holder */
  readonly holderId: string;
  /** Cryptographically random value used to verify ownership */
  readonly token: string;
  /** Fencing token for downstream validation */
  readonly fencingToken: FencingToken;
  /** When the lock was acquired (local time) */
  readonly acquiredAt: number;
  /** When the lock will expire (local time, accounting for drift) */
  readonly validUntil: number;
  /** Configured TTL in milliseconds */
  readonly ttlMs: number;
}

/**
 * Result of a lock acquisition attempt
 */
export type LockResult =
  | { success: true; lock: Lock }
  | { success: false; reason: LockFailureReason; retryAfterMs?: number };

/**
 * Reasons why lock acquisition might fail
 */
export enum LockFailureReason {
  /** Another holder owns the lock */
  ALREADY_HELD = "ALREADY_HELD",
  /** Redis operation failed */
  REDIS_ERROR = "REDIS_ERROR",
  /** Network timeout during acquisition */
  TIMEOUT = "TIMEOUT",
  /** Clock drift exceeded safe threshold */
  CLOCK_DRIFT_EXCEEDED = "CLOCK_DRIFT_EXCEEDED",
  /** Insufficient quorum (for multi-node Redlock) */
  QUORUM_NOT_REACHED = "QUORUM_NOT_REACHED",
  /** Lock was released during acquisition (race) */
  LOCK_RELEASED = "LOCK_RELEASED",
}

/**
 * Configuration options for the lock manager
 */
export interface LockManagerConfig {
  /**
   * Default TTL for locks in milliseconds.
   * CRITICAL: Must be significantly longer than expected operation time.
   * If your operation takes 5s, TTL should be at least 30s.
   */
  defaultTtlMs: number;

  /**
   * Clock drift factor (0-1). The lock validity is reduced by this factor
   * to account for clock differences between nodes.
   * Recommended: 0.01 (1%) - conservative value for most environments.
   *
   * Validity = TTL - (TTL * driftFactor) - elapsedAcquisitionTime
   */
  clockDriftFactor: number;

  /**
   * Maximum retry attempts when acquiring a lock
   */
  maxRetries: number;

  /**
   * Base delay between retries in milliseconds.
   * Actual delay uses exponential backoff with jitter.
   */
  retryDelayMs: number;

  /**
   * Maximum delay between retries in milliseconds
   */
  maxRetryDelayMs: number;

  /**
   * Timeout for individual Redis operations in milliseconds
   */
  operationTimeoutMs: number;

  /**
   * Unique identifier for this node/process.
   * Used to identify lock holder in logs and debugging.
   */
  nodeId: string;

  /**
   * Enable automatic heartbeat for long-running operations.
   * WARNING: This doesn't guarantee safety - always validate fencing tokens.
   */
  enableAutoHeartbeat: boolean;

  /**
   * Interval for automatic heartbeat in milliseconds.
   * Should be significantly less than TTL (recommend TTL/3).
   */
  heartbeatIntervalMs: number;
}

/**
 * Options for acquiring a specific lock
 */
export interface AcquireLockOptions {
  /** TTL for this specific lock (overrides default) */
  ttlMs?: number;
  /** Maximum time to wait for lock acquisition */
  waitTimeoutMs?: number;
  /** Whether to retry if lock is held by another */
  waitForLock?: boolean;
  /** Custom metadata to associate with the lock */
  metadata?: Record<string, unknown>;
}

/**
 * Options for extending a lock
 */
export interface ExtendLockOptions {
  /** New TTL from now (not added to remaining) */
  ttlMs: number;
}

/**
 * Abstract Redis client interface for dependency injection.
 * Implementations must provide these operations atomically.
 */
export interface RedisClientInterface {
  /**
   * SET with NX (only if not exists) and PX (expire in ms).
   * This MUST be atomic (single Redis command).
   */
  setNX(key: string, value: string, ttlMs: number): Promise<boolean>;

  /**
   * GET a key's value
   */
  get(key: string): Promise<string | null>;

  /**
   * Execute a Lua script atomically.
   * CRITICAL: All lock operations that check-then-modify MUST use this.
   */
  eval(script: string, keys: string[], args: string[]): Promise<unknown>;

  /**
   * INCR operation for fencing token generation
   */
  incr(key: string): Promise<number>;

  /**
   * Check if Redis is connected and responsive
   */
  ping(): Promise<boolean>;

  /**
   * Get current Redis server time [seconds, microseconds]
   * Used to detect clock skew between client and server.
   */
  time(): Promise<[number, number]>;

  /**
   * Graceful disconnect
   */
  disconnect(): Promise<void>;
}

/**
 * Logger interface for structured logging
 */
export interface Logger {
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, context?: Record<string, unknown>): void;
}

/**
 * Events emitted by the lock manager
 */
export interface LockManagerEvents {
  lockAcquired: (lock: Lock) => void;
  lockReleased: (resource: string, holderId: string) => void;
  lockExtended: (lock: Lock) => void;
  lockExpired: (resource: string, holderId: string) => void;
  heartbeatFailed: (lock: Lock, error: Error) => void;
  clockDriftDetected: (driftMs: number) => void;
}

/**
 * Custom error types for the lock manager
 */
export class LockError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly resource?: string,
    public readonly cause?: Error,
  ) {
    super(message);
    this.name = "LockError";
    Error.captureStackTrace?.(this, LockError);
  }
}

export class LockNotHeldError extends LockError {
  constructor(resource: string, holderId: string) {
    super(
      `Lock on resource '${resource}' is not held by '${holderId}'`,
      "LOCK_NOT_HELD",
      resource,
    );
    this.name = "LockNotHeldError";
  }
}

export class LockExpiredError extends LockError {
  constructor(resource: string, expiredAt: number) {
    super(
      `Lock on resource '${resource}' expired at ${new Date(expiredAt).toISOString()}`,
      "LOCK_EXPIRED",
      resource,
    );
    this.name = "LockExpiredError";
  }
}

export class ClockDriftError extends LockError {
  constructor(driftMs: number, maxDriftMs: number) {
    super(
      `Clock drift of ${driftMs}ms exceeds maximum allowed ${maxDriftMs}ms`,
      "CLOCK_DRIFT_EXCEEDED",
    );
    this.name = "ClockDriftError";
  }
}

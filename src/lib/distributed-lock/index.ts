/**
 * Distributed Lock Manager Module
 *
 * Safe Redis-based distributed locking with fencing tokens.
 *
 * @example
 * ```typescript
 * import {
 *   DistributedLockManager,
 *   type Lock,
 *   type LockResult,
 * } from './lib/distributed-lock';
 *
 * const lockManager = new DistributedLockManager(redisClient, {
 *   nodeId: 'worker-1',
 *   defaultTtlMs: 30000,
 * });
 *
 * const result = await lockManager.acquire('resource:123');
 * if (result.success) {
 *   // Use result.lock.fencingToken for downstream validation
 * }
 * ```
 */

export { DistributedLockManager } from "./DistributedLockManager";

export type {
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

export {
  LockError,
  LockNotHeldError,
  LockExpiredError,
  ClockDriftError,
  LockFailureReason,
} from "./types";

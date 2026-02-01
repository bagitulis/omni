/**
 * Distributed Lock Manager - Examples and Documentation
 *
 * This file provides comprehensive examples of how to use the Distributed Lock Manager
 * and documents critical considerations for safe distributed locking.
 */

import type { RedisClientInterface, Lock, Logger } from "./types";
import { DistributedLockManager, LockFailureReason } from "./index";

// =============================================================================
// EXAMPLE 1: Basic Redis Client Adapter
// =============================================================================

/**
 * Example Redis client adapter using ioredis.
 * Adapt this to your Redis client library (ioredis, node-redis, etc.)
 */
export function createRedisAdapter(
  /* ioredis client instance */
  client: {
    set: (
      key: string,
      value: string,
      ...args: string[]
    ) => Promise<"OK" | null>;
    get: (key: string) => Promise<string | null>;
    eval: (
      script: string,
      numKeys: number,
      ...args: string[]
    ) => Promise<unknown>;
    incr: (key: string) => Promise<number>;
    ping: () => Promise<"PONG">;
    time: () => Promise<[string, string]>;
    disconnect: () => Promise<"OK">;
  },
): RedisClientInterface {
  return {
    async setNX(key: string, value: string, ttlMs: number): Promise<boolean> {
      // SET key value PX ttlMs NX
      const result = await client.set(key, value, "PX", String(ttlMs), "NX");
      return result === "OK";
    },

    async get(key: string): Promise<string | null> {
      return client.get(key);
    },

    async eval(
      script: string,
      keys: string[],
      args: string[],
    ): Promise<unknown> {
      return client.eval(script, keys.length, ...keys, ...args);
    },

    async incr(key: string): Promise<number> {
      return client.incr(key);
    },

    async ping(): Promise<boolean> {
      const result = await client.ping();
      return result === "PONG";
    },

    async time(): Promise<[number, number]> {
      const [seconds, microseconds] = await client.time();
      return [parseInt(seconds, 10), parseInt(microseconds, 10)];
    },

    async disconnect(): Promise<void> {
      await client.disconnect();
    },
  };
}

// =============================================================================
// EXAMPLE 2: Basic Lock Usage
// =============================================================================

async function basicLockExample(
  lockManager: DistributedLockManager,
): Promise<void> {
  // Acquire a lock
  const result = await lockManager.acquire("user:123:profile", {
    ttlMs: 10000, // 10 second lock
    waitForLock: true, // Retry if another holder has it
    waitTimeoutMs: 5000, // Give up after 5 seconds of waiting
  });

  if (!result.success) {
    console.log(`Failed to acquire lock: ${result.reason}`);
    if (result.retryAfterMs) {
      console.log(`Retry after ${result.retryAfterMs}ms`);
    }
    return;
  }

  const lock = result.lock;
  console.log(`Lock acquired with fencing token: ${lock.fencingToken.value}`);

  try {
    // Perform your critical section work
    await performCriticalOperation(lock.fencingToken.value);
  } finally {
    // ALWAYS release the lock in a finally block
    const released = await lockManager.release(lock);
    console.log(`Lock released: ${released}`);
  }
}

// =============================================================================
// EXAMPLE 3: Using withLock Helper
// =============================================================================

async function withLockExample(
  lockManager: DistributedLockManager,
): Promise<number> {
  // The withLock helper handles acquire/release automatically
  return lockManager.withLock(
    "counter:global",
    async (lock) => {
      // Critical section - pass fencing token to downstream!
      const currentValue = await getCounter();
      const newValue = currentValue + 1;
      await setCounter(newValue, lock.fencingToken.value);
      return newValue;
    },
    { ttlMs: 5000 },
  );
}

// =============================================================================
// EXAMPLE 4: Lock Extension for Long Operations
// =============================================================================

async function longOperationExample(
  lockManager: DistributedLockManager,
): Promise<void> {
  const result = await lockManager.acquire("batch:process", {
    ttlMs: 30000, // 30 seconds initial
  });

  if (!result.success) {
    throw new Error(`Failed to acquire lock: ${result.reason}`);
  }

  let lock = result.lock;

  try {
    // Process items in batches, extending lock as needed
    const items = await getItemsToProcess();

    for (let i = 0; i < items.length; i++) {
      // Check remaining validity
      const remaining = lockManager.validateLock(lock);
      if (remaining < 5000) {
        // Less than 5 seconds remaining - extend
        console.log("Extending lock...");
        lock = await lockManager.extend(lock, { ttlMs: 30000 });
      }

      await processItem(items[i], lock.fencingToken.value);
    }
  } finally {
    await lockManager.release(lock);
  }
}

// =============================================================================
// EXAMPLE 5: Handling Lock Events
// =============================================================================

function setupEventHandlers(lockManager: DistributedLockManager): void {
  lockManager.on("lockAcquired", (lock) => {
    console.log(`Lock acquired: ${lock.resource} by ${lock.holderId}`);
  });

  lockManager.on("lockReleased", (resource, holderId) => {
    console.log(`Lock released: ${resource} by ${holderId}`);
  });

  lockManager.on("lockExpired", (resource, holderId) => {
    console.warn(`Lock expired: ${resource} by ${holderId}`);
  });

  lockManager.on("heartbeatFailed", (lock, error) => {
    console.error(`Heartbeat failed for ${lock.resource}: ${error.message}`);
    // Consider alerting here - the operation may be in an inconsistent state
  });

  lockManager.on("clockDriftDetected", (driftMs) => {
    console.warn(`Clock drift detected: ${driftMs}ms`);
    // Consider alerting - large drift can compromise lock safety
  });
}

// =============================================================================
// EXAMPLE 6: Custom Logger Integration
// =============================================================================

function createStructuredLogger(): Logger {
  return {
    debug(message: string, context?: Record<string, unknown>): void {
      console.debug(JSON.stringify({ level: "debug", message, ...context }));
    },
    info(message: string, context?: Record<string, unknown>): void {
      console.info(JSON.stringify({ level: "info", message, ...context }));
    },
    warn(message: string, context?: Record<string, unknown>): void {
      console.warn(JSON.stringify({ level: "warn", message, ...context }));
    },
    error(message: string, context?: Record<string, unknown>): void {
      console.error(JSON.stringify({ level: "error", message, ...context }));
    },
  };
}

// =============================================================================
// EXAMPLE 7: Safe Downstream Usage with Fencing Tokens
// =============================================================================

/**
 * CRITICAL: This example shows how to use fencing tokens in downstream systems.
 *
 * The database/storage layer MUST validate fencing tokens to prevent stale
 * lock holders from corrupting data.
 */
async function safeDatabaseOperation(
  lockManager: DistributedLockManager,
): Promise<void> {
  const result = await lockManager.acquire("account:123:balance");

  if (!result.success) {
    throw new Error("Failed to acquire lock");
  }

  const lock = result.lock;

  try {
    // The fencing token MUST be passed to and validated by the database
    // This is typically done with conditional writes:
    //
    // UPDATE accounts
    // SET balance = $new_balance, last_fencing_token = $token
    // WHERE id = 123 AND last_fencing_token < $token
    //
    // This ensures that if we're a stale lock holder (our lock expired and
    // someone else acquired it), our write will fail.

    await updateAccountBalance(
      123,
      1000,
      lock.fencingToken.value, // Pass to database!
    );
  } finally {
    await lockManager.release(lock);
  }
}

// =============================================================================
// STUB FUNCTIONS (for compilation)
// =============================================================================

async function performCriticalOperation(_fencingToken: bigint): Promise<void> {
  // Your critical section code
}

async function getCounter(): Promise<number> {
  return 0;
}

async function setCounter(
  _value: number,
  _fencingToken: bigint,
): Promise<void> {
  // Use conditional write with fencing token
}

async function getItemsToProcess(): Promise<string[]> {
  return [];
}

async function processItem(
  _item: string,
  _fencingToken: bigint,
): Promise<void> {
  // Process with fencing token validation
}

async function updateAccountBalance(
  _accountId: number,
  _balance: number,
  _fencingToken: bigint,
): Promise<void> {
  // Use: WHERE last_fencing_token < $fencingToken
}

// =============================================================================
// DOCUMENTATION: REDLOCK ALGORITHM CONSIDERATIONS
// =============================================================================

/**
 * ## Redlock Algorithm Considerations
 *
 * This implementation is based on the Redlock algorithm principles but uses
 * a SINGLE Redis instance. Here's what you need to know:
 *
 * ### Why Single-Instance?
 *
 * The full Redlock algorithm requires N independent Redis instances (typically 5)
 * and acquiring a lock on a majority (N/2 + 1) of them. This provides resilience
 * against individual node failures.
 *
 * However, Martin Kleppmann's analysis ("How to do distributed locking") showed
 * that Redlock has fundamental issues that multiple nodes don't solve:
 *
 * 1. **GC Pauses**: A lock holder can be paused (GC, swapping, etc.) for longer
 *    than the TTL. When it resumes, it still thinks it has the lock.
 *
 * 2. **Network Delays**: Messages can be delayed. The lock holder might receive
 *    a "lock acquired" response AFTER the lock has already expired on Redis.
 *
 * 3. **Clock Issues**: Even with multiple nodes, clock issues can cause problems.
 *
 * ### Our Approach: Fencing Tokens
 *
 * Instead of trying to make the lock itself "more reliable" with multiple nodes,
 * we focus on making the DOWNSTREAM SYSTEMS safe:
 *
 * 1. Every lock acquisition generates a monotonically increasing fencing token.
 *
 * 2. This token MUST be passed to all downstream systems (databases, APIs, etc.)
 *
 * 3. Downstream systems MUST reject operations from tokens lower than the last
 *    accepted token.
 *
 * This approach provides safety regardless of whether the lock holder is stale.
 *
 * ### When to Use Multiple Redis Instances
 *
 * Consider implementing full Redlock with multiple instances if:
 *
 * 1. High availability is critical (single Redis = single point of failure)
 * 2. You can accept the complexity of managing multiple Redis instances
 * 3. Your downstream systems ALSO implement fencing token validation
 *
 * Note: Multiple instances WITHOUT fencing tokens does NOT make locks safe!
 *
 * ### References
 *
 * - Martin Kleppmann: "How to do distributed locking"
 *   https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html
 *
 * - Antirez (Redis creator) response: "Is Redlock safe?"
 *   http://antirez.com/news/101
 *
 * - "Designing Data-Intensive Applications" by Martin Kleppmann, Chapter 8
 */

// =============================================================================
// DOCUMENTATION: KNOWN LIMITATIONS AND FAILURE MODES
// =============================================================================

/**
 * ## Known Limitations and Failure Modes
 *
 * ### 1. GC Pause Vulnerability
 *
 * **Scenario:**
 * 1. Process A acquires lock with 30s TTL
 * 2. Process A enters a 35-second GC pause
 * 3. Lock expires on Redis
 * 4. Process B acquires the same lock
 * 5. Process A resumes, still thinks it has the lock
 * 6. Both A and B operate on the protected resource
 *
 * **Mitigation:**
 * - Use fencing tokens in all downstream operations
 * - Check lock validity before EVERY critical operation
 * - Use conservative TTLs relative to GC characteristics
 * - Monitor and alert on long GC pauses
 *
 * ### 2. Network Partition
 *
 * **Scenario:**
 * 1. Process A acquires lock
 * 2. Network partition occurs between A and Redis
 * 3. A cannot extend the lock (heartbeat fails)
 * 4. Lock expires on Redis
 * 5. Process B acquires the lock
 * 6. A is still processing, unaware of partition
 *
 * **Mitigation:**
 * - Fencing tokens are the ONLY reliable protection
 * - Stop processing if heartbeat fails (but too late if already in progress)
 * - Use shorter TTLs with more frequent heartbeats
 *
 * ### 3. Redis Failover
 *
 * **Scenario:**
 * 1. Process A acquires lock on Redis primary
 * 2. Primary crashes BEFORE replicating to replica
 * 3. Replica is promoted to primary
 * 4. Lock does not exist on new primary
 * 5. Process B acquires the "same" lock
 *
 * **Mitigation:**
 * - Use Redis WAIT command for synchronous replication (adds latency)
 * - Use Redis Cluster with proper replica configuration
 * - Accept that some locks may be "lost" on failover
 * - Fencing tokens still protect downstream systems
 *
 * ### 4. Clock Skew
 *
 * **Scenario:**
 * 1. Process A acquires lock with TTL 30s
 * 2. A's clock is 5 seconds ahead of Redis
 * 3. A thinks lock is valid for 25 more seconds
 * 4. Lock actually expires in 20 seconds
 * 5. Race condition window of 5 seconds
 *
 * **Mitigation:**
 * - Use NTP with tight synchronization
 * - Configure clockDriftFactor (default 1%)
 * - We reduce validity by drift factor + acquisition time
 * - Monitor clock drift with checkClockDrift()
 *
 * ### 5. Split Brain with Auto-Heartbeat
 *
 * **Scenario:**
 * 1. Process A acquires lock, heartbeat running
 * 2. Network issue causes heartbeat to fail silently
 * 3. A thinks heartbeat succeeded (fire-and-forget issue)
 * 4. Lock expires
 * 5. B acquires lock
 * 6. A's next heartbeat succeeds (reconnected)
 * 7. Both think they have the lock
 *
 * **Mitigation:**
 * - Auto-heartbeat is DISABLED by default for this reason
 * - If enabled, heartbeatFailed event is emitted
 * - ALWAYS use fencing tokens regardless of heartbeat
 * - Prefer explicit extend() calls with error handling
 *
 * ### 6. Fencing Token Overflow
 *
 * **Scenario:**
 * 1. Fencing token counter uses bigint
 * 2. After 2^63 operations... (not a real concern)
 *
 * **Reality:**
 * - At 1 million locks per second, it takes 292,471 years
 * - Not a practical limitation
 *
 * ### 7. Resource Starvation
 *
 * **Scenario:**
 * 1. Many processes compete for the same lock
 * 2. Some processes may never acquire the lock
 *
 * **Mitigation:**
 * - Use fair queuing at the application level if needed
 * - Consider using Redis streams for fair ordering
 * - Set reasonable waitTimeoutMs
 * - Monitor lock acquisition times
 */

// =============================================================================
// DOCUMENTATION: PRODUCTION CHECKLIST
// =============================================================================

/**
 * ## Production Checklist
 *
 * Before deploying to production, ensure:
 *
 * [ ] Fencing token validation is implemented in ALL downstream systems
 * [ ] Redis persistence is configured (AOF recommended for durability)
 * [ ] Redis Sentinel or Cluster is configured for high availability
 * [ ] NTP is configured and monitored on all nodes
 * [ ] Lock TTLs are at least 10x expected operation time
 * [ ] Monitoring is set up for:
 *     - Lock acquisition latency
 *     - Lock acquisition failures
 *     - Lock extension failures
 *     - Clock drift warnings
 * [ ] Alerting is configured for heartbeatFailed events
 * [ ] Graceful shutdown calls lockManager.shutdown()
 * [ ] Process crash recovery is tested
 * [ ] Network partition scenarios are tested
 * [ ] Load testing with concurrent lock acquisition is done
 */

export {
  basicLockExample,
  withLockExample,
  longOperationExample,
  setupEventHandlers,
  createStructuredLogger,
  safeDatabaseOperation,
};

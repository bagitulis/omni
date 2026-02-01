/**
 * Distributed Lock Manager - Unit Tests
 *
 * These tests demonstrate the testable design with injectable Redis client.
 * Uses a mock Redis implementation for deterministic testing.
 */

import {
  DistributedLockManager,
  LockFailureReason,
  LockNotHeldError,
  LockExpiredError,
} from "./index";
import type { RedisClientInterface, Logger } from "./types";

// =============================================================================
// MOCK REDIS CLIENT
// =============================================================================

/**
 * In-memory mock Redis client for testing.
 * Simulates Redis behavior without network operations.
 */
export class MockRedisClient implements RedisClientInterface {
  private store: Map<string, { value: string; expiresAt: number }> = new Map();
  private counters: Map<string, number> = new Map();
  private serverTimeOffset = 0;
  private shouldFail = false;
  private failureCount = 0;
  private maxFailures = 0;

  /**
   * Simulate network failure for the next N operations
   */
  simulateFailure(count: number): void {
    this.shouldFail = true;
    this.failureCount = 0;
    this.maxFailures = count;
  }

  /**
   * Simulate clock skew between client and server
   */
  setServerTimeOffset(offsetMs: number): void {
    this.serverTimeOffset = offsetMs;
  }

  /**
   * Clear all stored data
   */
  clear(): void {
    this.store.clear();
    this.counters.clear();
    this.shouldFail = false;
    this.failureCount = 0;
    this.serverTimeOffset = 0;
  }

  /**
   * Get the raw value for a key (for test assertions)
   */
  getRaw(key: string): string | null {
    const entry = this.store.get(key);
    if (!entry) return null;
    if (Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return null;
    }
    return entry.value;
  }

  private checkFailure(): void {
    if (this.shouldFail) {
      this.failureCount++;
      if (this.failureCount <= this.maxFailures) {
        throw new Error("Simulated Redis failure");
      }
      this.shouldFail = false;
    }
  }

  async setNX(key: string, value: string, ttlMs: number): Promise<boolean> {
    this.checkFailure();

    // Check for existing non-expired key
    const existing = this.store.get(key);
    if (existing && Date.now() < existing.expiresAt) {
      return false;
    }

    this.store.set(key, {
      value,
      expiresAt: Date.now() + ttlMs,
    });
    return true;
  }

  async get(key: string): Promise<string | null> {
    this.checkFailure();

    const entry = this.store.get(key);
    if (!entry) return null;

    if (Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return null;
    }

    return entry.value;
  }

  async eval(script: string, keys: string[], args: string[]): Promise<unknown> {
    this.checkFailure();

    // Parse and execute common Lua script patterns
    if (script.includes("DEL") && script.includes("GET")) {
      // Release script
      const key = keys[0];
      const expectedValue = args[0];
      const entry = this.store.get(key);

      if (
        entry &&
        entry.value === expectedValue &&
        Date.now() < entry.expiresAt
      ) {
        this.store.delete(key);
        return 1;
      }
      return 0;
    }

    if (script.includes("PEXPIRE") && script.includes("GET")) {
      // Extend script
      const key = keys[0];
      const expectedValue = args[0];
      const newTtlMs = parseInt(args[1], 10);
      const entry = this.store.get(key);

      if (
        entry &&
        entry.value === expectedValue &&
        Date.now() < entry.expiresAt
      ) {
        entry.expiresAt = Date.now() + newTtlMs;
        return 1;
      }
      return 0;
    }

    if (script.includes("INCR")) {
      // Fencing token script
      const key = keys[0];
      const current = this.counters.get(key) ?? 0;
      const newValue = current + 1;
      this.counters.set(key, newValue);
      return newValue;
    }

    throw new Error(
      `Unsupported Lua script in mock: ${script.substring(0, 50)}`,
    );
  }

  async incr(key: string): Promise<number> {
    this.checkFailure();

    const current = this.counters.get(key) ?? 0;
    const newValue = current + 1;
    this.counters.set(key, newValue);
    return newValue;
  }

  async ping(): Promise<boolean> {
    this.checkFailure();
    return true;
  }

  async time(): Promise<[number, number]> {
    this.checkFailure();

    const now = Date.now() + this.serverTimeOffset;
    const seconds = Math.floor(now / 1000);
    const microseconds = (now % 1000) * 1000;
    return [seconds, microseconds];
  }

  async disconnect(): Promise<void> {
    this.clear();
  }
}

// =============================================================================
// TEST LOGGER
// =============================================================================

/**
 * Silent logger for tests (or capture logs for assertions)
 */
export class TestLogger implements Logger {
  public logs: {
    level: string;
    message: string;
    context?: Record<string, unknown>;
  }[] = [];

  debug(message: string, context?: Record<string, unknown>): void {
    this.logs.push({ level: "debug", message, context });
  }

  info(message: string, context?: Record<string, unknown>): void {
    this.logs.push({ level: "info", message, context });
  }

  warn(message: string, context?: Record<string, unknown>): void {
    this.logs.push({ level: "warn", message, context });
  }

  error(message: string, context?: Record<string, unknown>): void {
    this.logs.push({ level: "error", message, context });
  }

  clear(): void {
    this.logs = [];
  }

  hasLog(level: string, messageContains: string): boolean {
    return this.logs.some(
      (log) => log.level === level && log.message.includes(messageContains),
    );
  }
}

// =============================================================================
// TEST HELPER
// =============================================================================

function createTestManager(): {
  manager: DistributedLockManager;
  redis: MockRedisClient;
  logger: TestLogger;
} {
  const redis = new MockRedisClient();
  const logger = new TestLogger();
  const manager = new DistributedLockManager(
    redis,
    {
      nodeId: "test-node",
      defaultTtlMs: 5000,
      clockDriftFactor: 0.01,
      maxRetries: 3,
      retryDelayMs: 50,
      maxRetryDelayMs: 200,
      operationTimeoutMs: 1000,
      enableAutoHeartbeat: false,
      heartbeatIntervalMs: 1000,
    },
    logger,
  );
  return { manager, redis, logger };
}

// =============================================================================
// TESTS
// =============================================================================

/**
 * Test: Basic lock acquisition and release
 */
async function testBasicLockAcquisition(): Promise<void> {
  const { manager, redis } = createTestManager();

  const result = await manager.acquire("test:resource");

  if (!result.success) {
    throw new Error(
      `Expected lock acquisition to succeed, got: ${result.reason}`,
    );
  }

  // Verify lock was created in Redis
  const storedValue = redis.getRaw("dlock:test:resource");
  if (!storedValue) {
    throw new Error("Lock should exist in Redis");
  }

  // Verify fencing token is positive
  if (result.lock.fencingToken.value <= 0n) {
    throw new Error("Fencing token should be positive");
  }

  // Release the lock
  const released = await manager.release(result.lock);
  if (!released) {
    throw new Error("Lock release should succeed");
  }

  // Verify lock was removed from Redis
  const storedAfterRelease = redis.getRaw("dlock:test:resource");
  if (storedAfterRelease !== null) {
    throw new Error("Lock should be removed from Redis after release");
  }

  await manager.shutdown();
  console.log("PASS testBasicLockAcquisition");
}

/**
 * Test: Lock contention - second acquire should fail
 */
async function testLockContention(): Promise<void> {
  const { manager, redis } = createTestManager();

  // First acquire succeeds
  const result1 = await manager.acquire("test:resource");
  if (!result1.success) {
    throw new Error("First acquire should succeed");
  }

  // Second acquire (without waiting) should fail
  const result2 = await manager.acquire("test:resource", {
    waitForLock: false,
  });

  if (result2.success) {
    throw new Error("Second acquire should fail when lock is held");
  }

  if (result2.reason !== LockFailureReason.ALREADY_HELD) {
    throw new Error(`Expected ALREADY_HELD, got: ${result2.reason}`);
  }

  await manager.release(result1.lock);
  await manager.shutdown();
  console.log("PASS testLockContention");
}

/**
 * Test: Lock extension
 */
async function testLockExtension(): Promise<void> {
  const { manager } = createTestManager();

  const result = await manager.acquire("test:resource", { ttlMs: 2000 });
  if (!result.success) {
    throw new Error("Acquire should succeed");
  }

  const originalValidUntil = result.lock.validUntil;

  // Wait a bit
  await new Promise((resolve) => setTimeout(resolve, 100));

  // Extend the lock
  const extendedLock = await manager.extend(result.lock, { ttlMs: 5000 });

  if (extendedLock.validUntil <= originalValidUntil) {
    throw new Error("Extended lock should have later validUntil");
  }

  // Fencing token should remain the same
  if (extendedLock.fencingToken.value !== result.lock.fencingToken.value) {
    throw new Error("Fencing token should not change on extension");
  }

  await manager.release(extendedLock);
  await manager.shutdown();
  console.log("PASS testLockExtension");
}

/**
 * Test: Release with wrong token fails
 */
async function testReleaseWrongToken(): Promise<void> {
  const { manager, redis } = createTestManager();

  const result = await manager.acquire("test:resource");
  if (!result.success) {
    throw new Error("Acquire should succeed");
  }

  // Manually modify the lock token
  const fakeLock = { ...result.lock, token: "wrong-token" };

  // Try to release with wrong token - should return false
  const released = await manager.release(fakeLock);
  if (released) {
    throw new Error("Release with wrong token should fail");
  }

  // Original lock should still exist
  const storedValue = redis.getRaw("dlock:test:resource");
  if (!storedValue) {
    throw new Error("Lock should still exist after failed release");
  }

  // Clean up with correct token
  await manager.release(result.lock);
  await manager.shutdown();
  console.log("PASS testReleaseWrongToken");
}

/**
 * Test: Fencing tokens are monotonically increasing
 */
async function testFencingTokensIncreasing(): Promise<void> {
  const { manager } = createTestManager();

  const tokens: bigint[] = [];

  for (let i = 0; i < 5; i++) {
    const result = await manager.acquire(`test:resource:${i}`);
    if (!result.success) {
      throw new Error(`Acquire ${i} should succeed`);
    }
    tokens.push(result.lock.fencingToken.value);
    await manager.release(result.lock);
  }

  // Verify tokens are strictly increasing
  for (let i = 1; i < tokens.length; i++) {
    if (tokens[i] <= tokens[i - 1]) {
      throw new Error(
        `Token ${i} (${tokens[i]}) should be greater than token ${i - 1} (${tokens[i - 1]})`,
      );
    }
  }

  await manager.shutdown();
  console.log("PASS testFencingTokensIncreasing");
}

/**
 * Test: Redis failure handling
 */
async function testRedisFailureHandling(): Promise<void> {
  const { manager, redis, logger } = createTestManager();

  // Simulate Redis failure
  redis.simulateFailure(1);

  const result = await manager.acquire("test:resource");

  if (result.success) {
    throw new Error("Acquire should fail when Redis fails");
  }

  if (result.reason !== LockFailureReason.REDIS_ERROR) {
    throw new Error(`Expected REDIS_ERROR, got: ${result.reason}`);
  }

  // Should have logged the error
  if (!logger.hasLog("error", "Redis error")) {
    throw new Error("Should log Redis error");
  }

  await manager.shutdown();
  console.log("PASS testRedisFailureHandling");
}

/**
 * Test: withLock helper
 */
async function testWithLockHelper(): Promise<void> {
  const { manager } = createTestManager();

  let lockReceived: typeof import("./types").Lock | null = null;

  const result = await manager.withLock(
    "test:resource",
    async (lock) => {
      lockReceived = lock;
      return 42;
    },
    { ttlMs: 5000 },
  );

  if (result !== 42) {
    throw new Error(`Expected 42, got: ${result}`);
  }

  if (!lockReceived) {
    throw new Error("Lock should have been passed to callback");
  }

  // Lock should be released after withLock completes
  if (manager.isHeld("test:resource")) {
    throw new Error("Lock should be released after withLock");
  }

  await manager.shutdown();
  console.log("PASS testWithLockHelper");
}

/**
 * Test: withLock releases on exception
 */
async function testWithLockReleasesOnException(): Promise<void> {
  const { manager } = createTestManager();

  try {
    await manager.withLock(
      "test:resource",
      async () => {
        throw new Error("Intentional failure");
      },
      { ttlMs: 5000 },
    );
    throw new Error("Should have thrown");
  } catch (error) {
    if ((error as Error).message !== "Intentional failure") {
      throw error;
    }
  }

  // Lock should be released even after exception
  if (manager.isHeld("test:resource")) {
    throw new Error("Lock should be released after exception");
  }

  await manager.shutdown();
  console.log("PASS testWithLockReleasesOnException");
}

/**
 * Test: Lock validity check
 */
async function testLockValidityCheck(): Promise<void> {
  const { manager } = createTestManager();

  const result = await manager.acquire("test:resource", { ttlMs: 1000 });
  if (!result.success) {
    throw new Error("Acquire should succeed");
  }

  // Should be valid initially
  const remaining = manager.validateLock(result.lock);
  if (remaining <= 0) {
    throw new Error("Lock should be valid initially");
  }

  await manager.release(result.lock);
  await manager.shutdown();
  console.log("PASS testLockValidityCheck");
}

/**
 * Test: Event emission
 */
async function testEventEmission(): Promise<void> {
  const { manager } = createTestManager();

  const events: string[] = [];

  manager.on("lockAcquired", () => events.push("acquired"));
  manager.on("lockReleased", () => events.push("released"));

  const result = await manager.acquire("test:resource");
  if (!result.success) {
    throw new Error("Acquire should succeed");
  }

  await manager.release(result.lock);

  if (!events.includes("acquired")) {
    throw new Error("Should emit lockAcquired event");
  }

  if (!events.includes("released")) {
    throw new Error("Should emit lockReleased event");
  }

  await manager.shutdown();
  console.log("PASS testEventEmission");
}

// =============================================================================
// RUN ALL TESTS
// =============================================================================

export async function runAllTests(): Promise<void> {
  console.log("Running Distributed Lock Manager tests...\n");

  const tests = [
    testBasicLockAcquisition,
    testLockContention,
    testLockExtension,
    testReleaseWrongToken,
    testFencingTokensIncreasing,
    testRedisFailureHandling,
    testWithLockHelper,
    testWithLockReleasesOnException,
    testLockValidityCheck,
    testEventEmission,
  ];

  let passed = 0;
  let failed = 0;

  for (const test of tests) {
    try {
      await test();
      passed++;
    } catch (error) {
      failed++;
      console.error(`FAIL ${test.name}: ${(error as Error).message}`);
    }
  }

  console.log(`\nResults: ${passed} passed, ${failed} failed`);

  if (failed > 0) {
    process.exit(1);
  }
}

// Uncomment to run tests directly:
// runAllTests().catch(console.error);

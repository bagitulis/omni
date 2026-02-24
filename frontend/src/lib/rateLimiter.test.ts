import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  RateLimiter,
  createApiRateLimiter,
  createRateLimiterWithRetry,
} from "./rateLimiter";

describe("RateLimiter", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  describe("constructor validation", () => {
    it("throws when maxTokens is zero", () => {
      expect(() => new RateLimiter({ maxTokens: 0, refillRate: 1 })).toThrow(
        "maxTokens must be a positive number",
      );
    });

    it("throws when maxTokens is negative", () => {
      expect(() => new RateLimiter({ maxTokens: -5, refillRate: 1 })).toThrow(
        "maxTokens must be a positive number",
      );
    });

    it("throws when refillRate is zero", () => {
      expect(() => new RateLimiter({ maxTokens: 10, refillRate: 0 })).toThrow(
        "refillRate must be a positive number",
      );
    });

    it("throws when refillRate is negative", () => {
      expect(() => new RateLimiter({ maxTokens: 10, refillRate: -1 })).toThrow(
        "refillRate must be a positive number",
      );
    });

    it("creates instance with valid config", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 2 });
      expect(limiter).toBeInstanceOf(RateLimiter);
    });
  });

  describe("getMaxTokens / getRefillRate", () => {
    it("returns configured maxTokens", () => {
      const limiter = new RateLimiter({ maxTokens: 20, refillRate: 5 });
      expect(limiter.getMaxTokens()).toBe(20);
    });

    it("returns configured refillRate", () => {
      const limiter = new RateLimiter({ maxTokens: 20, refillRate: 5 });
      expect(limiter.getRefillRate()).toBe(5);
    });
  });

  describe("tryConsume", () => {
    it("throws when tokens argument is zero", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(() => limiter.tryConsume(0)).toThrow(
        "tokens must be a positive number",
      );
    });

    it("throws when tokens argument is negative", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(() => limiter.tryConsume(-1)).toThrow(
        "tokens must be a positive number",
      );
    });

    it("allows consuming tokens when bucket is full", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(limiter.tryConsume(5)).toBe(true);
    });

    it("allows consuming all tokens at once", () => {
      const limiter = new RateLimiter({ maxTokens: 5, refillRate: 1 });
      expect(limiter.tryConsume(5)).toBe(true);
    });

    it("rejects when requesting more tokens than available", () => {
      const limiter = new RateLimiter({ maxTokens: 3, refillRate: 1 });
      expect(limiter.tryConsume(4)).toBe(false);
    });

    it("uses default of 1 token when no argument provided", () => {
      const limiter = new RateLimiter({ maxTokens: 1, refillRate: 1 });
      expect(limiter.tryConsume()).toBe(true);
      expect(limiter.tryConsume()).toBe(false);
    });

    it("depletes bucket after many consecutive calls", () => {
      const limiter = new RateLimiter({ maxTokens: 3, refillRate: 1 });
      expect(limiter.tryConsume()).toBe(true);
      expect(limiter.tryConsume()).toBe(true);
      expect(limiter.tryConsume()).toBe(true);
      expect(limiter.tryConsume()).toBe(false);
    });
  });

  describe("getAvailableTokens", () => {
    it("returns maxTokens for fresh limiter", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(limiter.getAvailableTokens()).toBe(10);
    });

    it("reflects tokens after consumption", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      limiter.tryConsume(3);
      expect(limiter.getAvailableTokens()).toBeCloseTo(7, 1);
    });

    it("does not exceed maxTokens after time advance", () => {
      const limiter = new RateLimiter({ maxTokens: 5, refillRate: 10 });
      vi.advanceTimersByTime(10000); // advance 10 seconds
      expect(limiter.getAvailableTokens()).toBe(5);
    });

    it("refills tokens over time", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 2 });
      // Drain the bucket
      limiter.tryConsume(10);
      // Advance 2 seconds — should refill 4 tokens (2 tokens/sec * 2 sec)
      vi.advanceTimersByTime(2000);
      expect(limiter.getAvailableTokens()).toBeGreaterThanOrEqual(3.9);
      expect(limiter.getAvailableTokens()).toBeLessThanOrEqual(4.1);
    });
  });

  describe("reset", () => {
    it("restores bucket to maxTokens after depletion", () => {
      const limiter = new RateLimiter({ maxTokens: 5, refillRate: 1 });
      limiter.tryConsume(5);
      expect(limiter.getAvailableTokens()).toBeCloseTo(0, 1);
      limiter.reset();
      expect(limiter.getAvailableTokens()).toBe(5);
    });
  });

  describe("getTimeUntilTokens", () => {
    it("returns 0 when tokens are already available", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(limiter.getTimeUntilTokens(5)).toBe(0);
    });

    it("returns 0 when exactly enough tokens are available", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 1 });
      expect(limiter.getTimeUntilTokens(10)).toBe(0);
    });

    it("returns positive ms when bucket is depleted", () => {
      const limiter = new RateLimiter({ maxTokens: 5, refillRate: 1 });
      limiter.tryConsume(5);
      // Need 1 more token, refill rate is 1/sec → ~1000ms
      const wait = limiter.getTimeUntilTokens(1);
      expect(wait).toBeGreaterThan(0);
      expect(wait).toBeLessThanOrEqual(1000);
    });

    it("returns time proportional to tokens needed and refill rate", () => {
      const limiter = new RateLimiter({ maxTokens: 10, refillRate: 2 });
      limiter.tryConsume(10);
      // Need 4 tokens at 2/sec = 2 seconds = 2000ms (ceiling)
      const wait = limiter.getTimeUntilTokens(4);
      expect(wait).toBeGreaterThanOrEqual(2000);
    });
  });
});

describe("createApiRateLimiter", () => {
  it("creates limiter with default burst multiplier of 2", () => {
    const limiter = createApiRateLimiter(5);
    expect(limiter.getMaxTokens()).toBe(10); // 5 * 2
    expect(limiter.getRefillRate()).toBe(5);
  });

  it("respects custom burst multiplier", () => {
    const limiter = createApiRateLimiter(4, 3);
    expect(limiter.getMaxTokens()).toBe(12); // 4 * 3
    expect(limiter.getRefillRate()).toBe(4);
  });
});

describe("createRateLimiterWithRetry", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("returns limiter and getRetryAfter helper", () => {
    const { limiter, getRetryAfter } = createRateLimiterWithRetry({
      maxTokens: 5,
      refillRate: 1,
    });
    expect(limiter).toBeInstanceOf(RateLimiter);
    expect(typeof getRetryAfter).toBe("function");
  });

  it("getRetryAfter returns 0 when tokens are available", () => {
    const { getRetryAfter } = createRateLimiterWithRetry({
      maxTokens: 5,
      refillRate: 1,
    });
    expect(getRetryAfter()).toBe(0);
  });

  it("getRetryAfter returns positive seconds when depleted", () => {
    const { limiter, getRetryAfter } = createRateLimiterWithRetry({
      maxTokens: 2,
      refillRate: 1,
    });
    limiter.tryConsume(2);
    const retryAfter = getRetryAfter(1);
    expect(retryAfter).toBeGreaterThanOrEqual(1);
  });

  it("getRetryAfter defaults to requesting 1 token", () => {
    const { limiter, getRetryAfter } = createRateLimiterWithRetry({
      maxTokens: 1,
      refillRate: 1,
    });
    limiter.tryConsume(1);
    const retryAfter = getRetryAfter();
    expect(retryAfter).toBeGreaterThanOrEqual(1);
  });
});

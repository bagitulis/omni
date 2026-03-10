import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import {
  formatCurrency,
  formatPrice,
  formatNumber,
  formatDate,
  formatTime,
  truncateText,
  isEmpty,
  deepClone,
  mergeObjects,
  hasProperty,
  getNestedValue,
  sleep,
  getResourceClass,
  debounce,
  throttle,
} from "./helpers";

describe("formatCurrency", () => {
  it("formats positive number as IDR currency", () => {
    const result = formatCurrency(100000);
    expect(result).toContain("100.000");
    expect(result).toContain("Rp");
  });

  it("formats zero as IDR currency", () => {
    const result = formatCurrency(0);
    expect(result).toContain("0");
  });

  it("handles falsy value (treats as 0)", () => {
    // amount=0 is falsy, goes to fallback 0
    const result = formatCurrency(0);
    expect(result).toBeTruthy();
  });

  it("formats large numbers with thousand separators", () => {
    const result = formatCurrency(1000000);
    expect(result).toContain("1.000.000");
  });

  it("formats negative numbers", () => {
    const result = formatCurrency(-50000);
    expect(result).toContain("50.000");
  });
});

describe("formatPrice", () => {
  it("formats a positive number", () => {
    const result = formatPrice(1500);
    expect(result).toContain("1.500");
  });

  it("formats null as 0", () => {
    const result = formatPrice(null);
    expect(result).toBe("0");
  });

  it("formats undefined as 0", () => {
    const result = formatPrice(undefined);
    expect(result).toBe("0");
  });

  it("formats zero", () => {
    const result = formatPrice(0);
    expect(result).toBe("0");
  });

  it("formats large number with separators", () => {
    const result = formatPrice(1000000);
    expect(result).toContain("1.000.000");
  });
});

describe("formatNumber", () => {
  it("formats positive number", () => {
    const result = formatNumber(1234);
    expect(result).toContain("1.234");
  });

  it("formats zero", () => {
    expect(formatNumber(0)).toBe("0");
  });

  it("formats large number", () => {
    const result = formatNumber(9999999);
    expect(result).toContain("9.999.999");
  });
});

describe("formatDate", () => {
  it("formats a valid date string", () => {
    const result = formatDate("2024-01-15T10:30:00Z");
    expect(result).toBeTruthy();
    expect(typeof result).toBe("string");
  });

  it("formats a Date object", () => {
    const date = new Date("2024-06-01T00:00:00Z");
    const result = formatDate(date);
    expect(result).toBeTruthy();
    expect(typeof result).toBe("string");
  });

  it("returns original string on invalid date", async () => {
    const { logger } = await import("@/lib/logger");
    const result = formatDate("not-a-date");
    expect(result).toBe("not-a-date");
    expect(logger.error).toHaveBeenCalledWith(
      "Error formatting date",
      expect.objectContaining({ error: expect.any(RangeError) }),
    );
  });
});

describe("formatTime", () => {
  it("formats a valid timestamp", () => {
    const result = formatTime("2024-01-15T10:30:00Z");
    expect(result).toBeTruthy();
    expect(typeof result).toBe("string");
  });
});

describe("truncateText", () => {
  it("returns empty string for null/undefined/empty", () => {
    expect(truncateText("")).toBe("");
    expect(truncateText(null as unknown as string)).toBe("");
    expect(truncateText(undefined as unknown as string)).toBe("");
  });

  it("does not truncate text within limit", () => {
    expect(truncateText("hello", 10)).toBe("hello");
  });

  it("truncates text exceeding max length with ellipsis", () => {
    const result = truncateText("hello world", 5);
    expect(result).toBe("hello...");
  });

  it("uses default max length of 100", () => {
    const longText = "a".repeat(101);
    const result = truncateText(longText);
    expect(result).toBe("a".repeat(100) + "...");
  });

  it("exact max length is not truncated", () => {
    const text = "a".repeat(100);
    expect(truncateText(text)).toBe(text);
  });
});

describe("isEmpty", () => {
  it("returns true for null", () => {
    expect(isEmpty(null)).toBe(true);
  });

  it("returns true for undefined", () => {
    expect(isEmpty(undefined)).toBe(true);
  });

  it("returns true for empty string", () => {
    expect(isEmpty("")).toBe(true);
  });

  it("returns true for whitespace-only string", () => {
    expect(isEmpty("   ")).toBe(true);
  });

  it("returns false for non-empty string", () => {
    expect(isEmpty("hello")).toBe(false);
  });

  it("returns true for empty array", () => {
    expect(isEmpty([])).toBe(true);
  });

  it("returns false for non-empty array", () => {
    expect(isEmpty([1, 2])).toBe(false);
  });

  it("returns true for empty object", () => {
    expect(isEmpty({})).toBe(true);
  });

  it("returns false for non-empty object", () => {
    expect(isEmpty({ a: 1 })).toBe(false);
  });

  it("returns false for number 0", () => {
    expect(isEmpty(0)).toBe(false);
  });

  it("returns false for boolean false", () => {
    expect(isEmpty(false)).toBe(false);
  });
});

describe("deepClone", () => {
  it("clones a primitive (returns same value)", () => {
    expect(deepClone(42)).toBe(42);
    expect(deepClone("hello")).toBe("hello");
    expect(deepClone(null)).toBe(null);
  });

  it("deep clones a plain object", () => {
    const original = { a: 1, b: { c: 2 } };
    const clone = deepClone(original);
    expect(clone).toEqual(original);
    expect(clone).not.toBe(original);
    expect(clone.b).not.toBe(original.b);
  });

  it("deep clones an array", () => {
    const original = [1, [2, 3], { x: 4 }];
    const clone = deepClone(original);
    expect(clone).toEqual(original);
    expect(clone).not.toBe(original);
    expect(clone[1]).not.toBe(original[1]);
  });

  it("clones a Date object", () => {
    const date = new Date("2024-01-15");
    const clone = deepClone(date);
    expect(clone).toEqual(date);
    expect(clone).not.toBe(date);
    expect(clone instanceof Date).toBe(true);
  });

  it("handles nested objects", () => {
    const obj = { a: { b: { c: { d: 99 } } } };
    const clone = deepClone(obj);
    clone.a.b.c.d = 0;
    expect(obj.a.b.c.d).toBe(99);
  });
});

describe("mergeObjects", () => {
  it("merges source into target", () => {
    const result = mergeObjects<Record<string, unknown>>({ a: 1 }, { b: 2 });
    expect(result).toEqual({ a: 1, b: 2 });
  });
  it("source values overwrite target values", () => {
    const result = mergeObjects<Record<string, unknown>>(
      { a: 1, b: 2 },
      { b: 99 },
    );
    expect(result).toEqual({ a: 1, b: 99 });
  });
  it("handles multiple sources", () => {
    const result = mergeObjects<Record<string, unknown>>(
      { a: 1 },
      { b: 2 },
      { c: 3 },
    );
    expect(result).toEqual({ a: 1, b: 2, c: 3 });
  });
  it("does not mutate the target", () => {
    const target: Record<string, unknown> = { a: 1 };
    mergeObjects(target, { b: 2 });
    expect(target).toEqual({ a: 1 });
  });
});

describe("hasProperty", () => {
  it("returns true when object has property", () => {
    expect(hasProperty({ a: 1 }, "a")).toBe(true);
  });

  it("returns false when object does not have property", () => {
    expect(hasProperty({ a: 1 }, "b")).toBe(false);
  });

  it("returns false for null", () => {
    expect(hasProperty(null, "a")).toBe(false);
  });

  it("returns false for non-object (string)", () => {
    expect(hasProperty("hello", "length")).toBe(false);
  });

  it("returns true for inherited property via in operator", () => {
    const obj = Object.create({ inherited: true });
    expect(hasProperty(obj, "inherited")).toBe(true);
  });
});

describe("getNestedValue", () => {
  it("gets top-level value", () => {
    expect(getNestedValue({ a: 1 }, "a")).toBe(1);
  });

  it("gets nested value with dot notation", () => {
    expect(getNestedValue({ a: { b: { c: 42 } } }, "a.b.c")).toBe(42);
  });

  it("returns defaultValue when path not found", () => {
    expect(getNestedValue({ a: 1 }, "b.c", "default")).toBe("default");
  });

  it("returns undefined when no defaultValue and path not found", () => {
    expect(getNestedValue({ a: 1 }, "b")).toBeUndefined();
  });

  it("returns defaultValue when intermediate path is null", () => {
    expect(getNestedValue({ a: null }, "a.b", "fallback")).toBe("fallback");
  });
});

describe("sleep", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("resolves after specified milliseconds", async () => {
    const resolved = vi.fn();
    const promise = sleep(1000).then(resolved);
    expect(resolved).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    await promise;
    expect(resolved).toHaveBeenCalled();
  });

  it("does not resolve before timeout", async () => {
    const resolved = vi.fn();
    sleep(500).then(resolved);
    vi.advanceTimersByTime(499);
    expect(resolved).not.toHaveBeenCalled();
  });
});

describe("getResourceClass", () => {
  it("returns critical for > 80%", () => {
    expect(getResourceClass(81)).toBe("critical");
    expect(getResourceClass(100)).toBe("critical");
  });

  it("returns warning for > 60% and <= 80%", () => {
    expect(getResourceClass(61)).toBe("warning");
    expect(getResourceClass(80)).toBe("warning");
  });

  it("returns healthy for <= 60%", () => {
    expect(getResourceClass(60)).toBe("healthy");
    expect(getResourceClass(0)).toBe("healthy");
  });
});

describe("debounce", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  it("delays function execution by specified wait time", () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 300);
    debounced();
    expect(fn).not.toHaveBeenCalled();
    vi.advanceTimersByTime(300);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("only calls function once for rapid successive calls", () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 300);
    debounced();
    debounced();
    debounced();
    vi.advanceTimersByTime(300);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("passes arguments to the underlying function", () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);
    debounced("arg1", "arg2");
    vi.advanceTimersByTime(100);
    expect(fn).toHaveBeenCalledWith("arg1", "arg2");
  });
});

describe("throttle", () => {
  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  it("calls function immediately on first invocation", () => {
    const fn = vi.fn();
    const throttled = throttle(fn, 1000);
    throttled();
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("does not call function again within throttle limit", () => {
    const fn = vi.fn();
    const throttled = throttle(fn, 1000);
    throttled();
    throttled();
    throttled();
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("allows call after limit has passed", async () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttled = throttle(fn, 100);
    throttled();
    expect(fn).toHaveBeenCalledTimes(1);
    vi.advanceTimersByTime(101);
    throttled();
    expect(fn).toHaveBeenCalledTimes(2);
  });
});

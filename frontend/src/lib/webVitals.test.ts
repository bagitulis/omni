import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { initPerformanceMonitoring } from "./webVitals";

// Mock web-vitals module
const mockOnCLS = vi.fn();
const mockOnINP = vi.fn();
const mockOnFCP = vi.fn();
const mockOnLCP = vi.fn();
const mockOnTTFB = vi.fn();

vi.mock("web-vitals", () => ({
  onCLS: mockOnCLS,
  onINP: mockOnINP,
  onFCP: mockOnFCP,
  onLCP: mockOnLCP,
  onTTFB: mockOnTTFB,
}));

// Mock the logger
vi.mock("@/lib/logger", () => ({
  logger: {
    info: vi.fn(),
    error: vi.fn(),
    warn: vi.fn(),
  },
}));

describe("initPerformanceMonitoring", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("registers all five web vital handlers with default callback", async () => {
    initPerformanceMonitoring();
    // Allow dynamic import to resolve
    await vi.dynamicImportSettled();

    expect(mockOnCLS).toHaveBeenCalledTimes(1);
    expect(mockOnINP).toHaveBeenCalledTimes(1);
    expect(mockOnFCP).toHaveBeenCalledTimes(1);
    expect(mockOnLCP).toHaveBeenCalledTimes(1);
    expect(mockOnTTFB).toHaveBeenCalledTimes(1);
  });

  it("passes the same callback to all five handlers", async () => {
    initPerformanceMonitoring();
    await vi.dynamicImportSettled();

    const clsCallback = mockOnCLS.mock.calls[0][0];
    const inpCallback = mockOnINP.mock.calls[0][0];
    const fcpCallback = mockOnFCP.mock.calls[0][0];
    const lcpCallback = mockOnLCP.mock.calls[0][0];
    const ttfbCallback = mockOnTTFB.mock.calls[0][0];

    // All handlers receive the same callback function reference
    expect(clsCallback).toBe(inpCallback);
    expect(clsCallback).toBe(fcpCallback);
    expect(clsCallback).toBe(lcpCallback);
    expect(clsCallback).toBe(ttfbCallback);
  });

  it("accepts a custom callback and passes it to all handlers", async () => {
    const customCallback = vi.fn();
    initPerformanceMonitoring(customCallback);
    await vi.dynamicImportSettled();

    expect(mockOnCLS).toHaveBeenCalledWith(customCallback);
    expect(mockOnINP).toHaveBeenCalledWith(customCallback);
    expect(mockOnFCP).toHaveBeenCalledWith(customCallback);
    expect(mockOnLCP).toHaveBeenCalledWith(customCallback);
    expect(mockOnTTFB).toHaveBeenCalledWith(customCallback);
  });

  it("does not throw when called multiple times", () => {
    expect(() => {
      initPerformanceMonitoring();
      initPerformanceMonitoring();
    }).not.toThrow();
  });

  it("does not throw when called without arguments", () => {
    expect(() => initPerformanceMonitoring()).not.toThrow();
  });

  it("custom callback is a function", async () => {
    const myCallback = vi.fn();
    initPerformanceMonitoring(myCallback);
    await vi.dynamicImportSettled();

    const receivedCallback = mockOnCLS.mock.calls[0][0];
    expect(typeof receivedCallback).toBe("function");
  });
});

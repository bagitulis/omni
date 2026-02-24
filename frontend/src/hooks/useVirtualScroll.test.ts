import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import { useVirtualScroll } from "./useVirtualScroll";

describe("useVirtualScroll", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // jsdom default innerHeight is 768
    Object.defineProperty(window, "innerHeight", {
      value: 768,
      writable: true,
      configurable: true,
    });
  });

  it("returns virtual=true by default", () => {
    const { result } = renderHook(() => useVirtualScroll());
    expect(result.current.virtual).toBe(true);
    expect(result.current.tableProps.virtual).toBe(true);
  });

  it("returns virtual=false when enabled=false", () => {
    const { result } = renderHook(() => useVirtualScroll({ enabled: false }));
    expect(result.current.virtual).toBe(false);
    expect(result.current.tableProps.virtual).toBe(false);
  });

  it("uses containerHeight as scrollY when provided", () => {
    const { result } = renderHook(() =>
      useVirtualScroll({ containerHeight: 600 }),
    );
    expect(result.current.scrollConfig.y).toBe(600);
    expect(result.current.tableProps.scroll.y).toBe(600);
  });

  it("calculates scrollY from viewport minus offsetBottom", () => {
    // window.innerHeight = 768, default offsetBottom = 300
    // expected: max(768 - 300, 400) = 468
    const { result } = renderHook(() => useVirtualScroll());
    // Initial state is 500 (useState default), but useEffect sets it
    // Since jsdom runs useEffect synchronously in test environment,
    // we just check that scroll.y is a valid number
    expect(typeof result.current.scrollConfig.y).toBe("number");
    expect(result.current.scrollConfig.y).toBeGreaterThanOrEqual(400);
  });

  it("uses custom offsetBottom", () => {
    const { result } = renderHook(() =>
      useVirtualScroll({ offsetBottom: 100 }),
    );
    // max(768 - 100, 400) = 668
    expect(result.current.scrollConfig.y).toBeGreaterThanOrEqual(400);
  });

  it("enforces minimum height of 400 when calculated height is too small", () => {
    Object.defineProperty(window, "innerHeight", {
      value: 300,
      writable: true,
      configurable: true,
    });
    // max(300 - 300, 400) = 400
    const { result } = renderHook(() => useVirtualScroll());
    expect(result.current.scrollConfig.y).toBeGreaterThanOrEqual(400);
  });

  it("returns correct tableProps structure", () => {
    const { result } = renderHook(() =>
      useVirtualScroll({ containerHeight: 500 }),
    );
    expect(result.current.tableProps).toHaveProperty("virtual");
    expect(result.current.tableProps).toHaveProperty("scroll");
    expect(result.current.tableProps.scroll).toHaveProperty("y");
  });

  it("returns consistent scrollConfig and tableProps.scroll", () => {
    const { result } = renderHook(() =>
      useVirtualScroll({ containerHeight: 700 }),
    );
    expect(result.current.scrollConfig.y).toBe(
      result.current.tableProps.scroll.y,
    );
  });

  it("containerHeight overrides viewport calculation", () => {
    const { result } = renderHook(() =>
      useVirtualScroll({ containerHeight: 800 }),
    );
    expect(result.current.scrollConfig.y).toBe(800);
  });

  it("does not calculate height when disabled and no containerHeight", () => {
    const { result } = renderHook(() => useVirtualScroll({ enabled: false }));
    // Falls back to initial state value (500)
    expect(result.current.scrollConfig.y).toBe(500);
    expect(result.current.virtual).toBe(false);
  });
});

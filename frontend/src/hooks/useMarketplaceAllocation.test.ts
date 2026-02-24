import { beforeEach, describe, expect, it } from "vitest";
import {
  calculateAllocation,
  useMarketplaceAllocation,
} from "./useMarketplaceAllocation";

describe("calculateAllocation", () => {
  beforeEach(() => {});

  it("returns zeros for total <= 0", () => {
    expect(calculateAllocation(0, false)).toEqual({
      shopee: 0,
      tiktok: 0,
      lazada: 0,
    });
    expect(calculateAllocation(-5, false)).toEqual({
      shopee: 0,
      tiktok: 0,
      lazada: 0,
    });
  });

  it("returns zeros when no platforms available", () => {
    const result = calculateAllocation(100, false, {
      shopee: false,
      tiktok: false,
      lazada: false,
    });
    expect(result).toEqual({ shopee: 0, tiktok: 0, lazada: 0 });
  });

  it("AUTO mode: all available platforms get total", () => {
    const result = calculateAllocation(100, true);
    expect(result).toEqual({ shopee: 100, tiktok: 100, lazada: 100 });
  });

  it("AUTO mode: unavailable platform gets 0", () => {
    const result = calculateAllocation(100, true, { shopee: false });
    expect(result.shopee).toBe(0);
    expect(result.tiktok).toBe(100);
    expect(result.lazada).toBe(100);
  });

  it("non-AUTO mode 3 platforms: shopee ~60%, tiktok ~30%, lazada remainder", () => {
    const result = calculateAllocation(100, false);
    expect(result.shopee).toBeGreaterThanOrEqual(60);
    expect(result.tiktok).toBeGreaterThanOrEqual(30);
    expect(result.shopee + result.tiktok + result.lazada).toBe(100);
  });

  it("non-AUTO mode only shopee: gets full total", () => {
    const result = calculateAllocation(50, false, {
      shopee: true,
      tiktok: false,
      lazada: false,
    });
    expect(result).toEqual({ shopee: 50, tiktok: 0, lazada: 0 });
  });

  it("non-AUTO mode shopee+tiktok: shopee ~60%, tiktok remainder", () => {
    const result = calculateAllocation(100, false, {
      shopee: true,
      tiktok: true,
      lazada: false,
    });
    expect(result.shopee).toBeGreaterThan(0);
    expect(result.tiktok).toBeGreaterThan(0);
    expect(result.lazada).toBe(0);
    expect(result.shopee + result.tiktok).toBe(100);
  });

  it("returns zeros for NaN total", () => {
    const result = calculateAllocation(NaN, false);
    expect(result).toEqual({ shopee: 0, tiktok: 0, lazada: 0 });
  });
});

describe("useMarketplaceAllocation", () => {
  it("returns calculateAllocation function", () => {
    const result = useMarketplaceAllocation();
    expect(typeof result.calculateAllocation).toBe("function");
  });

  it("calculateAllocation from hook works correctly", () => {
    const { calculateAllocation: calc } = useMarketplaceAllocation();
    expect(calc(0, false)).toEqual({ shopee: 0, tiktok: 0, lazada: 0 });
  });
});

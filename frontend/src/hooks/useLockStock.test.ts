import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";

vi.mock("@/api/lockedOrders", () => ({
  getLockedOrders: vi.fn(),
}));

import { useLockStock } from "./useLockStock";
import * as lockedOrdersApi from "@/api/lockedOrders";

describe("useLockStock", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns initial state with empty lockedStockMap", () => {
    const { result } = renderHook(() => useLockStock());
    expect(result.current.lockedStockMap).toEqual({});
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.hasLockedOrders).toBe(false);
  });

  it("getLockedQty returns 0 for unknown SKU", () => {
    const { result } = renderHook(() => useLockStock());
    expect(result.current.getLockedQty("SKU-UNKNOWN")).toBe(0);
  });

  it("calculateSellableStock returns inventoryTotal when no locked qty", () => {
    const { result } = renderHook(() => useLockStock());
    expect(result.current.calculateSellableStock("SKU-A", 100)).toBe(100);
  });

  it("calculateSellableStock returns 0 when locked qty exceeds inventory", () => {
    const { result } = renderHook(() => useLockStock());
    // With empty map, locked=0, so sellable = max(0, 5-0) = 5
    expect(result.current.calculateSellableStock("SKU-A", 5)).toBe(5);
  });

  it("fetchLockedStock calls getLockedOrders", async () => {
    vi.mocked(lockedOrdersApi.getLockedOrders).mockResolvedValue([]);
    const { result } = renderHook(() => useLockStock());
    await act(async () => {
      await result.current.fetchLockedStock();
    });
    expect(lockedOrdersApi.getLockedOrders).toHaveBeenCalledOnce();
  });

  it("fetchLockedStock returns empty map when no locked orders", async () => {
    vi.mocked(lockedOrdersApi.getLockedOrders).mockResolvedValue([]);
    const { result } = renderHook(() => useLockStock());
    let map: Record<string, number> = {};
    await act(async () => {
      map = await result.current.fetchLockedStock();
    });
    expect(map).toEqual({});
  });

  it("fetchLockedStock returns empty map on error", async () => {
    vi.mocked(lockedOrdersApi.getLockedOrders).mockRejectedValue(
      new Error("API Error"),
    );
    const { result } = renderHook(() => useLockStock());
    let map: Record<string, number> = {};
    await act(async () => {
      map = await result.current.fetchLockedStock();
    });
    expect(map).toEqual({});
  });
});

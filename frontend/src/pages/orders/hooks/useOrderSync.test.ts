import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
}));

vi.mock("@/api/orders", () => ({
  isSyncableOrderTab: vi.fn((tab: string) =>
    ["unprocess", "processed", "shipped", "completed", "cancelled"].includes(
      tab,
    ),
  ),
  syncOrdersByCategory: vi.fn(),
  lockOrdersToday: vi.fn(),
  syncOrdersToday: vi.fn(),
}));

import { useOrderSync } from "./useOrderSync";
import * as ordersApi from "@/api/orders";
import { message } from "antd";

describe("useOrderSync", () => {
  const refetch = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(ordersApi.syncOrdersByCategory).mockResolvedValue(undefined);
    vi.mocked(ordersApi.syncOrdersToday).mockResolvedValue([]);
    vi.mocked(ordersApi.lockOrdersToday).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls syncOrdersByCategory for syncable tabs on mount", async () => {
    const { result } = renderHook(() =>
      useOrderSync("unprocess", "shopee", refetch, false),
    );

    await waitFor(() => {
      expect(result.current.isSyncing).toBe(false);
    });

    expect(ordersApi.syncOrdersByCategory).toHaveBeenCalledWith(
      "unprocess",
      "shopee",
    );
    expect(refetch).toHaveBeenCalled();
  });

  it("calls syncOrdersToday for 'today' tab", async () => {
    vi.mocked(ordersApi.isSyncableOrderTab).mockImplementation(
      (tab: string) => tab !== "today" && tab !== "locked",
    );

    renderHook(() => useOrderSync("today", "all", refetch, false));

    await waitFor(() => {
      expect(ordersApi.syncOrdersToday).toHaveBeenCalled();
    });
  });

  it("calls lockOrdersToday for 'locked' tab", async () => {
    vi.mocked(ordersApi.isSyncableOrderTab).mockImplementation(
      (tab: string) => tab !== "today" && tab !== "locked",
    );

    renderHook(() => useOrderSync("locked", "all", refetch, false));

    await waitFor(() => {
      expect(ordersApi.lockOrdersToday).toHaveBeenCalled();
    });
  });

  it("shows error message when sync fails", async () => {
    vi.mocked(ordersApi.syncOrdersByCategory).mockRejectedValue(
      new Error("Sync failed"),
    );

    renderHook(() => useOrderSync("unprocess", "shopee", refetch, false));

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith("Sync failed");
    });
  });

  it("shows fallback error message for non-Error rejection", async () => {
    vi.mocked(ordersApi.syncOrdersByCategory).mockRejectedValue("unknown");

    renderHook(() => useOrderSync("unprocess", "shopee", refetch, false));

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith("Failed to sync orders");
    });
  });

  it("returns isSyncing=false initially and after sync", async () => {
    const { result } = renderHook(() =>
      useOrderSync("unprocess", "shopee", refetch, false),
    );

    await waitFor(() => {
      expect(result.current.isSyncing).toBe(false);
    });
  });

  it("exposes syncActiveTab function", async () => {
    const { result } = renderHook(() =>
      useOrderSync("unprocess", "shopee", refetch, false),
    );
    await waitFor(() => {
      expect(result.current.isSyncing).toBe(false);
    });
    expect(typeof result.current.syncActiveTab).toBe("function");
  });

  it("syncActiveTab calls syncOrdersByCategory when called manually", async () => {
    vi.mocked(ordersApi.syncOrdersByCategory).mockResolvedValue(undefined);

    const { result } = renderHook(() =>
      useOrderSync("unprocess", "shopee", refetch, false),
    );

    // Wait for initial sync
    await waitFor(() => expect(result.current.isSyncing).toBe(false));

    vi.clearAllMocks();
    vi.mocked(ordersApi.syncOrdersByCategory).mockResolvedValue(undefined);

    await act(async () => {
      await result.current.syncActiveTab("processed");
    });

    expect(ordersApi.syncOrdersByCategory).toHaveBeenCalledWith(
      "processed",
      "shopee",
    );
  });

  it("sets up auto-refresh interval when autoRefresh=true and tab is syncable", async () => {
    vi.useFakeTimers();
    vi.mocked(ordersApi.syncOrdersByCategory).mockResolvedValue(undefined);

    const { unmount } = renderHook(() =>
      useOrderSync("unprocess", "shopee", refetch, true),
    );

    await act(async () => {
      await Promise.resolve();
    });
    expect(ordersApi.syncOrdersByCategory).toHaveBeenCalledTimes(1);

    await act(async () => {
      vi.advanceTimersByTime(30_000);
      await Promise.resolve();
    });

    // autoRefresh interval should have fired
    expect(ordersApi.syncOrdersByCategory).toHaveBeenCalledTimes(2);

    unmount();
  });

  it("does not set up interval for non-syncable tab even if autoRefresh=true", async () => {
    vi.useFakeTimers();
    vi.mocked(ordersApi.isSyncableOrderTab).mockReturnValue(false);
    vi.mocked(ordersApi.lockOrdersToday).mockResolvedValue([]);

    const { unmount } = renderHook(() =>
      useOrderSync("locked", "all", refetch, true),
    );

    await act(async () => {
      await Promise.resolve();
    });
    expect(ordersApi.lockOrdersToday).toHaveBeenCalledTimes(1);

    await act(async () => {
      vi.advanceTimersByTime(60_000);
      await Promise.resolve();
    });

    // interval should NOT fire for non-syncable tab
    expect(ordersApi.syncOrdersByCategory).not.toHaveBeenCalled();
    expect(ordersApi.lockOrdersToday).toHaveBeenCalledTimes(1);

    unmount();
  });
});

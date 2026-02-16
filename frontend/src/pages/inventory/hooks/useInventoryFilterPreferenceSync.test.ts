import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useInventoryFilterPreferenceSync } from "./useInventoryFilterPreferenceSync";

interface HookProps {
  preferencesLoaded: boolean;
  visibleColumns: string[];
  lockedColumns: string[];
  columnFilters: Record<string, string>;
  searchQuery: string;
}

describe("useInventoryFilterPreferenceSync", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it("does not save immediately when preferences are first loaded", () => {
    const saveFilterPreferences = vi.fn();

    const { rerender } = renderHook(
      (props: HookProps) =>
        useInventoryFilterPreferenceSync({
          ...props,
          saveFilterPreferences,
          debounceMs: 10,
        }),
      {
        initialProps: {
          preferencesLoaded: false,
          visibleColumns: ["SKU"],
          lockedColumns: [],
          columnFilters: {},
          searchQuery: "",
        },
      },
    );

    rerender({
      preferencesLoaded: true,
      visibleColumns: ["SKU"],
      lockedColumns: [],
      columnFilters: {},
      searchQuery: "",
    });

    act(() => {
      vi.advanceTimersByTime(20);
    });

    expect(saveFilterPreferences).not.toHaveBeenCalled();
  });

  it("saves once after payload changes and skips identical rerenders", () => {
    const saveFilterPreferences = vi.fn(
      (_payload: unknown, options?: { onSuccess?: () => void }) => {
        options?.onSuccess?.();
      },
    );

    const { rerender } = renderHook(
      (props: HookProps) =>
        useInventoryFilterPreferenceSync({
          ...props,
          saveFilterPreferences,
          debounceMs: 10,
        }),
      {
        initialProps: {
          preferencesLoaded: true,
          visibleColumns: ["SKU"],
          lockedColumns: [],
          columnFilters: {},
          searchQuery: "",
        },
      },
    );

    rerender({
      preferencesLoaded: true,
      visibleColumns: ["SKU"],
      lockedColumns: [],
      columnFilters: {},
      searchQuery: "shirt",
    });

    act(() => {
      vi.advanceTimersByTime(20);
    });

    expect(saveFilterPreferences).toHaveBeenCalledTimes(1);

    rerender({
      preferencesLoaded: true,
      visibleColumns: ["SKU"],
      lockedColumns: [],
      columnFilters: {},
      searchQuery: "shirt",
    });

    act(() => {
      vi.advanceTimersByTime(20);
    });

    expect(saveFilterPreferences).toHaveBeenCalledTimes(1);
  });
});

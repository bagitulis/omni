import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { useColumnManager } from "./useColumnManager";
import type { ColumnConfig } from "@/types/shared";

// Mock localStorage
const mockLocalStorage = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, "localStorage", {
  value: mockLocalStorage,
  writable: true,
});

describe("useColumnManager", () => {
  const storageKey = "test-columns";
  const defaultColumns: ColumnConfig[] = [
    { key: "id", title: "ID", visible: true, locked: true, order: 0 },
    { key: "name", title: "Name", visible: true, locked: true, order: 1 },
    { key: "price", title: "Price", visible: true, order: 2 },
    { key: "stock", title: "Stock", visible: true, order: 3 },
    { key: "status", title: "Status", visible: false, order: 4 },
  ];

  beforeEach(() => {
    mockLocalStorage.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should initialize with default columns", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    expect(result.current.columns).toEqual(defaultColumns);
    expect(result.current.visibleColumns).toHaveLength(4); // Only visible columns
  });

  it("should load columns from localStorage on mount", () => {
    const storedColumns: ColumnConfig[] = [
      { key: "id", title: "ID", visible: true, locked: true, order: 0 },
      { key: "name", title: "Name", visible: true, locked: true, order: 1 },
      { key: "price", title: "Price", visible: false, order: 2 }, // Hidden by user
      { key: "stock", title: "Stock", visible: true, order: 3 },
      { key: "status", title: "Status", visible: true, order: 4 }, // Shown by user
    ];

    mockLocalStorage.setItem(storageKey, JSON.stringify(storedColumns));

    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    expect(result.current.columns[2].visible).toBe(false); // price hidden
    expect(result.current.columns[4].visible).toBe(true); // status shown
    expect(result.current.visibleColumns).toHaveLength(4); // id, name, stock, status
  });

  it("should toggle column visibility", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    act(() => {
      result.current.toggleVisibility("price");
    });

    expect(result.current.columns[2].visible).toBe(false);
    expect(result.current.visibleColumns).toHaveLength(3); // One less visible
  });

  it("should not toggle locked column visibility", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    act(() => {
      result.current.toggleVisibility("name"); // locked column
    });

    expect(result.current.columns[1].visible).toBe(true); // Still visible
  });

  it("should update column order", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    const reordered = [
      result.current.columns[1], // name
      result.current.columns[0], // id
      result.current.columns[2], // price
      result.current.columns[3], // stock
      result.current.columns[4], // status
    ];

    act(() => {
      result.current.updateOrder(reordered);
    });

    expect(result.current.columns[0].key).toBe("name");
    expect(result.current.columns[0].order).toBe(0);
    expect(result.current.columns[1].key).toBe("id");
    expect(result.current.columns[1].order).toBe(1);
  });

  it("should reset to default columns", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    // Modify columns
    act(() => {
      result.current.toggleVisibility("price");
      result.current.toggleVisibility("status");
    });

    expect(result.current.columns[2].visible).toBe(false); // price hidden
    expect(result.current.columns[4].visible).toBe(true); // status shown

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.columns).toEqual(defaultColumns);
    expect(result.current.columns[2].visible).toBe(true); // price back to default
    expect(result.current.columns[4].visible).toBe(false); // status back to default
  });

  it("should persist changes to localStorage", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    act(() => {
      result.current.toggleVisibility("price");
    });

    const stored = mockLocalStorage.getItem(storageKey);
    expect(stored).toBeTruthy();

    const parsed = JSON.parse(stored!) as ColumnConfig[];
    expect(parsed[2].visible).toBe(false); // price hidden
  });

  it("should handle localStorage errors gracefully (load)", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    mockLocalStorage.setItem(storageKey, "invalid-json");

    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    expect(result.current.columns).toEqual(defaultColumns); // Falls back to defaults
    expect(consoleErrorSpy).toHaveBeenCalledWith(
      expect.stringContaining("Failed to load column preferences"),
      expect.any(Error),
    );

    consoleErrorSpy.mockRestore();
  });

  it("should handle localStorage errors gracefully (save)", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    // Mock setItem to throw error
    const originalSetItem = mockLocalStorage.setItem;
    mockLocalStorage.setItem = () => {
      throw new Error("QuotaExceededError");
    };

    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    act(() => {
      result.current.toggleVisibility("price");
    });

    expect(consoleErrorSpy).toHaveBeenCalledWith(
      expect.stringContaining("Failed to save column preferences"),
      expect.any(Error),
    );

    // Restore
    mockLocalStorage.setItem = originalSetItem;
    consoleErrorSpy.mockRestore();
  });

  it("should merge new columns from defaults with stored preferences", () => {
    const storedColumns: ColumnConfig[] = [
      { key: "id", title: "ID", visible: true, locked: true, order: 0 },
      { key: "name", title: "Name", visible: false, locked: true, order: 1 }, // User hid this
      { key: "price", title: "Price", visible: true, order: 2 },
      // 'stock' and 'status' not in stored (new columns added to app)
    ];

    mockLocalStorage.setItem(storageKey, JSON.stringify(storedColumns));

    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    // User preference preserved for existing columns
    expect(result.current.columns[1].visible).toBe(false); // name hidden

    // New columns added from defaults
    expect(result.current.columns[3].key).toBe("stock");
    expect(result.current.columns[3].visible).toBe(true); // Default visibility

    expect(result.current.columns[4].key).toBe("status");
    expect(result.current.columns[4].visible).toBe(false); // Default visibility
  });

  it("should return visibleColumns sorted by order", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    // Reorder: move price to end
    const reordered = [
      result.current.columns[0], // id
      result.current.columns[1], // name
      result.current.columns[3], // stock (moved up)
      result.current.columns[2], // price (moved down)
      result.current.columns[4], // status
    ];

    act(() => {
      result.current.updateOrder(reordered);
    });

    const visibleKeys = result.current.visibleColumns.map((col) => col.key);
    expect(visibleKeys).toEqual(["id", "name", "stock", "price"]); // Sorted by new order, status hidden
  });

  it("should handle empty defaultColumns", () => {
    const { result } = renderHook(() => useColumnManager([], storageKey));

    expect(result.current.columns).toEqual([]);
    expect(result.current.visibleColumns).toEqual([]);
  });

  it("should update visibleColumns reactively when toggling visibility", () => {
    const { result } = renderHook(() =>
      useColumnManager(defaultColumns, storageKey),
    );

    const initialVisible = result.current.visibleColumns.length;

    act(() => {
      result.current.toggleVisibility("price");
    });

    expect(result.current.visibleColumns.length).toBe(initialVisible - 1);

    act(() => {
      result.current.toggleVisibility("status");
    });

    expect(result.current.visibleColumns.length).toBe(initialVisible); // One hidden, one shown
  });
});

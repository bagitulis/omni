import { useState, useEffect, useCallback } from "react";
import type { ColumnConfig } from "@/types/shared";

/**
 * Hook for managing column visibility, ordering, and persistence
 *
 * Features:
 * - Toggle column visibility (with locked column protection)
 * - Reorder columns by drag-and-drop or programmatic update
 * - Reset to default configuration
 * - Persist preferences to localStorage with namespace
 *
 * @param defaultColumns - Initial column configuration
 * @param storageKey - localStorage key for persistence (e.g., "product-table-columns")
 * @returns Column manager state and actions
 */
export function useColumnManager(
  defaultColumns: ColumnConfig[],
  storageKey: string,
) {
  // Initialize state from localStorage or default
  const [columns, setColumns] = useState<ColumnConfig[]>(() => {
    try {
      const stored = localStorage.getItem(storageKey);
      if (stored) {
        const parsed = JSON.parse(stored) as ColumnConfig[];
        // Merge with defaults to handle new columns added after user saved preference
        return mergeColumns(defaultColumns, parsed);
      }
    } catch (error) {
      console.error(
        `Failed to load column preferences from ${storageKey}:`,
        error,
      );
    }
    return defaultColumns;
  });

  // Persist to localStorage whenever columns change
  useEffect(() => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(columns));
    } catch (error) {
      console.error(
        `Failed to save column preferences to ${storageKey}:`,
        error,
      );
    }
  }, [columns, storageKey]);

  /**
   * Toggle visibility of a single column
   * Locked columns cannot be hidden
   */
  const toggleVisibility = useCallback((columnKey: string) => {
    setColumns((prev) =>
      prev.map((col) => {
        if (col.key === columnKey && !col.locked) {
          return { ...col, visible: !col.visible };
        }
        return col;
      }),
    );
  }, []);

  /**
   * Update column order
   * @param newColumns - Reordered column array
   */
  const updateOrder = useCallback((newColumns: ColumnConfig[]) => {
    // Recalculate order property based on array index
    const reordered = newColumns.map((col, index) => ({
      ...col,
      order: index,
    }));
    setColumns(reordered);
  }, []);

  /**
   * Reset columns to default configuration
   */
  const reset = useCallback(() => {
    setColumns(defaultColumns);
  }, [defaultColumns]);

  /**
   * Get only visible columns, sorted by order
   */
  const visibleColumns = columns
    .filter((col) => col.visible)
    .sort((a, b) => a.order - b.order);

  return {
    columns,
    visibleColumns,
    toggleVisibility,
    updateOrder,
    reset,
  };
}

/**
 * Merge default columns with stored user preferences
 * - Preserves user's visibility and order preferences
 * - Adds new columns from defaults if not in stored config
 * - Removes columns from stored config if not in defaults (deleted columns)
 */
function mergeColumns(
  defaults: ColumnConfig[],
  stored: ColumnConfig[],
): ColumnConfig[] {
  const storedMap = new Map(stored.map((col) => [col.key, col]));

  // Start with defaults, override with stored preferences
  const merged = defaults.map((defaultCol) => {
    const storedCol = storedMap.get(defaultCol.key);
    if (storedCol) {
      // Preserve user preferences, but keep default metadata (title, locked, width)
      return {
        ...defaultCol,
        visible: storedCol.visible,
        order: storedCol.order,
      };
    }
    return defaultCol;
  });

  return merged;
}

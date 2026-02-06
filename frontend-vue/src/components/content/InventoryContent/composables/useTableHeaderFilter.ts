import { ref } from "vue";
import { useInventoryConfig } from "./useInventoryConfig";
import { FilterValuesExtractor } from "../FilterValuesExtractor";
import {
  clearSortState,
  loadSortState,
  saveSortState,
  SortDirection,
} from "./useSortPersistence";

/**
 * useTableHeaderFilter Composable
 * RESPONSIBILITY: Manage filter/sort in table header
 * - Extract unique values untuk filter dropdown
 * - Track active filters dan sort state
 * - Apply filter/sort langsung tanpa panel
 */
export function useTableHeaderFilter() {
  const config = useInventoryConfig();
  const activeFilterColumn = ref<string | null>(null);
  const columnValues = ref<Record<string, string[]>>({});
  const persistedSort = loadSortState();
  const columnSorts = ref<Record<string, SortDirection>>(
    persistedSort.column ? { [persistedSort.column]: persistedSort.direction } : {}
  );
  let lastInitializedColumns: string[] = [];

  /**
   * Initialize column filter values untuk dropdown
   * MUST pass data explicitly - no stale closure!
   */
  const initializeColumnValues = (columns: any[], data: any[]) => {
    const columnNames = columns
      .filter((col) => config.isColumnVisible(col.column_name))
      .map((col) => col.column_name);

    console.log("🔧 initializeColumnValues called:", {
      columnsCount: columnNames.length,
      dataCount: data?.length || 0,
      columnNames: columnNames.slice(0, 3),
    });

    // Validate required parameters
    if (!data || data.length === 0) {
      console.error("   ❌ ERROR: data is required and must not be empty!");
      return;
    }

    // Skip if columns haven't changed AND already initialized
    if (
      JSON.stringify(lastInitializedColumns) === JSON.stringify(columnNames) &&
      Object.keys(columnValues.value).length > 0
    ) {
      console.log("   ⏭️ Skipping - already initialized");
      return;
    }

    lastInitializedColumns = columnNames;

    console.log("   📊 Using data:", data.length, "items");

    if (columnNames.length > 0) {
      columnValues.value = FilterValuesExtractor.extractMultipleColumnValues(
        data,
        columnNames
      );

      console.log("   ✅ Values extracted:", {
        totalColumns: Object.keys(columnValues.value).length,
        sampleColumn: columnNames[0],
        sampleValues: columnValues.value[columnNames[0]]?.length || 0,
      });
    } else {
      console.warn("   ⚠️ Cannot initialize - no columns");
    }
  };

  /**
   * Apply filter untuk satu kolom
   */
  const applyFilter = (columnName: string, selectedValues: string[]) => {
    if (selectedValues.length === 0) {
      config.setColumnFilter(columnName, "");
    } else {
      config.setColumnFilter(columnName, JSON.stringify(selectedValues));
    }
    activeFilterColumn.value = null;
  };

  /**
   * Clear filter untuk satu kolom
   */
  const clearFilter = (columnName: string) => {
    config.setColumnFilter(columnName, "");
    activeFilterColumn.value = null;
  };

  /**
   * Apply sort untuk satu kolom
   * Clear all other sorts (single column sort only)
   */
  const applySort = (columnName: string, direction: SortDirection) => {
    // Clear all other column sorts (single-column sort only)
    Object.keys(columnSorts.value).forEach((col) => {
      if (col !== columnName) {
        columnSorts.value[col] = null;
      }
    });

    columnSorts.value[columnName] = direction;
    console.log(`🔄 Sort applied in headerFilter:`, { columnName, direction });

    if (direction === null) {
      clearSortState();
    } else {
      saveSortState({ column: columnName, direction });
    }
  };

  /**
   * Check apakah kolom punya active filter
   */
  const hasFilter = (columnName: string): boolean => {
    const filterValue = config.getColumnFilters()[columnName];
    return !!(filterValue && String(filterValue).trim());
  };

  /**
   * Get filter count untuk badge
   */
  const getFilterCount = (columnName: string): number => {
    const filterValue = config.getColumnFilters()[columnName];
    if (!filterValue) return 0;
    try {
      const parsed = JSON.parse(String(filterValue));
      return Array.isArray(parsed) ? parsed.length : 1;
    } catch {
      return 1;
    }
  };

  /**
   * Get current sort direction untuk kolom
   */
  const getSortDirection = (columnName: string): SortDirection => {
    return columnSorts.value[columnName] || null;
  };

  return {
    activeFilterColumn,
    columnValues,
    columnSorts,
    initializeColumnValues,
    applyFilter,
    clearFilter,
    applySort,
    hasFilter,
    getFilterCount,
    getSortDirection,
  };
}

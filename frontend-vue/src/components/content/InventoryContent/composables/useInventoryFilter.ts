import { useInventoryConfig } from "./useInventoryConfig";
import { useInventoryFilterEngine } from "./useInventoryFilterEngine";

/**
 * RESPONSIBILITY: High-level filter API for components
 * - Bridge between component state and config/filter engine
 * - Sync API & localStorage
 * - Apply filters to inventory data
 */

export function useInventoryFilter(state: any) {
  const config = useInventoryConfig();
  const filterEngine = useInventoryFilterEngine();

  /**
   * Load config from server with retry logic
   * Called on mount to restore saved preferences
   */
  async function loadFilterPreferences() {
    try {
      await config.loadFromServerWithRetry(5);
      syncStateFromConfig();
    } catch {
      syncStateFromConfig(); // Use localStorage as fallback
    }
  }

  /**
   * Sync component state from config singleton
   * Call when config changes or on mount
   */
  function syncStateFromConfig() {
    state.columnFilters = config.getColumnFilters();
    state.searchQuery = config.getSearchQuery();

    // CRITICAL: Initialize columnVisibility from config's visibleColumns
    // Do this BEFORE schemaColumns is loaded, so loadStats can filter correctly
    const visibleCols = config.getVisibleColumns();
    if (visibleCols && visibleCols.length > 0) {
      // Build columnVisibility object from visible columns list
      visibleCols.forEach((col: string) => {
        state.columnVisibility[col] = true;
      });
    }

    // Also sync visibility if schemaColumns already available
    if (state.schemaColumns && state.schemaColumns.length > 0) {
      state.schemaColumns.forEach((col: any) => {
        state.columnVisibility[col.column_name] = config.isColumnVisible(
          col.column_name
        );
      });
    }

    // CRITICAL: Also validate and sync locked columns
    // Locked columns must be valid subset of currently visible columns
    const lockedCols = config.getLockedColumns();

    // Filter locked columns to only include those that are also visible
    const validLockedCols = lockedCols.filter((col) =>
      visibleCols.includes(col)
    );

    // If there are invalid locked columns, fix them
    if (validLockedCols.length !== lockedCols.length) {
      // Update config to only have valid locked columns
      config.setLockedColumns(validLockedCols);
    }
  }

  /**
   * Apply search + column filters to inventory list
   * Called when rendering table
   */
  function getFilteredInventoryList() {
    return filterEngine.applyAll(
      state.inventoryList || [],
      state.searchQuery || "",
      state.columnFilters || {}
    );
  }

  function clearAllFilters() {
    console.log("🧹 clearAllFilters called");
    
    config.clearAllFilters();
    state.columnFilters = {};
    state.searchQuery = "";
    state.showFilterPanel = false;
    
    // CRITICAL: Also clear sort state
    state.sortColumn = null;
    state.sortDirection = null;
    
    // CRITICAL: Reset pagination to first page
    state.currentOffset = 0;
    
    console.log("✅ All filters, sort, and pagination cleared");
  }

  function showAllColumns() {
    const allColumnNames = state.schemaColumns.map(
      (col: any) => col.column_name
    );
    config.showAllColumns(allColumnNames);
    syncStateFromConfig();
  }

  function getVisibleColumnsCount() {
    return (
      state.schemaColumns.filter((col: any) =>
        config.isColumnVisible(col.column_name)
      ).length || 1
    );
  }

  function hasActiveFilters() {
    return config.hasActiveFilters();
  }

  function isColumnVisible(columnName: string): boolean {
    return config.isColumnVisible(columnName);
  }

  return {
    loadFilterPreferences,
    syncStateFromConfig,
    getFilteredInventoryList,
    clearAllFilters,
    showAllColumns,
    getVisibleColumnsCount,
    hasActiveFilters,
    isColumnVisible,
  };
}

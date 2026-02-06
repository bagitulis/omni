import { ref, reactive } from "vue";
import { useFilterPersistence } from "./useFilterPersistence";

/**
 * RESPONSIBILITY: Unified Column Configuration State Management
 * - Manage visibility, locked, filters, search state
 * - Coordinate with persistence layer
 * - Single source of truth for column config
 */

let instance: ReturnType<typeof createConfigManager> | null = null;

function createConfigManager() {
  const persistence = useFilterPersistence();
  const { STORAGE_KEY, loadFromStorage, saveToStorage } = persistence;

  // Initialize from localStorage with defaults
  const visibleColumns = ref<Set<string>>(
    loadFromStorage(STORAGE_KEY.VISIBLE, new Set()) as Set<string>
  );
  const lockedColumns = ref<Set<string>>(
    loadFromStorage(STORAGE_KEY.LOCKED, new Set()) as Set<string>
  );
  const columnFilters = reactive<Record<string, string>>(
    loadFromStorage(STORAGE_KEY.FILTERS, {})
  );
  const searchQuery = ref<string>(loadFromStorage(STORAGE_KEY.SEARCH, ""));
  const columnOrder = ref<string[]>(
    loadFromStorage(STORAGE_KEY.ORDER, []) as string[]
  );

  // ==================== VISIBILITY ====================

  function setColumnVisible(column: string, visible: boolean) {
    if (visible) {
      visibleColumns.value.add(column);
    } else {
      visibleColumns.value.delete(column);
    }
    saveToStorage(STORAGE_KEY.VISIBLE, visibleColumns.value);
    queueSync();
  }

  function setVisibleColumns(columns: string[]) {
    visibleColumns.value = new Set(columns);
    saveToStorage(STORAGE_KEY.VISIBLE, visibleColumns.value);
    queueSync();
  }

  function isColumnVisible(column: string): boolean {
    return visibleColumns.value.has(column);
  }

  function getVisibleColumns(): string[] {
    return Array.from(visibleColumns.value);
  }

  // ==================== LOCKED COLUMNS ====================

  function setColumnLocked(column: string, locked: boolean) {
    if (locked) {
      lockedColumns.value.add(column);
    } else {
      lockedColumns.value.delete(column);
    }
    saveToStorage(STORAGE_KEY.LOCKED, lockedColumns.value);
    queueSync();
  }

  function setLockedColumns(columns: string[]) {
    lockedColumns.value = new Set(columns);
    saveToStorage(STORAGE_KEY.LOCKED, lockedColumns.value);
    queueSync();
  }

  function isColumnLocked(column: string): boolean {
    return lockedColumns.value.has(column);
  }

  function getLockedColumns(): string[] {
    return Array.from(lockedColumns.value);
  }

  // ==================== FILTERS ====================

  function setColumnFilter(column: string, value: string) {
    columnFilters[column] = value;
    saveToStorage(STORAGE_KEY.FILTERS, columnFilters);
    queueSync();
  }

  function clearColumnFilter(column: string) {
    delete columnFilters[column];
    saveToStorage(STORAGE_KEY.FILTERS, columnFilters);
    queueSync();
  }

  function getColumnFilter(column: string): string {
    return columnFilters[column] || "";
  }

  function getColumnFilters(): Record<string, string> {
    return { ...columnFilters };
  }

  // ==================== SEARCH ====================

  function setSearchQuery(query: string) {
    searchQuery.value = query;
    saveToStorage(STORAGE_KEY.SEARCH, searchQuery.value);
    queueSync();
  }

  function getSearchQuery(): string {
    return searchQuery.value;
  }

  // ==================== COLUMN ORDER ====================

  function setColumnOrder(order: string[]) {
    columnOrder.value = order;
    saveToStorage(STORAGE_KEY.ORDER, order);
    queueSync();
  }

  function getColumnOrder(): string[] {
    return columnOrder.value;
  }

  // ==================== UTILITIES ====================

  function hasActiveFilters(): boolean {
    return (
      searchQuery.value.trim() !== "" ||
      Object.values(columnFilters).some((v) => v.trim() !== "")
    );
  }

  function clearAllFilters() {
    console.log("🧹 config.clearAllFilters called");
    console.log("   Before - searchQuery:", searchQuery.value);
    console.log(
      "   Before - columnFilters:",
      Object.keys(columnFilters).length
    );

    searchQuery.value = "";
    Object.keys(columnFilters).forEach((key) => delete columnFilters[key]);
    saveToStorage(STORAGE_KEY.SEARCH, "");
    saveToStorage(STORAGE_KEY.FILTERS, {});
    queueSync();

    console.log("   After - searchQuery:", searchQuery.value);
    console.log("   After - columnFilters:", Object.keys(columnFilters).length);
    console.log("✅ Config filters cleared");
  }

  function showAllColumns(allColumns: string[]) {
    visibleColumns.value = new Set(allColumns);
    saveToStorage(STORAGE_KEY.VISIBLE, visibleColumns.value);
    queueSync();
  }

  // ==================== TOGGLE METHODS (for backward compat) ====================

  function toggleColumnVisibility(column: string) {
    setColumnVisible(column, !isColumnVisible(column));
  }

  function toggleColumnLock(column: string) {
    const isLocked = lockedColumns.value.has(column);

    setColumnLocked(column, !isLocked);
  }

  // ==================== SYNC ====================

  async function performSync() {
    const lockedArray = Array.from(lockedColumns.value);
    const visibleArray = Array.from(visibleColumns.value);

    const payload = {
      columnFilters: { ...columnFilters },
      visibleColumns: visibleArray,
      lockedColumns: lockedArray,
      searchQuery: searchQuery.value,
      columnOrder: columnOrder.value,
    };

    return persistence.syncToAPI(payload);
  }

  function queueSync() {
    // DEBUG: Remove debounce - sync immediately for testing
    console.log("⏱️ Removing debounce - syncing immediately");
    performSync();
    // if (syncTimeout !== null) {
    //   clearTimeout(syncTimeout);
    // }
    // // Debounce sync to 300ms to batch rapid changes
    // syncTimeout = window.setTimeout(() => {
    //   performSync();
    //   syncTimeout = null;
    // }, 300);
  }

  // Legacy backward compat wrapper - call performSync
  async function syncToAPI(payload?: any, retries?: number): Promise<boolean> {
    if (payload) {
      // If payload provided, use it directly
      return persistence.syncToAPI(payload, retries || 3);
    }
    // No payload = use current state
    return performSync();
  }

  // Legacy method for backward compat
  async function loadFromAPI() {
    return persistence.loadFromAPI();
  }

  async function loadFromServerWithRetry(maxRetries = 5): Promise<boolean> {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const data = (await Promise.race([
          loadFromAPI(),
          new Promise(
            (_, reject) =>
              setTimeout(() => reject(new Error("Load timeout")), 6000) // 6 second timeout
          ),
        ])) as any;

        if (data) {
          // Apply loaded data - ensure all properties are restored
          const loadedVisible = new Set<string>(data.visibleColumns || []);
          const loadedLocked = new Set<string>(data.lockedColumns || []);

          visibleColumns.value = loadedVisible;
          lockedColumns.value = loadedLocked;
          // Only override columnOrder if API has data, otherwise keep localStorage
          if (data.columnOrder && data.columnOrder.length > 0) {
            columnOrder.value = data.columnOrder;
            saveToStorage(STORAGE_KEY.ORDER, columnOrder.value);
          }
          Object.assign(columnFilters, data.columnFilters || {});
          searchQuery.value = data.searchQuery || "";

          // Persist to localStorage - ensure they stick
          saveToStorage(STORAGE_KEY.VISIBLE, loadedVisible);
          saveToStorage(STORAGE_KEY.LOCKED, loadedLocked);
          saveToStorage(STORAGE_KEY.FILTERS, columnFilters);
          saveToStorage(STORAGE_KEY.SEARCH, searchQuery.value);

          return true;
        }

        if (attempt < maxRetries) {
          const delayMs = 500 * attempt;
          await new Promise((r) => setTimeout(r, delayMs));
        }
      } catch {
        if (attempt < maxRetries) {
          const delayMs = 500 * attempt;
          await new Promise((r) => setTimeout(r, delayMs));
        }
      }
    }

    // 🔒 CRITICAL: Restore from localStorage as fallback
    const storedVisible = loadFromStorage(STORAGE_KEY.VISIBLE, new Set());
    const storedLocked = loadFromStorage(STORAGE_KEY.LOCKED, new Set());

    visibleColumns.value = storedVisible;
    lockedColumns.value = storedLocked;

    return false;
  }

  return {
    // State
    visibleColumns,
    lockedColumns,
    columnFilters,
    searchQuery,
    columnOrder,

    // Visibility methods
    setColumnVisible,
    toggleColumnVisibility,
    setVisibleColumns,
    isColumnVisible,
    getVisibleColumns,

    // Lock methods
    setColumnLocked,
    toggleColumnLock,
    setLockedColumns,
    isColumnLocked,
    getLockedColumns,

    // Filter methods
    setColumnFilter,
    clearColumnFilter,
    getColumnFilter,
    getColumnFilters,

    // Search methods
    setSearchQuery,
    getSearchQuery,

    // Column order methods
    setColumnOrder,
    getColumnOrder,

    // Utility methods
    hasActiveFilters,
    clearAllFilters,
    showAllColumns,

    // Sync methods (legacy + new)
    performSync,
    syncToAPI,
    loadFromAPI,
    loadFromServerWithRetry,
  };
}

export function useInventoryConfig() {
  if (!instance) {
    instance = createConfigManager();
  }
  return instance;
}

export function resetInventoryConfig() {
  instance = null;
}

import { ref, computed } from "vue";
import { useColumnConfiguration } from "./useColumnConfiguration";
import { useInventoryConfig } from "./useInventoryConfig";

/**
 * useColumnManager Composable
 * RESPONSIBILITY: Unified column visibility & lock management
 * - Load columns from config and Google Sheets
 * - Toggle visibility and lock states via useInventoryConfig
 * - Direct sync to global state
 */

export function useColumnManager() {
  const { loadAvailableColumns, loadSelectedColumns, getLastColumnLetter } =
    useColumnConfiguration();
  const {
    visibleColumns,
    isColumnLocked,
    toggleColumnLock,
    setLockedColumns,
    setVisibleColumns,
    getVisibleColumns,
    getLockedColumns,
  } = useInventoryConfig();

  const loading = ref(false);
  const availableColumns = ref<any[]>([]);
  const errorMessage = ref("");

  const loadData = async () => {
    loading.value = true;
    try {
      availableColumns.value = await loadAvailableColumns();
      // NOTE: Do NOT load /columns/selected here
      // The visible columns come from /filter-preferences which is loaded earlier
      // in the mount sequence via useInventoryFilter.loadFilterPreferences()
      // Calling setVisibleColumns here would override the user's saved preferences

      // Only set visible columns if config is still empty
      const currentVisible = getVisibleColumns();
      if (currentVisible.length === 0) {
        const selectedColumns = await loadSelectedColumns();
        setVisibleColumns(selectedColumns);
      }
    } catch (error) {
      console.error("❌ Error loading columns:", error);
      errorMessage.value = "Failed to load columns. Check your connection.";
    } finally {
      loading.value = false;
    }
  };

  const isSelected = (columnName: string) =>
    visibleColumns.value.has(columnName);

  const toggleColumn = (columnName: string) => {
    const isCurrentlySelected = isSelected(columnName);
    const newVisible = isCurrentlySelected
      ? Array.from(visibleColumns.value).filter((col) => col !== columnName)
      : [...Array.from(visibleColumns.value), columnName];
    setVisibleColumns(newVisible);
  };

  const selectAll = () => {
    const allNames = availableColumns.value.map((col: any) => col.name);
    setVisibleColumns(allNames);
  };

  const clearAll = () => {
    setVisibleColumns([]);
  };

  const toggleLock = (columnName: string) => {
    toggleColumnLock(columnName);
  };

  const lockAll = () => {
    const visibleCols = getVisibleColumns();
    console.log(`🔒 Locking ${visibleCols.length} columns`);
    setLockedColumns(visibleCols);
  };

  const unlockAll = () => {
    console.log("🔓 Unlocking all columns");
    setLockedColumns([]);
  };

  const visibleColumnCount = computed(() => getVisibleColumns().length);
  const lockedCount = computed(() => getLockedColumns().length);

  return {
    loading,
    availableColumns,
    errorMessage,
    visibleColumnCount,
    lockedCount,
    loadData,
    isSelected,
    toggleColumn,
    selectAll,
    clearAll,
    toggleLock,
    lockAll,
    unlockAll,
    isColumnLocked,
    getLastColumnLetter,
  };
}

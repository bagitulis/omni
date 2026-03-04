import { ref } from "vue";
import { useColumnConfiguration } from "./useColumnConfiguration";

/**
 * useInventoryColumnManager Composable
 * RESPONSIBILITY: Manage inventory column selection state
 */
export function useInventoryColumnManager() {
  const {
    loadAvailableColumns,
    loadSelectedColumns,
    saveColumnConfiguration,
    getLastColumnLetter,
  } = useColumnConfiguration();

  const loadingColumns = ref(false);
  const savingColumns = ref(false);
  const availableColumns = ref<any[]>([]);
  const selectedColumns = ref<string[]>([]);
  const errorMessage = ref("");

  async function loadData() {
    loadingColumns.value = true;
    errorMessage.value = "";
    try {
      availableColumns.value = await loadAvailableColumns();
      selectedColumns.value = await loadSelectedColumns();
    } catch (error: any) {
      errorMessage.value = error.message || "Failed to load columns";
    } finally {
      loadingColumns.value = false;
    }
  }

  function isSelected(columnName: string): boolean {
    return selectedColumns.value.includes(columnName);
  }

  async function toggleColumn(columnName: string) {
    const idx = selectedColumns.value.indexOf(columnName);
    if (idx === -1) {
      selectedColumns.value.push(columnName);
    } else {
      selectedColumns.value.splice(idx, 1);
    }
    await saveColumns();
  }

  async function selectAll() {
    selectedColumns.value = availableColumns.value.map((col: any) => col.name);
    await saveColumns();
  }

  async function clearAll() {
    selectedColumns.value = [];
    await saveColumns();
  }

  async function saveColumns() {
    savingColumns.value = true;
    try {
      await saveColumnConfiguration(selectedColumns.value);
    } catch (error: any) {
      errorMessage.value = error.message || "Failed to save columns";
    } finally {
      savingColumns.value = false;
    }
  }

  return {
    loadingColumns,
    savingColumns,
    availableColumns,
    selectedColumns,
    errorMessage,
    loadData,
    isSelected,
    toggleColumn,
    selectAll,
    clearAll,
    getLastColumnLetter,
  };
}

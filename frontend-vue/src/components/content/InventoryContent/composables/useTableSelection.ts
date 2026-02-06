import { ref, computed, watch } from "vue";

/**
 * Row selection logic for inventory table
 * Handles selecting/deselecting individual rows and all rows
 */
export function useTableSelection(getTotalRows: () => number) {
  const selectedRows = ref<Set<number>>(new Set());
  const totalRows = computed(() => getTotalRows());

  const isAllSelected = computed(() => {
    return selectedRows.value.size === totalRows.value && totalRows.value > 0;
  });

  const isRowSelected = (rowIndex: number) => {
    return selectedRows.value.has(rowIndex);
  };

  const toggleRowSelection = (rowIndex: number) => {
    if (rowIndex < 0 || rowIndex >= totalRows.value) return;

    if (selectedRows.value.has(rowIndex)) {
      selectedRows.value.delete(rowIndex);
    } else {
      selectedRows.value.add(rowIndex);
    }
  };

  const toggleSelectAll = () => {
    if (isAllSelected.value) {
      selectedRows.value.clear();
    } else {
      selectedRows.value.clear();
      for (let i = 0; i < totalRows.value; i++) {
        selectedRows.value.add(i);
      }
    }
  };

  // Trim selections when data length shrinks (e.g., after filter)
  watch(
    totalRows,
    (newTotal) => {
      selectedRows.value = new Set(
        Array.from(selectedRows.value).filter((idx) => idx < newTotal)
      );
    },
    { immediate: true }
  );

  return {
    selectedRows,
    isAllSelected,
    isRowSelected,
    toggleRowSelection,
    toggleSelectAll,
  };
}

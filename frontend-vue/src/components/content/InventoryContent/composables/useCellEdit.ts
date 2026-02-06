import { ref } from "vue";

/**
 * Cell editing state management
 * Handles starting, saving, and canceling cell edits
 */
export function useCellEdit() {
  const editingCell = ref<{ itemIndex: number; column: string } | null>(null);
  const editValue = ref("");

  const isEditingCell = (itemIndex: number, column: string) => {
    return (
      editingCell.value?.itemIndex === itemIndex &&
      editingCell.value?.column === column
    );
  };

  const startEdit = (
    itemIndex: number,
    column: string,
    initialValue: string
  ) => {
    editingCell.value = { itemIndex, column };
    editValue.value = initialValue;
  };

  const saveEdit = (
    itemIndex: number,
    column: string,
    onSave: (data: { itemIndex: number; column: string; value: string }) => void
  ) => {
    if (editingCell.value) {
      onSave({
        itemIndex,
        column,
        value: editValue.value,
      });
      editingCell.value = null;
      editValue.value = "";
    }
  };

  const cancelEdit = () => {
    editingCell.value = null;
    editValue.value = "";
  };

  return {
    editingCell,
    editValue,
    isEditingCell,
    startEdit,
    saveEdit,
    cancelEdit,
  };
}

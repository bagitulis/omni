import {
  getAuthHeaders,
  getApiBaseUrl,
} from "../../../../utils/apiHeaders";

export function useInventoryForm(state: any, alertHandlers: any, reload: any) {
  const API_BASE_URL = getApiBaseUrl("/inventory");

  function editItem(item: any) {
    state.editingItem = item;
    state.formData = { ...item };
    state.showAddForm = true;
  }

  function closeAddForm() {
    state.showAddForm = false;
    state.editingItem = null;
    state.formData = {};
  }

  async function saveItem() {
    try {
      const url = state.editingItem
        ? `${API_BASE_URL}/${state.formData[state.keyColumn]}`
        : `${API_BASE_URL}`;

      const method = state.editingItem ? "PUT" : "POST";

      const response = await fetch(url, {
        method,
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify(state.formData),
      });

      const result = await response.json();

      if (result.status === "SUCCESS") {
        alertHandlers.success("Success", result.message);
        closeAddForm();
        await reload.inventoryData();
      } else {
        alertHandlers.error("Error", result.message);
      }
    } catch (error: any) {
      console.error("Save error:", error);
      alertHandlers.error("Save Error", error.message);
    }
  }

  async function deleteItem(item: any, keyColumn: string) {
    if (confirm(`Delete "${item[keyColumn]}"?`)) {
      try {
        const response = await fetch(`${API_BASE_URL}/${item[keyColumn]}`, {
          method: "DELETE",
          headers: getAuthHeaders(),
        });

        const result = await response.json();

        if (result.status === "SUCCESS") {
          alertHandlers.success("Success", result.message);
          await reload.inventoryData();
        } else {
          alertHandlers.error("Error", result.message);
        }
      } catch (error: any) {
        console.error("Delete error:", error);
        alertHandlers.error("Delete Error", error.message);
      }
    }
  }

  async function saveEditCell(rowIdx: number, columnName: string, item: any) {
    const key = `${rowIdx}-${columnName}`;
    const newValue = state.editingCell[key];

    if (!newValue || newValue === item[columnName]) {
      cancelEditCell(rowIdx, columnName);
      return;
    }

    try {
      const updateData = { ...item, [columnName]: newValue };
      const response = await fetch(`${API_BASE_URL}/${item[state.keyColumn]}`, {
        method: "PUT",
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify(updateData),
      });

      const result = await response.json();

      if (result.status === "SUCCESS") {
        // Update item in list
        item[columnName] = newValue;
        // Clear editing state dengan proper Vue reactivity
        const newEditingCell = { ...state.editingCell };
        delete newEditingCell[key];
        state.editingCell = newEditingCell;
        console.log(`✅ Cell saved: ${key} = "${newValue}"`);
      } else {
        alertHandlers.error("Error", result.message);
      }
    } catch (error: any) {
      console.error("❌ Edit error:", error);
      alertHandlers.error("Edit Error", error.message);
    }
  }

  function startEditCell(
    rowIdx: number,
    columnName: string,
    currentValue: any
  ) {
    const key = `${rowIdx}-${columnName}`;
    // Initialize dengan currentValue jika ada, otherwise empty string
    const initialValue =
      currentValue !== null && currentValue !== undefined
        ? String(currentValue)
        : "";
    state.editingCell[key] = initialValue;
    // Force Vue to detect change
    state.editingCell = { ...state.editingCell };
    console.log(`✏️ Edit started: ${key} = "${initialValue}"`);
  }

  function cancelEditCell(rowIdx: number, columnName: string) {
    const key = `${rowIdx}-${columnName}`;
    const newEditingCell = { ...state.editingCell };
    delete newEditingCell[key];
    state.editingCell = newEditingCell;
  }

  return {
    editItem,
    saveItem,
    deleteItem,
    closeAddForm,
    saveEditCell,
    startEditCell,
    cancelEditCell,
  };
}

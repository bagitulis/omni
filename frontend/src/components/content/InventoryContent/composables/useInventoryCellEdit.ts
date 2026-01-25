import { reactive } from "vue";
import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";
import { parseItemData } from "./useInventoryItemHelpers";

interface EditingCell {
  rowIndex: number | null;
  column: string | null;
  value: any;
}

export function useInventoryCellEdit(
  state: any,
  onAlertError: (title: string, msg: string) => void,
) {
  const editingCell = reactive<EditingCell>({
    rowIndex: null,
    column: null,
    value: null,
  });

  const handleCellEditSave = (eventData: any) => {
    // Handle safe destructuring - check if parameters exist
    if (!eventData || typeof eventData !== "object") {
      onAlertError("Error", "Invalid cell edit data");
      return;
    }

    const { itemIndex, column, value, item } = eventData;

    if (
      itemIndex === undefined ||
      column === undefined ||
      value === undefined
    ) {
      onAlertError("Error", "Missing required cell edit parameters");
      return;
    }

    // Use passed item if available, otherwise fall back to filtered list lookup
    const itemToSave = item || state.filteredInventoryList[itemIndex];
    if (!itemToSave) {
      onAlertError("Error", "Item not found for editing");
      return;
    }
    saveCellDirectly(itemIndex, column, value, itemToSave);
  };

  const saveCellDirectly = async (
    _rowIdx: number,
    columnName: string,
    newValue: any,
    item: any,
  ) => {
    try {
      // Handle nested data field - supports both string (Node.js) and object (Go) formats
      let itemData = item;
      if (item.data !== undefined && item.data !== null) {
        if (typeof item.data === "string") {
          try {
            itemData = JSON.parse(item.data);
          } catch {
            itemData = item;
          }
        } else if (typeof item.data === "object") {
          itemData = item.data;
        }
      }

      const keyValue =
        itemData[state.configData.key_column] || itemData.SKU || item.keyValue;

      if (!keyValue) {
        onAlertError("Error", "Key value not found");
        return;
      }

      const updateData = { ...itemData, [columnName]: newValue };

      // Optimistic update - update state immediately
      const listIndex = state.inventoryList.findIndex((i: any) => {
        const iData = parseItemData(i);
        return (iData[state.configData.key_column] || iData.SKU) === keyValue;
      });

      if (listIndex !== -1) {
        // Update with new data
        if (state.inventoryList[listIndex].data) {
          state.inventoryList[listIndex].data = JSON.stringify(updateData);
        } else {
          state.inventoryList[listIndex] = {
            ...state.inventoryList[listIndex],
            [columnName]: newValue,
          };
        }
      }

      // Persist to backend
      const response = await fetch(
        `${getApiBaseUrl("/inventory")}/${keyValue}`,
        {
          method: "PUT",
          headers: {
            ...getAuthHeaders(),
            "Content-Type": "application/json",
          },
          body: JSON.stringify(updateData),
        },
      );

      if (!response.ok) {
        throw new Error(`Failed to update: ${response.status}`);
      }

      const result = await response.json();
      // Update state with server response if available
      if (result.data && listIndex !== -1) {
        state.inventoryList[listIndex] = result.data;
      }
    } catch (error) {
      onAlertError("Save Failed", String(error));
      // Note: optimistic update already done, error message shown, but state keeps the change
      // This is acceptable for better UX
    }
  };

  return {
    editingCell,
    handleCellEditSave,
    saveCellDirectly,
  };
}

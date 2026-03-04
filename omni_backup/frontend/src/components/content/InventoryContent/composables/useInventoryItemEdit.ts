import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

export function useInventoryItemEdit(state: any, onAlert: any) {
  const handleEditItem = (item: any) => {
    state.editingItem = item;
    state.formData = item;
    state.showAddForm = true;
  };

  const handleDeleteItem = async (item: any) => {
    if (!confirm("Confirm delete this item?")) {
      return;
    }

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
        onAlert.error("Error", "Key value not found");
        return;
      }

      const response = await fetch(
        `${getApiBaseUrl("/inventory")}/${keyValue}`,
        {
          method: "DELETE",
          headers: getAuthHeaders(),
        },
      );

      if (!response.ok) {
        throw new Error(`Failed to delete: ${response.status}`);
      }

      const index = state.inventoryList.findIndex((i: any) => {
        const iData = i.data ? JSON.parse(i.data) : i;
        return (iData[state.configData.key_column] || iData.SKU) === keyValue;
      });

      if (index !== -1) {
        state.inventoryList.splice(index, 1);
        onAlert.success("✅ Deleted", "Item successfully deleted");
      }
    } catch (error) {
      onAlert.error("Delete Error", String(error));
    }
  };

  return {
    handleEditItem,
    handleDeleteItem,
  };
}

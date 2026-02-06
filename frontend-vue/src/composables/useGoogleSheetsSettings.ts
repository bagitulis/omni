import { ref } from "vue";
import { useApi } from "./useApi";
import googleSheetsStorageService from "../services/googleSheetsStorageService";
import type { ValidationResult } from "./useGoogleSheetsLinks";

export function useGoogleSheetsSettings() {
  const api = useApi();

  const inventorySheet = ref("");
  const statusMsg = ref("");
  const statusType = ref<"success" | "error">("success");
  const saving = ref(false);

  const metadata = ref({
    inventory: null as ValidationResult | null,
    wallet: null as ValidationResult | null,
    shipping: null as ValidationResult | null,
    order: null as ValidationResult | null,
  });

  const loadSettings = async () => {
    try {
      const response = await api.get("/google/settings/saved-links");
      const linksData = response.data || response;

      return linksData;
    } catch (error: any) {
      throw error;
    }
  };

  const loadDetailedSettings = async () => {
    try {
      const response = await api.get("/google/settings/detailed");
      const detailedData = response.data || response;
      return detailedData;
    } catch (error: any) {
      throw error;
    }
  };

  const processDetailedSettings = (detailedData: any) => {
    const types = ["inventory", "wallet", "shipping", "order"] as const;
    types.forEach((type) => {
      const worksheetKey = `${type}_available_worksheets`;
      const sheetNameKey = `${type}_sheet_name`;
      const spreadsheetIdKey = `${type}_spreadsheet_id`;

      if (detailedData[worksheetKey] && detailedData[worksheetKey].length > 0) {
        metadata.value[type] = {
          spreadsheetId: detailedData[spreadsheetIdKey] || "",
          name: detailedData[spreadsheetIdKey] || "",
          sheets: detailedData[worksheetKey],
          type,
        };
        googleSheetsStorageService.saveSheetsCacheLocal(
          type,
          detailedData[worksheetKey],
          detailedData[spreadsheetIdKey]
        );
        if (type === "inventory") {
          inventorySheet.value = detailedData[sheetNameKey] || "";
        }
      }
    });
  };

  const loadFromCacheAsFallback = () => {
    const cachedSheets = googleSheetsStorageService.getSheetsCacheLocal();

    Object.entries(cachedSheets).forEach(([type, cache]) => {
      const typeKey = type as keyof typeof metadata.value;
      if (!metadata.value[typeKey] && cache?.data?.length > 0) {
        metadata.value[typeKey] = {
          spreadsheetId: cache.spreadsheetId,
          name: cache.spreadsheetId,
          sheets: cache.data,
          type,
        };
      }
    });
  };

  const updateInventorySheet = async (
    sheetName: string,
    inventoryId: string,
    sheets: any[]
  ) => {
    if (!sheetName) return;
    inventorySheet.value = sheetName;

    const selections = googleSheetsStorageService.getSelectedSheetsLocal();
    selections.inventory = sheetName;
    googleSheetsStorageService.saveSelectedSheetsLocal(selections);

    try {
      await api.post("/google/settings/update-detailed", {
        inventory_spreadsheet_id: inventoryId,
        inventory_sheet_name: sheetName,
        inventory_available_worksheets: sheets,
      });
      statusMsg.value = "✅ Inventory sheet configured successfully!";
      statusType.value = "success";
      setTimeout(() => {
        statusMsg.value = "";
      }, 3000);
    } catch (error: any) {
      statusMsg.value = `Failed: ${error.message}`;
      statusType.value = "error";
    }
  };

  return {
    inventorySheet,
    statusMsg,
    statusType,
    saving,
    metadata,
    loadSettings,
    loadDetailedSettings,
    processDetailedSettings,
    loadFromCacheAsFallback,
    updateInventorySheet,
  };
}

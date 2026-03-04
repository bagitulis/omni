import { ref, computed } from "vue";
import { useApi } from "./useApi";
import googleSheetsStorageService from "../services/googleSheetsStorageService";

export interface SpreadsheetLinks {
  inventory: string;
  wallet: string;
  shipping: string;
  order: string;
}

export interface SheetMetadata {
  name: string;
  sheetId: number;
  index: number;
  columnCount: number;
  rowCount: number;
}

export interface ValidationResult {
  spreadsheetId: string;
  name: string;
  sheets: SheetMetadata[];
  type: string;
}

export function useGoogleSheetsLinks() {
  const api = useApi();

  // Link state
  const links = ref<SpreadsheetLinks>({
    inventory: "",
    wallet: "",
    shipping: "",
    order: "",
  });
  const initialLinks = ref<SpreadsheetLinks>({...links.value});

  // Validation state
  const validating = ref({
    inventory: false,
    wallet: false,
    shipping: false,
    order: false,
  });

  const validated = ref({
    inventory: false,
    wallet: false,
    shipping: false,
    order: false,
  });

  const errors = ref({
    inventory: "",
    wallet: "",
    shipping: "",
    order: "",
  });

  // Metadata state
  const metadata = ref({
    inventory: null as ValidationResult | null,
    wallet: null as ValidationResult | null,
    shipping: null as ValidationResult | null,
    order: null as ValidationResult | null,
  });

  // Lock state
  const locked = ref({
    inventory: false,
    wallet: false,
    shipping: false,
    order: false,
  });

  const hasChanges = computed(() =>
    Object.keys(links.value).some(
      (key) => links.value[key as keyof SpreadsheetLinks] !== initialLinks.value[key as keyof SpreadsheetLinks]
    )
  );

  const checkLink = async (type: keyof SpreadsheetLinks) => {
    const linkValue = links.value[type];
    if (!linkValue) return;

    validating.value[type] = true;
    errors.value[type] = "";
    validated.value[type] = false;

    try {
      const response = await api.post("/google/settings/validate-link", {
        spreadsheetUrl: linkValue,
        type,
      });
      if (response.success) {
        metadata.value[type] = response.data;
        validated.value[type] = true;
        googleSheetsStorageService.saveSheetsCacheLocal(
          type,
          response.data.sheets || [],
          response.data.spreadsheetId
        );
        await saveMetadata(type, response.data);
      } else {
        errors.value[type] = response.error || "Validation failed";
      }
    } catch (err: any) {
      errors.value[type] = err.message || "Failed to validate";
    } finally {
      validating.value[type] = false;
    }
  };

  const saveMetadata = async (type: keyof SpreadsheetLinks, data: ValidationResult) => {
    try {
      const payload = {
        inventory_spreadsheet_id: metadata.value.inventory?.spreadsheetId || links.value.inventory?.split('/d/')[1] || "",
        wallet_spreadsheet_id: metadata.value.wallet?.spreadsheetId || links.value.wallet?.split('/d/')[1] || "",
        shipping_spreadsheet_id: metadata.value.shipping?.spreadsheetId || links.value.shipping?.split('/d/')[1] || "",
        order_spreadsheet_id: metadata.value.order?.spreadsheetId || links.value.order?.split('/d/')[1] || "",
        [`${type}_available_worksheets`]: data.sheets || [],
      };
      await api.post("/google/settings/update-detailed", payload);
      // Don't auto-lock here, let user decide when to lock
    } catch (error) {
      console.warn("Failed to save metadata:", error);
    }
  };

  const saveLinks = async () => {
    if (!hasChanges.value) return;
    await api.post("/google/settings/save-links", {
      inventory: links.value.inventory || null,
      wallet: links.value.wallet || null,
      shipping: links.value.shipping || null,
      order: links.value.order || null,
    });
    googleSheetsStorageService.saveLinksLocal(links.value);
    Object.keys(links.value).forEach((key) => {
      if (links.value[key as keyof SpreadsheetLinks]) {
        locked.value[key as keyof SpreadsheetLinks] = true;
      }
    });
    initialLinks.value = { ...links.value };
  };

  const unlock = (type: keyof SpreadsheetLinks) => {
    validated.value[type] = false;
    metadata.value[type] = null;
    errors.value[type] = "";
    locked.value[type] = false;
  };

  return {
    links,
    initialLinks,
    validating,
    validated,
    errors,
    metadata,
    locked,
    hasChanges,
    checkLink,
    saveLinks,
    unlock,
  };
}

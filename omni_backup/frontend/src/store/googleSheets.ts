import { defineStore } from "pinia";
import { ref, computed, Ref } from "vue";
import apiService from "../services/api";

export const useGoogleSheetsStore = defineStore("googleSheets", () => {
  // State
  const authStatus: Ref<any> = ref({
    authenticated: false,
    settings: {},
    hasCredentials: false,
  });

  const spreadsheets: Ref<any[]> = ref([]);
  const worksheets: Ref<any[]> = ref([]);
  const loading: Ref<boolean> = ref(false);
  const editMode: Ref<boolean> = ref(false);
  const tempSettings: Ref<Record<string, any>> = ref({});

  // Getters
  const isAuthenticated = computed(() => authStatus.value.authenticated);
  const currentSettings = computed(() => authStatus.value.settings);
  const inventorySpreadsheetId = computed(
    () => authStatus.value.settings?.inventory_spreadsheet_id || ""
  );
  const inventorySheetName = computed(
    () => authStatus.value.settings?.inventory_sheet_name || ""
  );

  // Actions
  async function loadAuthStatus(): Promise<any> {
    try {
      loading.value = true;
      // Add cache busting for auth status
      const cacheBuster = `_t=${Date.now()}`;
      const response = await apiService.get(
        `/google/auth/status?${cacheBuster}`
      );
      if (response.success) {
        authStatus.value = response.data;
        // Auth status loaded - suppress log

        if (response.data.authenticated) {
          await loadSpreadsheets();
        }

        return response;
      }
      return response;
    } catch (error) {
      console.error("Failed to load auth status:", error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  function sortSpreadsheetsByLastUsed(): void {
    const settings = authStatus.value.settings;

    if (!settings) return;

    const currentWalletId = settings.wallet_spreadsheet_id;
    const currentShippingId = settings.shipping_spreadsheet_id;

    spreadsheets.value.sort((a, b) => {
      if (a.id === currentWalletId) return -1;
      if (b.id === currentWalletId) return 1;

      if (a.id === currentShippingId) return -1;
      if (b.id === currentShippingId) return 1;

      return a.name.localeCompare(b.name);
    });
  }

  async function refreshSpreadsheets(): Promise<any> {
    try {
      loading.value = true;

      // Add cache busting for refresh
      const cacheBuster = `_t=${Date.now()}`;
      const response = await apiService.get(
        `/google/sheets/refresh?${cacheBuster}`
      );

      if (response.success) {
        spreadsheets.value = response.data;
        // Refreshed spreadsheets

        sortSpreadsheetsByLastUsed();

        return response;
      } else {
        throw new Error(response.error || "Failed to refresh spreadsheets");
      }
    } catch (error) {
      console.error("❌ Failed to refresh spreadsheets:", error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function loadSpreadsheets(): Promise<any> {
    try {
      loading.value = true;
      // Add cache busting to ensure fresh data
      const cacheBuster = `_t=${Date.now()}`;
      const response = await apiService.get(
        `/google/sheets/list?${cacheBuster}`
      );
      if (response.success) {
        spreadsheets.value = response.data;
        // Loaded spreadsheets

        sortSpreadsheetsByLastUsed();

        return response;
      }
      return response;
    } catch (error) {
      console.error("Failed to load spreadsheets:", error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function loadWorksheets(spreadsheetId: string): Promise<any> {
    try {
      loading.value = true;
      const response = await apiService.get(
        `/google/worksheets/${spreadsheetId}`
      );
      if (response.success) {
        worksheets.value = response.data;

        return response;
      }
      return response;
    } catch (error) {
      console.error("Failed to load worksheets:", error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function updateSettings(settings: Record<string, any>): Promise<any> {
    try {
      loading.value = true;
      const response = await apiService.post(
        "/google/settings/update",
        settings
      );
      if (response.success) {
        await loadAuthStatus();
        editMode.value = false;
        return response;
      }
      return response;
    } catch (error) {
      console.error("Failed to update settings:", error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function testConnection(): Promise<any> {
    try {
      const response = await apiService.get("/google/settings/test");

      return response;
    } catch (error) {
      console.error("Failed to test connection:", error);
      throw error;
    }
  }

  async function logout(): Promise<any> {
    try {
      const response = await apiService.get("/google/auth/logout");
      if (response.success) {
        await loadAuthStatus();
        return response;
      }
      return response;
    } catch (error) {
      console.error("Failed to logout:", error);
      throw error;
    }
  }

  function startEdit(): void {
    editMode.value = true;
    tempSettings.value = { ...currentSettings.value };
  }

  function cancelEdit(): void {
    editMode.value = false;
    tempSettings.value = {};
  }

  return {
    // State
    authStatus,
    spreadsheets,
    worksheets,
    loading,
    editMode,
    tempSettings,

    // Getters
    isAuthenticated,
    currentSettings,
    inventorySpreadsheetId,
    inventorySheetName,

    // Actions
    loadAuthStatus,
    loadSpreadsheets,
    refreshSpreadsheets,
    loadWorksheets,
    updateSettings,
    testConnection,
    logout,
    startEdit,
    cancelEdit,
  };
});

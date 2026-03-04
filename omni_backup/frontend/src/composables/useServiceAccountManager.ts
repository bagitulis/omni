/**
 * Composable for Google Service Account management logic
 */

import { ref } from "vue";
import { useApi } from "./useApi";

interface ServiceAccount {
  id: string;
  email: string;
  isActive: boolean;
  quotaUsed: number;
}

interface SheetInfo {
  spreadsheetId: string;
  name: string;
  sheets?: Array<{
    sheetId: number;
    title: string;
    rowCount: number;
  }>;
}

export function useServiceAccountManager() {
  const api = useApi();

  const spreadsheetUrl = ref("");
  const urlLocked = ref(false);
  const sheetInfo = ref<SheetInfo | null>(null);
  const accounts = ref<ServiceAccount[]>([]);
  const loading = ref(false);
  const detecting = ref(false);
  const switching = ref(false);
  const detectError = ref("");
  const statusMsg = ref("");
  const statusType = ref<"success" | "error">("success");

  const loadServiceAccounts = async () => {
    loading.value = true;
    try {
      const response = await api.get("/google/service-accounts");
      accounts.value = response.data?.serviceAccounts || [];
    } catch (error) {
      console.error("Failed to load accounts:", (error as any).message);
    } finally {
      loading.value = false;
    }
  };

  const detectAndSaveSheet = async () => {
    if (!spreadsheetUrl.value) return;

    detecting.value = true;
    detectError.value = "";

    try {
      const response = await api.post("/google/service-accounts/detect-sheet", {
        spreadsheetUrl: spreadsheetUrl.value,
      });

      sheetInfo.value = response.data;
      urlLocked.value = true;

      statusMsg.value = `✅ "${response.data.name}" detected!`;
      statusType.value = "success" as const;

      setTimeout(() => {
        statusMsg.value = "";
      }, 3000);
    } catch (error) {
      detectError.value =
        (error as any).response?.data?.error || "Failed to detect spreadsheet";
    } finally {
      detecting.value = false;
    }
  };

  const toggleUrlLock = () => {
    urlLocked.value = false;
    spreadsheetUrl.value = "";
    sheetInfo.value = null;
    detectError.value = "";
  };

  const switchAccount = async (accountId: string) => {
    switching.value = true;

    try {
      await api.post("/google/service-accounts/switch", { email: accountId });
      await loadServiceAccounts();

      statusMsg.value = "✅ Service account switched";
      statusType.value = "success" as const;

      setTimeout(() => {
        statusMsg.value = "";
      }, 2000);
    } catch {
      statusMsg.value = "❌ Failed to switch account";
      statusType.value = "error";
    } finally {
      switching.value = false;
    }
  };

  return {
    spreadsheetUrl,
    urlLocked,
    sheetInfo,
    accounts,
    loading,
    detecting,
    switching,
    detectError,
    statusMsg,
    statusType,
    loadServiceAccounts,
    detectAndSaveSheet,
    toggleUrlLock,
    switchAccount,
  };
}

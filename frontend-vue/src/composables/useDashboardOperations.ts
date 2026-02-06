import { ref } from "vue";
import { useAppStore } from "../store/app";
import { useToast } from "./useToast";

interface WalletParams {
  month: number;
  year: number;
  transaction_type: string;
}

export function useDashboardOperations() {
  const appStore = useAppStore();
  const toast = useToast();

  const lastUpdated = ref<string>(new Date().toLocaleString());
  const walletParams = ref<WalletParams>({
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
    transaction_type: "wallet_order_income",
  });

  const shippingParams = ref({
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
    option: 1,
  });

  const showToast = (
    severity: "error" | "success" | "info" | "warn",
    summary: string,
    detail: string
  ) => {
    const severityMap: Record<string, string> = {
      error: "error",
      success: "success",
      info: "info",
      warn: "warning",
    };

    toast.add({
      severity: severityMap[severity] as any,
      summary,
      detail,
      life: 3000,
    });
  };

  const refreshAll = async () => {
    await appStore.refreshAll();
    lastUpdated.value = new Date().toLocaleString();
  };

  const loadStatus = async () => {
    await appStore.loadStatus(true);
    lastUpdated.value = new Date().toLocaleString();
  };

  const executeOperation = async (
    operation: string,
    params: Record<string, any> = {}
  ) => {
    try {
      const result = await appStore.executeOperation(operation, params);
      if (result.success) {
        showToast("success", "Success", result.message);
      } else {
        showToast("error", "Error", result.message);
      }
      return result;
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
      throw error;
    }
  };

  const updateWalletParams = (newParams: Record<string, any>) => {
    walletParams.value = { ...walletParams.value, ...newParams };
  };

  const updateShippingParams = (newParams: Record<string, any>) => {
    shippingParams.value = { ...shippingParams.value, ...newParams };
  };

  return {
    lastUpdated,
    walletParams,
    shippingParams,
    showToast,
    refreshAll,
    loadStatus,
    executeOperation,
    updateWalletParams,
    updateShippingParams,
  };
}

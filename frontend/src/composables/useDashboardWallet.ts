import { useAppStore } from "../store/app";
import { useToast } from "./useToast";

export function useDashboardWallet(
  walletParams: any,
  closeModalFn: (modal: string) => void
) {
  const appStore = useAppStore();
  // const walletStore = useWalletStore();
  const toast = useToast();

  const showToast = (
    severity: string,
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

  const exportWalletToSheets = async () => {
    try {
      const result = await appStore.executeSheetsOperation(
        "wallet_to_sheets",
        walletParams.value
      );

      if (result.success) {
        showToast("success", "Success", result.message);
        closeModalFn("wallet");
      } else {
        showToast("error", "Error", result.error || result.message);
      }
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  return {
    exportWalletToSheets,
  };
}

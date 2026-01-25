import { useAppStore } from "../store/app";
import { useShippingStore } from "../store/shipping";
import { useToast } from "./useToast";

export function useDashboardShipping(
  shippingParams: any,
  showForm: any,
  showList: any,
  closeModalFn: (modal: string) => void
) {
  const appStore = useAppStore();
  const shippingStore = useShippingStore();
  const toast = useToast();

  const showToast = (severity: string, summary: string, detail: string) => {
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

  const selectShippingOption = (option: number) => {
    shippingParams.value.option = option;
    if (option === 1) {
      showForm.value = true;
      showList.value = false;
    } else {
      showForm.value = false;
      showList.value = true;
      loadShippingFiles();
    }
  };

  const initializeShippingForm = () => {
    // Automatically select option 1 (Get Order Numbers from Wallet)
    selectShippingOption(1);
  };

  const loadShippingFiles = async () => {
    try {
      await shippingStore.getShippingFiles();
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  const processShippingFile = async (filename: string) => {
    try {
      const result = await shippingStore.processShippingFile(filename);
      if (result.success) {
        showToast("success", "Success", result.message);
        closeModalFn("shipping");
      } else {
        showToast("error", "Error", result.message);
      }
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  const exportShippingToSheets = async () => {
    try {
      const result = await appStore.executeSheetsOperation(
        "shipping_fee_to_sheets",
        shippingParams.value
      );
      if (result.success) {
        showToast("success", "Success", result.message);
        closeModalFn("shipping");
      } else {
        showToast("error", "Error", result.message);
      }
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  return {
    selectShippingOption,
    loadShippingFiles,
    processShippingFile,
    exportShippingToSheets,
    initializeShippingForm,
  };
}

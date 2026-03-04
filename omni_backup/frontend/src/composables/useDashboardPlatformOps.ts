import { computed } from "vue";
import { useAppStore } from "../store/app";
import { useUIStore } from "../store/ui";
import { useToast } from "./useToast";

export function useDashboardPlatformOps() {
  const appStore = useAppStore();
  const uiStore = useUIStore();
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

  const handlePlatformOperation = async (
    operation: string,
    platform: string | null = null
  ) => {
    const targetPlatform = platform || uiStore.activePlatform;

    if (operation === "export_orders") {
      return { type: "show-modal", modal: "exportOrders", platform: targetPlatform };
    }

    if (operation === "update_price") {
      if (targetPlatform === "shopee") {
        return { type: "show-modal", modal: "price", platform: targetPlatform };
      }
      return executePriceUpdate();
    }

    return executeDefaultOperation(targetPlatform, operation);
  };

  const executeDefaultOperation = async (
    platform: string,
    operation: string
  ) => {
    try {
      const result = await appStore.executeOperation(
        `${platform}_operation`,
        { operation }
      );
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
    }
  };

  const executePriceUpdate = async () => {
    const platform = appStore.modals.price.platform;
    const currentPriceOption = computed(() => appStore.modals.price.option || "regular").value;

    try {
      const result = await appStore.executeOperation(
        `${platform}_operation`,
        { operation: "update_price", price_type: currentPriceOption }
      );
      if (result.success) {
        showToast("success", "Success", result.message);
      } else {
        showToast("error", "Error", result.message);
      }
      return result;
    } catch (error: any) {
      showToast("error", "Error", error.message);
    }
  };

  const executeExportOrder = async (exportType: string) => {
    const platform = appStore.modals.exportOrders.platform;
    if (!platform) {
      showToast("warn", "Warning", "Platform not selected");
      return;
    }
    try {
      const result = await appStore.exportOrders(platform, exportType);
      if (result.success) {
        showToast("success", "Success", result.message || "Export completed");
      } else {
        showToast("error", "Error", result.message || "Export failed");
      }
      return result;
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  const handleTokenOperation = async (operation: string) => {
    const platform = appStore.modals.token.platform;
    if (!platform) {
      showToast("warn", "Warning", "Platform not selected");
      return;
    }
    let code: string | null = null;

    if (operation === "update_code") {
      code = prompt(`Enter new authorization code for ${platform}:`);
      if (!code) return;
    }

    try {
      const result = await appStore.handleTokenOperation(
        platform,
        operation,
        code
      );
      if (result.success) {
        showToast("success", "Success", result.message || "Operation completed");
      } else {
        showToast("error", "Error", result.message || "Operation failed");
      }
      return result;
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : "Unknown error";
      showToast("error", "Error", errorMessage);
    }
  };

  return {
    handlePlatformOperation,
    executePriceUpdate,
    executeExportOrder,
    handleTokenOperation,
    executeDefaultOperation,
  };
}

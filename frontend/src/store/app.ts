import { defineStore } from "pinia";
import { ref, computed, Ref } from "vue";
import { useToast } from "@/composables/useToast";
import {
  useAppConnection,
  useAppLogging,
  useAppOperations,
} from "@/composables/useAppCore";
import apiService from "@/services/api";
import { useModalsStore } from "./modals";
import { useOrdersStore } from "./orders";
import { loadStatusWithTokens, syncMasterProduct } from "./appOperations";
import {
  addLog as sharedAddLog,
  executeOperation as sharedExecuteOperation,
} from "@/composables/useSharedAppActions";

export const useAppStore = defineStore("app", () => {
  // Import other stores (no circular deps now)
  const modalsStore = useModalsStore();
  const ordersStore = useOrdersStore();

  // Lazy load shipping/wallet stores to avoid any circular issues
  const getShippingStore = () => {
    const { useShippingStore } = require("./shipping");
    return useShippingStore();
  };
  const getWalletStore = () => {
    const { useWalletStore } = require("./wallet");
    return useWalletStore();
  };

  // Connection
  const {
    connectionStatus,
    isReconnecting,
    isConnected,
    startStatusPolling,
    cleanup,
  } = useAppConnection();

  // Logging
  const { logs, addLog, clearLogs } = useAppLogging();

  // State
  const status: Ref<any> = ref(null);
  const loading: Ref<boolean> = ref(false);
  const requestCount: Ref<number> = ref(0);
  const errorCount: Ref<number> = ref(0);
  const debugInfo: Ref<any> = ref(null);
  const spreadsheets: Ref<any[]> = ref([]);
  const lastStatusUpdate: Ref<Date | null> = ref(null);
  const lastTokenRefreshAt: Ref<number | null> = ref(null);

  const { updatePrice, exportOrders, handleTokenOperation } =
    useAppOperations(addLog);

  // Delegate modals to modalsStore
  const modals = computed(() => modalsStore.modals);
  const showModal = modalsStore.showModal;
  const closeModal = modalsStore.closeModal;

  // Delegate shipping (lazy loaded to avoid circular)
  const shippingFiles = computed(() => getShippingStore().shippingFiles);
  const getShippingFiles = () => getShippingStore().getShippingFiles();
  const processShippingFile = (f: string) =>
    getShippingStore().processShippingFile(f);
  const processShippingFee = (p: Record<string, any>) =>
    getShippingStore().processShippingFee(p);

  // Delegate wallet (lazy loaded to avoid circular)
  const getWalletTransactions = (p: Record<string, any>) =>
    getWalletStore().getWalletTransactions(p);

  // Delegate orders to ordersStore
  const executeOrderExport = ordersStore.executeOrderExport;

  async function initializeApp(): Promise<void> {
    // Non-blocking: Set initial state immediately for faster FCP
    loading.value = true;
    connectionStatus.value = "connecting";
    addLog("🔄 Connecting to backend server...");

    // Defer health check to after paint using requestIdleCallback
    const performHealthCheck = async () => {
      try {
        const health = await Promise.race([
          apiService.healthCheck(),
          new Promise<any>((_, reject) =>
            setTimeout(() => reject(new Error("Health check timeout")), 5000)
          ),
        ]);

        if (health.status === "healthy") {
          connectionStatus.value = "connected";
          addLog("✅ Backend connected successfully");
          // Defer status loading further to reduce blocking
          requestAnimationFrame(() => {
            loadStatus().catch(() => {});
            startStatusPolling(async () => {
              await loadStatus(false);
            });
          });
        } else {
          connectionStatus.value = "error";
          addLog("❌ Backend health check failed");
        }
      } catch (error: any) {
        connectionStatus.value = "error";
        addLog(`❌ Failed to connect to backend: ${error.message}`);
        scheduleReconnect();
      } finally {
        loading.value = false;
      }
    };

    // Use requestIdleCallback for non-critical health check
    // Falls back to setTimeout for browsers without support
    if (typeof requestIdleCallback !== "undefined") {
      requestIdleCallback(() => performHealthCheck(), { timeout: 1000 });
    } else {
      setTimeout(() => performHealthCheck(), 50);
    }
  }

  function scheduleReconnect(): void {
    const retryDelay = Math.min(5000 * 1.5, 30000);
    setTimeout(() => {
      if (connectionStatus.value !== "connected") initializeApp();
    }, retryDelay);
  }

  async function refreshSpreadsheets(): Promise<any> {
    try {
      loading.value = true;
      const response = await apiService.get("/google/sheets/refresh");
      if (response.success) {
        spreadsheets.value = response.data || [];
        console.log(`✅ Refreshed ${response.data.length} spreadsheets`);
        return response;
      }
      throw new Error(response.error || "Failed to refresh spreadsheets");
    } catch (error: any) {
      const { error: showError } = useToast();
      showError("Refresh Failed", error.message);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function loadStatus(
    shouldLog: boolean = true,
    refreshTokens: boolean = false
  ): Promise<any> {
    try {
      const response = await loadStatusWithTokens(
        status,
        lastTokenRefreshAt,
        connectionStatus,
        isReconnecting,
        loading,
        handleTokenOperation,
        addLog,
        shouldLog,
        refreshTokens
      );
      lastStatusUpdate.value = new Date();
      return response;
    } catch (error: any) {
      handleConnectionError(error);
      throw error;
    }
  }

  function handleConnectionError(error: any): void {
    if (connectionStatus.value === "connected") {
      connectionStatus.value = "error";
      addLog(`❌ Connection lost: ${error.message}`);
      setTimeout(() => {
        if (connectionStatus.value !== "connected") loadStatus(true);
      }, 3000);
    }
  }

  async function executeOperation(
    operation: string,
    params: Record<string, any> = {}
  ): Promise<any> {
    try {
      loading.value = true;
      requestCount.value++;

      const isStatusOperation =
        operation.includes("status") || operation === "get_token_status";

      if (!isStatusOperation) {
        addLog(`🔄 Executing operation: ${operation}`);
      }

      const response = await apiService.executeOperation(operation, params);

      if ((response as any).success) {
        if (!isStatusOperation) {
          logOperationSuccess(operation, response);
        }
        await loadStatus(false);
      } else {
        errorCount.value++;
        addLog(`❌ Operation ${operation} failed: ${(response as any).error}`);
      }
      return response;
    } catch (error: any) {
      handleOperationError(operation, error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  function logOperationSuccess(operation: string, response: any): void {
    addLog(`✅ Operation ${operation} completed successfully`);
    if (response.results) {
      for (const [platform, result] of Object.entries(response.results)) {
        if (result && typeof result === "object") {
          const res = result as any;
          addLog(
            `   ${platform}: ${res.processed || 0} processed, ${res.skipped || 0} skipped, ${res.failed || 0} failed`
          );
        }
      }
    }
  }

  function handleOperationError(operation: string, error: any): void {
    errorCount.value++;
    if (
      error.message.includes("Network error") ||
      error.message.includes("cannot connect")
    ) {
      connectionStatus.value = "error";
      addLog(`❌ Operation ${operation} error: ${error.message}`);
      setTimeout(() => {
        if (connectionStatus.value !== "connected") loadStatus(true);
      }, 3000);
    } else {
      addLog(`❌ Operation ${operation} error: ${error.message}`);
    }
  }

  async function executeSheetsOperation(
    operation: string,
    params: Record<string, any> = {}
  ): Promise<any> {
    try {
      loading.value = true;
      requestCount.value++;
      addLog(`🔄 Executing sheets operation: ${operation}`);

      const response = await apiService.executeSheetsOperation(
        operation,
        params
      );

      if (response.success) {
        addLog(`✅ Sheets operation ${operation} completed successfully`);
        await loadStatus(false);
      } else {
        errorCount.value++;
        addLog(`❌ Sheets operation ${operation} failed: ${response.error}`);
      }
      return response;
    } catch (error: any) {
      errorCount.value++;
      addLog(`❌ Sheets operation ${operation} error: ${error.message}`);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function debugCheckFunctions(): Promise<any> {
    try {
      const response = await apiService.debugCheckFunctions();
      debugInfo.value = response.data;
      return response;
    } catch (error: any) {
      addLog(`❌ Debug check failed: ${error.message}`);
      throw error;
    }
  }

  async function refreshAll(): Promise<void> {
    addLog("🔄 Manual refresh triggered");
    try {
      await loadStatus(true);
      await syncMasterProduct(addLog);
    } catch (error: any) {
      addLog(`⚠️ Error during refresh: ${error.message}`);
    }
  }

  return {
    status,
    loading,
    logs,
    connectionStatus,
    requestCount,
    errorCount,
    modals,
    shippingFiles,
    debugInfo,
    spreadsheets,
    lastStatusUpdate,
    isReconnecting,
    isConnected,
    initializeApp,
    loadStatus,
    executeOperation,
    showModal,
    closeModal,
    handleTokenOperation,
    getWalletTransactions,
    getShippingFiles,
    processShippingFile,
    processShippingFee,
    updatePrice,
    exportOrders,
    executeOrderExport,
    debugCheckFunctions,
    addLog,
    clearLogs,
    refreshAll,
    cleanup,
    executeSheetsOperation,
    refreshSpreadsheets,
  };
});

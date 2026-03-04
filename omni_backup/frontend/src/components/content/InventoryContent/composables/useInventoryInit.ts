/**
 * Inventory Initialization Composable
 * Handles the initialization sequence and lifecycle
 */

import { useInventoryData } from "./useInventoryData";
import { useInventorySyncSheets } from "./useInventorySyncSheets";
import { useInventoryForm } from "./useInventoryForm";
import { useInventoryFilter } from "./useInventoryFilter";
import { useInventoryConfig, resetInventoryConfig } from "./useInventoryConfig";
import { useInventoryAlerts } from "./useInventoryAlerts";
import { useInventoryCellEdit } from "./useInventoryCellEdit";
import { useInventoryBatchCheck } from "./useInventoryBatchCheck";
import { useInventoryItemEdit } from "./useInventoryItemEdit";
import { useInventoryStockUpdate } from "./useInventoryStockUpdate";
import { useInventoryPriceUpdate } from "./useInventoryPriceUpdate";
import { useInventoryWholesale } from "./useInventoryWholesale";
import { useInventoryPagination } from "./useInventoryPagination";
import { resetEditableLock } from "./useEditableLock";
import { useLockStock } from "./useLockStock";
import { loadSortState } from "./useSortPersistence";

interface InventoryState {
  loading: boolean;
  isInitialized: boolean;
  lockStock: {
    loading: boolean;
    error: string | null;
    map: Record<string, number>;
  };
  sortColumn: string | null;
  sortDirection: string | null;
  syncAlert: any;
  configData: { spreadsheet_id: string; sheet_name: string };
  // Pagination properties
  currentOffset: number;
  pageSize: number;
  totalRecords: number;
  // Wholesale properties
  showWholesaleModal: boolean;
  wholesaleSkus: string[];
  showWholesaleUpdateModal: boolean;
  wholesaleUpdateItems: Array<{
    sku: string;
    price: number;
    platform?: string;
  }>;
  showWholesaleMpqModal: boolean;
  wholesaleMpqItems: Array<{ sku: string; price: number; platform?: string }>;
  [key: string]: any;
}

export function useInventoryInit(state: InventoryState) {
  const alertMethods = useInventoryAlerts();
  const configMethods = useInventoryConfig();

  /**
   * Initialize all composable methods
   */
  function initializeMethods(): Record<string, any> {
    // Sync alert property
    Object.defineProperty(state, "syncAlert", {
      get: () => alertMethods.state.syncAlert,
      set: (val) => {
        alertMethods.state.syncAlert = val;
      },
      configurable: true,
    });

    const alertAPI = {
      success: alertMethods.showAlertSuccess,
      error: alertMethods.showAlertError,
      warning: alertMethods.showAlertWarning,
      info: alertMethods.showAlertInfo,
    };

    const dataMethods = useInventoryData(state);
    const reloadMethods = {
      inventoryData: () => dataMethods.loadInventoryData(),
    };

    return {
      dataMethods,
      syncMethods: useInventorySyncSheets(
        state,
        alertAPI,
        dataMethods,
        configMethods
      ),
      formMethods: useInventoryForm(state, alertAPI, reloadMethods),
      filterMethods: useInventoryFilter(state),
      cellEditMethods: useInventoryCellEdit(state, alertAPI.error),
      batchCheckMethods: useInventoryBatchCheck(state, alertAPI),
      itemEditMethods: useInventoryItemEdit(state, alertAPI),
      stockUpdateMethods: useInventoryStockUpdate(state),
      priceUpdateMethods: useInventoryPriceUpdate(state),
      wholesaleMethods: useInventoryWholesale(state),
      paginationMethods: useInventoryPagination(state),
    };
  }

  /**
   * Load persisted sort state
   */
  function loadPersistedSort(): void {
    const persistedSort = loadSortState();
    state.sortColumn = persistedSort.column;
    state.sortDirection = persistedSort.direction;
  }

  /**
   * Initialize inventory data sequence
   */
  async function initializeInventory(
    methods: Record<string, any>
  ): Promise<void> {
    await methods.dataMethods.loadConfiguration();
    await methods.filterMethods.loadFilterPreferences();
    methods.filterMethods.syncStateFromConfig();
    await methods.dataMethods.loadStats();
    state.isInitialized = true;
    await methods.dataMethods.loadInventoryData();
    await loadLockStockData();
    await methods.batchCheckMethods?.loadSavedPlatformStatus();
  }

  /**
   * Load lock stock data
   */
  async function loadLockStockData(): Promise<void> {
    const lockStockComposable = useLockStock();
    state.lockStock.loading = true;
    try {
      await lockStockComposable.fetchLockedStockData();
      state.lockStock.map = lockStockComposable.lockedStockMap.value;
      state.lockStock.error = null;
    } catch (error: any) {
      console.warn("⚠️ Failed to load lock stock data:", error);
      state.lockStock.error = error.message;
    } finally {
      state.lockStock.loading = false;
    }
  }

  /**
   * Show setup warning if config is missing
   */
  function showSetupWarningIfNeeded(): void {
    setTimeout(() => {
      if (!state.configData.spreadsheet_id || !state.configData.sheet_name) {
        alertMethods.showAlertWarning(
          "⚠️ Setup Diperlukan",
          "Please click '⚙️ Settings' to setup spreadsheet and inventory columns"
        );
      }
    }, 1500);
  }

  /**
   * Cleanup on unmount
   */
  function cleanup(searchDebounceTimer?: ReturnType<typeof setTimeout>): void {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
    resetInventoryConfig();
    resetEditableLock();
  }

  return {
    initializeMethods,
    loadPersistedSort,
    initializeInventory,
    loadLockStockData,
    showSetupWarningIfNeeded,
    cleanup,
  };
}

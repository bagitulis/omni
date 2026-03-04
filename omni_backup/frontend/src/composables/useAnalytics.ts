/**
 * Analytics Composable
 * Handles analytics API operations for price reconciliation
 * Single Responsibility: Analytics state and API calls
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";
import { useToast } from "./useToast";

export interface SyncStatus {
  synced: boolean;
  totalOrders: number;
  syncedAt: string | null;
}

export interface AnalyticsSettings {
  priceColumn: string;
  formulaDeduction: number;
  formulaMultiplier: number;
}

export interface PriceVariant {
  price: number;
  count: number;
  totalEscrow: number;
  avgEscrow: number;
  transactions: TransactionDetail[];
}

export interface TransactionDetail {
  orderSn: string;
  orderDate: string;
  price: number;
  escrowAmount: number;
  quantity: number;
}

export interface SkuGroup {
  sku: string;
  modelSku: string;
  itemName: string;
  modelName: string;
  inventoryPrice: number | null;
  expectedIncome: number | null;
  totalTransactions: number;
  uniqueUnitPrices: number[];
  uniqueActualIncomes: number[];
  priceVariants: PriceVariant[];
  hasMultiplePrices: boolean;
  hasPriceDifference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface ReconciliationSummary {
  totalSku: number;
  totalTransactions: number;
  skuOk: number;
  skuWithPriceDiff: number;
  skuNoInventory: number;
}

export interface ReconciliationResult {
  summary: ReconciliationSummary;
  skuGroups: SkuGroup[];
}

export interface ShippingFeeOrder {
  orderSn: string;
  orderDate: string | null;
  buyerPaid: number;
  actualFee: number;
  shopeeRebate: number;
  difference: number;
  buyerName: string | null;
  paymentMethod: string | null;
}

export interface ShippingFeeSummary {
  totalOrders: number;
  ordersWithDifference: number;
  totalProfit: number;
  totalLoss: number;
  netImpact: number;
}

export interface ShippingFeeResult {
  summary: ShippingFeeSummary;
  orders: ShippingFeeOrder[];
}

export function useAnalytics() {
  const api = useApi();
  const { success: showSuccess, error: showError } = useToast();

  // State
  const loading = ref(false);
  const syncing = ref(false);
  const syncStatus = ref<SyncStatus | null>(null);
  const settings = ref<AnalyticsSettings>({
    priceColumn: "HARGA",
    formulaDeduction: 1500,
    formulaMultiplier: 0.84,
  });
  const reconciliationResult = ref<ReconciliationResult | null>(null);
  const shippingFeeResult = ref<ShippingFeeResult | null>(null);
  const selectedMonth = ref(new Date().getMonth()); // 0-indexed
  const selectedYear = ref(new Date().getFullYear());

  // Computed
  const currentPeriod = computed(() => ({
    month: selectedMonth.value + 1, // API uses 1-indexed
    year: selectedYear.value,
  }));

  const canSync = computed(() => {
    const now = new Date();
    const isCurrentMonth =
      selectedMonth.value === now.getMonth() &&
      selectedYear.value === now.getFullYear();
    return !isCurrentMonth && !syncing.value;
  });

  const periodLabel = computed(() => {
    const months = [
      "Januari",
      "Februari",
      "Maret",
      "April",
      "Mei",
      "Juni",
      "Juli",
      "Agustus",
      "September",
      "Oktober",
      "November",
      "Desember",
    ];
    return `${months[selectedMonth.value]} ${selectedYear.value}`;
  });

  // Actions
  async function fetchSyncStatus(): Promise<void> {
    try {
      loading.value = true;
      const response = await api.get<{ success: boolean; data: SyncStatus }>(
        `/analytics/shopee/sync-status?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`
      );
      syncStatus.value = response.data;
    } catch (error) {
      console.error("Failed to fetch sync status:", error);
      syncStatus.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchSettings(): Promise<void> {
    try {
      const response = await api.get<{
        success: boolean;
        data: AnalyticsSettings;
      }>("/analytics/shopee/settings");
      settings.value = response.data;
    } catch (error) {
      console.error("Failed to fetch settings:", error);
    }
  }

  async function saveSettings(newSettings: AnalyticsSettings): Promise<void> {
    try {
      await api.post("/analytics/shopee/settings", newSettings);
      settings.value = newSettings;
      showSuccess("Settings saved successfully");
    } catch (error) {
      showError("Failed to save settings");
      throw error;
    }
  }

  async function syncEscrowData(forceResync = false): Promise<void> {
    if (!canSync.value) {
      showError("Cannot sync current month. Wait until month ends.");
      return;
    }

    try {
      syncing.value = true;
      const response = await api.post<{
        success: boolean;
        message: string;
        data: { totalOrders: number; totalItems: number };
      }>("/analytics/shopee/sync", {
        month: currentPeriod.value.month,
        year: currentPeriod.value.year,
        forceResync,
      });

      if (response.success) {
        showSuccess(response.message);
        await fetchSyncStatus();
      }
    } catch (error) {
      showError("Failed to sync escrow data");
      throw error;
    } finally {
      syncing.value = false;
    }
  }

  async function deleteSyncData(): Promise<void> {
    try {
      loading.value = true;
      await api.delete(
        `/analytics/shopee/sync?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`
      );
      showSuccess("Sync data deleted successfully");
      await fetchSyncStatus();
      reconciliationResult.value = null;
      shippingFeeResult.value = null;
    } catch (error) {
      showError("Failed to delete sync data");
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function fetchReconciliation(): Promise<void> {
    try {
      loading.value = true;
      const response = await api.get<{
        success: boolean;
        data: ReconciliationResult;
      }>(
        `/analytics/shopee/reconciliation?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`
      );
      reconciliationResult.value = response.data;
    } catch (error) {
      console.error("Failed to fetch reconciliation:", error);
      reconciliationResult.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchShippingFeeAnalysis(): Promise<void> {
    try {
      loading.value = true;
      const response = await api.get<{
        success: boolean;
        data: ShippingFeeResult;
      }>(
        `/analytics/shopee/shipping-fee?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`
      );
      shippingFeeResult.value = response.data;
    } catch (error) {
      console.error("Failed to fetch shipping fee analysis:", error);
      shippingFeeResult.value = null;
    } finally {
      loading.value = false;
    }
  }

  function setPeriod(month: number, year: number): void {
    selectedMonth.value = month;
    selectedYear.value = year;
  }

  function goToPreviousMonth(): void {
    if (selectedMonth.value === 0) {
      selectedMonth.value = 11;
      selectedYear.value--;
    } else {
      selectedMonth.value--;
    }
  }

  function goToNextMonth(): void {
    if (selectedMonth.value === 11) {
      selectedMonth.value = 0;
      selectedYear.value++;
    } else {
      selectedMonth.value++;
    }
  }

  // Initialize
  async function initialize(): Promise<void> {
    // Default to previous month
    goToPreviousMonth();
    await Promise.all([fetchSyncStatus(), fetchSettings()]);
  }

  return {
    // State
    loading,
    syncing,
    syncStatus,
    settings,
    reconciliationResult,
    shippingFeeResult,
    selectedMonth,
    selectedYear,

    // Computed
    currentPeriod,
    canSync,
    periodLabel,

    // Actions
    fetchSyncStatus,
    fetchSettings,
    saveSettings,
    syncEscrowData,
    deleteSyncData,
    fetchReconciliation,
    fetchShippingFeeAnalysis,
    setPeriod,
    goToPreviousMonth,
    goToNextMonth,
    initialize,
  };
}

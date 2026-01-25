/**
 * TikTok Analytics Composable
 * Handles TikTok analytics API operations for price and shipping reconciliation
 * Single Responsibility: TikTok analytics state and API calls
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";
import { useToast } from "./useToast";

export interface TiktokSyncStatus {
  synced: boolean;
  totalOrders: number;
  syncedAt: string | null;
}

export interface TiktokAnalyticsSettings {
  priceColumn: string;
  formulaDeduction: number;
  formulaMultiplier: number;
}

export interface TiktokSkuGroup {
  sku: string;
  sellerSku: string;
  productName: string;
  inventoryPrice: number | null;
  expectedIncome: number | null;
  totalTransactions: number;
  uniqueUnitPrices: number[];
  uniqueActualIncomes: number[];
  hasMultiplePrices: boolean;
  hasPriceDifference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface TiktokReconciliationSummary {
  totalSku: number;
  totalTransactions: number;
  skuOk: number;
  skuWithPriceDiff: number;
  skuNoInventory: number;
}

export interface TiktokReconciliationResult {
  summary: TiktokReconciliationSummary;
  skuGroups: TiktokSkuGroup[];
}

export interface TiktokShippingFeeOrder {
  orderId: string;
  orderDate: string | null;
  buyerPaid: number;
  actualFee: number;
  platformDiscount: number;
  difference: number;
  orderStatus: string | null;
  currency: string;
}

export interface TiktokShippingFeeSummary {
  totalOrders: number;
  ordersWithDifference: number;
  totalProfit: number;
  totalLoss: number;
  netImpact: number;
}

export interface TiktokShippingFeeResult {
  summary: TiktokShippingFeeSummary;
  orders: TiktokShippingFeeOrder[];
}

export function useTiktokAnalytics() {
  const api = useApi();
  const { success: showSuccess, error: showError } = useToast();

  // State
  const loading = ref(false);
  const syncing = ref(false);
  const syncStatus = ref<TiktokSyncStatus | null>(null);
  const settings = ref<TiktokAnalyticsSettings>({
    priceColumn: "HARGA",
    formulaDeduction: 1500,
    formulaMultiplier: 0.84,
  });
  const reconciliationResult = ref<TiktokReconciliationResult | null>(null);
  const shippingFeeResult = ref<TiktokShippingFeeResult | null>(null);

  // Default to PREVIOUS month so sync button is enabled by default
  // (Current month cannot be synced - escrow data not finalized)
  const now = new Date();
  const prevMonth = now.getMonth() === 0 ? 11 : now.getMonth() - 1; // 0-indexed
  const prevYear =
    now.getMonth() === 0 ? now.getFullYear() - 1 : now.getFullYear();
  const selectedMonth = ref(prevMonth);
  const selectedYear = ref(prevYear);

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
      const response = await api.get<{
        success: boolean;
        data: TiktokSyncStatus;
      }>(
        `/analytics/tiktok/sync-status?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
      );
      syncStatus.value = response.data;
    } catch (error) {
      console.error("Failed to fetch TikTok sync status:", error);
      syncStatus.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function fetchSettings(): Promise<void> {
    try {
      const response = await api.get<{
        success: boolean;
        data: TiktokAnalyticsSettings;
      }>("/analytics/tiktok/settings");
      settings.value = response.data;
    } catch (error) {
      console.error("Failed to fetch TikTok settings:", error);
    }
  }

  async function saveSettings(
    newSettings: TiktokAnalyticsSettings,
  ): Promise<void> {
    try {
      await api.post("/analytics/tiktok/settings", newSettings);
      settings.value = newSettings;
      showSuccess("TikTok settings saved successfully");
    } catch (error) {
      showError("Failed to save TikTok settings");
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
        data: { totalOrders: number; totalItems: number; failedOrders: number };
      }>("/analytics/tiktok/sync", {
        month: currentPeriod.value.month,
        year: currentPeriod.value.year,
        forceResync,
      });

      if (response.success) {
        showSuccess(response.message);
        await fetchSyncStatus();
      }
    } catch (error) {
      showError("Failed to sync TikTok escrow data");
      throw error;
    } finally {
      syncing.value = false;
    }
  }

  async function deleteSyncData(): Promise<void> {
    try {
      loading.value = true;
      await api.delete(
        `/analytics/tiktok/sync?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
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
        data: TiktokReconciliationResult;
      }>(
        `/analytics/tiktok/reconciliation?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
      );
      reconciliationResult.value = response.data;
    } catch (error) {
      console.error("Failed to fetch TikTok reconciliation:", error);
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
        data: TiktokShippingFeeResult;
      }>(
        `/analytics/tiktok/shipping-fee?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
      );
      shippingFeeResult.value = response.data;
    } catch (error) {
      console.error("Failed to fetch TikTok shipping fee analysis:", error);
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

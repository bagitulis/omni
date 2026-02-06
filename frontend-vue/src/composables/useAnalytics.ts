/**
 * Analytics Composable
 * Handles analytics API operations for Shopee price reconciliation
 * Single Responsibility: Analytics state and API calls
 */

import { ref, computed } from "vue";
import { useApi } from "./useApi";
import { useToast } from "./useToast";

export interface SyncStatus {
  synced: boolean;
  total_orders: number;
  synced_at: string | null;
}

export interface AnalyticsSettings {
  price_column: string;
  formula_deduction: number;
  formula_multiplier: number;
}

export interface PriceVariant {
  price: number;
  count: number;
  total_escrow: number;
  avg_escrow: number;
  transactions: TransactionDetail[];
}

export interface TransactionDetail {
  order_sn: string;
  order_date: string;
  price: number;
  escrow_amount: number;
  quantity: number;
}

export interface SkuGroup {
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  inventory_price: number | null;
  expected_income: number | null;
  total_transactions: number;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  price_variants: PriceVariant[];
  has_multiple_prices: boolean;
  has_price_difference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface ReconciliationSummary {
  total_sku: number;
  total_transactions: number;
  sku_ok: number;
  sku_with_price_diff: number;
  sku_no_inventory: number;
}

export interface ReconciliationResult {
  summary: ReconciliationSummary;
  sku_groups: SkuGroup[];
}

export interface ShippingFeeOrder {
  order_sn: string;
  order_date: string | null;
  buyer_paid: number;
  actual_fee: number;
  shopee_rebate: number;
  difference: number;
  buyer_name: string | null;
  payment_method: string | null;
}

export interface ShippingFeeSummary {
  total_orders: number;
  orders_with_difference: number;
  total_profit: number;
  total_loss: number;
  net_impact: number;
}

export interface ShippingFeeResult {
  summary: ShippingFeeSummary;
  orders: ShippingFeeOrder[];
}

// Job progress interface for background sync
export interface JobProgress {
  id: string;
  type: string;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  progress_percent: number;
  progress_message: string;
  total_items: number;
  processed_items: number;
  error_message?: string;
  result_data?: string;
  started_at?: string;
  completed_at?: string;
}

export function useAnalytics() {
  const api = useApi();
  const { success: showSuccess, error: showError } = useToast();

  // State
  const loading = ref(false);
  const syncing = ref(false);
  const syncStatus = ref<SyncStatus | null>(null);
  const settings = ref<AnalyticsSettings>({
    price_column: "HARGA",
    formula_deduction: 1500,
    formula_multiplier: 0.84,
  });
  const reconciliationResult = ref<ReconciliationResult | null>(null);
  const shippingFeeResult = ref<ShippingFeeResult | null>(null);

  // Job progress state
  const currentJobId = ref<string | null>(null);
  const jobProgress = ref<JobProgress | null>(null);
  const pollingInterval = ref<ReturnType<typeof setInterval> | null>(null);

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

  // Job progress computed
  const syncProgressPercent = computed(
    () => jobProgress.value?.progress_percent ?? 0,
  );
  const syncProgressMessage = computed(
    () => jobProgress.value?.progress_message ?? "",
  );
  const isSyncRunning = computed(() => jobProgress.value?.status === "running");

  // Stop polling
  function stopPolling(): void {
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value);
      pollingInterval.value = null;
    }
  }

  // Poll job status
  async function pollJobStatus(jobId: string): Promise<void> {
    try {
      const response = await api.get<{
        success: boolean;
        job: JobProgress;
      }>(`/jobs/${jobId}`);

      if (response.success && response.job) {
        jobProgress.value = response.job;

        // Check if job completed or failed
        if (response.job.status === "completed") {
          stopPolling();
          syncing.value = false;
          currentJobId.value = null;
          showSuccess("Shopee escrow sync completed successfully");
          await fetchSyncStatus();
        } else if (response.job.status === "failed") {
          stopPolling();
          syncing.value = false;
          currentJobId.value = null;
          showError(response.job.error_message || "Escrow sync failed");
        } else if (response.job.status === "cancelled") {
          stopPolling();
          syncing.value = false;
          currentJobId.value = null;
          showError("Escrow sync was cancelled");
        }
      }
    } catch (error) {
      console.error("Failed to poll job status:", error);
      // Don't stop polling on error, retry
    }
  }

  // Start polling for job progress
  function startPolling(jobId: string): void {
    stopPolling(); // Clear any existing polling
    currentJobId.value = jobId;
    jobProgress.value = {
      id: jobId,
      type: "shopee_escrow_sync",
      status: "pending",
      progress_percent: 0,
      progress_message: "Starting sync...",
      total_items: 0,
      processed_items: 0,
    };

    // Poll immediately, then every 2 seconds
    pollJobStatus(jobId);
    pollingInterval.value = setInterval(() => pollJobStatus(jobId), 2000);
  }

  // Actions
  async function fetchSyncStatus(): Promise<void> {
    try {
      loading.value = true;
      const response = await api.get<{ success: boolean; data: SyncStatus }>(
        `/analytics/shopee/sync-status?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
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
      jobProgress.value = null;

      const response = await api.post<{
        success: boolean;
        message: string;
        data?: {
          job_id: string;
          total_orders?: number;
          total_items?: number;
        };
      }>("/analytics/shopee/sync", {
        month: currentPeriod.value.month,
        year: currentPeriod.value.year,
        force_resync: forceResync,
      });

      if (response.success) {
        // Check if this is async job response
        if (response.data?.job_id) {
          showSuccess("Escrow sync started. Monitoring progress...");
          startPolling(response.data.job_id);
        } else {
          // Sync completed immediately (no background job)
          showSuccess(response.message);
          syncing.value = false;
          await fetchSyncStatus();
        }
      }
    } catch (error) {
      syncing.value = false;
      jobProgress.value = null;
      showError("Failed to sync escrow data");
      throw error;
    }
  }

  async function deleteSyncData(): Promise<void> {
    try {
      loading.value = true;
      await api.delete(
        `/analytics/shopee/sync?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
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
        `/analytics/shopee/reconciliation?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
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
        `/analytics/shopee/shipping-fee?month=${currentPeriod.value.month}&year=${currentPeriod.value.year}`,
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

  // Cancel ongoing sync
  function cancelSync(): void {
    stopPolling();
    syncing.value = false;
    currentJobId.value = null;
    jobProgress.value = null;
  }

  // Initialize
  async function initialize(): Promise<void> {
    // Default to previous month
    goToPreviousMonth();
    await Promise.all([fetchSyncStatus(), fetchSettings()]);
  }

  // Cleanup on unmount
  function cleanup(): void {
    stopPolling();
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

    // Job progress state
    currentJobId,
    jobProgress,
    syncProgressPercent,
    syncProgressMessage,
    isSyncRunning,

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
    cancelSync,
    initialize,
    cleanup,
  };
}

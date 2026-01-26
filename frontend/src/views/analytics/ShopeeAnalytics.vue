<template>
  <div class="dashboard">
    <div class="main-layout">
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="uiStore.activeTab"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />
      <main class="main-content">
        <div class="analytics-page">
          <div class="analytics-header">
            <div class="header-left">
              <h1>🛒 Shopee Report</h1>
              <p class="subtitle">Price & Shipping Fee Analysis</p>
            </div>
            <div class="header-right">
              <button
                class="btn-icon"
                @click="showSettings = true"
                type="button"
                aria-label="Open analytics settings"
              >
                <span aria-hidden="true">⚙️</span>
              </button>
            </div>
          </div>
          <div class="tab-navigation">
            <button
              :class="['tab-btn', { active: analyticsTab === 'price' }]"
              @click="switchTab('price')"
            >
              💰 Price Analysis
            </button>
            <button
              :class="['tab-btn', { active: analyticsTab === 'shipping' }]"
              @click="switchTab('shipping')"
            >
              🚚 Shipping Fee
            </button>
          </div>
          <div class="period-selector">
            <div class="period-controls">
              <div class="select-group">
                <label for="shopee-month-select">Month</label>
                <select
                  id="shopee-month-select"
                  v-model="selectedMonth"
                  class="period-select"
                >
                  <option v-for="m in months" :key="m.value" :value="m.value">
                    {{ m.label }}
                  </option>
                </select>
              </div>
              <div class="select-group">
                <label for="shopee-year-select">Year</label>
                <select
                  id="shopee-year-select"
                  v-model="selectedYear"
                  class="period-select"
                >
                  <option v-for="y in years" :key="y" :value="y">
                    {{ y }}
                  </option>
                </select>
              </div>
            </div>
            <div class="sync-status">
              <span v-if="syncStatus?.synced" class="status-badge synced"
                >✓ Synced ({{ syncStatus.total_orders }} orders)</span
              >
              <span v-else class="status-badge not-synced">⚠ Not synced</span>
              <span v-if="syncStatus?.synced_at" class="sync-date">{{
                formatDate(syncStatus.synced_at)
              }}</span>
            </div>
          </div>
          <!-- Progress bar for background sync -->
          <div v-if="syncing && jobProgress" class="sync-progress-container">
            <div class="sync-progress-info">
              <span class="progress-label">{{ syncProgressMessage }}</span>
              <span class="progress-percent">{{ syncProgressPercent }}%</span>
            </div>
            <div class="sync-progress-bar">
              <div
                class="sync-progress-fill"
                :style="{ width: syncProgressPercent + '%' }"
              ></div>
            </div>
            <div v-if="jobProgress.total_items > 0" class="progress-items">
              {{ jobProgress.processed_items }} /
              {{ jobProgress.total_items }} items
            </div>
          </div>
          <div class="action-bar">
            <div class="action-left">
              <button
                class="btn-primary"
                @click="handleSync(false)"
                :disabled="!canSync || syncing"
              >
                <span v-if="syncing">⏳ Syncing...</span
                ><span v-else>🔄 Sync Escrow Data</span>
              </button>
              <button
                v-if="syncStatus?.synced"
                class="btn-secondary"
                @click="handleSync(true)"
                :disabled="syncing"
              >
                ♻️ Force Resync
              </button>
              <button
                v-if="syncStatus?.synced"
                class="btn-secondary"
                @click="handleAnalyze"
                :disabled="loading"
              >
                <span v-if="loading">⏳ Analyzing...</span
                ><span v-else
                  >📈 Analyze
                  {{ analyticsTab === "price" ? "Prices" : "Shipping" }}</span
                >
              </button>
              <button
                v-if="syncStatus?.synced"
                class="btn-danger"
                @click="handleDeleteSync"
                :disabled="loading || syncing"
              >
                <span aria-hidden="true">🗑️</span> Delete Sync
              </button>
            </div>
            <div class="action-right">
              <button
                v-if="analyticsTab === 'price' && reconciliationResult"
                class="btn-export"
                @click="exportPriceToCSV"
              >
                📥 Export CSV
              </button>
              <button
                v-if="analyticsTab === 'shipping' && shippingFeeResult"
                class="btn-export"
                @click="exportShippingToCSV"
              >
                📥 Export CSV
              </button>
            </div>
          </div>
          <template v-if="analyticsTab === 'price'">
            <AnalyticsSummaryCards
              v-if="reconciliationResult"
              :summary="reconciliationResult.summary"
            />
            <AnalyticsResultsTable
              v-if="reconciliationResult"
              :skuGroups="reconciliationResult.sku_groups"
            />
          </template>
          <template v-if="analyticsTab === 'shipping'">
            <ShippingFeeSummary
              v-if="shippingFeeResult"
              :summary="shippingFeeResult.summary"
            />
            <ShippingFeeTable
              v-if="shippingFeeResult"
              :orders="shippingFeeResult.orders"
            />
          </template>
          <div v-if="!hasData && !loading && !syncing" class="empty-state">
            <div class="empty-icon">📭</div>
            <h3>No Data Available</h3>
            <p>Select a period and sync escrow data to start analysis</p>
          </div>
          <div v-if="loading || syncing" class="loading-state">
            <div class="spinner"></div>
            <p>
              {{ syncing ? "Syncing escrow data..." : "Loading analysis..." }}
            </p>
          </div>
          <AnalyticsSettingsModal
            v-if="showSettings"
            :settings="settings"
            @close="showSettings = false"
            @save="handleSaveSettings"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useAnalytics } from "@/composables/useAnalytics";
import {
  MONTHS,
  getAvailableYears,
  formatAnalyticsDate,
  downloadCSV,
} from "@/utils/analyticsHelpers";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import AnalyticsSettingsModal from "@/components/analytics/AnalyticsSettingsModal.vue";
import AnalyticsSummaryCards from "@/components/analytics/AnalyticsSummaryCards.vue";
import AnalyticsResultsTable from "@/components/analytics/AnalyticsResultsTable.vue";
import ShippingFeeSummary from "@/components/analytics/ShippingFeeSummary.vue";
import ShippingFeeTable from "@/components/analytics/ShippingFeeTable.vue";

const router = useRouter();
const route = useRoute();
const uiStore = useUIStore();
uiStore.setActiveTab("analytics");

const handleTabChange = (tab: string): void => {
  uiStore.setActiveTab(tab as any);
  if (tab !== "report" && tab !== "analytics") router.push("/");
};
const handlePlatformChange = (platform: string): void =>
  uiStore.setActivePlatform(platform as any);

const {
  loading,
  syncing,
  syncStatus,
  settings,
  reconciliationResult,
  shippingFeeResult,
  selectedMonth,
  selectedYear,
  canSync,
  jobProgress,
  syncProgressPercent,
  syncProgressMessage,
  fetchSyncStatus,
  saveSettings,
  syncEscrowData,
  deleteSyncData,
  fetchReconciliation,
  fetchShippingFeeAnalysis,
  initialize,
  cleanup,
} = useAnalytics();

const showSettings = ref(false);
const analyticsTab = ref<"price" | "shipping">("price");
const hasData = computed(() =>
  analyticsTab.value === "price"
    ? !!reconciliationResult.value
    : !!shippingFeeResult.value,
);
const months = MONTHS;
const years = computed(() => getAvailableYears());
const formatDate = formatAnalyticsDate;

async function handleSync(force: boolean) {
  await syncEscrowData(force);
  if (syncStatus.value?.synced) {
    analyticsTab.value === "price"
      ? await fetchReconciliation()
      : await fetchShippingFeeAnalysis();
  }
}

async function handleDeleteSync() {
  if (confirm("Are you sure you want to delete all sync data for this period?"))
    await deleteSyncData();
}

async function handleAnalyze() {
  analyticsTab.value === "price"
    ? await fetchReconciliation()
    : await fetchShippingFeeAnalysis();
}

function updateURLQuery(tabValue: string): void {
  router
    .push({ path: route.path, query: { ...route.query, tab: tabValue } })
    .catch(() => {});
}

function switchTab(tab: "price" | "shipping") {
  analyticsTab.value = tab;
  updateURLQuery(tab);
  if (syncStatus.value?.synced) {
    if (tab === "price" && !reconciliationResult.value) fetchReconciliation();
    else if (tab === "shipping" && !shippingFeeResult.value)
      fetchShippingFeeAnalysis();
  }
}

async function handleSaveSettings(newSettings: typeof settings.value) {
  await saveSettings(newSettings);
  showSettings.value = false;
}

function exportPriceToCSV() {
  if (!reconciliationResult.value) return;
  const rows = [
    ["Status", "SKU", "Item Name", "Inventory Price", "Expected Income", "Qty"],
  ];
  reconciliationResult.value.sku_groups.forEach((sku) =>
    rows.push([
      sku.status,
      sku.model_sku || sku.sku,
      `"${sku.item_name.replace(/"/g, '""')}"`,
      String(sku.inventory_price || 0),
      String(sku.expected_income || 0),
      String(sku.total_transactions),
    ]),
  );
  downloadCSV(
    rows,
    `shopee-price-${selectedYear.value}-${selectedMonth.value + 1}.csv`,
  );
}

function exportShippingToCSV() {
  if (!shippingFeeResult.value) return;
  const rows = [
    [
      "Order Date",
      "Order SN",
      "Buyer Paid",
      "Actual Fee",
      "Rebate",
      "Difference",
      "Payment",
    ],
  ];
  shippingFeeResult.value.orders.forEach((o) =>
    rows.push([
      o.order_date || "",
      o.order_sn,
      String(o.buyer_paid),
      String(o.actual_fee),
      String(o.shopee_rebate),
      String(o.difference),
      o.payment_method || "",
    ]),
  );
  downloadCSV(
    rows,
    `shopee-shipping-${selectedYear.value}-${selectedMonth.value + 1}.csv`,
  );
}

watch([selectedMonth, selectedYear], async () => {
  await fetchSyncStatus();
  reconciliationResult.value = null;
  shippingFeeResult.value = null;
});

watch(
  () => route.query.tab,
  (newTab) => {
    if (
      (newTab === "price" || newTab === "shipping") &&
      analyticsTab.value !== newTab
    ) {
      analyticsTab.value = newTab;
      if (syncStatus.value?.synced) {
        newTab === "price" && !reconciliationResult.value
          ? fetchReconciliation()
          : newTab === "shipping" &&
            !shippingFeeResult.value &&
            fetchShippingFeeAnalysis();
      }
    }
  },
);

onMounted(async () => {
  const appStore = useAppStore();

  // Properly wait for app initialization (like Inventory.vue)
  if (appStore.connectionStatus === "connecting") {
    await appStore.initializeApp();
  } else if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
    // Wait a bit for connection
    await new Promise<void>((resolve) => {
      const check = () => {
        if (appStore.connectionStatus === "connected") resolve();
        else if (appStore.connectionStatus === "error") resolve();
        else setTimeout(check, 100);
      };
      setTimeout(check, 100);
    });
  }

  const tabFromQuery = route.query.tab as string;
  if (tabFromQuery === "price" || tabFromQuery === "shipping")
    analyticsTab.value = tabFromQuery;
  else updateURLQuery(analyticsTab.value);

  await initialize();
  if (syncStatus.value?.synced) {
    analyticsTab.value === "price"
      ? await fetchReconciliation()
      : await fetchShippingFeeAnalysis();
  }
});

onUnmounted(() => {
  cleanup();
});
</script>

<style scoped>
@import "./ShopeeAnalytics.styles.css";

.dashboard {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
.main-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.main-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

/* Progress bar styles */
.sync-progress-container {
  background: var(--bg-secondary, #f5f5f5);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
}
.sync-progress-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}
.progress-label {
  color: var(--text-secondary, #666);
}
.progress-percent {
  font-weight: 600;
  color: var(--primary-color, #3b82f6);
}
.sync-progress-bar {
  height: 8px;
  background: var(--border-color, #e0e0e0);
  border-radius: 4px;
  overflow: hidden;
}
.sync-progress-fill {
  height: 100%;
  background: var(--primary-color, #3b82f6);
  border-radius: 4px;
  transition: width 0.3s ease;
}
.progress-items {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted, #999);
  text-align: right;
}
</style>

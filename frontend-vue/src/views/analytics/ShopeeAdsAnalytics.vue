<template>
  <div class="dashboard">
    <div class="main-layout">
      <!-- Left Sidebar -->
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="uiStore.activeTab"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />

      <!-- Main Content -->
      <main class="main-content">
        <div class="analytics-page">
          <!-- Header -->
          <div class="analytics-header">
            <div class="header-left">
              <h1><span aria-hidden="true">🛒</span> Shopee Ads Analytics</h1>
              <p class="subtitle">
                Track and analyze your advertising performance
              </p>
            </div>
          </div>

          <!-- Tab Navigation -->
          <div class="tab-navigation">
            <button
              :class="['tab-btn', { active: activeTab === 'dashboard' }]"
              @click="switchTab('dashboard')"
              type="button"
            >
              <span aria-hidden="true">📊</span> Dashboard
            </button>
            <button
              :class="['tab-btn', { active: activeTab === 'data' }]"
              @click="switchTab('data')"
              type="button"
            >
              <span aria-hidden="true">📋</span> Product Data
            </button>
            <button
              :class="['tab-btn', { active: activeTab === 'upload' }]"
              @click="switchTab('upload')"
              type="button"
            >
              <span aria-hidden="true">📤</span> Upload
            </button>
            <button
              :class="['tab-btn', { active: activeTab === 'insights' }]"
              @click="switchTab('insights')"
              type="button"
            >
              <span aria-hidden="true">🤖</span> AI Insights
            </button>
          </div>

          <!-- Dashboard Tab -->
          <template v-if="activeTab === 'dashboard'">
            <ShopeeAdsDashboard
              :dashboard="dashboardSummary"
              :loading="loading"
              :format-currency="formatCurrency"
              :format-number="formatNumber"
              :format-roas="formatRoas"
              @refresh="handleDashboardRefresh"
            />
          </template>

          <!-- Data Tab -->
          <template v-if="activeTab === 'data'">
            <ShopeeAdsDataTable
              :data="productData"
              :loading="loading"
              :total-records="totalRecords"
              :format-currency="formatCurrency"
              :format-number="formatNumber"
              :format-roas="formatRoas"
              @refresh="handleDataRefresh"
            />
          </template>

          <!-- Upload Tab -->
          <template v-if="activeTab === 'upload'">
            <ShopeeAdsUpload
              ref="uploadRef"
              :uploading="uploading"
              :history="uploadHistory"
              @upload="handleUpload"
              @fetch-history="fetchUploadHistory"
            />
          </template>

          <!-- AI Insights Tab -->
          <template v-if="activeTab === 'insights'">
            <AdsReportViewer platform="shopee" />
          </template>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useUnifiedHeader } from "@/composables/useUnifiedHeader";
import { useShopeeAdsAnalytics } from "@/composables/useShopeeAdsAnalytics";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import ShopeeAdsDashboard from "@/components/analytics/ShopeeAdsDashboard.vue";
import ShopeeAdsDataTable from "@/components/analytics/ShopeeAdsDataTable.vue";
import ShopeeAdsUpload from "@/components/analytics/ShopeeAdsUpload.vue";
import AdsReportViewer from "@/components/analytics/AdsReportViewer.vue";

const router = useRouter();
const route = useRoute();
const uiStore = useUIStore();
const appStore = useAppStore();
const { setConnectionStatus } = useUnifiedHeader();

const {
  loading,
  uploading,
  dashboardSummary,
  productData,
  uploadHistory,
  totalRecords,
  fetchDashboard,
  fetchProductData,
  fetchUploadHistory,
  uploadFile,
  formatCurrency,
  formatNumber,
  formatRoas,
} = useShopeeAdsAnalytics();

// Local state
const activeTab = ref<"dashboard" | "data" | "upload" | "insights">(
  "dashboard",
);
const uploadRef = ref<InstanceType<typeof ShopeeAdsUpload> | null>(null);

type ValidTab = "dashboard" | "data" | "upload" | "insights";
const validTabs: ValidTab[] = ["dashboard", "data", "upload", "insights"];

// URL Query Sync
function updateURLQuery(tabValue: string): void {
  const query = { ...route.query, tab: tabValue };
  router.push({ path: route.path, query }).catch(() => {});
}

function switchTab(tab: ValidTab) {
  activeTab.value = tab;
  updateURLQuery(tab);
}

// Watch route query for tab changes
watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab && validTabs.includes(newTab as ValidTab)) {
      if (activeTab.value !== newTab) {
        activeTab.value = newTab as ValidTab;
      }
    }
  },
);

// Navigation handlers
type TabType =
  | "logs"
  | "settings"
  | "operation"
  | "product-management"
  | "order-management"
  | "inventory"
  | "script-monitor"
  | "analytics";
type PlatformType = "shopee" | "lazada" | "tiktok";

function handleTabChange(tab: string) {
  uiStore.setActiveTab(tab as TabType);
  if (tab !== "analytics") {
    router.push(`/${tab}`);
  }
}

function handlePlatformChange(platform: string) {
  uiStore.setActivePlatform(platform as PlatformType);
}

// Data handlers
async function handleDashboardRefresh() {
  await fetchDashboard();
}

async function handleDataRefresh(options: {
  orderBy: string;
  orderDir: string;
}) {
  await fetchProductData({ limit: 100, ...options });
}

async function handleUpload(file: File, periodLabel: string) {
  const result = await uploadFile(file, periodLabel);
  if (result) {
    uploadRef.value?.setUploadResult({
      success: true,
      message: "Upload successful!",
      details: {
        processed: result.inserted_rows || 0,
        skipped: result.skipped_rows || 0,
      },
    });
    await fetchDashboard();
  }
}

// Lifecycle
onMounted(async () => {
  uiStore.setActiveTab("analytics");

  // Initialize app connection if not already connected
  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }

  // Read tab from URL query
  const tabFromQuery = route.query.tab as string;
  if (tabFromQuery && validTabs.includes(tabFromQuery as ValidTab)) {
    activeTab.value = tabFromQuery as ValidTab;
  } else {
    // Set default tab to URL
    updateURLQuery(activeTab.value);
  }

  await Promise.all([
    fetchDashboard(),
    fetchProductData({ limit: 100, orderBy: "revenue", orderDir: "desc" }),
    fetchUploadHistory(),
  ]);
});

// Sync connection status to unified header
watch(
  () => appStore.connectionStatus,
  (newStatus) => {
    setConnectionStatus(newStatus);
  },
  { immediate: true },
);
</script>

<style scoped>
@import "./ShopeeAdsAnalytics.styles.css";

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
</style>

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
              <h1>📈 TikTok Ads Analytics</h1>
              <p class="subtitle">Creative Performance & ML Insights</p>
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
              <span aria-hidden="true">📋</span> Creative Data
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
            <TiktokAdsDashboard
              :dashboard="dashboard"
              :loading="loading"
              :format-currency="formatCurrency"
              :format-number="formatNumber"
              :format-roi="formatRoi"
            />
          </template>

          <!-- Data Tab -->
          <template v-if="activeTab === 'data'">
            <TiktokAdsDataTable
              :data="creativeData"
              :loading="loading"
              :total-records="totalRecords"
              :format-currency="formatCurrency"
              :format-number="formatNumber"
              :format-roi="formatRoi"
              @refresh="handleRefreshDataWithSort"
            />
          </template>

          <!-- Upload Tab -->
          <template v-if="activeTab === 'upload'">
            <TiktokAdsUpload
              :uploading="uploading"
              :uploads="uploadHistory"
              @upload="handleUpload"
            />
          </template>

          <!-- AI Insights Tab -->
          <template v-if="activeTab === 'insights'">
            <AdsReportViewer platform="tiktok" />
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
import { useTiktokAdsAnalytics } from "@/composables/useTiktokAdsAnalytics";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import TiktokAdsDashboard from "@/components/analytics/TiktokAdsDashboard.vue";
import TiktokAdsDataTable from "@/components/analytics/TiktokAdsDataTable.vue";
import TiktokAdsUpload from "@/components/analytics/TiktokAdsUpload.vue";
import AdsReportViewer from "@/components/analytics/AdsReportViewer.vue";

const router = useRouter();
const route = useRoute();
const uiStore = useUIStore();
const appStore = useAppStore();
const { setConnectionStatus } = useUnifiedHeader();

const {
  loading,
  uploading,
  dashboard,
  creativeData,
  uploadHistory,
  totalRecords,
  fetchDashboard,
  fetchCreativeData,
  fetchUploadHistory,
  uploadFile,
  formatCurrency,
  formatNumber,
  formatRoi,
} = useTiktokAdsAnalytics();

// Local state
const activeTab = ref<"dashboard" | "data" | "upload" | "insights">(
  "dashboard",
);

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
async function handleRefreshDataWithSort(options: {
  orderBy: string;
  orderDir: string;
}) {
  await fetchCreativeData({ limit: 100, ...options });
}

async function handleUpload(file: File, mode: "skip" | "update") {
  const result = await uploadFile(file, mode);
  if (result?.success) {
    // Switch to dashboard after successful upload
    activeTab.value = "dashboard";
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

  // Only fetch data for the active tab (lazy loading)
  if (activeTab.value === "dashboard") {
    await fetchDashboard();
  } else if (activeTab.value === "data") {
    await fetchCreativeData({
      limit: 100,
      orderBy: "revenue",
      orderDir: "desc",
    });
  } else if (activeTab.value === "upload") {
    await fetchUploadHistory();
  }
  // insights tab loads its own data via AdsReportViewer
});

// Sync connection status to unified header
watch(
  () => appStore.connectionStatus,
  (newStatus) => {
    setConnectionStatus(newStatus);
  },
  { immediate: true },
);

// Watch for tab changes and load data lazily
watch(activeTab, async (newTab) => {
  if (newTab === "dashboard" && !dashboard.value) {
    await fetchDashboard();
  } else if (newTab === "data" && creativeData.value.length === 0) {
    await fetchCreativeData({
      limit: 100,
      orderBy: "revenue",
      orderDir: "desc",
    });
  } else if (newTab === "upload" && uploadHistory.value.length === 0) {
    await fetchUploadHistory();
  }
});
</script>

<style scoped>
@import "./TiktokAdsAnalytics.styles.css";

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

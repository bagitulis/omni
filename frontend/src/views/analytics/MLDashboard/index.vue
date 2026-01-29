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
        <div class="ml-dashboard-page">
          <!-- Header -->
          <div class="page-header">
            <div class="header-left">
              <h1><span aria-hidden="true">🧠</span> ML Analytics Dashboard</h1>
              <p class="subtitle">AI-Powered Product Intelligence</p>
            </div>
            <div class="header-right">
              <button
                @click="refreshData"
                :disabled="loading"
                class="btn-refresh"
                type="button"
              >
                <span v-if="loading" aria-hidden="true">⏳</span>
                <span v-else aria-hidden="true">🔄</span>
                {{ loading ? "Loading..." : "Refresh" }}
              </button>
            </div>
          </div>

          <!-- Dashboard Grid -->
          <div class="dashboard-grid">
            <!-- Left Column: Health Card + Alerts -->
            <div class="left-column">
              <PortfolioHealthCard
                :health="portfolioHealth"
                :loading="loading"
                :format-currency="formatCurrency"
              />
              <AlertsPanel :alerts="alerts" />
            </div>

            <!-- Right Column: Product Table -->
            <div class="right-column">
              <ProductScoreTable
                :products="products"
                :loading="loading"
                :has-more="productsMeta.has_more"
                :format-currency="formatCurrency"
                @select="handleProductSelect"
                @load-more="loadMoreProducts"
                @sort="handleSort"
              />
            </div>
          </div>

          <!-- Product Detail Modal -->
          <ProductDetailModal
            :product="selectedProduct"
            :format-currency="formatCurrency"
            @close="selectedProduct = null"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import {
  useMLAnalytics,
  type MLProductAnalysis,
} from "@/composables/useMLAnalytics";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import {
  PortfolioHealthCard,
  ProductScoreTable,
  AlertsPanel,
} from "@/components/analytics/ml";
import ProductDetailModal from "./ProductDetailModal.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();

const {
  loading,
  portfolioHealth,
  products,
  productsMeta,
  alerts,
  fetchPortfolioHealth,
  fetchProducts,
  fetchAlerts,
  loadMoreProducts,
  formatCurrency,
} = useMLAnalytics();

const selectedProduct = ref<MLProductAnalysis | null>(null);

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

async function refreshData() {
  await Promise.all([
    fetchPortfolioHealth("tiktok"),
    fetchProducts({ limit: 20, sortBy: "unified_score", sortDir: "desc" }),
    fetchAlerts(),
  ]);
}

function handleProductSelect(product: MLProductAnalysis) {
  selectedProduct.value = product;
}

function handleSort(sortBy: string, sortDir: string) {
  fetchProducts({ limit: 20, sortBy, sortDir });
}

onMounted(async () => {
  uiStore.setActiveTab("analytics");

  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }

  await refreshData();
});
</script>

<style scoped>
@import "./MLDashboard.styles.css";
</style>

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
        <div class="classification-page">
          <!-- Header -->
          <div class="page-header">
            <div class="header-left">
              <h1><span aria-hidden="true">📊</span> Product Classification</h1>
              <p class="subtitle">
                Actionable recommendations for your products
              </p>
            </div>
            <div class="header-right">
              <button
                @click="fetchClassifiedProducts"
                :disabled="loading"
                class="btn-refresh"
                type="button"
              >
                {{ loading ? "Loading..." : "Refresh" }}
              </button>
            </div>
          </div>

          <!-- Summary Cards -->
          <div class="summary-cards">
            <div class="summary-card scale-up">
              <span class="count">{{ actionCounts.scale_up }}</span>
              <span class="label">Scale Up</span>
            </div>
            <div class="summary-card maintain">
              <span class="count">{{ actionCounts.maintain }}</span>
              <span class="label">Maintain</span>
            </div>
            <div class="summary-card reduce">
              <span class="count">{{ actionCounts.reduce }}</span>
              <span class="label">Reduce</span>
            </div>
            <div class="summary-card stop">
              <span class="count">{{ actionCounts.stop }}</span>
              <span class="label">Stop</span>
            </div>
          </div>

          <!-- Tabs -->
          <div class="tabs">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              :class="['tab', { active: activeTab === tab.key }]"
              @click="activeTab = tab.key"
              type="button"
            >
              {{ tab.label }}
              <span class="tab-count">{{ getTabCount(tab.key) }}</span>
            </button>
          </div>

          <!-- Products List -->
          <div v-if="loading" class="loading-state">
            <p>Loading products...</p>
          </div>

          <div v-else-if="activeProducts.length === 0" class="empty-state">
            <p>No products in this category</p>
          </div>

          <div v-else class="products-grid">
            <div
              v-for="product in activeProducts"
              :key="product.product_id"
              class="product-card"
            >
              <div class="card-header" :class="activeTab">
                <span class="action-badge">{{ product.action }}</span>
                <span class="source-badge">{{ product.source }}</span>
              </div>
              <div class="card-body">
                <h3>{{ product.product_name }}</h3>
                <div class="metrics">
                  <div class="metric">
                    <span class="label">Cost</span>
                    <span class="value cost">{{
                      formatCurrency(product.total_cost)
                    }}</span>
                  </div>
                  <div class="metric">
                    <span class="label">Revenue</span>
                    <span class="value revenue">{{
                      formatCurrency(product.total_revenue)
                    }}</span>
                  </div>
                  <div class="metric">
                    <span class="label">ROAS</span>
                    <span class="value roas">{{
                      formatRoas(product.roas)
                    }}</span>
                  </div>
                </div>
                <div class="status-row">
                  <span
                    class="trend"
                    :class="product.trend_direction.toLowerCase()"
                  >
                    {{ product.trend_direction }}
                  </span>
                  <span class="confidence">{{ product.confidence_level }}</span>
                </div>
                <p class="action-label">{{ product.action_label }}</p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useUnifiedAnalytics } from "@/composables/useUnifiedAnalytics";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();

const {
  loading,
  classifiedProducts,
  actionCounts,
  fetchClassifiedProducts,
  fetchKPI,
  formatCurrency,
  formatRoas,
} = useUnifiedAnalytics();

const tabs = [
  { key: "scale_up", label: "Scale Up" },
  { key: "maintain", label: "Maintain" },
  { key: "reduce", label: "Reduce" },
  { key: "stop", label: "Stop" },
];

const activeTab = ref<string>("scale_up");

const activeProducts = computed(() => {
  if (!classifiedProducts.value) return [];
  return (
    classifiedProducts.value[
      activeTab.value as keyof typeof classifiedProducts.value
    ] || []
  );
});

function getTabCount(key: string): number {
  return actionCounts.value[key as keyof typeof actionCounts.value] || 0;
}

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
  if (tab !== "analytics") router.push(`/${tab}`);
}

function handlePlatformChange(platform: string) {
  uiStore.setActivePlatform(platform as PlatformType);
}

onMounted(async () => {
  uiStore.setActiveTab("analytics");
  if (appStore.connectionStatus !== "connected") appStore.initializeApp();
  await Promise.all([fetchClassifiedProducts(), fetchKPI()]);
});
</script>

<style scoped>
@import "./ProductClassification.styles.css";
</style>

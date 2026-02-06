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
        <div class="hub-page">
          <!-- Header -->
          <div class="page-header">
            <div class="header-left">
              <h1><span aria-hidden="true">📈</span> Analytics Hub</h1>
              <p class="subtitle">Unified insights across all platforms</p>
            </div>
            <div class="header-right">
              <button
                @click="refreshCache"
                :disabled="loading"
                class="btn-refresh"
                type="button"
              >
                {{ loading ? "Refreshing..." : "Refresh Data" }}
              </button>
            </div>
          </div>

          <!-- KPI Cards -->
          <div class="kpi-grid">
            <div class="kpi-card">
              <span class="kpi-icon">📦</span>
              <div class="kpi-data">
                <span class="kpi-value">{{ formatNumber(totalProducts) }}</span>
                <span class="kpi-label">Total Products</span>
              </div>
            </div>
            <div class="kpi-card">
              <span class="kpi-icon">💰</span>
              <div class="kpi-data">
                <span class="kpi-value">{{
                  formatCurrency(summary?.combined?.total_revenue || 0)
                }}</span>
                <span class="kpi-label">Total Revenue</span>
              </div>
            </div>
            <div class="kpi-card">
              <span class="kpi-icon">📊</span>
              <div class="kpi-data">
                <span class="kpi-value">{{ formatRoas(avgRoas) }}</span>
                <span class="kpi-label">Average ROAS</span>
              </div>
            </div>
            <div class="kpi-card">
              <span class="kpi-icon">🎯</span>
              <div class="kpi-data">
                <span class="kpi-value">{{
                  formatNumber(summary?.combined?.total_orders || 0)
                }}</span>
                <span class="kpi-label">Total Orders</span>
              </div>
            </div>
          </div>

          <!-- Quick Actions -->
          <div class="section">
            <h2>Quick Actions</h2>
            <div class="actions-grid">
              <router-link to="/analytics/simulator" class="action-card">
                <span class="action-icon">🎯</span>
                <h3>Budget Simulator</h3>
                <p>Predict ROAS and optimize ad spend</p>
              </router-link>
              <router-link to="/analytics/classification" class="action-card">
                <span class="action-icon">📊</span>
                <h3>Product Classification</h3>
                <p>Stop, Scale, Maintain recommendations</p>
              </router-link>
              <router-link to="/analytics/ml" class="action-card">
                <span class="action-icon">🧠</span>
                <h3>ML Dashboard</h3>
                <p>AI-powered product intelligence</p>
              </router-link>
              <router-link to="/analytics/ai-reports" class="action-card">
                <span class="action-icon">🤖</span>
                <h3>AI Reports</h3>
                <p>ML-generated analytics reports</p>
              </router-link>
            </div>
          </div>

          <!-- Platform Comparison -->
          <div class="section">
            <h2>Platform Comparison</h2>
            <div class="platforms-grid">
              <div class="platform-card tiktok">
                <div class="platform-header">
                  <span class="platform-icon">🎵</span>
                  <span class="platform-name">TikTok</span>
                </div>
                <div class="platform-stats">
                  <div class="stat">
                    <span class="label">Revenue</span>
                    <span class="value">{{
                      formatCurrency(summary?.tiktok?.total_revenue || 0)
                    }}</span>
                  </div>
                  <div class="stat">
                    <span class="label">Cost</span>
                    <span class="value">{{
                      formatCurrency(summary?.tiktok?.total_cost || 0)
                    }}</span>
                  </div>
                  <div class="stat">
                    <span class="label">ROAS</span>
                    <span class="value">{{
                      formatRoas(summary?.tiktok?.avg_roas || 0)
                    }}</span>
                  </div>
                </div>
                <router-link to="/analytics/tiktok-ads" class="platform-link">
                  View Details
                </router-link>
              </div>

              <div class="platform-card shopee">
                <div class="platform-header">
                  <span class="platform-icon">🛒</span>
                  <span class="platform-name">Shopee</span>
                </div>
                <div class="platform-stats">
                  <div class="stat">
                    <span class="label">Revenue</span>
                    <span class="value">{{
                      formatCurrency(summary?.shopee?.total_revenue || 0)
                    }}</span>
                  </div>
                  <div class="stat">
                    <span class="label">Cost</span>
                    <span class="value">{{
                      formatCurrency(summary?.shopee?.total_cost || 0)
                    }}</span>
                  </div>
                  <div class="stat">
                    <span class="label">ROAS</span>
                    <span class="value">{{
                      formatRoas(summary?.shopee?.avg_roas || 0)
                    }}</span>
                  </div>
                </div>
                <router-link to="/analytics/shopee-ads" class="platform-link">
                  View Details
                </router-link>
              </div>
            </div>
          </div>

          <!-- Action Summary -->
          <div class="section">
            <h2>Action Summary</h2>
            <div class="action-summary">
              <div class="action-item scale-up">
                <span class="count">{{ actionCounts.scale_up }}</span>
                <span class="label">Scale Up</span>
              </div>
              <div class="action-item maintain">
                <span class="count">{{ actionCounts.maintain }}</span>
                <span class="label">Maintain</span>
              </div>
              <div class="action-item reduce">
                <span class="count">{{ actionCounts.reduce }}</span>
                <span class="label">Reduce</span>
              </div>
              <div class="action-item stop">
                <span class="count">{{ actionCounts.stop }}</span>
                <span class="label">Stop</span>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useUnifiedHeader } from "@/composables/useUnifiedHeader";
import { useUnifiedAnalytics } from "@/composables/useUnifiedAnalytics";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();
const { setConnectionStatus } = useUnifiedHeader();

const {
  loading,
  summary,
  totalProducts,
  avgRoas,
  actionCounts,
  fetchAll,
  refreshCache,
  formatCurrency,
  formatRoas,
  formatNumber,
} = useUnifiedAnalytics();

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
  await fetchAll();
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
@import "./AnalyticsHub.styles.css";
</style>

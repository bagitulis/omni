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
          <div
            v-if="selectedProduct"
            class="modal-overlay"
            @click="selectedProduct = null"
          >
            <div class="modal-content" @click.stop>
              <div class="modal-header">
                <h2>
                  {{
                    selectedProduct.product_name || selectedProduct.product_id
                  }}
                </h2>
                <button
                  @click="selectedProduct = null"
                  class="btn-close"
                  type="button"
                >
                  ✕
                </button>
              </div>
              <div class="modal-body">
                <div class="detail-grid">
                  <!-- Score Section -->
                  <div class="detail-section">
                    <h4>Unified Score</h4>
                    <div
                      class="score-large"
                      :class="scoreClass(selectedProduct.unified_score)"
                    >
                      {{ selectedProduct.unified_score.toFixed(0) }}
                    </div>
                    <div class="score-breakdown">
                      <div class="breakdown-item">
                        <span>ROAS Score</span>
                        <span>{{ selectedProduct.roas_score.toFixed(0) }}</span>
                      </div>
                      <div class="breakdown-item">
                        <span>Trend Score</span>
                        <span>{{
                          selectedProduct.trend_score.toFixed(0)
                        }}</span>
                      </div>
                      <div class="breakdown-item">
                        <span>Momentum Score</span>
                        <span>{{
                          selectedProduct.momentum_score.toFixed(0)
                        }}</span>
                      </div>
                      <div class="breakdown-item">
                        <span>Volatility Score</span>
                        <span>{{
                          selectedProduct.volatility_score.toFixed(0)
                        }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- Financial Section -->
                  <div class="detail-section">
                    <h4>Financial Metrics</h4>
                    <div class="metric-row">
                      <span>Total Cost</span>
                      <span class="cost">{{
                        formatCurrency(selectedProduct.total_cost)
                      }}</span>
                    </div>
                    <div class="metric-row">
                      <span>Total Revenue</span>
                      <span class="revenue">{{
                        formatCurrency(selectedProduct.total_revenue)
                      }}</span>
                    </div>
                    <div class="metric-row">
                      <span>Total Profit</span>
                      <span
                        :class="
                          selectedProduct.total_profit >= 0 ? 'revenue' : 'cost'
                        "
                      >
                        {{ formatCurrency(selectedProduct.total_profit) }}
                      </span>
                    </div>
                    <div class="metric-row">
                      <span>ROAS</span>
                      <span class="roas"
                        >{{ selectedProduct.roas.toFixed(2) }}x</span
                      >
                    </div>
                  </div>

                  <!-- Recommendation Section -->
                  <div class="detail-section full-width">
                    <h4>Recommendation</h4>
                    <div class="recommendation-box">
                      <ActionBadge :action="selectedProduct.action" />
                      <p class="rec-label">
                        {{ selectedProduct.action_label }}
                      </p>
                      <p
                        class="rec-budget"
                        v-if="selectedProduct.budget_change_pct !== 0"
                      >
                        Budget Change:
                        <strong
                          :class="
                            selectedProduct.budget_change_pct > 0
                              ? 'positive'
                              : 'negative'
                          "
                        >
                          {{ selectedProduct.budget_change_pct > 0 ? "+" : ""
                          }}{{ selectedProduct.budget_change_pct }}%
                        </strong>
                      </p>
                      <p class="rec-confidence">
                        Confidence:
                        <strong>{{ selectedProduct.confidence_level }}</strong>
                        ({{
                          (selectedProduct.success_probability * 100).toFixed(
                            0,
                          )
                        }}% success probability)
                      </p>
                    </div>
                  </div>

                  <!-- Status Section -->
                  <div class="detail-section full-width">
                    <h4>Status Indicators</h4>
                    <div class="status-row">
                      <div class="status-item">
                        <span class="status-label">Category</span>
                        <span
                          class="category-badge"
                          :class="
                            'cat-' + selectedProduct.category.toLowerCase()
                          "
                        >
                          {{ selectedProduct.category }}
                        </span>
                      </div>
                      <div class="status-item">
                        <span class="status-label">Trend</span>
                        <span
                          class="trend-badge"
                          :class="
                            'trend-' +
                            selectedProduct.trend_direction.toLowerCase()
                          "
                        >
                          {{ selectedProduct.trend_direction }}
                        </span>
                      </div>
                      <div class="status-item">
                        <span class="status-label">Fatigue</span>
                        <span
                          class="fatigue-badge"
                          :class="
                            'fatigue-' +
                            selectedProduct.fatigue_status.toLowerCase()
                          "
                        >
                          {{ selectedProduct.fatigue_status }}
                        </span>
                      </div>
                      <div class="status-item">
                        <span class="status-label">Churn Risk</span>
                        <span
                          >{{
                            selectedProduct.churn_risk_score.toFixed(0)
                          }}%</span
                        >
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
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
  ActionBadge,
} from "@/components/analytics/ml";

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

// Navigation
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

// Data actions
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

function scoreClass(score: number): string {
  if (score >= 70) return "score-excellent";
  if (score >= 55) return "score-good";
  if (score >= 40) return "score-fair";
  return "score-poor";
}

// Lifecycle
onMounted(async () => {
  uiStore.setActiveTab("analytics");

  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }

  await refreshData();
});
</script>

<style scoped>
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
  background: #f3f4f6;
}

.ml-dashboard-page {
  padding: 24px;
  max-width: 1600px;
  margin: 0 auto;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left h1 {
  margin: 0;
  font-size: 28px;
  color: #1f2937;
}

.subtitle {
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 14px;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-refresh:hover:not(:disabled) {
  background: #2563eb;
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Dashboard Grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 24px;
}

.left-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.right-column {
  min-width: 0;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: white;
  border-radius: 16px;
  max-width: 700px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  color: #1f2937;
}

.btn-close {
  padding: 8px 12px;
  background: #f3f4f6;
  border: none;
  border-radius: 8px;
  font-size: 18px;
  cursor: pointer;
}

.btn-close:hover {
  background: #e5e7eb;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.detail-section {
  background: #f9fafb;
  padding: 16px;
  border-radius: 12px;
}

.detail-section.full-width {
  grid-column: 1 / -1;
}

.detail-section h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.score-large {
  font-size: 48px;
  font-weight: 700;
  text-align: center;
  margin-bottom: 16px;
}

.score-excellent {
  color: #059669;
}
.score-good {
  color: #3b82f6;
}
.score-fair {
  color: #f59e0b;
}
.score-poor {
  color: #dc2626;
}

.score-breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.breakdown-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #4b5563;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #e5e7eb;
  font-size: 14px;
}

.metric-row:last-child {
  border-bottom: none;
}

.cost {
  color: #dc2626;
  font-weight: 600;
}
.revenue {
  color: #059669;
  font-weight: 600;
}
.roas {
  color: #3b82f6;
  font-weight: 600;
}

.recommendation-box {
  text-align: center;
}

.rec-label {
  margin: 12px 0 8px;
  font-size: 16px;
  font-weight: 500;
  color: #374151;
}

.rec-budget {
  margin: 8px 0;
  font-size: 14px;
  color: #6b7280;
}

.rec-budget .positive {
  color: #059669;
}
.rec-budget .negative {
  color: #dc2626;
}

.rec-confidence {
  font-size: 13px;
  color: #9ca3af;
}

.status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-label {
  font-size: 12px;
  color: #9ca3af;
}

.category-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.cat-star {
  background: #d1fae5;
  color: #059669;
}
.cat-growth {
  background: #dbeafe;
  color: #2563eb;
}
.cat-stable {
  background: #f3f4f6;
  color: #6b7280;
}
.cat-watch {
  background: #fef3c7;
  color: #d97706;
}
.cat-problem {
  background: #fee2e2;
  color: #dc2626;
}

.trend-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.trend-up {
  background: #d1fae5;
  color: #059669;
}
.trend-stable {
  background: #f3f4f6;
  color: #6b7280;
}
.trend-down {
  background: #fee2e2;
  color: #dc2626;
}

.fatigue-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.fatigue-fresh {
  background: #d1fae5;
  color: #059669;
}
.fatigue-aging {
  background: #fef3c7;
  color: #d97706;
}
.fatigue-fatigued {
  background: #fed7aa;
  color: #c2410c;
}
.fatigue-dead {
  background: #fee2e2;
  color: #dc2626;
}

/* Responsive */
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .left-column {
    order: 2;
  }

  .right-column {
    order: 1;
  }
}

@media (max-width: 768px) {
  .ml-dashboard-page {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>

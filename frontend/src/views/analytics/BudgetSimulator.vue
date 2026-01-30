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
        <div class="simulator-page">
          <!-- Header -->
          <div class="page-header">
            <div class="header-left">
              <h1><span aria-hidden="true">🎯</span> Budget Simulator</h1>
              <p class="subtitle">Predict ROAS and optimize your ad spend</p>
            </div>
          </div>

          <div class="simulator-grid">
            <!-- Input Section -->
            <div class="input-section">
              <div class="card">
                <h3>Simulation Parameters</h3>

                <!-- Product Selector -->
                <div class="form-group">
                  <label>Select Product</label>
                  <select
                    v-model="selectedProductId"
                    @change="onProductChange"
                    :disabled="loadingProducts"
                    class="form-control"
                  >
                    <option value="">-- Select Product --</option>
                    <option
                      v-for="p in products"
                      :key="p.product_id"
                      :value="p.product_id"
                    >
                      {{ p.product_name }} ({{ formatRoas(p.avg_roas) }})
                    </option>
                  </select>
                </div>

                <!-- Target ROAS -->
                <div class="form-group">
                  <label>Target ROAS</label>
                  <div class="input-with-suffix">
                    <input
                      type="number"
                      v-model.number="targetRoas"
                      min="0.1"
                      max="50"
                      step="0.1"
                      class="form-control"
                    />
                    <span class="suffix">x</span>
                  </div>
                  <p class="hint">Current: {{ currentRoasText }}</p>
                </div>

                <!-- Budget Per Day -->
                <div class="form-group">
                  <label>Budget per Day</label>
                  <div class="input-with-prefix">
                    <span class="prefix">Rp</span>
                    <input
                      type="number"
                      v-model.number="budgetPerDay"
                      min="10000"
                      step="10000"
                      class="form-control"
                    />
                  </div>
                </div>

                <!-- Period -->
                <div class="form-group">
                  <label>Simulation Period</label>
                  <select v-model.number="periodDays" class="form-control">
                    <option :value="7">7 days</option>
                    <option :value="14">14 days</option>
                    <option :value="30">30 days</option>
                  </select>
                </div>

                <!-- Calculate Button -->
                <button
                  @click="runSimulation"
                  :disabled="loading || !selectedProduct"
                  class="btn-primary"
                  type="button"
                >
                  <span v-if="loading">Calculating...</span>
                  <span v-else>Calculate Projection</span>
                </button>

                <p v-if="error" class="error-text">{{ error }}</p>
              </div>
            </div>

            <!-- Result Section -->
            <div class="result-section">
              <div v-if="!hasResult" class="empty-result">
                <span aria-hidden="true">📊</span>
                <p>Select a product and run simulation</p>
              </div>

              <template v-else>
                <!-- Feasibility Card -->
                <div class="result-card" :class="feasibilityColor">
                  <div class="result-header">
                    <h4>Feasibility</h4>
                    <span class="badge" :class="feasibilityColor">
                      {{ simulationResult?.feasibility }}
                    </span>
                  </div>
                  <div class="confidence">
                    {{
                      formatPercent(simulationResult?.confidence_percent || 0)
                    }}
                    confidence
                  </div>
                </div>

                <!-- ROAS Comparison -->
                <div class="result-card">
                  <h4>ROAS Projection</h4>
                  <div class="roas-comparison">
                    <div class="roas-item">
                      <span class="label">Current</span>
                      <span class="value">
                        {{ formatRoas(simulationResult?.current_roas || 0) }}
                      </span>
                    </div>
                    <div class="arrow">
                      <span>{{ trendIcon }}</span>
                    </div>
                    <div class="roas-item">
                      <span class="label">Projected</span>
                      <span class="value projected">
                        {{ formatRoas(simulationResult?.projected_roas || 0) }}
                      </span>
                    </div>
                  </div>
                  <p class="trend-text">
                    Trend:
                    <strong>{{ simulationResult?.trend_prediction }}</strong>
                  </p>
                </div>

                <!-- Optimal Budget -->
                <div class="result-card">
                  <h4>Optimal Budget</h4>
                  <p class="optimal-value">
                    {{ formatCurrency(simulationResult?.optimal_budget || 0) }}
                    <span class="per-day">/day</span>
                  </p>
                </div>

                <!-- Recommendation -->
                <div class="result-card recommendation">
                  <h4>Recommendation</h4>
                  <p class="rec-text">{{ simulationResult?.recommendation }}</p>
                </div>

                <!-- Alternatives -->
                <div
                  v-if="simulationResult?.alternatives?.length"
                  class="result-card"
                >
                  <h4>Alternatives</h4>
                  <ul class="alternatives-list">
                    <li
                      v-for="(alt, idx) in simulationResult.alternatives"
                      :key="idx"
                    >
                      <template v-if="alt.target_roas && alt.required_budget">
                        ROAS {{ formatRoas(alt.target_roas) }} needs
                        {{ formatCurrency(alt.required_budget) }}/day
                      </template>
                      <template v-else-if="alt.budget && alt.expected_roas">
                        {{ formatCurrency(alt.budget) }}/day yields
                        {{ formatRoas(alt.expected_roas) }}
                      </template>
                    </li>
                  </ul>
                </div>
              </template>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useUnifiedHeader } from "@/composables/useUnifiedHeader";
import { useBudgetSimulation } from "@/composables/useBudgetSimulation";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();
const { setConnectionStatus } = useUnifiedHeader();

const {
  loading,
  loadingProducts,
  error,
  products,
  selectedProduct,
  targetRoas,
  budgetPerDay,
  periodDays,
  simulationResult,
  hasResult,
  feasibilityColor,
  trendIcon,
  fetchProducts,
  runSimulation,
  selectProduct,
  formatCurrency,
  formatRoas,
  formatPercent,
} = useBudgetSimulation();

const selectedProductId = ref("");

const currentRoasText = computed(() => {
  if (!selectedProduct.value) return "-";
  return formatRoas(selectedProduct.value.avg_roas);
});

function onProductChange() {
  const p = products.value.find(
    (x) => x.product_id === selectedProductId.value,
  );
  if (p) selectProduct(p);
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
  if (tab !== "analytics") {
    router.push(`/${tab}`);
  }
}

function handlePlatformChange(platform: string) {
  uiStore.setActivePlatform(platform as PlatformType);
}

onMounted(async () => {
  uiStore.setActiveTab("analytics");
  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }
  await fetchProducts();
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
@import "./BudgetSimulator.styles.css";
</style>

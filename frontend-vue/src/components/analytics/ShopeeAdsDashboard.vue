<template>
  <div class="dashboard-container">
    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <span aria-hidden="true">⏳</span> Loading dashboard...
    </div>

    <!-- No Data State -->
    <div v-else-if="!dashboard || !hasData" class="empty-state">
      <div class="empty-icon" aria-hidden="true">📊</div>
      <h2>No Data Available</h2>
      <p>Upload Shopee Ads data to see analytics</p>
    </div>

    <!-- Dashboard Content -->
    <template v-else>
      <!-- Summary Cards Row 1 -->
      <div class="summary-cards">
        <div class="card">
          <div class="card-label">Total Cost</div>
          <div class="card-value cost">
            {{ formatCurrency(dashboard.total_cost) }}
          </div>
        </div>
        <div class="card">
          <div class="card-label">Total Revenue</div>
          <div class="card-value revenue">
            {{ formatCurrency(dashboard.total_revenue) }}
          </div>
        </div>
        <div class="card">
          <div class="card-label">Total Orders</div>
          <div class="card-value">
            {{ formatNumber(dashboard.total_orders) }}
          </div>
        </div>
        <div class="card">
          <div class="card-label">Average ROAS</div>
          <div class="card-value" :class="roasClass(dashboard.avg_roas)">
            {{ formatRoas(dashboard.avg_roas) }}x
          </div>
        </div>
      </div>

      <!-- Summary Cards Row 2 -->
      <div class="summary-cards">
        <div class="card">
          <div class="card-label">Impressions</div>
          <div class="card-value">
            {{ formatNumber(dashboard.total_impressions) }}
          </div>
        </div>
        <div class="card">
          <div class="card-label">Clicks</div>
          <div class="card-value">
            {{ formatNumber(dashboard.total_clicks) }}
          </div>
        </div>
        <div class="card">
          <div class="card-label">CTR</div>
          <div class="card-value">{{ formatPercent(dashboard.avg_ctr) }}</div>
        </div>
        <div class="card">
          <div class="card-label">Conversion Rate</div>
          <div class="card-value">
            {{ formatPercent(dashboard.avg_conversion_rate) }}
          </div>
        </div>
      </div>

      <!-- Bidding Mode Comparison -->
      <div class="section">
        <h2>Bidding Mode Performance</h2>
        <div class="comparison-grid">
          <div
            v-for="stat in dashboard.bidding_mode_stats"
            :key="stat.bidding_mode"
            class="comparison-card"
          >
            <div class="type-header">
              <span class="type-icon" aria-hidden="true">{{
                getBiddingIcon(stat.bidding_mode)
              }}</span>
              <span class="type-name">{{ stat.bidding_mode }}</span>
            </div>
            <div class="type-metrics">
              <div class="metric">
                <span class="metric-label">Cost</span>
                <span class="metric-value">{{
                  formatCurrency(stat.cost)
                }}</span>
              </div>
              <div class="metric">
                <span class="metric-label">Revenue</span>
                <span class="metric-value">{{
                  formatCurrency(stat.revenue)
                }}</span>
              </div>
              <div class="metric">
                <span class="metric-label">ROAS</span>
                <span class="metric-value" :class="roasClass(stat.roas)"
                  >{{ formatRoas(stat.roas) }}x</span
                >
              </div>
              <div class="metric">
                <span class="metric-label">Products</span>
                <span class="metric-value">{{
                  formatNumber(stat.product_count || 0)
                }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Products -->
      <div class="section">
        <h2>Top 10 Products by Revenue</h2>
        <div class="table-container">
          <table class="data-table" aria-label="Top performing products">
            <thead>
              <tr>
                <th scope="col">#</th>
                <th scope="col">Product</th>
                <th scope="col">Cost</th>
                <th scope="col">Revenue</th>
                <th scope="col">Orders</th>
                <th scope="col">ROAS</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(product, index) in dashboard.top_products"
                :key="product.product_id"
              >
                <td>{{ index + 1 }}</td>
                <td class="product-name">
                  {{ truncateName(product.product_name) }}
                </td>
                <td>{{ formatCurrency(product.cost) }}</td>
                <td>{{ formatCurrency(product.revenue) }}</td>
                <td>{{ formatNumber(product.orders) }}</td>
                <td :class="roasClass(product.roas)">
                  {{ formatRoas(product.roas) }}x
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { DashboardSummary } from "@/composables/useShopeeAdsAnalytics";

const props = defineProps<{
  dashboard: DashboardSummary | null;
  loading: boolean;
  formatCurrency: (value: number) => string;
  formatNumber: (value: number) => string;
  formatRoas: (value: number) => string;
}>();

const hasData = computed(
  () => props.dashboard && props.dashboard.total_cost > 0,
);

function formatPercent(value: number): string {
  if (value == null) return "0.00%";
  return `${(value * 100).toFixed(2)}%`;
}

function roasClass(roas: number): string {
  if (roas == null) return "roas-poor";
  if (roas >= 5) return "roas-excellent";
  if (roas >= 2) return "roas-good";
  if (roas >= 1) return "roas-ok";
  return "roas-poor";
}

function getBiddingIcon(mode: string | null | undefined): string {
  if (!mode) return "📊";
  if (mode.includes("GMV")) return "🎯";
  if (mode.includes("Auto")) return "🤖";
  if (mode.includes("Manual")) return "⚙️";
  return "📊";
}

function truncateName(name: string | null | undefined): string {
  if (!name) return "-";
  return name.length > 40 ? name.substring(0, 40) + "..." : name;
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.card {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.card-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 8px;
}

.card-value {
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.card-value.cost {
  color: #dc2626;
}

.card-value.revenue {
  color: #047857;
}

.section {
  margin-top: 32px;
}

.section h2 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #1f2937;
}

.comparison-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.comparison-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.type-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.type-icon {
  font-size: 24px;
}

.type-name {
  font-weight: 600;
  color: #1f2937;
}

.type-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.metric {
  display: flex;
  flex-direction: column;
}

.metric-label {
  font-size: 11px;
  color: #6b7280;
}

.metric-value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background: #f53d2d;
  color: white;
  font-weight: 600;
  font-size: 12px;
}

.data-table td {
  font-size: 13px;
  color: #374151;
}

.data-table tbody tr:hover {
  background: #f9fafb;
}

.product-name {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.roas-excellent {
  color: #047857;
  font-weight: 600;
}

.roas-good {
  color: #059669;
}

.roas-ok {
  color: #b45309;
}

.roas-poor {
  color: #dc2626;
}

@media (max-width: 768px) {
  .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>

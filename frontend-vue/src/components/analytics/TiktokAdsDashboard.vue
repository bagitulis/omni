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
      <p>Upload TikTok Ads data to see analytics</p>
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
          <div class="card-label">Average ROI</div>
          <div class="card-value" :class="roiClass(dashboard.avg_roi)">
            {{ formatRoi(dashboard.avg_roi) }}x
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

      <!-- Creative Type Comparison -->
      <div class="section">
        <h2>Creative Type Performance</h2>
        <div class="comparison-grid">
          <div
            v-for="stat in dashboard.creative_type_comparison"
            :key="stat.creative_type"
            class="comparison-card"
          >
            <div class="type-header">
              <span class="type-icon" aria-hidden="true">{{
                stat.creative_type === "Video" ? "🎬" : "🖼️"
              }}</span>
              <span class="type-name">{{ stat.creative_type }}</span>
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
                <span class="metric-label">ROI</span>
                <span class="metric-value" :class="roiClass(stat.roi)"
                  >{{ formatRoi(stat.roi) }}x</span
                >
              </div>
              <div class="metric">
                <span class="metric-label">Cost/Order</span>
                <span class="metric-value">{{
                  formatCurrency(stat.cost_per_order)
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
                <th scope="col">Product ID</th>
                <th scope="col">Cost</th>
                <th scope="col">Revenue</th>
                <th scope="col">Orders</th>
                <th scope="col">ROI</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(product, index) in dashboard.top_products"
                :key="product.product_id"
              >
                <td>{{ index + 1 }}</td>
                <td class="product-id">{{ product.product_id }}</td>
                <td>{{ formatCurrency(product.cost) }}</td>
                <td>{{ formatCurrency(product.revenue) }}</td>
                <td>{{ formatNumber(product.orders) }}</td>
                <td :class="roiClass(product.roi)">
                  {{ formatRoi(product.roi) }}x
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
import type { DashboardSummary } from "@/composables/useTiktokAdsAnalytics";

const props = defineProps<{
  dashboard: DashboardSummary | null;
  loading: boolean;
  formatCurrency: (value: number) => string;
  formatNumber: (value: number) => string;
  formatRoi: (value: number) => string;
}>();

const hasData = computed(
  () => props.dashboard && props.dashboard.total_cost > 0,
);

function formatPercent(value: number): string {
  return `${(value * 100).toFixed(2)}%`;
}

function roiClass(roi: number): string {
  if (roi >= 5) return "roi-excellent";
  if (roi >= 2) return "roi-good";
  if (roi >= 1) return "roi-ok";
  return "roi-poor";
}
</script>

<style scoped src="./TiktokAdsDashboard.css"></style>

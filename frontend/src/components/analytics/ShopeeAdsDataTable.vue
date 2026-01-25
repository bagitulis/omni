<template>
  <div class="data-table-container">
    <!-- Filters -->
    <div class="filters">
      <div class="filter-group">
        <label for="bidding-mode-filter">Bidding Mode</label>
        <select
          id="bidding-mode-filter"
          v-model="filterMode"
          class="filter-select"
        >
          <option value="">All</option>
          <option value="GMV Max ROAS">GMV Max ROAS</option>
          <option value="Auto">Auto</option>
          <option value="Manual">Manual</option>
        </select>
      </div>
      <div class="filter-group">
        <label for="sort-by-filter">Sort By</label>
        <select id="sort-by-filter" v-model="sortBy" class="filter-select">
          <option value="revenue">Revenue</option>
          <option value="cost">Cost</option>
          <option value="roas">ROAS</option>
          <option value="conversions">Orders</option>
        </select>
      </div>
      <button
        class="btn-refresh"
        type="button"
        @click="handleRefresh"
        aria-label="Refresh data"
      >
        <span aria-hidden="true">🔄</span> Refresh
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <span aria-hidden="true">⏳</span> Loading data...
    </div>

    <!-- No Data State -->
    <div v-else-if="filteredData.length === 0" class="empty-state">
      <p>No product data found</p>
    </div>

    <!-- Data Table -->
    <div v-else class="table-wrapper">
      <table class="product-table" aria-label="Product data table">
        <thead>
          <tr>
            <th scope="col">Product</th>
            <th scope="col">Bidding Mode</th>
            <th scope="col">Cost</th>
            <th scope="col">Revenue</th>
            <th scope="col">Orders</th>
            <th scope="col">ROAS</th>
            <th scope="col">CTR</th>
            <th scope="col">Period</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filteredData" :key="row.id">
            <td class="product-cell">
              <div class="product-name">
                {{ truncateName(row.productName) }}
              </div>
              <div class="product-id">{{ row.productId }}</div>
            </td>
            <td>
              <span class="mode-badge" :class="modeClass(row.biddingMode)">
                {{ row.biddingMode || "Unknown" }}
              </span>
            </td>
            <td>{{ formatCurrency(row.cost) }}</td>
            <td class="revenue">{{ formatCurrency(row.revenue) }}</td>
            <td>{{ formatNumber(row.conversions) }}</td>
            <td :class="roasClass(row.roas)">{{ formatRoas(row.roas) }}x</td>
            <td>{{ formatPercent(row.ctr) }}</td>
            <td class="period">{{ row.periodLabel }}</td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination Info -->
      <div class="pagination-info">
        Showing {{ filteredData.length }} of {{ totalRecords }} records
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { ProductData } from "@/composables/useShopeeAdsAnalytics";

const props = defineProps<{
  data: ProductData[];
  loading: boolean;
  totalRecords: number;
  formatCurrency: (value: number) => string;
  formatNumber: (value: number) => string;
  formatRoas: (value: number) => string;
}>();

const emit = defineEmits<{
  refresh: [options: { orderBy: string; orderDir: string }];
}>();

// Local state
const filterMode = ref("");
const sortBy = ref("revenue");

// Computed
const filteredData = computed(() => {
  let result = [...props.data];

  // Filter by bidding mode
  if (filterMode.value) {
    result = result.filter((r) => r.biddingMode?.includes(filterMode.value));
  }

  // Sort
  result.sort((a, b) => {
    switch (sortBy.value) {
      case "revenue":
        return b.revenue - a.revenue;
      case "cost":
        return b.cost - a.cost;
      case "roas":
        return b.roas - a.roas;
      case "conversions":
        return b.conversions - a.conversions;
      default:
        return 0;
    }
  });

  return result;
});

// Methods
function handleRefresh() {
  emit("refresh", { orderBy: sortBy.value, orderDir: "desc" });
}

function formatPercent(value: number): string {
  return `${(value * 100).toFixed(2)}%`;
}

function modeClass(mode: string | null): string {
  if (!mode) return "mode-unknown";
  if (mode.includes("GMV")) return "mode-gmv";
  if (mode.includes("Auto")) return "mode-auto";
  return "mode-manual";
}

function roasClass(roas: number): string {
  if (roas >= 5) return "roas-excellent";
  if (roas >= 2) return "roas-good";
  if (roas >= 1) return "roas-ok";
  return "roas-poor";
}

function truncateName(name: string): string {
  return name.length > 35 ? name.substring(0, 35) + "..." : name;
}
</script>

<style scoped>
.data-table-container {
  padding: 20px;
}

.filters {
  display: flex;
  gap: 16px;
  align-items: flex-end;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-group label {
  font-size: 12px;
  color: #6b7280;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  min-width: 150px;
}

.btn-refresh {
  padding: 8px 16px;
  background: #f53d2d;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-refresh:hover {
  background: #d93025;
}

.loading-state,
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
}

.table-wrapper {
  overflow-x: auto;
}

.product-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.product-table th,
.product-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.product-table th {
  background: #f53d2d;
  color: white;
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
}

.product-table td {
  font-size: 13px;
  color: #374151;
}

.product-table tbody tr:hover {
  background: #f9fafb;
}

.product-cell {
  max-width: 250px;
}

.product-name {
  font-weight: 500;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-id {
  font-size: 11px;
  color: #9ca3af;
}

.mode-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.mode-gmv {
  background: #dcfce7;
  color: #166534;
}

.mode-auto {
  background: #dbeafe;
  color: #1e40af;
}

.mode-manual {
  background: #fef3c7;
  color: #92400e;
}

.mode-unknown {
  background: #f3f4f6;
  color: #6b7280;
}

.revenue {
  color: #059669;
  font-weight: 500;
}

.period {
  font-size: 11px;
  color: #6b7280;
}

.roas-excellent {
  color: #059669;
  font-weight: 600;
}

.roas-good {
  color: #10b981;
}

.roas-ok {
  color: #f59e0b;
}

.roas-poor {
  color: #dc2626;
}

.pagination-info {
  margin-top: 16px;
  text-align: center;
  font-size: 13px;
  color: #6b7280;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-select {
    width: 100%;
  }
}
</style>

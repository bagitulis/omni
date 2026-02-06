<template>
  <div class="data-table-container">
    <!-- Filters -->
    <div class="filters">
      <div class="filter-group">
        <label for="creative-type-filter">Creative Type</label>
        <select
          id="creative-type-filter"
          v-model="filterType"
          class="filter-select"
        >
          <option value="">All</option>
          <option value="Video">Video</option>
          <option value="Kartu produk">Kartu Produk</option>
        </select>
      </div>
      <div class="filter-group">
        <label for="sort-by-filter">Sort By</label>
        <select id="sort-by-filter" v-model="sortBy" class="filter-select">
          <option value="revenue">Revenue</option>
          <option value="cost">Cost</option>
          <option value="roi">ROI</option>
          <option value="orders">Orders</option>
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
      <p>No creative data found</p>
    </div>

    <!-- Data Table -->
    <div v-else class="table-wrapper">
      <table class="creative-table" aria-label="Creative data table">
        <thead>
          <tr>
            <th scope="col">Campaign</th>
            <th scope="col">Product ID</th>
            <th scope="col">Type</th>
            <th scope="col">Cost</th>
            <th scope="col">Revenue</th>
            <th scope="col">Orders</th>
            <th scope="col">ROI</th>
            <th scope="col">CTR</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filteredData" :key="row.id">
            <td class="campaign-cell">
              <div class="campaign-name">{{ row.campaign_name }}</div>
              <div class="campaign-id">{{ row.campaign_id }}</div>
            </td>
            <td class="product-id">{{ row.product_id }}</td>
            <td>
              <span class="type-badge" :class="typeClass(row.creative_type)">
                {{ row.creative_type }}
              </span>
            </td>
            <td>{{ formatCurrency(row.cost) }}</td>
            <td class="revenue">{{ formatCurrency(row.gross_revenue) }}</td>
            <td>{{ formatNumber(row.orders_sku) }}</td>
            <td :class="roiClass(row.roi)">{{ formatRoi(row.roi) }}x</td>
            <td>{{ formatPercent(row.ctr) }}</td>
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
import type { CreativeData } from "@/composables/useTiktokAdsAnalytics";

const props = defineProps<{
  data: CreativeData[];
  loading: boolean;
  totalRecords: number;
  formatCurrency: (value: number) => string;
  formatNumber: (value: number) => string;
  formatRoi: (value: number) => string;
}>();

const emit = defineEmits<{
  refresh: [options: { orderBy: string; orderDir: string }];
}>();

// Local state
const filterType = ref("");
const sortBy = ref("revenue");

// Computed
const filteredData = computed(() => {
  let result = [...props.data];

  // Filter by type
  if (filterType.value) {
    result = result.filter((r) => r.creative_type === filterType.value);
  }

  // Sort
  result.sort((a, b) => {
    switch (sortBy.value) {
      case "revenue":
        return b.gross_revenue - a.gross_revenue;
      case "cost":
        return b.cost - a.cost;
      case "roi":
        return b.roi - a.roi;
      case "orders":
        return b.orders_sku - a.orders_sku;
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

function typeClass(type: string): string {
  return type === "Video" ? "type-video" : "type-card";
}

function roiClass(roi: number): string {
  if (roi >= 5) return "roi-excellent";
  if (roi >= 2) return "roi-good";
  if (roi >= 1) return "roi-ok";
  return "roi-poor";
}
</script>

<style scoped>
.data-table-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.filters {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-group label {
  font-size: 0.75rem;
  color: #6b7280;
}

.filter-select {
  padding: 0.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  background: #ffffff;
  font-size: 0.875rem;
  min-width: 140px;
}

.btn-refresh {
  padding: 0.5rem 1rem;
  background: #1d4ed8;
  color: #ffffff;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-refresh:hover {
  background: #1e40af;
}

.loading-state,
.empty-state {
  padding: 2rem;
  text-align: center;
  color: #6b7280;
}

.table-wrapper {
  overflow-x: auto;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
}

.creative-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.creative-table th,
.creative-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.creative-table th {
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
  white-space: nowrap;
}

.campaign-cell {
  max-width: 200px;
}

.campaign-name {
  font-weight: 500;
  color: #1f2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.campaign-id {
  font-size: 0.75rem;
  color: #6b7280;
  font-family: monospace;
}

.product-id {
  font-family: monospace;
  font-size: 0.75rem;
}

.type-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.type-video {
  background: #dbeafe;
  color: #1e40af;
}

.type-card {
  background: #fef3c7;
  color: #92400e;
}

.revenue {
  color: #166534;
  font-weight: 500;
}

.roi-excellent {
  color: #166534;
  font-weight: 600;
}

.roi-good {
  color: #1d4ed8;
}

.roi-ok {
  color: #92400e;
}

.roi-poor {
  color: #991b1b;
}

.pagination-info {
  padding: 0.75rem;
  font-size: 0.75rem;
  color: #6b7280;
  text-align: right;
}

/* Dark mode */
:root.dark .filter-select {
  background: #1f2937;
  border-color: #374151;
  color: #f9fafb;
}

:root.dark .table-wrapper {
  background: #1f2937;
  border-color: #374151;
}

:root.dark .creative-table th {
  background: #374151;
  color: #f9fafb;
}

:root.dark .creative-table td {
  border-color: #374151;
}

:root.dark .campaign-name {
  color: #f9fafb;
}
</style>

<template>
  <div class="score-table-container">
    <!-- Header -->
    <div class="table-header">
      <h3>Product Analysis</h3>
      <div class="table-controls">
        <select v-model="sortBy" class="sort-select" aria-label="Sort by">
          <option value="unified_score">Score</option>
          <option value="revenue">Revenue</option>
          <option value="roas">ROAS</option>
          <option value="profit">Profit</option>
        </select>
        <button
          @click="toggleSortDir"
          class="sort-dir-btn"
          type="button"
          :aria-label="
            sortDir === 'desc' ? 'Sort descending' : 'Sort ascending'
          "
        >
          {{ sortDir === "desc" ? "↓" : "↑" }}
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading && products.length === 0" class="loading-state">
      <div v-for="i in 5" :key="i" class="skeleton-row"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="products.length === 0" class="empty-state">
      <Icon name="document" size="xl" class="empty-icon" />
      <p class="empty-title">No products to analyze</p>
      <p class="empty-hint">Upload TikTok Ads data to see product analysis</p>
    </div>

    <!-- Table -->
    <div v-else class="table-wrapper">
      <table class="score-table" aria-label="Product score analysis">
        <thead>
          <tr>
            <th scope="col" class="col-product">Product</th>
            <th scope="col" class="col-score">Score</th>
            <th scope="col" class="col-category">Category</th>
            <th scope="col" class="col-action">Action</th>
            <th scope="col" class="col-roas">ROAS</th>
            <th scope="col" class="col-revenue">Revenue</th>
            <th scope="col" class="col-alerts">Alerts</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="product in products"
            :key="product.product_id"
            class="product-row"
            @click="$emit('select', product)"
          >
            <td class="col-product">
              <div class="product-info">
                <span class="product-name">{{
                  truncateName(product.product_name)
                }}</span>
                <span class="product-id">{{ product.product_id }}</span>
              </div>
            </td>
            <td class="col-score">
              <div class="score-cell">
                <span
                  class="score-value"
                  :class="scoreClass(product.unified_score)"
                >
                  {{ product.unified_score.toFixed(0) }}
                </span>
                <div class="score-bar">
                  <div
                    class="score-fill"
                    :style="{ width: product.unified_score + '%' }"
                    :class="scoreClass(product.unified_score)"
                  ></div>
                </div>
              </div>
            </td>
            <td class="col-category">
              <span
                class="category-badge"
                :class="'cat-' + product.category.toLowerCase()"
              >
                {{ product.category }}
              </span>
            </td>
            <td class="col-action">
              <ActionBadge :action="product.action" />
            </td>
            <td class="col-roas" :class="roasClass(product.roas)">
              {{ product.roas.toFixed(2) }}x
            </td>
            <td class="col-revenue">
              {{ formatCurrency(product.total_revenue) }}
            </td>
            <td class="col-alerts">
              <Icon
                v-if="product.has_fatigue_warning"
                name="zap"
                size="sm"
                class="alert-icon fatigue"
                title="Creative Fatigue"
              />
              <Icon
                v-else-if="product.has_churn_risk"
                name="warning"
                size="sm"
                class="alert-icon churn"
                title="Churn Risk"
              />
              <span v-else class="no-alerts">-</span>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Load More -->
      <div v-if="hasMore" class="load-more">
        <button
          @click="$emit('load-more')"
          :disabled="loading"
          class="load-more-btn"
          type="button"
        >
          {{ loading ? "Loading..." : "Load More" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import type { MLProductAnalysis } from "@/composables/useMLAnalytics";
import Icon from "@/components/ui/Icon.vue";
import ActionBadge from "./ActionBadge.vue";

const props = defineProps<{
  products: MLProductAnalysis[];
  loading: boolean;
  hasMore: boolean;
  formatCurrency: (val: number) => string;
}>();

const emit = defineEmits<{
  (e: "select", product: MLProductAnalysis): void;
  (e: "load-more"): void;
  (e: "sort", sortBy: string, sortDir: string): void;
}>();

const sortBy = ref("unified_score");
const sortDir = ref("desc");

function toggleSortDir() {
  sortDir.value = sortDir.value === "desc" ? "asc" : "desc";
}

watch([sortBy, sortDir], () => {
  emit("sort", sortBy.value, sortDir.value);
});

function truncateName(name: string): string {
  if (!name) return "Unknown";
  return name.length > 40 ? name.substring(0, 40) + "..." : name;
}

function scoreClass(score: number): string {
  if (score >= 70) return "score-excellent";
  if (score >= 55) return "score-good";
  if (score >= 40) return "score-fair";
  return "score-poor";
}

function roasClass(roas: number): string {
  if (roas >= 3) return "roas-excellent";
  if (roas >= 2) return "roas-good";
  if (roas >= 1) return "roas-ok";
  return "roas-poor";
}
</script>

<style scoped>
.score-table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
}

.table-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.table-controls {
  display: flex;
  gap: 8px;
}

.sort-select {
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background: white;
}

.sort-dir-btn {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-weight: bold;
}

.loading-state {
  padding: 20px;
}

.skeleton-row {
  height: 48px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 6px;
  margin-bottom: 8px;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
}

.empty-icon :deep(.icon) {
  width: 48px;
  height: 48px;
  color: #6b7280;
  margin: 0 auto 12px;
  display: block;
}

.empty-state .empty-title {
  font-size: 16px;
  font-weight: 500;
  color: #374151;
  margin: 0 0 8px;
}

.empty-state .empty-hint {
  font-size: 13px;
  color: #9ca3af;
  margin: 0;
}

.table-wrapper {
  overflow-x: auto;
}

.score-table {
  width: 100%;
  border-collapse: collapse;
}

.score-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.score-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
}

.product-row {
  cursor: pointer;
  transition: background 0.15s;
}

.product-row:hover {
  background: #f9fafb;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.product-name {
  font-size: 14px;
  font-weight: 500;
  color: #1f2937;
}

.product-id {
  font-size: 12px;
  color: #9ca3af;
  font-family: monospace;
}

.score-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 100px;
}

.score-value {
  font-size: 14px;
  font-weight: 700;
  min-width: 28px;
}

.score-bar {
  flex: 1;
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
}

.score-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s;
}

.score-excellent {
  color: #059669;
  background-color: #059669;
}
.score-good {
  color: #3b82f6;
  background-color: #3b82f6;
}
.score-fair {
  color: #f59e0b;
  background-color: #f59e0b;
}
.score-poor {
  color: #dc2626;
  background-color: #dc2626;
}

.category-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
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

.roas-excellent {
  color: #059669;
  font-weight: 600;
}
.roas-good {
  color: #3b82f6;
  font-weight: 600;
}
.roas-ok {
  color: #f59e0b;
  font-weight: 600;
}
.roas-poor {
  color: #dc2626;
  font-weight: 600;
}

.col-revenue {
  font-family: monospace;
  font-size: 13px;
}

.alert-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 4px;
}

.alert-icon :deep(.icon) {
  width: 16px;
  height: 16px;
}

.alert-icon.fatigue :deep(.icon) {
  color: #f59e0b;
}

.alert-icon.churn :deep(.icon) {
  color: #ef4444;
}

.no-alerts {
  color: #d1d5db;
}

.load-more {
  padding: 16px;
  text-align: center;
  border-top: 1px solid #e5e7eb;
}

.load-more-btn {
  padding: 10px 24px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.load-more-btn:hover:not(:disabled) {
  background: #2563eb;
}

.load-more-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

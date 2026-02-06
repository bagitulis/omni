<template>
  <div class="import-list-container">
    <!-- Search & Filter Bar -->
    <div class="filter-bar">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Cari produk Shopee..."
          class="search-input"
          @input="debouncedSearch"
        />
        <span v-if="searchQuery" class="clear-icon" @click="clearSearch"
          >✕</span
        >
      </div>
      <button @click="$emit('refresh')" class="btn-refresh" :disabled="loading">
        <span class="refresh-icon" :class="{ spinning: loading }">↻</span>
        Refresh
      </button>
    </div>

    <!-- Product List Panel -->
    <div class="product-list-panel">
      <div class="panel-header">
        <h3>Produk Shopee</h3>
        <span class="product-count">{{ filteredProducts.length }} produk</span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Memuat produk...</p>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredProducts.length === 0" class="empty-state">
        <span class="empty-icon">📦</span>
        <p v-if="searchQuery">Tidak ada produk yang cocok dengan pencarian</p>
        <p v-else>Belum ada produk Shopee di database</p>
      </div>

      <!-- Product Cards -->
      <div v-else class="product-cards-list">
        <ImportProductCard
          v-for="product in filteredProducts"
          :key="product.item_id"
          :product="product"
          :selected="selectedProduct?.item_id === product.item_id"
          @click="$emit('select', product)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { ShopeeProduct } from "./ImportSelector.types";
import ImportProductCard from "./ImportProductCard.vue";

const props = defineProps<{
  products: ShopeeProduct[];
  loading: boolean;
  selectedProduct: ShopeeProduct | null;
}>();

defineEmits<{
  (e: "select", product: ShopeeProduct): void;
  (e: "refresh"): void;
}>();

const searchQuery = ref("");
let searchTimeout: number | null = null;

const filteredProducts = computed(() => {
  if (!searchQuery.value.trim()) return props.products;

  const query = searchQuery.value.toLowerCase();
  return props.products.filter(
    (p) =>
      p.name.toLowerCase().includes(query) ||
      p.item_id.toString().includes(query),
  );
});

const debouncedSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = window.setTimeout(() => {
    // Search is reactive via computed
  }, 300);
};

const clearSearch = () => {
  searchQuery.value = "";
};
</script>

<style scoped>
.import-list-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
  min-height: 0;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 1rem;
  align-items: center;
  padding: 1rem;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
  border-radius: 0.75rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  flex-shrink: 0;
}

.search-input-wrapper {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 1rem;
  font-size: 1.125rem;
  opacity: 0.6;
}

.search-input {
  width: 100%;
  padding: 0.875rem 3rem 0.875rem 3rem;
  border: 2px solid transparent;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 0.5rem;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #ff6b2c;
  background: white;
  box-shadow: 0 0 0 3px rgba(255, 107, 44, 0.1);
}

.clear-icon {
  position: absolute;
  right: 1rem;
  cursor: pointer;
  font-size: 1.125rem;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.clear-icon:hover {
  opacity: 1;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-refresh:hover:not(:disabled) {
  background: #ff5511;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 107, 44, 0.3);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-icon {
  font-size: 1.25rem;
  transition: transform 0.3s;
}

.refresh-icon.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Product List Panel */
.product-list-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e5e7eb;
  overflow: hidden;
  flex: 1;
  min-height: 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 2px solid #ff6b2c;
  flex-shrink: 0;
}

.panel-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #1a1a1a;
}

.product-count {
  font-size: 0.875rem;
  font-weight: 600;
  color: #6b7280;
  background: rgba(255, 107, 44, 0.1);
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
}

.product-cards-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
  color: #6b7280;
  flex: 1;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f4f6;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

@media (max-width: 640px) {
  .filter-bar {
    flex-direction: column;
  }

  .btn-refresh {
    width: 100%;
    justify-content: center;
  }
}
</style>

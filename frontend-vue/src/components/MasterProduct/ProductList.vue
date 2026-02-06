<template>
  <div class="product-list">
    <!-- Batch Update Toolbar -->
    <ProductBatchToolbar
      :selectedCount="selectedSkus.size"
      @clear="clearSelection"
      @open-price-modal="openBatchPriceModal"
      @open-stock-modal="openBatchStockModal"
    />

    <!-- Batch Update Modal -->
    <ProductBatchModal
      :show="showBatchModal"
      :type="batchModalType"
      :count="selectedSkus.size"
      :updating="batchUpdating"
      v-model:modelValue="batchValue"
      @close="closeBatchModal"
      @confirm="executeBatchUpdate"
    />

    <!-- States: Loading or Empty -->
    <ProductListStates :loading="loading" :isEmpty="products.length === 0" />

    <!-- Product Cards -->
    <div v-if="!loading && products.length > 0" class="product-cards">
      <ProductCard
        v-for="product in products"
        :key="product.id"
        :product="product"
        :expanded="expandedProducts.has(product.id)"
        :selectedSkus="selectedSkus"
        :isAllSelected="isAllProductSkusSelected(product)"
        :editingCell="editingCell"
        :savingCell="savingCell"
        :editedValue="editedValue"
        :skusNeedingSync="skusNeedingSync"
        :syncStatus="syncStatus"
        :syncErrorMessages="syncErrorMessages"
        :platformLoading="platformLoading"
        @delete="$emit('delete', product)"
        @toggle-expand="toggleExpand"
        @toggle-all-skus="toggleAllProductSkus"
        @toggle-sku-selection="toggleSkuSelection"
        @start-edit="startEdit"
        @save-edit="saveEdit"
        @cancel-edit="cancelEdit"
        @update:editedValue="(val) => (editedValue = val)"
        @sync-platform="syncToPlatform"
        @platform-click="handlePlatformClick"
      />
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button
        class="btn-pagination"
        :disabled="currentPage <= 1"
        @click="$emit('page-change', currentPage - 1)"
      >
        ← Sebelumnya
      </button>
      <span class="page-info">
        Halaman {{ currentPage }} dari {{ totalPages }}
      </span>
      <button
        class="btn-pagination"
        :disabled="currentPage >= totalPages"
        @click="$emit('page-change', currentPage + 1)"
      >
        Selanjutnya →
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useProductList } from "./composables/useProductList";
import { type MasterProduct, type ProductSku } from "./ProductList.types";
import ProductBatchToolbar from "./ProductBatchToolbar.vue";
import ProductBatchModal from "./ProductBatchModal.vue";
import ProductListStates from "./ProductListStates.vue";
import ProductCard from "./ProductCard.vue";

// Props
const props = defineProps<{
  products: MasterProduct[];
  loading?: boolean;
  currentPage?: number;
  totalPages?: number;
}>();

// Emits
const emit = defineEmits<{
  delete: [product: MasterProduct];
  "page-change": [page: number];
  "sku-updated": [sku: ProductSku];
}>();

// Composable
const {
  expandedProducts,
  skusNeedingSync,
  syncStatus,
  syncErrorMessages,
  platformLoading,
  editingCell,
  editedValue,
  savingCell,
  selectedSkus,
  showBatchModal,
  batchModalType,
  batchValue,
  batchUpdating,
  toggleExpand,
  syncToPlatform,
  handlePlatformClick,
  startEdit,
  cancelEdit,
  saveEdit,
  toggleSkuSelection,
  isAllProductSkusSelected,
  toggleAllProductSkus,
  clearSelection,
  openBatchPriceModal,
  openBatchStockModal,
  closeBatchModal,
  executeBatchUpdate,
} = useProductList(props, emit);
</script>

<style scoped>
.product-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.product-cards {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  margin-top: 1rem;
}

.btn-pagination {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border, #e5e7eb);
  background: var(--surface, white);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.btn-pagination:hover:not(:disabled) {
  background: var(--surface-hover, #f3f4f6);
}

.btn-pagination:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
}
</style>

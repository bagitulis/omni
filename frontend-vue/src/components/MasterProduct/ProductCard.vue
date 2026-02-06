<template>
  <div class="product-card">
    <ProductCardHeader
      :product="product"
      :expanded="expanded"
      @toggle-expand="$emit('toggle-expand', product.id)"
      @delete="$emit('delete', product)"
    />

    <!-- Expandable Variant Table -->
    <transition name="expand">
      <div v-if="expanded" class="variant-table-wrapper-container">
        <ProductVariantTable
          :skus="product.skus || []"
          :selectedSkus="selectedSkus"
          :isAllSelected="isAllSelected"
          :editingCell="editingCell"
          :savingCell="savingCell"
          :editedValue="editedValue"
          :skusNeedingSync="skusNeedingSync"
          :syncStatus="syncStatus"
          :syncErrorMessages="syncErrorMessages"
          :platformLoading="platformLoading"
          @toggle-all="$emit('toggle-all-skus', product, $event)"
          @toggle-row="$emit('toggle-sku-selection', $event)"
          @start-edit="(...args) => $emit('start-edit', ...args)"
          @save-edit="(sku) => $emit('save-edit', sku, product)"
          @cancel-edit="$emit('cancel-edit')"
          @update:editedValue="$emit('update:editedValue', $event)"
          @sync="(...args) => $emit('sync-platform', product.id, ...args)"
          @platform-click="$emit('platform-click', ...args)"
        />
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { type MasterProduct, type ProductSku } from "./ProductList.types";
import ProductCardHeader from "./ProductCardHeader.vue";
import ProductVariantTable from "./ProductVariantTable.vue";

defineProps<{
  product: MasterProduct;
  expanded: boolean;
  selectedSkus: Set<number>;
  isAllSelected: boolean;
  editingCell: { skuId: number; field: "price" | "stock" } | null;
  savingCell: { skuId: number; field: "price" | "stock" } | null;
  editedValue: number;
  skusNeedingSync: Map<number, Set<string>>;
  syncStatus: Map<string, "syncing" | "success" | "error">;
  syncErrorMessages: Map<string, string>;
  platformLoading: Map<string, boolean>;
}>();

defineEmits<{
  delete: [product: MasterProduct];
  "toggle-expand": [productId: number];
  "toggle-all-skus": [product: MasterProduct, event: Event];
  "toggle-sku-selection": [skuId: number];
  "start-edit": [skuId: number, field: "price" | "stock", value: number];
  "save-edit": [sku: ProductSku, product: MasterProduct];
  "cancel-edit": [];
  "update:editedValue": [value: number];
  "sync-platform": [productId: number, skuId: number, platform: string];
  "platform-click": [sku: ProductSku, platform: string];
}>();
</script>

<style scoped>
.product-card {
  background: var(--surface, white);
  border-radius: 0.625rem;
  border: 1px solid var(--border, #e5e7eb);
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}

.product-card:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
  border-color: #d1d5db;
  transform: translateY(-1px);
}

.expand-enter-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  max-height: 800px;
}

.expand-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.6, 1);
  max-height: 800px;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
  transform: translateY(-10px);
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  transform: translateY(0);
}
</style>

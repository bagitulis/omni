<template>
  <div class="variant-table-wrapper">
    <table class="variant-table">
      <thead>
        <tr>
          <th class="checkbox-col">
            <input
              type="checkbox"
              :checked="isAllSelected"
              @change="$emit('toggle-all', $event)"
              title="Pilih semua varian"
            />
          </th>
          <th>SKU</th>
          <th>Varian</th>
          <th class="text-left">Harga</th>
          <th class="text-left">Stok</th>
          <th>Platform</th>
        </tr>
      </thead>
      <tbody>
        <ProductVariantRow
          v-for="sku in skus"
          :key="sku.id"
          :sku="sku"
          :selected="selectedSkus.has(sku.id)"
          :editingCell="editingCell"
          :savingCell="savingCell"
          :editedValue="editedValue"
          :skusNeedingSync="skusNeedingSync"
          :syncStatus="syncStatus"
          :syncErrorMessages="syncErrorMessages"
          :platformLoading="platformLoading"
          @toggle-row="$emit('toggle-row', $event)"
          @start-edit="(...args) => $emit('start-edit', ...args)"
          @save-edit="$emit('save-edit', $event)"
          @cancel-edit="$emit('cancel-edit')"
          @update:editedValue="$emit('update:editedValue', $event)"
          @sync="(...args) => $emit('sync', ...args)"
          @platform-click="(...args) => $emit('platform-click', ...args)"
        />
        <tr v-if="!skus?.length">
          <td colspan="6" class="no-skus">Belum ada SKU</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { type ProductSku } from "./ProductList.types";
import ProductVariantRow from "./ProductVariantRow.vue";

defineProps<{
  skus: ProductSku[];
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
  "toggle-all": [event: Event];
  "toggle-row": [skuId: number];
  "start-edit": [skuId: number, field: "price" | "stock", value: number];
  "save-edit": [sku: ProductSku];
  "cancel-edit": [];
  "update:editedValue": [value: number];
  sync: [skuId: number, platform: string];
  "platform-click": [sku: ProductSku, platform: string];
}>();
</script>

<style scoped>
.variant-table-wrapper {
  border-top: 1px solid var(--border, #e5e7eb);
  overflow: hidden;
  background: linear-gradient(to bottom, #fafbfc 0%, #ffffff 100%);
}

.variant-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.variant-table thead {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  backdrop-filter: blur(10px);
}

.variant-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #495057;
  border-bottom: 2px solid var(--border, #e5e7eb);
  text-transform: uppercase;
  letter-spacing: 0.025em;
  font-size: 0.75rem;
}

.variant-table tbody {
  background: white;
}

.checkbox-col {
  width: 40px;
  text-align: center;
}

.checkbox-col input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.text-left {
  text-align: left;
}

.no-skus {
  text-align: center;
  color: var(--text-secondary, #6b7280);
  padding: 2rem !important;
  font-style: italic;
}
</style>

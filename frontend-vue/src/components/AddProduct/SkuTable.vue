<template>
  <div class="sku-table">
    <div class="sku-header">
      <h4 class="sku-title">SKU Details</h4>
      <button
        v-if="!hasVariants"
        type="button"
        class="btn btn-secondary btn-sm"
        @click="$emit('add-sku')"
      >
        <i class="pi pi-plus" aria-hidden="true"></i> Add SKU
      </button>
    </div>

    <!-- Bulk edit for variants -->
    <div v-if="hasVariants && skus.length > 1" class="bulk-edit">
      <label for="bulk-price">Set all prices:</label>
      <input
        id="bulk-price"
        type="number"
        :value="bulkPrice"
        placeholder="Price"
        min="0"
        @input="bulkPrice = parseFloat(($event.target as HTMLInputElement).value) || 0"
      />
      <label for="bulk-stock">Set all stock:</label>
      <input
        id="bulk-stock"
        type="number"
        :value="bulkStock"
        placeholder="Stock"
        min="0"
        @input="bulkStock = parseInt(($event.target as HTMLInputElement).value) || 0"
      />
      <button type="button" class="btn btn-secondary btn-sm" @click="applyBulk">
        Apply to All
      </button>
    </div>

    <div class="table-wrapper">
      <table class="data-table" role="grid">
        <thead>
          <tr>
            <th v-if="hasVariants" scope="col">Variant</th>
            <th scope="col">Seller SKU *</th>
            <th scope="col">Price *</th>
            <th scope="col">Stock *</th>
            <th v-if="!hasVariants" scope="col" class="actions-col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(sku, index) in skus" :key="index">
            <td v-if="hasVariants" class="variant-cell">
              {{ sku.variantLabel || `Variant ${index + 1}` }}
            </td>
            <td>
              <input
                :id="`sku-code-${index}`"
                type="text"
                :value="sku.sellerSku"
                placeholder="SKU-001"
                required
                @input="updateSku(index, 'sellerSku', ($event.target as HTMLInputElement).value)"
              />
            </td>
            <td>
              <input
                :id="`sku-price-${index}`"
                type="number"
                :value="sku.price"
                placeholder="0"
                min="0"
                required
                @input="updateSku(index, 'price', parseFloat(($event.target as HTMLInputElement).value) || 0)"
              />
            </td>
            <td>
              <input
                :id="`sku-stock-${index}`"
                type="number"
                :value="sku.stock"
                placeholder="0"
                min="0"
                required
                @input="updateSku(index, 'stock', parseInt(($event.target as HTMLInputElement).value) || 0)"
              />
            </td>
            <td v-if="!hasVariants" class="actions-col">
              <button
                v-if="skus.length > 1"
                type="button"
                class="btn btn-danger btn-sm"
                :aria-label="`Remove SKU ${sku.sellerSku || index + 1}`"
                @click="$emit('remove-sku', index)"
              >
                <i class="pi pi-trash" aria-hidden="true"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <p v-if="skus.length === 0" class="empty-message">
      No SKUs added yet. Click "Add SKU" to create one.
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

interface Sku {
  sellerSku: string;
  price: number;
  stock: number;
  variantLabel?: string;
}

defineProps<{
  skus: Sku[];
  hasVariants?: boolean;
}>();

const emit = defineEmits<{
  "add-sku": [];
  "remove-sku": [index: number];
  "update-sku": [index: number, field: string, value: string | number];
  "bulk-update": [field: string, value: number];
}>();

const bulkPrice = ref(0);
const bulkStock = ref(0);

function updateSku(index: number, field: string, value: string | number) {
  emit("update-sku", index, field, value);
}

function applyBulk() {
  if (bulkPrice.value > 0) emit("bulk-update", "price", bulkPrice.value);
  if (bulkStock.value > 0) emit("bulk-update", "stock", bulkStock.value);
}
</script>

<style scoped>
.sku-table {
  margin-top: 16px;
}

.sku-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.sku-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.bulk-edit {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 6px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.bulk-edit label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}

.bulk-edit input {
  width: 100px;
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
}

.table-wrapper {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table th,
.data-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
}

.data-table td input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
}

.data-table td input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.variant-cell {
  font-weight: 500;
  color: #1f2937;
  white-space: nowrap;
}

.actions-col {
  width: 80px;
  text-align: center;
}

.empty-message {
  text-align: center;
  padding: 24px;
  color: #6b7280;
  font-size: 14px;
}

.btn { padding: 10px 16px; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 14px; display: inline-flex; align-items: center; gap: 8px; }
.btn-sm { padding: 6px 12px; font-size: 12px; }
.btn-secondary { background: #e5e7eb; color: #374151; }
.btn-secondary:hover { background: #d1d5db; }
.btn-danger { background: #ef4444; color: white; }
.btn-danger:hover { background: #dc2626; }
</style>

<template>
  <div class="tab-content">
    <div class="content-header">
      <h3>Get Product Detail</h3>
      <div class="detail-inputs">
        <input
          v-model="localItemId"
          :type="inputType"
          :placeholder="placeholder"
          class="search-input"
          @input="$emit('update:itemId', localItemId)"
        />
        <button
          @click="$emit('fetch-detail')"
          :disabled="!localItemId || loading"
          class="btn btn-primary"
        >
          <i :class="loading ? 'pi pi-spin pi-spinner' : 'pi pi-search'"></i>
          {{ loading ? "Loading..." : "Search" }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-spinner">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Loading product detail...</p>
    </div>

    <div v-else-if="product" class="product-detail-card">
      <div class="detail-header">
        <h3>{{ product.name || "Product" }}</h3>
        <span :class="['status-badge', product.status?.toLowerCase()]">
          {{ product.status }}
        </span>
      </div>
      <div class="detail-grid">
        <div class="detail-section">
          <h4>Basic Info</h4>
          <p><strong>Item ID:</strong> {{ product.itemId }}</p>
          <p v-if="product.brand">
            <strong>Brand:</strong> {{ product.brand }}
          </p>
          <p><strong>Status:</strong> {{ product.status }}</p>
        </div>
        <div class="detail-section">
          <h4>Pricing & Stock</h4>
          <div v-if="product.skus && product.skus.length > 0">
            <div v-for="(sku, idx) in product.skus" :key="idx" class="sku-item">
              <p>
                <strong>SKU {{ Number(idx) + 1 }}:</strong> {{ sku.sellerSku }}
              </p>
              <p v-if="sku.variantName">
                <strong>Variant:</strong> {{ sku.variantName }}
              </p>
              <p>
                <strong>Price:</strong>
                {{ sku.price ? "Rp " + formatNumber(sku.price) : "-" }}
              </p>
              <p><strong>Stock:</strong> {{ sku.quantity || 0 }}</p>
            </div>
          </div>
          <p v-else>No SKU information</p>
        </div>
        <div class="detail-section full-width">
          <h4>Description</h4>
          <p v-if="product.description" class="description-text">
            {{ stripHtml(product.description) }}
          </p>
          <p v-else>-</p>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <i class="pi pi-inbox"></i>
      <p>No product selected</p>
      <small>{{ emptyMessage }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { formatNumber, stripHtml } from "@/utils/productManagerUtils";

interface Sku {
  sellerSku?: string;
  variantName?: string;
  price?: number;
  quantity?: number;
}

interface Product {
  name?: string;
  itemId?: string | number;
  brand?: string;
  status?: string;
  description?: string;
  skus?: Sku[];
}

const props = defineProps<{
  itemId: string;
  loading: boolean;
  product: Product | null;
  placeholder?: string;
  inputType?: string;
  emptyMessage?: string;
}>();

defineEmits<{
  "update:itemId": [value: string];
  "fetch-detail": [];
}>();

const localItemId = ref(props.itemId);

watch(
  () => props.itemId,
  (newVal) => {
    localItemId.value = newVal;
  },
);
</script>

<style scoped>
/* Inherits from ProductManagerCommon.styles.css */
.sku-item {
  border: 1px solid #ecf0f1;
  padding: 12px;
  border-radius: 4px;
  margin: 10px 0;
  background: #f9fafb;
}
</style>

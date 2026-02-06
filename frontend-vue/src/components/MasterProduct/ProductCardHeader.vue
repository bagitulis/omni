<template>
  <div class="card-header">
    <ProductThumbnail
      :imageUrl="getFirstImage(product)"
      :altText="product.title"
    />

    <!-- Product Info -->
    <div class="product-info">
      <h3 class="product-title">{{ product.title }}</h3>
      <div class="product-meta">
        <span class="sku-count"> {{ product.skus?.length || 0 }} SKU </span>
        <template v-if="product.skus?.length === 1">
          <span class="product-price-inline">{{
            formatPrice(product.skus[0].price)
          }}</span>
          <span class="product-stock-inline"
            >Stok: {{ product.skus[0].stock }}</span
          >
        </template>
        <span class="product-status" :class="getStatusClass(product.status)">
          {{ getStatusLabel(product.status) }}
        </span>
      </div>
      <div class="product-actions">
        <router-link
          :to="`/master-products/${product.id}`"
          class="btn-action btn-edit"
        >
          Edit
        </router-link>
        <button
          class="btn-action btn-delete"
          @click.stop="$emit('delete', product)"
        >
          Hapus
        </button>
      </div>
    </div>

    <!-- Expand Toggle -->
    <button
      v-if="product.skus && product.skus.length > 1"
      class="expand-toggle"
      :class="{ expanded: expanded }"
      @click="$emit('toggle-expand')"
      :aria-label="expanded ? 'Collapse variants' : 'Expand variants'"
    >
      <span class="chevron">▼</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { type MasterProduct } from "./ProductList.types";
import ProductThumbnail from "./ProductThumbnail.vue";

defineProps<{
  product: MasterProduct;
  expanded: boolean;
}>();

defineEmits<{
  delete: [product: MasterProduct];
  "toggle-expand": [];
}>();

const getFirstImage = (product: MasterProduct): string | null => {
  if (!product.images || product.images.length === 0) return null;
  return product.images[0] || null;
};

const getStatusClass = (status: string): string => {
  switch (status) {
    case "active":
      return "status-active";
    case "draft":
      return "status-draft";
    case "archived":
      return "status-archived";
    default:
      return "status-draft";
  }
};

const getStatusLabel = (status: string): string => {
  switch (status) {
    case "active":
      return "Aktif";
    case "draft":
      return "Draft";
    case "archived":
      return "Diarsipkan";
    default:
      return status;
  }
};

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price);
};
</script>

<style scoped>
.card-header {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  align-items: flex-start;
}

.product-info {
  flex: 1;
  min-width: 0;
}

.product-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0 0 0.5rem 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-meta {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.75rem;
}

.sku-count {
  font-size: 0.875rem;
  color: #6b7280;
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  padding: 0.25rem 0.625rem;
  border-radius: 0.375rem;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.product-price-inline {
  font-size: 0.875rem;
  font-weight: 600;
  color: #059669;
  padding: 0.25rem 0.5rem;
  background: #ecfdf5;
  border-radius: 0.25rem;
}

.product-stock-inline {
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
  padding: 0.25rem 0.5rem;
  background: #f9fafb;
  border-radius: 0.25rem;
}

.product-status {
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
}

.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-draft {
  background: #fef3c7;
  color: #92400e;
}

.status-archived {
  background: #f3f4f6;
  color: #6b7280;
}

.product-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  border: none;
}

.btn-edit {
  background: var(--primary, #3b82f6);
  color: white;
}

.btn-edit:hover {
  background: var(--primary-hover, #2563eb);
}

.btn-delete {
  background: transparent;
  color: #dc2626;
  border: 1px solid #dc2626;
}

.btn-delete:hover {
  background: #fef2f2;
}

.expand-toggle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid var(--border, #e5e7eb);
  background: linear-gradient(135deg, #ffffff 0%, #f9fafb 100%);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.expand-toggle:hover {
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  border-color: #d1d5db;
  transform: scale(1.05);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.expand-toggle:active {
  transform: scale(0.95);
}

.chevron {
  font-size: 0.875rem;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  color: #6b7280;
  font-weight: bold;
}

.expand-toggle.expanded {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-color: #2563eb;
}

.expand-toggle.expanded .chevron {
  transform: rotate(180deg);
  color: white;
}

.expand-toggle.expanded:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  border-color: #1d4ed8;
}

@media (max-width: 640px) {
  .card-header {
    flex-wrap: wrap;
  }
}
</style>

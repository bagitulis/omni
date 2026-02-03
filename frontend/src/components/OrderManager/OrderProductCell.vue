<template>
  <div class="product-col">
    <div v-for="(item, i) in items" :key="i" class="product-item">
      <div class="product-image-wrapper">
        <ProductThumbnail
          :src="item.product_image"
          :alt="item.product_name"
          :platform="platform"
          size="medium"
        />
      </div>

      <div class="product-info">
        <div class="product-header">
          <span class="product-name" :title="item.product_name">
            {{ item.product_name }}
          </span>
          <span class="product-qty">x{{ item.qty }}</span>
        </div>

        <div
          v-if="item.variation_name"
          class="product-variant"
          :title="item.variation_name"
        >
          Var: {{ item.variation_name }}
        </div>

        <div v-if="item.sku" class="product-sku">SKU: {{ item.sku }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ProductThumbnail from "../Common/ProductThumbnail.vue";

export interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  price?: number;
  product_image?: string;
}

interface Props {
  items: OrderItem[];
  platform: string;
}

defineProps<Props>();
</script>

<style scoped>
@import "./styles/variables.css";

.product-col {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-sm);
}

.product-item {
  display: flex;
  gap: var(--om-spacing-sm);
  align-items: flex-start;
  padding: 4px;
  border-radius: var(--om-radius-sm);
  transition: background-color var(--om-transition-fast);
}

.product-item:hover {
  background-color: var(--om-bg-hover);
}

.product-image-wrapper {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.product-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--om-spacing-xs);
}

.product-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.product-qty {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  font-weight: 600;
  background: var(--om-bg-secondary);
  padding: 1px 6px;
  border-radius: var(--om-radius-full);
  white-space: nowrap;
}

.product-variant,
.product-sku {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-sku {
  color: var(--om-text-disabled);
  font-family: monospace;
}
</style>

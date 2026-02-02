<template>
  <div class="product-col">
    <div v-for="(item, i) in items" :key="i" class="product-item">
      <div class="product-image">
        <img
          v-if="getProductImage(item)"
          :src="getProductImage(item)"
          :alt="item.product_name"
          @error="onImageError(item)"
          loading="lazy"
        />
        <div v-else class="image-placeholder">
          <i class="pi pi-image"></i>
        </div>
      </div>
      <div class="product-info">
        <div class="product-name" :title="item.product_name">
          {{ item.product_name }}
        </div>
        <span class="product-qty">x{{ item.qty }}</span>
        <div v-if="item.variation_name" class="product-variant">
          Variasi: {{ item.variation_name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

export interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number; // Backend sends 'qty' field
  price?: number;
  product_image?: string;
}

interface Props {
  items: OrderItem[];
  platform: string;
}

const props = defineProps<Props>();

const imageErrors = ref<Set<string>>(new Set());

const getItemKey = (item: OrderItem) => item.sku || item.product_name;

const getProductImage = (item: OrderItem) => {
  const key = getItemKey(item);
  if (imageErrors.value.has(key) || !item.product_image) {
    return null;
  }
  const url = item.product_image;

  // Shopee CDN: add _tn suffix for 60x60 thumbnail
  // Format: https://cf.shopee.co.id/file/{image_id} -> add _tn
  if (props.platform === "shopee" && url.includes("cf.shopee")) {
    // If already has _tn suffix, use as-is
    if (url.includes("_tn")) {
      return url;
    }
    // Add _tn suffix for thumbnail (no extension needed)
    return `${url}_tn`;
  }

  return url;
};

const onImageError = (item: OrderItem) => {
  imageErrors.value.add(getItemKey(item));
};
</script>

<style scoped>
@import "./OrderManager.theme.css";

/* Product Column */
.product-col {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-md);
}

.product-item {
  display: flex;
  gap: var(--om-spacing-sm);
  align-items: flex-start;
}

.product-image {
  width: 60px;
  height: 60px;
  border-radius: var(--om-radius-sm);
  overflow: hidden;
  background: var(--om-bg-secondary);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--om-text-disabled);
  font-size: 1.5rem;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.product-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  /* Allow 2 lines */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
  white-space: normal;
}

.product-qty {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  font-weight: 500;
}

.product-variant {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Responsive Design */
@media (max-width: 1200px) {
  .product-name {
    font-size: var(--om-font-xs);
  }
}

@media (max-width: 768px) {
  .product-col {
    border: 1px solid var(--om-border);
    border-radius: var(--om-radius-sm);
    padding: var(--om-spacing-sm);
    background: var(--om-bg-secondary);
  }

  .product-item {
    gap: var(--om-spacing-xs);
  }

  .product-image {
    width: 50px;
    height: 50px;
  }
}
</style>

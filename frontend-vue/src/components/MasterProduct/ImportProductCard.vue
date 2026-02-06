<template>
  <div class="product-card" :class="{ selected }" @click="$emit('click')">
    <img
      v-if="getProductImage(product)"
      :src="getProductImage(product) || ''"
      :alt="product.name"
      class="product-image"
    />
    <div v-else class="product-image-placeholder">
      <span>📷</span>
    </div>
    <div class="product-card-info">
      <h4 class="product-name">{{ product.name }}</h4>
      <div class="product-meta">
        <span class="platform-badge">🟠 Shopee</span>
        <span class="sku-count">{{ getSkuCount(product) }} SKU</span>
      </div>
      <div class="product-price">
        {{ formatPrice(getProductPrice(product)) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ShopeeProduct } from "./ImportSelector.types";

const props = defineProps<{
  product: ShopeeProduct;
  selected: boolean;
}>();

defineEmits<{
  (e: "click"): void;
}>();

const getProductImage = (product: ShopeeProduct): string | null => {
  if (product.images && product.images.length > 0) {
    return product.images[0];
  }
  return null;
};

const getSkuCount = (product: ShopeeProduct): number => {
  return product.models?.length || 0;
};

const getProductPrice = (product: ShopeeProduct): number => {
  return product.price_info?.current_price || product.price || 0;
};

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price || 0);
};
</script>

<style scoped>
.product-card {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.product-card:hover {
  border-color: #ff6b2c;
  box-shadow: 0 4px 12px rgba(255, 107, 44, 0.15);
  transform: translateX(4px);
}

.product-card.selected {
  border-color: #ff6b2c;
  background: linear-gradient(135deg, #fff5f0 0%, #ffffff 100%);
  box-shadow: 0 4px 16px rgba(255, 107, 44, 0.2);
}

.product-image,
.product-image-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 0.5rem;
  object-fit: cover;
  flex-shrink: 0;
  background: #f3f4f6;
}

.product-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  opacity: 0.5;
}

.product-card-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.product-name {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.product-meta {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.platform-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.25rem 0.625rem;
  background: linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%);
  color: white;
  border-radius: 0.25rem;
}

.sku-count {
  font-size: 0.8125rem;
  color: #6b7280;
  font-weight: 500;
}

.product-price {
  font-size: 1rem;
  font-weight: 700;
  color: #ff6b2c;
}
</style>

<template>
  <div class="product-item-row">
    <!-- Image -->
    <div class="product-image-wrapper">
      <ProductThumbnail
        :src="product.product_image"
        :alt="product.product_name"
        :platform="platform"
        size="medium"
      />
    </div>

    <!-- Details -->
    <div class="product-details">
      <div class="product-name" :title="product.product_name">
        {{ product.product_name }}
      </div>
      <div class="product-meta">
        <span v-if="product.variation_name" class="meta-tag variant">
          {{ product.variation_name }}
        </span>
        <span v-if="product.sku" class="meta-tag sku">
          {{ product.sku }}
        </span>
      </div>
    </div>

    <!-- Pricing & Qty -->
    <div class="product-pricing">
      <div class="price-qty-group">
        <span class="qty-badge">x{{ product.qty }}</span>
        <span v-if="product.price" class="price-text">
          {{ formatPrice(product.price) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ProductThumbnail from "../Common/ProductThumbnail.vue";

interface OrderProduct {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  price?: number;
  product_image?: string;
}

interface Props {
  product: OrderProduct;
  platform: string;
}

defineProps<Props>();

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};
</script>

<style scoped>
@import "./OrderManager.theme.css";

.product-item-row {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-sm);
  border-bottom: 1px solid var(--om-border);
  background: var(--om-bg-primary);
  transition: background-color var(--om-transition-fast);
}

.product-item-row:hover {
  background-color: var(--om-bg-hover);
}

.product-item-row:last-child {
  border-bottom: none;
}

.product-image-wrapper {
  flex-shrink: 0;
}

/* Details */
.product-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.product-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-meta {
  display: flex;
  gap: var(--om-spacing-xs);
  flex-wrap: wrap;
}

.meta-tag {
  font-size: var(--om-font-xs);
  padding: 2px 6px;
  border-radius: var(--om-radius-sm);
  background: var(--om-bg-secondary);
  color: var(--om-text-secondary);
}

.meta-tag.sku {
  font-family: monospace;
  background: #f0f0f0;
}

/* Pricing */
.product-pricing {
  text-align: right;
  min-width: 100px;
}

.price-qty-group {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.qty-badge {
  font-size: var(--om-font-xs);
  font-weight: 600;
  color: var(--om-text-primary);
  background: var(--om-bg-secondary);
  padding: 2px 8px;
  border-radius: var(--om-radius-full);
}

.price-text {
  font-size: var(--om-font-sm);
  font-weight: 600;
  color: var(--om-primary);
}

@media (max-width: 768px) {
  .product-item-row {
    flex-wrap: wrap;
  }

  .product-details {
    min-width: calc(100% - 70px);
  }

  .product-pricing {
    width: 100%;
    display: flex;
    justify-content: flex-end;
    margin-top: -10px;
  }

  .price-qty-group {
    flex-direction: row;
    align-items: center;
    gap: var(--om-spacing-sm);
  }
}
</style>

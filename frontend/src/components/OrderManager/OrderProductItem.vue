<template>
  <div class="product-item">
    <!-- Product Image -->
    <div class="product-image">
      <img
        v-if="imageUrl"
        :src="imageUrl"
        :alt="product.product_name"
        @error="onImageError"
        loading="lazy"
      />
      <div v-else class="image-placeholder">
        <i class="pi pi-image"></i>
      </div>
    </div>

    <!-- Product Details -->
    <div class="product-details">
      <span class="product-name" :title="product.product_name">
        {{ product.product_name }}
      </span>
      <span v-if="product.variation_name" class="product-variant">
        {{ product.variation_name }}
      </span>
      <span class="product-sku">SKU: {{ product.sku }}</span>
    </div>

    <!-- Quantity -->
    <div class="product-qty">
      <span class="qty-label">Qty</span>
      <span class="qty-value">{{ product.qty }}</span>
    </div>

    <!-- Price (optional) -->
    <div v-if="product.price" class="product-price">
      {{ formatPrice(product.price) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

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

const props = defineProps<Props>();

const imageError = ref(false);

const imageUrl = computed(() => {
  if (imageError.value || !props.product.product_image) {
    return null;
  }
  // Convert to webp thumbnail if possible (60x60)
  const url = props.product.product_image;
  // Shopee images can use _tn suffix for thumbnails
  if (props.platform === "shopee" && !url.includes("_tn")) {
    return url.replace(/\.(jpg|jpeg|png)$/i, "_tn.$1");
  }
  return url;
});

const onImageError = () => {
  imageError.value = true;
};

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

.product-item {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-md);
  padding: var(--om-spacing-sm) 0;
  border-bottom: 1px solid var(--om-border);
}

.product-item:last-child {
  border-bottom: none;
}

.product-image {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  border-radius: var(--om-radius-sm);
  overflow: hidden;
  background: var(--om-bg-secondary);
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

.product-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.product-name {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-variant {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
}

.product-sku {
  font-size: var(--om-font-xs);
  color: var(--om-text-disabled);
  font-family: monospace;
}

.product-qty {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 50px;
}

.qty-label {
  font-size: var(--om-font-xs);
  color: var(--om-text-secondary);
}

.qty-value {
  font-size: var(--om-font-base);
  font-weight: 600;
  color: var(--om-text-primary);
}

.product-price {
  min-width: 100px;
  text-align: right;
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-primary);
}

@media (max-width: 768px) {
  .product-item {
    flex-wrap: wrap;
  }

  .product-details {
    flex: 1 1 calc(100% - 80px);
    order: 1;
  }

  .product-image {
    order: 0;
  }

  .product-qty,
  .product-price {
    order: 2;
  }
}
</style>

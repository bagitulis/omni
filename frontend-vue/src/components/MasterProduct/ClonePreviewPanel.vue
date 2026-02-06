<template>
  <div class="panel" :class="{ conflict: isConflict }">
    <div class="panel-header">
      <span class="platform-badge" :class="platform">
        {{ platformEmoji }} {{ platformLabel }}
      </span>
      <span class="label">{{ panelLabel }}</span>
    </div>
    <div class="panel-body">
      <div class="product-image">
        <img
          v-if="product?.images?.[0]"
          :src="product.images[0]"
          alt="Product"
        />
        <div v-else class="no-image">No Image</div>
      </div>
      <div class="product-info">
        <h3
          class="product-name"
          :class="{ truncated: adjustments?.title_will_truncate }"
        >
          {{ displayTitle }}
        </h3>
        <div v-if="adjustments?.title_will_truncate" class="char-count">
          {{ displayTitle.length }}/{{ adjustments.title_limit }} karakter
        </div>
        <div class="product-details">
          <div class="detail-row">
            <span class="label">SKU:</span>
            <span class="value">{{ product?.sku || "-" }}</span>
          </div>
          <div
            class="detail-row"
            :class="{ different: hasDifference('price') }"
          >
            <span class="label">Harga:</span>
            <span class="value price">{{ formatPrice(product?.price) }}</span>
          </div>
          <div
            class="detail-row"
            :class="{ different: hasDifference('stock') }"
          >
            <span class="label">Stok:</span>
            <span class="value">{{ product?.stock || 0 }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type {
  ProductSummary,
  AdjustmentInfo,
  Difference,
} from "./ClonePreview.types";

const props = defineProps<{
  platform: string;
  platformLabel: string;
  panelLabel: string;
  product?: ProductSummary;
  isConflict?: boolean;
  adjustments?: AdjustmentInfo;
  differences?: Difference[];
}>();

const platformEmoji = computed(() => {
  const emojis: Record<string, string> = {
    shopee: "",
    lazada: "",
    tiktok: "",
  };
  return emojis[props.platform] || "";
});

const displayTitle = computed(() => {
  if (props.adjustments?.adjusted_title) {
    return props.adjustments.adjusted_title;
  }
  return props.product?.name || "";
});

const formatPrice = (price?: number) => {
  if (!price) return "Rp 0";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};

const hasDifference = (field: string) => {
  return props.differences?.some((d) => d.field === field);
};
</script>

<style scoped>
.panel {
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 0.75rem;
  overflow: hidden;
}

.panel.conflict {
  border-color: #f59e0b;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.panel-header .label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
}

.platform-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.25rem 0.75rem;
  border-radius: 0.25rem;
  color: white;
}

.platform-badge.shopee {
  background: linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%);
}

.platform-badge.lazada {
  background: linear-gradient(135deg, #0f146d 0%, #1a237e 100%);
}

.platform-badge.tiktok {
  background: linear-gradient(135deg, #000000 0%, #333333 100%);
}

.panel-body {
  padding: 1rem;
}

.product-image {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 0.5rem;
  overflow: hidden;
  background: #f3f4f6;
  margin-bottom: 1rem;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 0.875rem;
}

.product-name {
  margin: 0 0 0.5rem 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-name.truncated {
  color: #f59e0b;
}

.char-count {
  font-size: 0.75rem;
  color: #f59e0b;
  margin-bottom: 0.5rem;
}

.product-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.detail-row .label {
  color: #6b7280;
}

.detail-row .value {
  font-weight: 600;
  color: #1a1a1a;
}

.detail-row .value.price {
  color: #ff6b2c;
}

.detail-row.different {
  background: #fef3c7;
  padding: 0.25rem 0.5rem;
  margin: -0.25rem -0.5rem;
  border-radius: 0.25rem;
}
</style>

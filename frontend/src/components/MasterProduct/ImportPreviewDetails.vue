<template>
  <div class="preview-details">
    <!-- Product Image -->
    <div class="preview-image-container">
      <img
        v-if="previewData.images?.[0]"
        :src="previewData.images[0]"
        :alt="previewData.title"
        class="preview-image"
      />
    </div>

    <!-- Product Title & Description -->
    <div class="preview-info">
      <h4 class="preview-title">{{ previewData.title }}</h4>
      <p class="preview-description">{{ previewData.description }}</p>
    </div>

    <!-- Source Info -->
    <div class="preview-source">
      <div class="source-item">
        <span class="label">Platform:</span>
        <span class="value">🟠 {{ previewData.source_platform }}</span>
      </div>
      <div class="source-item">
        <span class="label">Item ID:</span>
        <span class="value">{{ previewData.source_item_id }}</span>
      </div>
    </div>

    <!-- SKUs Preview -->
    <div class="preview-skus">
      <h5>SKUs yang akan diimport ({{ previewData.skus?.length || 0 }})</h5>
      <div class="sku-list">
        <div
          v-for="(sku, index) in previewData.skus"
          :key="index"
          class="sku-item"
        >
          <div class="sku-header">
            <span class="sku-code">{{ sku.seller_sku }}</span>
            <span class="sku-price">{{ formatPrice(sku.price) }}</span>
          </div>
          <div class="sku-details">
            <span class="sku-variant">{{ sku.variant_name || "Default" }}</span>
            <span class="sku-stock">Stok: {{ sku.stock }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Import Button -->
    <button @click="$emit('import')" class="btn-import" :disabled="importing">
      <span v-if="importing" class="spinner-sm"></span>
      {{ importing ? "Mengimport..." : "Import ke Master Product" }}
    </button>
  </div>
</template>

<script setup lang="ts">
import type { ImportPreviewResponse } from "@/services/masterProductService";

defineProps<{
  previewData: ImportPreviewResponse;
  importing: boolean;
}>();

defineEmits<{
  (e: "import"): void;
}>();

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
.preview-details {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.preview-image-container {
  width: 100%;
  aspect-ratio: 16/9;
  border-radius: 0.5rem;
  overflow: hidden;
  background: #f3f4f6;
  flex-shrink: 0;
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.preview-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: #1a1a1a;
}

.preview-description {
  margin: 0;
  color: #6b7280;
  font-size: 0.9375rem;
  line-height: 1.6;
}

.preview-source {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.5rem;
}

.source-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.source-item .label {
  color: #6b7280;
  font-weight: 500;
}

.source-item .value {
  color: #1a1a1a;
  font-weight: 600;
}

.preview-skus h5 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 700;
  color: #1a1a1a;
}

.sku-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.sku-item {
  padding: 1rem;
  background: linear-gradient(135deg, #f9fafb 0%, #ffffff 100%);
  border: 1px solid #e5e7eb;
  border-left: 3px solid #ff6b2c;
  border-radius: 0.5rem;
}

.sku-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.sku-code {
  font-family: "Courier New", monospace;
  font-weight: 700;
  color: #1a1a1a;
  font-size: 0.875rem;
}

.sku-price {
  font-weight: 700;
  color: #ff6b2c;
  font-size: 0.9375rem;
}

.sku-details {
  display: flex;
  justify-content: space-between;
  font-size: 0.8125rem;
  color: #6b7280;
}

.btn-import {
  width: 100%;
  padding: 1rem;
  background: linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%);
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: auto;
}

.btn-import:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(255, 107, 44, 0.4);
}

.btn-import:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner-sm {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

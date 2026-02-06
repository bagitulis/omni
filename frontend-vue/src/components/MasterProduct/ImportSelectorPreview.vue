<template>
  <div class="preview-panel" :class="{ empty: !selectedProduct }">
    <!-- Empty State -->
    <div v-if="!selectedProduct" class="preview-empty">
      <span class="preview-icon">👈</span>
      <h3>Pilih Produk</h3>
      <p>Pilih produk dari daftar untuk melihat preview import</p>
    </div>

    <!-- Content State -->
    <div v-else class="preview-content">
      <div class="preview-header">
        <h3>Preview Import</h3>
        <button @click="$emit('close')" class="btn-close">✕</button>
      </div>

      <!-- Preview Loading -->
      <div v-if="loading" class="preview-loading">
        <div class="spinner"></div>
        <p>Memuat preview...</p>
      </div>

      <!-- Preview Data -->
      <ImportPreviewDetails
        v-else-if="previewData"
        :preview-data="previewData"
        :importing="importing"
        @import="$emit('import')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ShopeeProduct } from "./ImportSelector.types";
import type { ImportPreviewResponse } from "@/services/masterProductService";
import ImportPreviewDetails from "./ImportPreviewDetails.vue";

defineProps<{
  selectedProduct: ShopeeProduct | null;
  previewData: ImportPreviewResponse | null;
  loading: boolean;
  importing: boolean;
}>();

defineEmits<{
  (e: "close"): void;
  (e: "import"): void;
}>();
</script>

<style scoped>
.preview-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e5e7eb;
  overflow: hidden;
  height: 100%;
}

.preview-panel.empty {
  border: 2px dashed #d1d5db;
}

.preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 3rem;
  text-align: center;
}

.preview-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.preview-empty h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: #1a1a1a;
}

.preview-empty p {
  margin: 0;
  color: #6b7280;
}

.preview-content {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
  color: white;
  flex-shrink: 0;
}

.preview-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.25rem;
  transition: all 0.2s;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: rotate(90deg);
}

.preview-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
  color: #6b7280;
  flex: 1;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f4f6;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1024px) {
  .preview-panel {
    min-height: 500px;
  }
}
</style>

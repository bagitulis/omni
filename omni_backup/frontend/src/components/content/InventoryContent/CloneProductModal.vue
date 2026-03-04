<template>
  <Modal :is-open="show" title="Clone Product" @close="handleClose">
    <div class="clone-modal-content">
      <!-- Product Info -->
      <div class="product-info">
        <div class="info-row">
          <span class="label">SKU:</span>
          <span class="value">{{ sku }}</span>
        </div>
        <div class="info-row">
          <span class="label">Source:</span>
          <span class="platform-badge" :class="sourcePlatform">
            {{ getPlatformInfo(sourcePlatform).name }}
          </span>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <i class="pi pi-spin pi-spinner" aria-hidden="true"></i>
        <span>Loading product data...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state" role="alert">
        <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
        <span>{{ error }}</span>
      </div>

      <!-- Product Data Loaded -->
      <template v-else-if="productData">
        <div class="product-preview">
          <h4>Product: {{ productData.title }}</h4>
          <div v-if="productData.images.length" class="image-preview">
            <img :src="productData.images[0]" :alt="productData.title" />
          </div>
          <div class="sku-count">
            {{ productData.skus.length }} SKU variant(s)
          </div>
        </div>

        <!-- Target Selection -->
        <div class="target-selection">
          <h4>Clone to:</h4>
          <div v-if="availableTargets.length === 0" class="no-targets">
            <i class="pi pi-check-circle" aria-hidden="true"></i>
            Product already exists in all platforms
          </div>
          <div v-else class="target-list">
            <label
              v-for="target in availableTargets"
              :key="target"
              class="target-option"
              :class="{ selected: selectedTargets.includes(target) }"
            >
              <input
                type="checkbox"
                :checked="selectedTargets.includes(target)"
                @change="toggleTarget(target)"
              />
              <span class="platform-badge" :class="target">
                {{ getPlatformInfo(target).name }}
              </span>
              <span class="status-badge not-found">Not yet created</span>
            </label>
          </div>
        </div>
      </template>
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="handleClose">
        Cancel
      </button>
      <button
        type="button"
        class="btn btn-primary"
        :disabled="!canProceed"
        @click="handleStartClone"
      >
        <i class="pi pi-copy" aria-hidden="true"></i>
        Clone to Selected Platform{{ selectedTargets.length > 1 ? 's' : '' }}
      </button>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Modal from "@/components/Modal.vue";
import type { Platform, CloneProductData } from "./composables/useProductClone";

const props = defineProps<{
  show: boolean;
  sku: string;
  sourcePlatform: Platform;
  loading: boolean;
  error: string | null;
  productData: CloneProductData | null;
  availableTargets: Platform[];
  selectedTargets: Platform[];
}>();

const emit = defineEmits<{
  close: [];
  toggleTarget: [platform: Platform];
  startClone: [targetPlatform: Platform];
}>();

const canProceed = computed(() => 
  !props.loading && 
  !props.error && 
  props.productData && 
  props.selectedTargets.length > 0
);

function getPlatformInfo(platform: Platform) {
  const info: Record<Platform, { name: string }> = {
    shopee: { name: "Shopee" },
    lazada: { name: "Lazada" },
    tiktok: { name: "TikTok" },
  };
  return info[platform];
}

function toggleTarget(platform: Platform) {
  emit("toggleTarget", platform);
}

function handleClose() {
  emit("close");
}

function handleStartClone() {
  if (props.selectedTargets.length > 0) {
    emit("startClone", props.selectedTargets[0]);
  }
}
</script>

<style scoped>
.clone-modal-content {
  min-height: 200px;
}

.product-info {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 600;
  color: #666;
  min-width: 60px;
}

.value {
  font-family: monospace;
  background: #e9ecef;
  padding: 2px 8px;
  border-radius: 4px;
}

.platform-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 600;
}

.platform-badge.shopee {
  background: #fff5f5;
  color: #f53d2d;
}

.platform-badge.lazada {
  background: #e0e0ff;
  color: #0f146d;
}

.platform-badge.tiktok {
  background: #e8e8e8;
  color: #333;
}

.loading-state,
.error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 24px;
  color: #666;
}

.error-state {
  background: #fef2f2;
  color: #991b1b;
  border-radius: 8px;
}

.loading-state i {
  font-size: 24px;
  color: #3b82f6;
}

.product-preview {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.product-preview h4 {
  margin: 0 0 12px 0;
  color: #333;
}

.image-preview {
  width: 80px;
  height: 80px;
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 8px;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.sku-count {
  font-size: 13px;
  color: #666;
}

.target-selection {
  margin-top: 16px;
}

.target-selection h4 {
  margin: 0 0 12px 0;
  color: #333;
}

.no-targets {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
  background: #d4edda;
  color: #155724;
  border-radius: 8px;
}

.target-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.target-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.target-option:hover {
  border-color: #3b82f6;
  background: #f8f9fa;
}

.target-option.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.target-option input {
  width: 18px;
  height: 18px;
}

.status-badge {
  margin-left: auto;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
}

.status-badge.not-found {
  background: #fef2f2;
  color: #991b1b;
}

/* Footer buttons */
.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #d1d5db;
}
</style>

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="clone-preview-overlay" @click.self="close">
      <div class="clone-preview-modal">
        <!-- Header -->
        <div class="modal-header">
          <h2>Preview Clone</h2>
          <button @click="close" class="btn-close">&times;</button>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="loading-container">
          <div class="spinner"></div>
          <p>Memuat preview...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="error-container">
          <span class="error-icon">!</span>
          <p>{{ error }}</p>
          <button @click="loadPreview" class="btn-retry">Coba Lagi</button>
        </div>

        <!-- Preview Content -->
        <div v-else-if="previewData" class="preview-content">
          <!-- Conflict Warning -->
          <div v-if="previewData.has_conflict" class="conflict-warning">
            <span class="warning-icon">!</span>
            <div class="warning-text">
              <strong>Produk sudah ada di {{ targetPlatformLabel }}</strong>
              <p>SKU ini sudah terdaftar. Pilih untuk update atau buat baru.</p>
            </div>
          </div>

          <!-- Adjustment Info -->
          <div
            v-if="previewData.adjustments?.title_will_truncate"
            class="adjustment-info"
          >
            <span class="info-icon">i</span>
            <div class="info-text">
              <strong>Judul akan dipotong</strong>
              <p>
                {{ targetPlatformLabel }} memiliki batas
                {{ previewData.adjustments.title_limit }} karakter
              </p>
            </div>
          </div>

          <!-- Two Column Comparison -->
          <div class="comparison-grid">
            <!-- Source Panel -->
            <div class="panel source-panel">
              <div class="panel-header">
                <span class="platform-badge" :class="sourcePlatform">
                  {{ platformEmoji(sourcePlatform) }} {{ sourcePlatformLabel }}
                </span>
                <span class="label">Sumber</span>
              </div>
              <div class="panel-body">
                <div class="product-image">
                  <img
                    v-if="previewData.source_product?.images?.[0]"
                    :src="previewData.source_product.images[0]"
                    alt="Product"
                  />
                  <div v-else class="no-image">No Image</div>
                </div>
                <div class="product-info">
                  <h3 class="product-name">
                    {{ previewData.source_product?.name }}
                  </h3>
                  <div class="product-details">
                    <div class="detail-row">
                      <span class="label">SKU:</span>
                      <span class="value">{{
                        previewData.source_product?.sku || "-"
                      }}</span>
                    </div>
                    <div class="detail-row">
                      <span class="label">Harga:</span>
                      <span class="value price">{{
                        formatPrice(previewData.source_product?.price)
                      }}</span>
                    </div>
                    <div class="detail-row">
                      <span class="label">Stok:</span>
                      <span class="value">{{
                        previewData.source_product?.stock || 0
                      }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Arrow -->
            <div class="arrow-container">
              <span class="arrow">&rarr;</span>
            </div>

            <!-- Target Panel -->
            <div
              class="panel target-panel"
              :class="{ conflict: previewData.has_conflict }"
            >
              <div class="panel-header">
                <span class="platform-badge" :class="targetPlatform">
                  {{ platformEmoji(targetPlatform) }} {{ targetPlatformLabel }}
                </span>
                <span class="label">{{
                  previewData.has_conflict ? "Sudah Ada" : "Target"
                }}</span>
              </div>
              <div class="panel-body">
                <!-- If conflict, show existing product -->
                <template
                  v-if="previewData.has_conflict && previewData.target_product"
                >
                  <div class="product-image">
                    <img
                      v-if="previewData.target_product?.images?.[0]"
                      :src="previewData.target_product.images[0]"
                      alt="Product"
                    />
                    <div v-else class="no-image">No Image</div>
                  </div>
                  <div class="product-info">
                    <h3 class="product-name">
                      {{ previewData.target_product?.name }}
                    </h3>
                    <div class="product-details">
                      <div class="detail-row">
                        <span class="label">SKU:</span>
                        <span class="value">{{
                          previewData.target_product?.sku || "-"
                        }}</span>
                      </div>
                      <div
                        class="detail-row"
                        :class="{ different: hasDifference('price') }"
                      >
                        <span class="label">Harga:</span>
                        <span class="value price">{{
                          formatPrice(previewData.target_product?.price)
                        }}</span>
                      </div>
                      <div
                        class="detail-row"
                        :class="{ different: hasDifference('stock') }"
                      >
                        <span class="label">Stok:</span>
                        <span class="value">{{
                          previewData.target_product?.stock || 0
                        }}</span>
                      </div>
                    </div>
                  </div>
                </template>

                <!-- If no conflict, show preview -->
                <template v-else>
                  <div class="product-image">
                    <img
                      v-if="previewData.source_product?.images?.[0]"
                      :src="previewData.source_product.images[0]"
                      alt="Product"
                    />
                    <div v-else class="no-image">No Image</div>
                  </div>
                  <div class="product-info">
                    <h3
                      class="product-name"
                      :class="{
                        truncated: previewData.adjustments?.title_will_truncate,
                      }"
                    >
                      {{
                        previewData.adjustments?.adjusted_title ||
                        previewData.source_product?.name
                      }}
                    </h3>
                    <div
                      v-if="previewData.adjustments?.title_will_truncate"
                      class="char-count"
                    >
                      {{
                        (previewData.adjustments?.adjusted_title || "").length
                      }}/{{ previewData.adjustments?.title_limit }} karakter
                    </div>
                    <div class="product-details">
                      <div class="detail-row">
                        <span class="label">SKU:</span>
                        <span class="value">{{
                          previewData.source_product?.sku || "-"
                        }}</span>
                      </div>
                      <div class="detail-row">
                        <span class="label">Harga:</span>
                        <span class="value price">{{
                          formatPrice(previewData.source_product?.price)
                        }}</span>
                      </div>
                      <div class="detail-row">
                        <span class="label">Stok:</span>
                        <span class="value">{{
                          previewData.source_product?.stock || 0
                        }}</span>
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </div>

          <!-- Differences List -->
          <div
            v-if="previewData.differences?.length"
            class="differences-section"
          >
            <h4>Perbedaan</h4>
            <div class="differences-list">
              <div
                v-for="diff in previewData.differences"
                :key="diff.field"
                class="diff-item"
              >
                <span class="diff-field">{{
                  formatFieldName(diff.field)
                }}</span>
                <span class="diff-source">{{ diff.source_value }}</span>
                <span class="diff-arrow">&rarr;</span>
                <span class="diff-target">{{ diff.target_value }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="modal-footer">
          <button @click="close" class="btn btn-secondary">Batal</button>
          <button
            v-if="previewData?.has_conflict"
            @click="handleUpdate"
            class="btn btn-warning"
            :disabled="cloning"
          >
            <span v-if="cloning" class="spinner-sm"></span>
            Update Existing
          </button>
          <button
            @click="handleClone"
            class="btn btn-primary"
            :disabled="cloning"
          >
            <span v-if="cloning" class="spinner-sm"></span>
            {{
              previewData?.has_conflict
                ? "Clone Anyway"
                : "Clone ke " + targetPlatformLabel
            }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import apiService from "@/services/api";

// Types
interface ProductSummary {
  platform: string;
  item_id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
  images: string[];
}

interface Difference {
  field: string;
  source_value: string;
  target_value: string;
}

interface AdjustmentInfo {
  title_will_truncate: boolean;
  desc_will_truncate: boolean;
  original_title?: string;
  adjusted_title?: string;
  title_limit: number;
  desc_limit: number;
}

interface PreviewData {
  has_conflict: boolean;
  source_product?: ProductSummary;
  target_product?: ProductSummary;
  differences?: Difference[];
  adjustments?: AdjustmentInfo;
}

// Props
const props = defineProps<{
  modelValue: boolean;
  sourcePlatform: string;
  sourceItemId: string;
  targetPlatform: string;
  sku?: string;
}>();

// Emits
const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "clone", data: { updateExisting: boolean }): void;
}>();

// State
const loading = ref(false);
const error = ref("");
const previewData = ref<PreviewData | null>(null);
const cloning = ref(false);

// Computed
const sourcePlatformLabel = computed(() => {
  const labels: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok",
  };
  return labels[props.sourcePlatform] || props.sourcePlatform;
});

const targetPlatformLabel = computed(() => {
  const labels: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok",
  };
  return labels[props.targetPlatform] || props.targetPlatform;
});

// Methods
const platformEmoji = (platform: string) => {
  const emojis: Record<string, string> = {
    shopee: "",
    lazada: "",
    tiktok: "",
  };
  return emojis[platform] || "";
};

const formatPrice = (price?: number) => {
  if (!price) return "Rp 0";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};

const formatFieldName = (field: string) => {
  const names: Record<string, string> = {
    name: "Nama",
    price: "Harga",
    stock: "Stok",
    image_count: "Jumlah Gambar",
  };
  return names[field] || field;
};

const hasDifference = (field: string) => {
  return previewData.value?.differences?.some((d) => d.field === field);
};

const loadPreview = async () => {
  loading.value = true;
  error.value = "";

  try {
    const params = new URLSearchParams({
      source_platform: props.sourcePlatform,
      target_platform: props.targetPlatform,
      source_item_id: props.sourceItemId,
    });
    if (props.sku) {
      params.append("sku", props.sku);
    }

    const response = await apiService.get<{
      success: boolean;
      data: PreviewData;
    }>(`/clone/preview?${params.toString()}`);

    if (response.success) {
      previewData.value = response.data;
    } else {
      throw new Error("Failed to load preview");
    }
  } catch (err: any) {
    console.error("Error loading preview:", err);
    error.value = err.message || "Gagal memuat preview";
  } finally {
    loading.value = false;
  }
};

const close = () => {
  emit("update:modelValue", false);
};

const handleClone = () => {
  cloning.value = true;
  emit("clone", { updateExisting: false });
};

const handleUpdate = () => {
  cloning.value = true;
  emit("clone", { updateExisting: true });
};

// Watch for modal open
watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal) {
      loadPreview();
    } else {
      previewData.value = null;
      error.value = "";
      cloning.value = false;
    }
  },
);
</script>

<style scoped>
.clone-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.clone-preview-modal {
  background: white;
  border-radius: 1rem;
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
  color: white;
  border-radius: 1rem 1rem 0 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}

.btn-close {
  width: 36px;
  height: 36px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: rotate(90deg);
}

/* Loading & Error States */
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  gap: 1rem;
  color: #6b7280;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #f3f4f6;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
  margin-right: 0.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-icon {
  width: 48px;
  height: 48px;
  background: #fee2e2;
  color: #ef4444;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
}

.btn-retry {
  padding: 0.75rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  font-weight: 600;
}

/* Preview Content */
.preview-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

/* Warning & Info Banners */
.conflict-warning,
.adjustment-info {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}

.conflict-warning {
  background: #fef3c7;
  border: 1px solid #f59e0b;
}

.adjustment-info {
  background: #e0f2fe;
  border: 1px solid #0ea5e9;
}

.warning-icon,
.info-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  flex-shrink: 0;
}

.warning-icon {
  background: #f59e0b;
  color: white;
}

.info-icon {
  background: #0ea5e9;
  color: white;
}

.warning-text strong,
.info-text strong {
  display: block;
  margin-bottom: 0.25rem;
}

.warning-text p,
.info-text p {
  margin: 0;
  font-size: 0.875rem;
  opacity: 0.8;
}

/* Comparison Grid */
.comparison-grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1rem;
  align-items: start;
}

.arrow-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 100px;
}

.arrow {
  font-size: 2rem;
  color: #9ca3af;
}

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

/* Differences Section */
.differences-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.differences-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.875rem;
  font-weight: 700;
  color: #1a1a1a;
}

.differences-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.diff-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.diff-field {
  font-weight: 600;
  min-width: 100px;
}

.diff-source {
  color: #6b7280;
}

.diff-arrow {
  color: #9ca3af;
}

.diff-target {
  color: #f59e0b;
  font-weight: 600;
}

/* Footer */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
  border-radius: 0 0 1rem 1rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #d1d5db;
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
}

.btn-primary {
  background: linear-gradient(135deg, #ff6b2c 0%, #ff5511 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 107, 44, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .comparison-grid {
    grid-template-columns: 1fr;
  }

  .arrow-container {
    padding: 0.5rem 0;
  }

  .arrow {
    transform: rotate(90deg);
  }
}
</style>

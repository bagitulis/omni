<template>
  <div class="import-selector">
    <!-- Search & Filter Bar -->
    <div class="filter-bar">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Cari produk Shopee..."
          class="search-input"
          @input="debouncedSearch"
        />
        <span v-if="searchQuery" class="clear-icon" @click="clearSearch"
          >✕</span
        >
      </div>
      <button @click="fetchProducts" class="btn-refresh" :disabled="loading">
        <span class="refresh-icon" :class="{ spinning: loading }">↻</span>
        Refresh
      </button>
    </div>

    <!-- Two-Column Layout -->
    <div class="import-grid">
      <!-- Left: Product List -->
      <div class="product-list-panel">
        <div class="panel-header">
          <h3>Produk Shopee</h3>
          <span class="product-count"
            >{{ filteredProducts.length }} produk</span
          >
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Memuat produk...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredProducts.length === 0" class="empty-state">
          <span class="empty-icon">📦</span>
          <p v-if="searchQuery">Tidak ada produk yang cocok dengan pencarian</p>
          <p v-else>Belum ada produk Shopee di database</p>
        </div>

        <!-- Product Cards -->
        <div v-else class="product-cards-list">
          <div
            v-for="product in filteredProducts"
            :key="product.item_id"
            class="product-card"
            :class="{ selected: selectedProduct?.item_id === product.item_id }"
            @click="selectProduct(product)"
          >
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
        </div>
      </div>

      <!-- Right: Preview Panel -->
      <div class="preview-panel" :class="{ empty: !selectedProduct }">
        <div v-if="!selectedProduct" class="preview-empty">
          <span class="preview-icon">👈</span>
          <h3>Pilih Produk</h3>
          <p>Pilih produk dari daftar untuk melihat preview import</p>
        </div>

        <div v-else class="preview-content">
          <div class="preview-header">
            <h3>Preview Import</h3>
            <button @click="closePreview" class="btn-close">✕</button>
          </div>

          <!-- Preview Loading -->
          <div v-if="previewLoading" class="preview-loading">
            <div class="spinner"></div>
            <p>Memuat preview...</p>
          </div>

          <!-- Preview Data -->
          <div v-else-if="previewData" class="preview-details">
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
              <h5>
                SKUs yang akan diimport ({{ previewData.skus?.length || 0 }})
              </h5>
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
                    <span class="sku-variant">{{
                      sku.variant_name || "Default"
                    }}</span>
                    <span class="sku-stock">Stok: {{ sku.stock }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Import Button -->
            <button
              @click="confirmImport"
              class="btn-import"
              :disabled="importing"
            >
              <span v-if="importing" class="spinner-sm"></span>
              {{ importing ? "Mengimport..." : "Import ke Master Product" }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Messages -->
    <transition name="fade">
      <div v-if="successMessage" class="message success">
        <span class="message-icon">✓</span>
        {{ successMessage }}
      </div>
    </transition>
    <transition name="fade">
      <div v-if="errorMessage" class="message error">
        <span class="message-icon">✕</span>
        {{ errorMessage }}
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import apiService from "@/services/api";
import masterProductService from "@/services/masterProductService";
import type { ImportPreviewResponse } from "@/services/masterProductService";

// Types
interface ShopeeProduct {
  id: number;
  item_id: number;
  name: string;
  description?: string;
  images?: string[];
  price?: number;
  price_info?: {
    current_price?: number;
  };
  models?: Array<{
    model_id: number;
    model_sku: string;
    tier_index?: number[];
  }>;
}

// State
const products = ref<ShopeeProduct[]>([]);
const selectedProduct = ref<ShopeeProduct | null>(null);
const previewData = ref<ImportPreviewResponse | null>(null);
const loading = ref(false);
const previewLoading = ref(false);
const importing = ref(false);
const searchQuery = ref("");
const successMessage = ref("");
const errorMessage = ref("");
let searchTimeout: number | null = null;

// Computed
const filteredProducts = computed(() => {
  if (!searchQuery.value.trim()) return products.value;

  const query = searchQuery.value.toLowerCase();
  return products.value.filter(
    (p) =>
      p.name.toLowerCase().includes(query) ||
      p.item_id.toString().includes(query),
  );
});

// Methods
const fetchProducts = async () => {
  loading.value = true;
  errorMessage.value = "";

  try {
    const response = await apiService.get<{
      success: boolean;
      data: ShopeeProduct[];
    }>("/shopee/db/products");

    if (response.success) {
      products.value = response.data || [];
    } else {
      throw new Error("Failed to fetch products");
    }
  } catch (error: any) {
    console.error("Error fetching Shopee products:", error);
    errorMessage.value = error.message || "Gagal memuat produk Shopee";
  } finally {
    loading.value = false;
  }
};

const selectProduct = async (product: ShopeeProduct) => {
  selectedProduct.value = product;
  previewData.value = null;
  previewLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await masterProductService.previewImport(
      "shopee",
      product.item_id.toString(),
    );
    previewData.value = response;
  } catch (error: any) {
    console.error("Error loading preview:", error);
    errorMessage.value = error.message || "Gagal memuat preview";
    previewData.value = null;
  } finally {
    previewLoading.value = false;
  }
};

const closePreview = () => {
  selectedProduct.value = null;
  previewData.value = null;
};

const confirmImport = async () => {
  if (!selectedProduct.value) return;

  importing.value = true;
  errorMessage.value = "";
  successMessage.value = "";

  try {
    const response = await masterProductService.importProduct(
      "shopee",
      selectedProduct.value.item_id.toString(),
    );

    successMessage.value = `Berhasil import produk dengan ${response.skus_imported} SKU!`;

    // Refresh product list and close preview
    setTimeout(() => {
      closePreview();
      fetchProducts();
      successMessage.value = "";
    }, 2000);
  } catch (error: any) {
    console.error("Error importing product:", error);
    errorMessage.value = error.message || "Gagal import produk";
  } finally {
    importing.value = false;
  }
};

const debouncedSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = window.setTimeout(() => {
    // Search is reactive via computed
  }, 300);
};

const clearSearch = () => {
  searchQuery.value = "";
};

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

// Lifecycle
onMounted(() => {
  fetchProducts();
});
</script>

<style scoped>
.import-selector {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 1rem;
  align-items: center;
  padding: 1rem;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
  border-radius: 0.75rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.search-input-wrapper {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 1rem;
  font-size: 1.125rem;
  opacity: 0.6;
}

.search-input {
  width: 100%;
  padding: 0.875rem 3rem 0.875rem 3rem;
  border: 2px solid transparent;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 0.5rem;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #ff6b2c;
  background: white;
  box-shadow: 0 0 0 3px rgba(255, 107, 44, 0.1);
}

.clear-icon {
  position: absolute;
  right: 1rem;
  cursor: pointer;
  font-size: 1.125rem;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.clear-icon:hover {
  opacity: 1;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-refresh:hover:not(:disabled) {
  background: #ff5511;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 107, 44, 0.3);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-icon {
  font-size: 1.25rem;
  transition: transform 0.3s;
}

.refresh-icon.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Import Grid */
.import-grid {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 1.5rem;
  flex: 1;
  min-height: 0;
}

/* Product List Panel */
.product-list-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e5e7eb;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 2px solid #ff6b2c;
}

.panel-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #1a1a1a;
}

.product-count {
  font-size: 0.875rem;
  font-weight: 600;
  color: #6b7280;
  background: rgba(255, 107, 44, 0.1);
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
}

.product-cards-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

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

/* Preview Panel */
.preview-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e5e7eb;
  overflow: hidden;
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
}

.btn-import:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(255, 107, 44, 0.4);
}

.btn-import:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Loading States */
.loading-state,
.preview-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
  color: #6b7280;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f4f6;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  text-align: center;
  color: #6b7280;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* Messages */
.message {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.message.success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.message.error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.message-icon {
  font-size: 1.25rem;
  font-weight: 700;
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(1rem);
}

/* Responsive */
@media (max-width: 1024px) {
  .import-grid {
    grid-template-columns: 1fr;
  }

  .preview-panel {
    min-height: 500px;
  }
}

@media (max-width: 640px) {
  .filter-bar {
    flex-direction: column;
  }

  .btn-refresh {
    width: 100%;
    justify-content: center;
  }
}
</style>

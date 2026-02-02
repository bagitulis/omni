<template>
  <div class="import-selector">
    <!-- Main Grid -->
    <div class="import-grid">
      <!-- Left: Product List -->
      <ImportSelectorList
        :products="products"
        :loading="loading"
        :selected-product="selectedProduct"
        @select="selectProduct"
        @refresh="fetchProducts"
      />

      <!-- Right: Preview Panel -->
      <ImportSelectorPreview
        :selected-product="selectedProduct"
        :preview-data="previewData"
        :loading="previewLoading"
        :importing="importing"
        @close="closePreview"
        @import="confirmImport"
      />
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
import { ref, onMounted } from "vue";
import apiService from "@/services/api";
import masterProductService from "@/services/masterProductService";
import type { ImportPreviewResponse } from "@/services/masterProductService";
import type { ShopeeProduct } from "./ImportSelector.types";

import ImportSelectorList from "./ImportSelectorList.vue";
import ImportSelectorPreview from "./ImportSelectorPreview.vue";

// State
const products = ref<ShopeeProduct[]>([]);
const selectedProduct = ref<ShopeeProduct | null>(null);
const previewData = ref<ImportPreviewResponse | null>(null);
const loading = ref(false);
const previewLoading = ref(false);
const importing = ref(false);
const successMessage = ref("");
const errorMessage = ref("");

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

.import-grid {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 1.5rem;
  flex: 1;
  min-height: 0;
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

@media (max-width: 1024px) {
  .import-grid {
    grid-template-columns: 1fr;
  }
}
</style>

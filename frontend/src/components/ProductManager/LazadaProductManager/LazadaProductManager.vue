<template>
  <div class="product-manager-container lazada-product-manager">
    <!-- Header -->
    <div class="manager-header">
      <h2>
        <i class="pi pi-shopping-cart" aria-hidden="true"></i> Lazada Product
        Manager
      </h2>
      <div class="header-actions">
        <button
          @click="showAddProductModal = true"
          class="btn btn-success"
        >
          <i class="pi pi-plus" aria-hidden="true"></i>
          Add Product
        </button>
        <button
          @click="fetchProductsFromAPI"
          :disabled="loading"
          class="btn btn-primary"
        >
          <i
            :class="loading ? 'pi pi-spin pi-spinner' : 'pi pi-cloud-download'"
            aria-hidden="true"
          ></i>
          {{ loading ? "Syncing..." : "Sync from API" }}
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs-container">
      <div class="tabs-header">
        <button
          @click="
            activeTab = 'db-products';
            updateURLQuery('db-products');
            reloadFromDB();
          "
          :class="['tab-btn', { active: activeTab === 'db-products' }]"
        >
          <i class="pi pi-database"></i>
          Master Product
          <span class="tab-badge">{{ products.length }}</span>
        </button>
        <button
          @click="
            activeTab = 'product-detail';
            updateURLQuery('product-detail');
          "
          :class="['tab-btn', { active: activeTab === 'product-detail' }]"
        >
          <i class="pi pi-file-pdf"></i>
          Item Specification
        </button>
      </div>
    </div>

    <!-- Tab Content -->
    <div class="tabs-content">
      <!-- DB Products Tab -->
      <div v-if="activeTab === 'db-products'" class="tab-content">
        <div class="content-header">
          <div class="search-and-filter">
            <div class="search-box">
              <label for="lazada-product-search" class="sr-only"
                >Search products</label
              >
              <input
                id="lazada-product-search"
                v-model="searchQuery"
                type="text"
                placeholder="Search by Item ID, SKU ID, SKU Name, Product Name, or Variant..."
                class="search-input"
              />
            </div>
            <div class="filter-controls">
              <button
                @click="showFilterPanel = !showFilterPanel"
                :class="[
                  'btn-filter',
                  { active: showFilterPanel || hasActiveFilters },
                ]"
              >
                <i class="pi pi-sliders-h"></i>
                Filters {{ hasActiveFilters ? "✓" : "" }}
              </button>
            </div>
          </div>

          <FilterPanel
            :showFilterPanel="showFilterPanel"
            :columnLabels="config.columnLabels"
            :columnFilters="columnFilters"
            :visibleColumns="visibleColumns"
            :filterableFields="config.filterableFields"
            :availableStatuses="config.availableStatuses"
            @update:columnFilters="columnFilters = $event"
            @update:visibleColumns="visibleColumns = $event"
            @reset-visibility="resetColumnVisibility"
            @clear-filters="clearAllFilters"
          />
        </div>

        <ProductTable
          :products="paginatedProducts"
          :columnFields="config.columnFields"
          :columnLabels="config.columnLabels"
          :visibleColumns="visibleColumns"
          :currentPage="currentPage"
          :totalPages="totalPages"
          :loading="loading"
          :keyFields="['item_id', 'sku_id']"
          @next-page="currentPage++"
          @prev-page="currentPage--"
        >
          <template #cell-status="{ value }">
            <span :class="['status-badge', value?.toLowerCase()]">
              {{ value }}
            </span>
          </template>
        </ProductTable>
      </div>

      <!-- Product Detail Tab -->
      <div v-if="activeTab === 'product-detail'" class="tab-content">
        <div class="content-header">
          <h3>Get Product Detail</h3>
          <div class="detail-inputs">
            <input
              v-model="detailItemId"
              type="number"
              placeholder="Enter Item ID"
              class="search-input"
            />
            <button
              @click="handleFetchProductDetail"
              :disabled="!detailItemId || detailLoading"
              class="btn btn-primary"
            >
              <i
                :class="
                  detailLoading ? 'pi pi-spin pi-spinner' : 'pi pi-search'
                "
              ></i>
              {{ detailLoading ? "Loading..." : "Search" }}
            </button>
          </div>
        </div>

        <div v-if="detailLoading" class="loading-spinner">
          <i class="pi pi-spin pi-spinner"></i>
          <p>Loading product detail...</p>
        </div>

        <div v-else-if="selectedProduct" class="product-detail-card">
          <div class="detail-header">
            <h3>{{ selectedProduct.name || "Product" }}</h3>
            <span
              :class="['status-badge', selectedProduct.status?.toLowerCase()]"
            >
              {{ selectedProduct.status }}
            </span>
          </div>
          <div class="detail-grid">
            <div class="detail-section">
              <h4>Basic Info</h4>
              <p><strong>Item ID:</strong> {{ selectedProduct.itemId }}</p>
              <p><strong>Brand:</strong> {{ selectedProduct.brand || "-" }}</p>
            </div>
            <div class="detail-section">
              <h4>Pricing & Stock</h4>
              <div
                v-if="selectedProduct.skus && selectedProduct.skus.length > 0"
              >
                <div
                  v-for="(sku, idx) in selectedProduct.skus"
                  :key="idx"
                  class="sku-item"
                >
                  <p>
                    <strong>SKU {{ Number(idx) + 1 }}:</strong>
                    {{ sku.sellerSku }}
                  </p>
                  <p v-if="sku.variantName">
                    <strong>Variant:</strong> {{ sku.variantName }}
                  </p>
                  <p>
                    <strong>Price:</strong>
                    {{ sku.price ? "Rp " + formatNumber(sku.price) : "-" }}
                  </p>
                  <p><strong>Stock:</strong> {{ sku.quantity }}</p>
                </div>
              </div>
              <p v-else>No SKU information</p>
            </div>
            <div class="detail-section full-width">
              <h4>Description</h4>
              <p v-if="selectedProduct.description" class="description-text">
                {{ stripHtml(selectedProduct.description) }}
              </p>
              <p v-else>-</p>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <i class="pi pi-inbox"></i>
          <p>No product selected</p>
          <small>Enter an Item ID and click Search to view details</small>
        </div>
      </div>
    </div>

    <!-- Error Message -->
    <div v-if="errorMessage" class="error-banner">
      <i class="pi pi-exclamation-circle" aria-hidden="true"></i>
      {{ errorMessage }}
      <button
        @click="errorMessage = ''"
        class="btn-close"
        type="button"
        aria-label="Close error message"
      >
        <i class="pi pi-times" aria-hidden="true"></i>
      </button>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="success-banner">
      <i class="pi pi-check-circle"></i>
      {{ successMessage }}
    </div>

    <!-- Add Product Modal -->
    <AddProductModal
      :show="showAddProductModal"
      platform="lazada"
      @close="showAddProductModal = false"
      @success="handleProductCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useProductManager } from "@/composables/useProductManager";
import { useProductDetail } from "@/composables/useProductDetail";
import { useProductSync } from "@/composables/useProductSync";
import { formatNumber, stripHtml } from "@/utils/productManagerUtils";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";
import FilterPanel from "../FilterPanel.vue";
import ProductTable from "../ProductTable.vue";
import AddProductModal from "../../AddProduct/AddProductModal.vue";
import "../ProductManagerCommon.styles.css";

const router = useRouter();
const route = useRoute();
const API_BASE_URL = getApiBaseUrl("/lazada");

const {
  products,
  errorMessage,
  successMessage,
  currentPage,
  showFilterPanel,
  searchQuery,
  columnFilters,
  visibleColumns,
  config,
  paginatedProducts,
  totalPages,
  hasActiveFilters,
  clearAllFilters,
  resetColumnVisibility,
  showError,
  showSuccess,
} = useProductManager({ platform: "lazada" });

const { loading, fetchProductsFromAPI: syncFromAPI } =
  useProductSync(API_BASE_URL);

// Load products from database directly
const loadDBProducts = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/db/products`, {
      headers: getAuthHeaders(),
    });
    const data = await response.json();
    if (data.success) {
      products.value = data.products || [];
      return { success: true, count: products.value.length };
    } else {
      return { success: false, error: data.error || "Failed to load products" };
    }
  } catch {
    return { success: false, error: "Error loading products" };
  }
};

const { detailItemId, detailLoading, selectedProduct, fetchProductDetail } =
  useProductDetail(API_BASE_URL);

const activeTab = ref("db-products");
const showAddProductModal = ref(false);

const updateURLQuery = (tabValue: string): void => {
  const query = { ...route.query, tab: tabValue };
  router
    .push({
      path: route.path,
      query,
    })
    .catch(() => {
      // Ignore navigation errors
    });
};

const fetchProductsFromAPI = async () => {
  const result = await syncFromAPI();
  if (result.success) {
    const { productCount, savedCount } = result;
    showSuccess(
      `✅ Synced ${productCount} products from API | Saved: ${savedCount} products`,
    );
    const dbResult = await loadDBProducts();
    if (dbResult.success) {
      currentPage.value = 1;
    } else {
      showError(dbResult.error || "Failed to load products after sync");
    }
  } else {
    showError(result.error || "Failed to sync products");
  }
};

const reloadFromDB = async () => {
  const result = await loadDBProducts();
  if (!result.success) {
    showError(result.error || "Failed to load products");
  } else {
    currentPage.value = 1;
  }
};

const handleFetchProductDetail = async () => {
  const result = await fetchProductDetail();
  if (result?.success) {
    showSuccess("Product loaded");
  } else {
    showError(result?.error || "Product not found");
  }
};

const handleProductCreated = async () => {
  showSuccess("Product created successfully on Lazada!");
  await reloadFromDB();
};

onMounted(() => {
  // Read from URL query parameter
  const tabFromQuery = route.query.tab as string;
  const validTabs = ["db-products", "product-detail"];

  if (tabFromQuery && validTabs.includes(tabFromQuery)) {
    activeTab.value = tabFromQuery;
  }

  if (products.value.length === 0) {
    loadDBProducts();
  }
});

// Watch for URL query changes (browser back/forward, direct URL access)
watch(
  () => route.query.tab,
  (newTab) => {
    const validTabs = ["db-products", "product-detail"];
    const tabValue = newTab as string;

    if (
      tabValue &&
      validTabs.includes(tabValue) &&
      tabValue !== activeTab.value
    ) {
      activeTab.value = tabValue;
    }
  },
);
</script>

<style scoped>
.btn-success {
  background: #10b981;
  color: white;
}
.btn-success:hover {
  background: #059669;
}
</style>

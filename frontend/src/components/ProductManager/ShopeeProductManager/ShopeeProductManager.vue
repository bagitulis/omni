<template>
  <div class="shopee-product-manager">
    <!-- Header -->
    <div class="manager-header">
      <h2><i class="pi pi-shopping-cart"></i> Shopee Product Manager</h2>
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
          <i class="pi pi-database"></i> Master Product
          <span class="tab-badge">{{ products.length }}</span>
        </button>
        <button
          @click="
            activeTab = 'product-detail';
            updateURLQuery('product-detail');
          "
          :class="['tab-btn', { active: activeTab === 'product-detail' }]"
        >
          <i class="pi pi-file-pdf"></i> Item Specification
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
              <label for="shopee-product-search" class="sr-only"
                >Search products</label
              >
              <input
                id="shopee-product-search"
                v-model="searchQuery"
                type="text"
                placeholder="Search by Item ID, Model ID, SKU, Product Name, or Model Name..."
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
                <i class="pi pi-sliders-h"></i> Filters
                {{ hasActiveFilters ? "✓" : "" }}
              </button>
            </div>
          </div>

          <FilterPanel
            :showFilterPanel="showFilterPanel"
            :columnLabels="config.columnLabels"
            :columnFilters="columnFilters"
            :visibleColumns="visibleColumns"
            :filterableFields="config.filterableFields"
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
          :keyFields="['itemId', 'modelId']"
          @next-page="currentPage++"
          @prev-page="currentPage--"
        />
      </div>

      <!-- Product Detail Tab -->
      <ProductDetailTab
        v-if="activeTab === 'product-detail'"
        v-model:itemId="detailItemId"
        :loading="detailLoading"
        :product="selectedProduct"
        placeholder="Enter Item ID"
        input-type="number"
        empty-message="Enter an Item ID and click Search to view details"
        @fetch-detail="handleFetchProductDetail"
      />
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
      platform="shopee"
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
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";
import FilterPanel from "../FilterPanel.vue";
import ProductTable from "../ProductTable.vue";
import ProductDetailTab from "../ProductDetailTab.vue";
import AddProductModal from "../../AddProduct/AddProductModal.vue";
import "../ProductManagerCommon.styles.css";

const router = useRouter();
const route = useRoute();
const API_BASE_URL = getApiBaseUrl("/shopee");

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
} = useProductManager({ platform: "shopee" });

const { loading, fetchProductsFromAPI: syncFromAPI } =
  useProductSync(API_BASE_URL);

// Load products from database and sync to products ref from useProductManager
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
      `✅ Synced ${productCount} products from API | Saved: ${savedCount} products to database`,
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
    showError(result.error || "Failed to load products from database");
  } else {
    currentPage.value = 1;
  }
};

const handleFetchProductDetail = async () => {
  const result = await fetchProductDetail();
  if (result?.success) {
    showSuccess("Product detail loaded successfully");
  } else {
    showError(result?.error || "Failed to fetch product detail");
  }
};

const handleProductCreated = async () => {
  showSuccess("Product created successfully on Shopee!");
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
/* Shopee-specific overrides - common styles in ProductManagerCommon.styles.css */
.shopee-product-manager {
  padding: 20px;
  background: #f8f9fa;
  min-height: 100vh;
}
.btn-primary {
  background: #1976d2;
}
.btn-primary:hover:not(:disabled) {
  background: #1565c0;
}
.tab-btn:hover {
  color: #2c3e50;
}
.tab-btn.active {
  color: #1976d2;
  border-bottom-color: #1976d2;
}
.tab-btn.active .tab-badge {
  background: #1976d2;
}
.search-input:focus {
  border-color: #1976d2;
  box-shadow: 0 0 0 3px rgba(25, 118, 210, 0.1);
}
.btn-filter.active {
  background-color: #1976d2;
  border-color: #1976d2;
}
.detail-inputs button {
  background: #1976d2;
}
.detail-inputs button:hover:not(:disabled) {
  background: #1565c0;
}
.btn-success {
  background: #10b981;
  color: white;
}
.btn-success:hover {
  background: #059669;
}
</style>

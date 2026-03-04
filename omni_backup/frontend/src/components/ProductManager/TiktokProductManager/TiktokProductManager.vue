<template>
  <div class="product-manager-container tiktok-product-manager">
    <!-- Header -->
    <div class="manager-header">
      <h2><i class="pi pi-shopping-bag"></i> TikTok Product Manager</h2>
      <div class="header-actions">
        <button
          @click="showAddProductModal = true"
          class="btn btn-success"
        >
          <i class="pi pi-plus"></i>
          Add Product
        </button>
        <button
          @click="fetchProductsFromTikTokAPI"
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
            activeTab = 'master';
            updateURLQuery('master');
            loadProductsFromDB();
          "
          :class="['tab-btn', { active: activeTab === 'master' }]"
        >
          <i class="pi pi-database"></i> Master Product
          <span class="tab-badge">{{ products.length }}</span>
        </button>
        <button
          @click="
            activeTab = 'specification';
            updateURLQuery('specification');
          "
          :class="['tab-btn', { active: activeTab === 'specification' }]"
        >
          <i class="pi pi-file-pdf"></i> Item Specification
        </button>
      </div>
    </div>

    <!-- Tab Content -->
    <div class="tabs-content">
      <!-- Master Tab -->
      <div v-if="activeTab === 'master'" class="tab-content">
        <div class="content-header">
          <div class="search-and-filter">
            <div class="search-box">
              <label for="tiktok-product-search" class="sr-only"
                >Search products</label
              >
              <input
                id="tiktok-product-search"
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
            <span :class="['status-badge', value?.toLowerCase()]">{{
              value
            }}</span>
          </template>
        </ProductTable>

        <!-- Sync Status Message -->
        <div v-if="products.length > 0" class="sync-status">
          <span class="status-info"
            >✅ {{ products.length }} product SKU(s) synced from TikTok</span
          >
        </div>
      </div>

      <!-- Item Specification Tab -->
      <ProductDetailTab
        v-if="activeTab === 'specification'"
        v-model:itemId="detailItemId"
        :loading="detailLoading"
        :product="selectedProduct"
        placeholder="Enter Product ID"
        input-type="text"
        empty-message="Enter a Product ID and click Search to view details"
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

    <div v-if="successMessage" class="success-banner">
      <i class="pi pi-check-circle"></i>
      {{ successMessage }}
    </div>

    <!-- Add Product Modal -->
    <AddProductModal
      :show="showAddProductModal"
      platform="tiktok"
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
import FilterPanel from "../FilterPanel.vue";
import ProductTable from "../ProductTable.vue";
import ProductDetailTab from "../ProductDetailTab.vue";
import AddProductModal from "@/components/AddProduct/AddProductModal.vue";
import tiktokDbService from "@/services/tiktokDbService";
import "../ProductManagerCommon.styles.css";

const router = useRouter();
const route = useRoute();

const {
  products,
  loading,
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
} = useProductManager({ platform: "tiktok" });

const { detailItemId, detailLoading, selectedProduct } = useProductDetail("");

const activeTab = ref("master");
const showAddProductModal = ref(false);

const handleProductCreated = (productId: string) => {
  showSuccess(`✅ Product created successfully! ID: ${productId}`);
  loadProductsFromDB(); // Refresh product list
};

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

const loadProductsFromDB = async () => {
  try {
    const response = await tiktokDbService.getMasterProductsFromDB();
    if (response.success && response.products) {
      products.value = response.products;
    }
  } catch {
    showError("Failed to load products from database");
  }
};

const fetchProductsFromTikTokAPI = async () => {
  loading.value = true;
  try {
    const response = await tiktokDbService.searchProducts();
    if (response.success) {
      await loadProductsFromDB();
      showSuccess(
        `✅ Synced ${response.pageCount} page(s) - Total ${response.totalProducts} active products`,
      );
    } else {
      showError(response.error || "Failed to sync products");
    }
  } catch {
    showError("Error syncing products");
  } finally {
    loading.value = false;
  }
};

const handleFetchProductDetail = async () => {
  if (!detailItemId.value) {
    showError("Please enter a Product ID");
    return;
  }

  detailLoading.value = true;
  try {
    const response = await tiktokDbService.getProductById(detailItemId.value);
    if (response.success) {
      selectedProduct.value = response.data || response;
      showSuccess("✅ Product loaded");
    } else {
      selectedProduct.value = null;
      showError("Product not found");
    }
  } catch {
    showError("Error fetching product details");
  } finally {
    detailLoading.value = false;
  }
};

onMounted(() => {
  // Read from URL query parameter
  const tabFromQuery = route.query.tab as string;
  const validTabs = ["master", "specification"];

  if (tabFromQuery && validTabs.includes(tabFromQuery)) {
    activeTab.value = tabFromQuery;
  }

  if (products.value.length === 0) {
    loadProductsFromDB();
  }
});

// Watch for URL query changes (browser back/forward, direct URL access)
watch(
  () => route.query.tab,
  (newTab) => {
    const validTabs = ["master", "specification"];
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
/* TikTok-specific overrides - common styles in ProductManagerCommon.styles.css */
.tiktok-product-manager {
  padding: 20px;
  background-color: #f5f5f5;
}
.btn-primary {
  background-color: #667eea;
}
.btn-primary:hover:not(:disabled) {
  background-color: #5568d3;
}
.btn-success {
  background-color: #10b981;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}
.btn-success:hover:not(:disabled) {
  background-color: #059669;
}
.tab-btn:hover,
.tab-btn.active {
  color: #667eea;
}
.tab-btn.active {
  border-bottom-color: #667eea;
}
.tab-btn.active .tab-badge {
  background: #667eea;
}
.search-input:focus {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}
.btn-filter.active {
  background-color: #667eea;
  border-color: #667eea;
}
.loading-spinner i {
  color: #667eea;
}
.detail-inputs button {
  background: #667eea;
}
.detail-inputs button:hover:not(:disabled) {
  background: #5568d3;
}
</style>

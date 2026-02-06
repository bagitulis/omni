<template>
  <div class="master-product-page">
    <div class="main-layout">
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="'master-products'"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />
      <main class="main-content">
        <div class="master-product-list">
          <div class="page-header">
            <h1>Master Produk</h1>
            <div class="header-actions">
              <router-link
                to="/master-products/import"
                class="btn btn-secondary"
              >
                Import dari Shopee
              </router-link>
              <router-link to="/master-products/add" class="btn btn-primary">
                Tambah Produk
              </router-link>
            </div>
          </div>

          <!-- Search & Filter -->
          <div class="filter-bar">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari produk..."
              class="search-input"
              @input="debouncedSearch"
            />
            <select
              v-model="statusFilter"
              class="status-select"
              @change="fetchProducts"
            >
              <option value="">Semua Status</option>
              <option value="active">Active</option>
              <option value="draft">Draft</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <!-- Loading State -->
          <div v-if="loading" class="loading-container">
            <div class="spinner"></div>
            <p>Memuat produk...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="error-container">
            <p class="error-message">{{ error }}</p>
            <button @click="fetchProducts" class="btn btn-primary">
              Coba Lagi
            </button>
          </div>

          <!-- Empty State -->
          <div v-else-if="products.length === 0" class="empty-container">
            <p>Belum ada produk master.</p>
            <router-link to="/master-products/add" class="btn btn-primary">
              Tambah Produk Pertama
            </router-link>
          </div>

          <!-- Product List -->
          <div v-else class="product-list-container">
            <ProductList
              :products="products"
              :loading="loading"
              :current-page="currentPage"
              :total-pages="totalPages"
              @delete="handleDelete"
              @page-change="handlePageChange"
            />
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useUnifiedHeader } from "@/composables/useUnifiedHeader";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import ProductList from "@/components/MasterProduct/ProductList.vue";
import masterProductService, {
  type MasterProduct,
} from "@/services/masterProductService";

const router = useRouter();
const uiStore = useUIStore();
const { setConnectionStatus } = useUnifiedHeader();

// Navigation Handlers
const handleTabChange = (tab: string) => {
  if (tab === "master-products") return;

  if (["settings", "product-management", "order-management"].includes(tab)) {
    // Navigate to dashboard with tab query
    router.push({ path: "/dashboard", query: { tab } });
    uiStore.setActiveTab(tab);
  } else {
    // Handle other tabs if they exist
    router.push({ name: tab });
  }
};

const handlePlatformChange = (platform: string) => {
  uiStore.setActivePlatform(platform);
};

// State
const loading = ref(false);
const error = ref<string | null>(null);
const products = ref<MasterProduct[]>([]);
const currentPage = ref(1);
const totalPages = ref(1);
const totalItems = ref(0);
const pageSize = ref(20);

// Filters
const searchQuery = ref("");
const statusFilter = ref("");

// Debounce timer
let searchTimeout: ReturnType<typeof setTimeout> | null = null;

// Fetch products from API
const fetchProducts = async () => {
  loading.value = true;
  error.value = null;
  setConnectionStatus("connecting");

  try {
    const response = await masterProductService.list({
      page: currentPage.value,
      limit: pageSize.value,
      status: statusFilter.value || undefined,
      search: searchQuery.value || undefined,
    });

    if (response.success) {
      products.value = response.data || [];
      if (response.meta) {
        totalItems.value = response.meta.total;
        totalPages.value = Math.ceil(response.meta.total / pageSize.value);
      }
      setConnectionStatus("connected");
    } else {
      error.value = response.error || "Gagal memuat produk";
      setConnectionStatus("error");
    }
  } catch (err: any) {
    console.error("Error fetching products:", err);
    error.value = err.message || "Gagal memuat produk";
    setConnectionStatus("error");
  } finally {
    loading.value = false;
  }
};

// Debounced search
const debouncedSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    fetchProducts();
  }, 300);
};

// Handle delete
const handleDelete = async (product: MasterProduct) => {
  if (!confirm(`Hapus produk "${product.title}"?`)) {
    return;
  }

  try {
    const response = await masterProductService.delete(product.id);
    if (response.success) {
      // Remove from local list
      products.value = products.value.filter((p) => p.id !== product.id);
      // Refresh if list is empty but there are more pages
      if (products.value.length === 0 && currentPage.value > 1) {
        currentPage.value--;
        fetchProducts();
      }
    } else {
      alert(response.error || "Gagal menghapus produk");
    }
  } catch (err: any) {
    alert(err.message || "Gagal menghapus produk");
  }
};

// Handle page change
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchProducts();
};

// Watch for filter changes
watch(statusFilter, () => {
  currentPage.value = 1;
});

// Initial fetch
onMounted(() => {
  fetchProducts();
});

// Cleanup on unmount
onUnmounted(() => {
  // Reset connection status when leaving the page
  setConnectionStatus("connecting");
});
</script>

<style scoped>
@import "../Dashboard.module.css";

.master-product-list {
  padding: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-weight: 500;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--primary, #3b82f6);
  color: white;
  border: none;
}

.btn-primary:hover {
  background: var(--primary-hover, #2563eb);
}

.btn-secondary {
  background: var(--surface, #f3f4f6);
  color: var(--text-primary, #1a1a1a);
  border: 1px solid var(--border, #e5e7eb);
}

.btn-secondary:hover {
  background: var(--surface-hover, #e5e7eb);
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.search-input {
  flex: 1;
  padding: 0.5rem 1rem;
  border: 1px solid var(--border, #e5e7eb);
  border-radius: 0.375rem;
  font-size: 0.875rem;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary, #3b82f6);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.status-select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border, #e5e7eb);
  border-radius: 0.375rem;
  font-size: 0.875rem;
  background: white;
  min-width: 150px;
}

/* Loading State */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  color: var(--text-secondary, #6b7280);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border, #e5e7eb);
  border-top-color: var(--primary, #3b82f6);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Error State */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
}

.error-message {
  color: var(--error, #ef4444);
  margin-bottom: 1rem;
}

/* Empty State */
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  color: var(--text-secondary, #6b7280);
}

.empty-container p {
  margin-bottom: 1rem;
}

.product-list-container {
  background: var(--surface, white);
  border-radius: 0.5rem;
  border: 1px solid var(--border, #e5e7eb);
  padding: 1rem;
  min-height: 400px;
}

@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
  }

  .btn {
    flex: 1;
    text-align: center;
  }

  .filter-bar {
    flex-direction: column;
  }

  .status-select {
    width: 100%;
  }
}
</style>

/**
 * Composable: useProductManager
 * Shared logic for all product manager components
 * Handles filtering, pagination, preferences, and formatting
 */

import { ref, computed, watch, onMounted, Ref } from "vue";
import { getPlatformConfig } from "@/utils/productManagerConfig";
import {
  createDefaultFilterState,
  createDefaultVisibleColumns,
  debounce,
} from "@/utils/productManagerUtils";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface UseProductManagerOptions {
  platform: string;
  pageSize?: number;
}

export const useProductManager = (options: UseProductManagerOptions) => {
  const { platform, pageSize = 10 } = options;
  const config = getPlatformConfig(platform);

  // ==================== STATE ====================
  const products: Ref<any[]> = ref([]);
  const loading: Ref<boolean> = ref(false);
  const errorMessage: Ref<string> = ref("");
  const successMessage: Ref<string> = ref("");
  const currentPage: Ref<number> = ref(1);
  const showFilterPanel: Ref<boolean> = ref(false);
  const searchQuery: Ref<string> = ref("");

  // Initialize filter states from config
  const columnFilters: Ref<Record<string, string>> = ref(
    createDefaultFilterState(config.filterableFields)
  );
  const visibleColumns: Ref<Record<string, boolean>> = ref(
    createDefaultVisibleColumns(config.columnFields)
  );

  // OPTIMIZED: Track if filter preferences have been loaded
  let filterPreferencesLoaded = false;

  // ==================== COMPUTED ====================
  const filteredProducts = computed(() => {
    let result = products.value;

    // Apply text search
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase();
      result = result.filter((product) =>
        config.searchFields.some((field) =>
          product[field]?.toString().toLowerCase().includes(query)
        )
      );
    }

    // Apply column filters
    result = result.filter((product) => {
      for (const [field, filterValue] of Object.entries(columnFilters.value)) {
        if (!filterValue) continue;

        const productValue = product[field];

        // Handle numeric filters (price, stock)
        if (
          field === "price" ||
          field === "quantity" ||
          field === "currentPrice" ||
          field === "originalPrice" ||
          field === "sellerStock" ||
          field === "shopeeStock"
        ) {
          const minVal = parseInt(String(filterValue)) || 0;
          if ((productValue || 0) < minVal) return false;
        }
        // Handle status filter (exact match)
        else if (field === "status") {
          if (productValue !== filterValue) return false;
        }
        // Handle text filters (contains)
        else {
          if (
            !productValue
              ?.toString()
              .toLowerCase()
              .includes(String(filterValue).toLowerCase())
          ) {
            return false;
          }
        }
      }
      return true;
    });

    return result;
  });

  const paginatedProducts = computed(() => {
    const start = (currentPage.value - 1) * pageSize;
    return filteredProducts.value.slice(start, start + pageSize);
  });

  const totalPages = computed(
    () => Math.ceil(filteredProducts.value.length / pageSize) || 1
  );

  const hasActiveFilters = computed(
    () => searchQuery.value || Object.values(columnFilters.value).some((v) => v)
  );

  // ==================== METHODS ====================
  const clearAllFilters = () => {
    searchQuery.value = "";
    columnFilters.value = createDefaultFilterState(config.filterableFields);
    currentPage.value = 1;
    saveFilterPreferences();
  };

  const resetColumnVisibility = () => {
    visibleColumns.value = createDefaultVisibleColumns(config.columnFields);
    saveFilterPreferences();
  };

  const saveFilterPreferences = debounce(async () => {
    try {
      await fetch(getApiBaseUrl("/filter-preferences"), {
        method: "POST",
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          platform: platform,
          page: "product",
          columnFilters: columnFilters.value,
          visibleColumns: visibleColumns.value,
          searchQuery: searchQuery.value,
        }),
      });
    } catch (error) {
      console.error("Error saving filter preferences:", error);
    }
  }, 1000);

  const loadFilterPreferences = async () => {
    // OPTIMIZED: Only load once per session
    if (filterPreferencesLoaded) return;

    try {
      const response = await fetch(
        `${getApiBaseUrl("/filter-preferences")}?platform=${platform}&page=product`,
        { headers: getAuthHeaders() }
      );
      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          columnFilters.value =
            result.data.columnFilters || columnFilters.value;
          visibleColumns.value =
            result.data.visibleColumns || visibleColumns.value;
          searchQuery.value = result.data.searchQuery || "";
          filterPreferencesLoaded = true;
        }
      }
    } catch (error) {
      console.error("Error loading filter preferences:", error);
    }
  };

  const showError = (message: string) => {
    errorMessage.value = message;
    setTimeout(() => {
      errorMessage.value = "";
    }, 4000);
  };

  const showSuccess = (message: string) => {
    successMessage.value = message;
    setTimeout(() => {
      successMessage.value = "";
    }, 3000);
  };

  // ==================== WATCHERS ====================
  watch(searchQuery, () => {
    currentPage.value = 1;
    saveFilterPreferences();
  });

  watch(
    () => ({ ...columnFilters.value }),
    () => {
      currentPage.value = 1;
      saveFilterPreferences();
    },
    { deep: true }
  );

  watch(
    () => visibleColumns.value,
    () => {
      saveFilterPreferences();
    },
    { deep: true }
  );

  // OPTIMIZED: Only load filter preferences when user opens the filter panel
  // This defers the 2,197ms API call until actually needed
  watch(
    () => showFilterPanel.value,
    (isOpen) => {
      if (isOpen && !filterPreferencesLoaded) {
        loadFilterPreferences();
      }
    }
  );

  // ==================== LIFECYCLE ====================
  onMounted(async () => {
    // OPTIMIZED: Load filter preferences in background (non-blocking)
    // Use setTimeout to push to end of event loop, doesn't block rendering
    setTimeout(() => {
      if (!filterPreferencesLoaded) {
        loadFilterPreferences();
      }
    }, 100);
  });

  return {
    // State
    products,
    loading,
    errorMessage,
    successMessage,
    currentPage,
    showFilterPanel,
    searchQuery,
    columnFilters,
    visibleColumns,

    // Config
    config,

    // Computed
    filteredProducts,
    paginatedProducts,
    totalPages,
    hasActiveFilters,

    // Methods
    clearAllFilters,
    resetColumnVisibility,
    saveFilterPreferences,
    loadFilterPreferences,
    showError,
    showSuccess,
  };
};

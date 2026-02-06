import { ref } from "vue";
import { getAuthHeaders } from "@/utils/apiHeaders";

// OPTIMIZED: Client-side cache untuk products
interface CacheEntry {
  data: any;
  timestamp: number;
}

export function useProductSync(apiBaseUrl: string) {
  const products = ref<any[]>([]);
  const loading = ref(false);
  const cache: Map<string, CacheEntry> = new Map();
  const CACHE_TTL = 0; // disable cache

  function isCacheValid(timestamp: number): boolean {
    return Date.now() - timestamp < CACHE_TTL;
  }

  function getCacheKey(url: string): string {
    return `products_${url}`;
  }

  async function fetchProductsFromAPI() {
    loading.value = true;
    try {
      const response = await fetch(`${apiBaseUrl}/products`, {
        method: "GET",
        headers: getAuthHeaders(),
      });
      const data = await response.json();
      if (data.success) {
        // OPTIMIZED: Clear products cache on API sync
        cache.delete(getCacheKey(`${apiBaseUrl}/db/products`));

        return {
          success: true,
          productCount: data.products?.length || 0,
          savedCount: data.detail_saved || 0,
        };
      } else {
        return {
          success: false,
          error: data.error || "Failed to sync products",
        };
      }
    } catch {
      return { success: false, error: "Error syncing products" };
    } finally {
      loading.value = false;
    }
  }

  async function loadProductsFromDB() {
    const cacheKey = getCacheKey(`${apiBaseUrl}/db/products`);

    // OPTIMIZED: Check cache first
    const cached = cache.get(cacheKey);
    if (cached && isCacheValid(cached.timestamp)) {
      products.value = cached.data.products || [];
      return { success: true, count: products.value.length };
    }

    loading.value = true;
    try {
      const response = await fetch(`${apiBaseUrl}/db/products`, {
        method: "GET",
        headers: getAuthHeaders(),
      });
      const data = await response.json();
      if (data.success) {
        products.value = data.products || [];

        return { success: true, count: products.value.length };
      } else {
        return {
          success: false,
          error: data.error || "Failed to load products",
        };
      }
    } catch {
      return { success: false, error: "Error loading products" };
    } finally {
      loading.value = false;
    }
  }

  return {
    products,
    loading,
    fetchProductsFromAPI,
    loadProductsFromDB,
  };
}

import { ref, computed } from "vue";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

/**
 * Composable untuk mengelola Lock Stock data
 * Lock Stock = Total - Quantity yang terkunci di Locked Orders
 * Fetches locked orders and calculates locked quantities per SKU
 */
export function useLockStock() {
  const lockedStockMap = ref<Record<string, number>>({});
  const lockStockLoading = ref(false);
  const lockStockError = ref<string | null>(null);

  const API_BASE = getApiBaseUrl();

  /**
   * Fetch locked orders data dari locked-today endpoint
   * Creates a map of SKU -> locked qty
   */
  async function fetchLockedStockData() {
    lockStockLoading.value = true;
    lockStockError.value = null;

    try {
      const response = await fetch(`${API_BASE}/orders/locked-today`, {
        method: "GET",
        headers: getAuthHeaders(),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      // Build map of SKU -> total locked qty
      const map: Record<string, number> = {};

      if (data.items && Array.isArray(data.items)) {
        for (const item of data.items) {
          const sku = item.sku;
          const qty = item.qty || 0;

          if (sku) {
            map[sku] = (map[sku] || 0) + qty;
          }
        }
      }

      lockedStockMap.value = map;
    } catch (error: any) {
      lockStockError.value = error.message || "Failed to load locked stock";
      console.error("❌ Error fetching locked stock:", error);
      lockedStockMap.value = {};
    } finally {
      lockStockLoading.value = false;
    }
  }

  /**
   * Calculate lock stock untuk satu item
   * Lock Stock = Total - Locked Qty
   * @param sku - SKU of product
   * @param total - Total quantity dari inventory
   * @returns Lock stock value (minimum 0)
   */
  function calculateLockStock(sku: string, total: number): number {
    const lockedQty = lockedStockMap.value[sku] || 0;
    const lockStock = total - lockedQty;
    return Math.max(0, lockStock);
  }

  /**
   * Get locked quantity untuk satu SKU
   * @param sku - SKU of product
   * @returns Locked quantity
   */
  function getLockedQty(sku: string): number {
    return lockedStockMap.value[sku] || 0;
  }

  /**
   * Check if SKU memiliki locked orders
   */
  const hasLockedOrders = computed(() => {
    return Object.keys(lockedStockMap.value).length > 0;
  });

  return {
    lockedStockMap,
    lockStockLoading,
    lockStockError,
    fetchLockedStockData,
    calculateLockStock,
    getLockedQty,
    hasLockedOrders,
  };
}

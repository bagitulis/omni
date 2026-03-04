/**
 * useStockUpdate Composable
 * RESPONSIBILITY: Handle stock update operations to platforms
 * Features:
 * - Update stock for checked/selected inventory items
 * - Batch update to Shopee, Lazada, TikTok
 * - Progress tracking and error handling
 */

import { ref, computed } from "vue";
import axios from "axios";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

const getApiBase = () => getApiBaseUrl("/inventory");

export interface StockUpdateItem {
  sku: string;
  stock: number;
  platforms?: ("shopee" | "lazada" | "tiktok")[];
}

export interface PlatformResult {
  success: boolean;
  itemId?: string;
  modelId?: string;
  skuId?: string;
  productId?: string;
  error?: string;
}

export interface StockUpdateResult {
  sku: string;
  success: boolean;
  platforms: {
    shopee?: PlatformResult;
    lazada?: PlatformResult;
    tiktok?: PlatformResult;
  };
  errors: string[];
}

export interface BatchUpdateResult {
  total: number;
  successful: number;
  failed: number;
  results: StockUpdateResult[];
}

export function useStockUpdate() {
  const isUpdating = ref(false);
  const progress = ref(0);
  const currentSku = ref("");
  const lastResult = ref<BatchUpdateResult | null>(null);
  const error = ref<string | null>(null);

  /**
   * Update stock for a single SKU
   */
  const updateSingleStock = async (
    sku: string,
    stock: number,
    platforms?: ("shopee" | "lazada" | "tiktok")[]
  ): Promise<StockUpdateResult | null> => {
    try {
      isUpdating.value = true;
      currentSku.value = sku;
      error.value = null;

      const response = await axios.post(
        `${getApiBase()}/update-stock`,
        { sku, stock, platforms },
        { headers: getAuthHeaders() }
      );

      return response.data.data as StockUpdateResult;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[useStockUpdate] Single update error:", err);
      return null;
    } finally {
      isUpdating.value = false;
      currentSku.value = "";
    }
  };

  /**
   * Batch update stock for multiple SKUs
   */
  const updateBatchStock = async (
    items: StockUpdateItem[]
  ): Promise<BatchUpdateResult | null> => {
    try {
      isUpdating.value = true;
      progress.value = 0;
      error.value = null;
      lastResult.value = null;

      const response = await axios.post(
        `${getApiBase()}/update-stock-batch`,
        { items },
        { headers: getAuthHeaders() }
      );

      const result = response.data.data as BatchUpdateResult;
      lastResult.value = result;
      progress.value = 100;

      // Log platform results summary
      if (result.results && result.results.length > 0) {
        result.results.forEach((res) => {
          const platforms = Object.entries(res.platforms || {})
            .filter(([, p]: any) => p?.success)
            .map(([name]) => name.toUpperCase())
            .join(", ");
          const status = res.success ? "✅" : "❌";
          console.log(
            `[Stock Update] ${status} ${res.sku} → ${platforms || "failed"}`
          );
        });
      }

      return result;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error(`[Stock Update] ❌ Request failed: ${error.value}`);
      return null;
    } finally {
      isUpdating.value = false;
    }
  };

  /**
   * Update stock for checked inventory items
   * @param inventoryItems - Array of inventory items from the table
   * @param checkField - Field name for checkbox (default: "checkbox" or "selected")
   * @param stockField - Field name for available stock (default: "available_stock")
   */
  const updateCheckedItems = async (
    inventoryItems: any[],
    checkField: string = "checkbox",
    stockField: string = "available_stock"
  ): Promise<BatchUpdateResult | null> => {
    // Filter checked items
    const checkedItems = inventoryItems.filter((item) => {
      const isChecked =
        item[checkField] === true ||
        item[checkField] === "true" ||
        item[checkField] === 1;
      return isChecked && item.sku;
    });

    if (checkedItems.length === 0) {
      error.value = "No items selected for update";
      return null;
    }

    // Map to stock update format
    const items: StockUpdateItem[] = checkedItems.map((item) => ({
      sku: item.sku,
      stock: parseInt(item[stockField] || item.available_stock || "0", 10),
    }));

    return updateBatchStock(items);
  };

  /**
   * Lookup platform IDs for a SKU (for debugging)
   */
  const lookupPlatformIds = async (sku: string) => {
    try {
      const response = await axios.post(
        `${getApiBase()}/lookup-platform-ids`,
        { sku },
        { headers: getAuthHeaders() }
      );
      return response.data.data;
    } catch (err: any) {
      console.error("[useStockUpdate] Lookup error:", err);
      return null;
    }
  };

  // Computed states
  const hasError = computed(() => error.value !== null);
  const successCount = computed(() => lastResult.value?.successful || 0);
  const failedCount = computed(() => lastResult.value?.failed || 0);

  return {
    // State
    isUpdating,
    progress,
    currentSku,
    lastResult,
    error,

    // Computed
    hasError,
    successCount,
    failedCount,

    // Methods
    updateSingleStock,
    updateBatchStock,
    updateCheckedItems,
    lookupPlatformIds,
  };
}

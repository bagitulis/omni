/**
 * usePriceUpdate Composable
 * RESPONSIBILITY: Handle price update operations to platforms
 * Features:
 * - Update price for checked/selected inventory items
 * - Batch update to Shopee, Lazada, TikTok
 * - Uses single HARGA column for all platforms
 * - Progress tracking and error handling
 */

import { ref, computed } from "vue";
import axios from "axios";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

const getApiBase = () => getApiBaseUrl("/inventory");

export interface PriceUpdateItem {
  sku: string;
  price: number;
  platforms?: ("shopee" | "lazada" | "tiktok")[];
}

export interface PlatformPriceResult {
  success: boolean;
  itemId?: string;
  modelId?: string;
  skuId?: string;
  productId?: string;
  error?: string;
}

export interface PriceUpdateResult {
  sku: string;
  success: boolean;
  platforms: {
    shopee?: PlatformPriceResult;
    lazada?: PlatformPriceResult;
    tiktok?: PlatformPriceResult;
  };
  errors: string[];
  skipped: string[];
}

export interface BatchPriceUpdateResult {
  total: number;
  successful: number;
  failed: number;
  skipped: number;
  results: PriceUpdateResult[];
}

export function usePriceUpdate() {
  const isUpdating = ref(false);
  const progress = ref(0);
  const currentSku = ref("");
  const lastResult = ref<BatchPriceUpdateResult | null>(null);
  const error = ref<string | null>(null);

  /**
   * Update price for a single SKU
   */
  const updateSinglePrice = async (
    sku: string,
    price: number,
    platforms?: ("shopee" | "lazada" | "tiktok")[]
  ): Promise<PriceUpdateResult | null> => {
    try {
      isUpdating.value = true;
      currentSku.value = sku;
      error.value = null;

      const response = await axios.post(
        `${getApiBase()}/update-price`,
        { sku, price, platforms },
        { headers: getAuthHeaders() }
      );

      return response.data.data as PriceUpdateResult;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[usePriceUpdate] Single update error:", err);
      return null;
    } finally {
      isUpdating.value = false;
      currentSku.value = "";
    }
  };

  /**
   * Batch update price for multiple SKUs
   */
  const updateBatchPrice = async (
    items: PriceUpdateItem[]
  ): Promise<BatchPriceUpdateResult | null> => {
    try {
      isUpdating.value = true;
      progress.value = 0;
      error.value = null;
      lastResult.value = null;

      const response = await axios.post(
        `${getApiBase()}/update-price-batch`,
        { items },
        { headers: getAuthHeaders() }
      );

      const result = response.data.data as BatchPriceUpdateResult;
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
            `[Price Update] ${status} ${res.sku} → ${platforms || "failed"}`
          );
        });
      }

      return result;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error(`[Price Update] ❌ Request failed: ${error.value}`);
      return null;
    } finally {
      isUpdating.value = false;
    }
  };

  /**
   * Update price for checked inventory items
   * Uses HARGA column for price value
   * @param inventoryItems - Array of inventory items from the table
   * @param checkField - Field name for checkbox (default: "checkbox" or "selected")
   * @param priceField - Field name for price (default: "HARGA")
   */
  const updateCheckedItems = async (
    inventoryItems: any[],
    checkField: string = "checkbox",
    priceField: string = "HARGA"
  ): Promise<BatchPriceUpdateResult | null> => {
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

    // Map to price update format - use HARGA column
    const items: PriceUpdateItem[] = checkedItems.map((item) => ({
      sku: item.sku,
      price: parseInt(item[priceField] || item.HARGA || item.harga || "0", 10),
    }));

    return updateBatchPrice(items);
  };

  // Computed states
  const hasError = computed(() => error.value !== null);
  const successCount = computed(() => lastResult.value?.successful || 0);
  const failedCount = computed(() => lastResult.value?.failed || 0);
  const skippedCount = computed(() => lastResult.value?.skipped || 0);

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
    skippedCount,

    // Methods
    updateSinglePrice,
    updateBatchPrice,
    updateCheckedItems,
  };
}

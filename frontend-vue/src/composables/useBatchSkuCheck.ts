import { ref } from "vue";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface BatchCheckResult {
  sku: string;
  lazada: boolean;
  shopee: boolean;
  tiktok: boolean;
}

interface BatchCheckResponse {
  success: boolean;
  total: number;
  checked: number;
  results: BatchCheckResult[];
}

/**
 * Composable for batch checking multiple SKUs to all platforms
 * @param skus - Array of SKUs to be checked
 * @returns Object dengan status loading dan hasil check
 */
export const useBatchSkuCheck = () => {
  const loading = ref(false);
  const results = ref<BatchCheckResult[]>([]);
  const error = ref<string | null>(null);

  /**
   * Batch check multiple SKUs
   */
  const checkSkus = async (skus: string[]): Promise<BatchCheckResult[]> => {
    if (!skus || skus.length === 0) {
      error.value = "SKU list is empty";
      return [];
    }

    loading.value = true;
    error.value = null;

    try {
      const response = await fetch(
        `${getApiBaseUrl("/inventory")}/batch-check-sku`,
        {
          method: "POST",
          headers: {
            ...getAuthHeaders(),
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ skus }),
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();

      // Support both legacy and new format
      // Legacy: { success, total, checked, results }
      // New: { success, data: { total, checked, results } }
      const responseData = data.data || data;
      const isSuccess = data.success === true;

      if (isSuccess && responseData.results) {
        results.value = responseData.results;
        return responseData.results;
      } else {
        throw new Error("Invalid response format");
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : "Unknown error";
      error.value = errorMsg;
      console.error("❌ Batch SKU check failed:", errorMsg);
      return [];
    } finally {
      loading.value = false;
    }
  };

  /**
   * Get result untuk satu SKU
   */
  const getSkuResult = (sku: string): BatchCheckResult | null => {
    return results.value.find((r) => r.sku === sku) || null;
  };

  /**
   * Clear results
   */
  const clearResults = () => {
    results.value = [];
    error.value = null;
  };

  /**
   * Save check results ke database
   */
  const saveResults = async (
    checkResults: BatchCheckResult[]
  ): Promise<boolean> => {
    if (!checkResults || checkResults.length === 0) {
      error.value = "No results to save";
      return false;
    }

    loading.value = true;
    error.value = null;

    try {
      const response = await fetch(
        `${getApiBaseUrl("/inventory")}/batch-save-platform-status`,
        {
          method: "POST",
          headers: {
            ...getAuthHeaders(),
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ results: checkResults }),
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();
      return data.success;
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : "Unknown error";
      error.value = errorMsg;
      console.error("❌ Failed to save results:", errorMsg);
      return false;
    } finally {
      loading.value = false;
    }
  };

  /**
   * Load status dari database
   */
  const loadSavedStatus = async (): Promise<BatchCheckResult[]> => {
    loading.value = true;
    error.value = null;

    try {
      const response = await fetch(
        `${getApiBaseUrl("/inventory")}/platform-status`,
        {
          method: "GET",
          headers: getAuthHeaders(),
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();

      // Support both legacy and new format
      const responseData = data.data || data;
      const isSuccess = data.success === true;

      if (isSuccess && Array.isArray(responseData.results)) {
        results.value = responseData.results;
        return responseData.results;
      } else {
        throw new Error("Invalid response format");
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : "Unknown error";
      error.value = errorMsg;
      console.error("⚠️ Failed to load saved status:", errorMsg);
      return [];
    } finally {
      loading.value = false;
    }
  };

  return {
    loading,
    results,
    error,
    checkSkus,
    getSkuResult,
    clearResults,
    saveResults,
    loadSavedStatus,
  };
};

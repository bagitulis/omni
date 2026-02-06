import { usePriceUpdate } from "@/composables/usePriceUpdate";
import { useInventoryAlerts } from "./useInventoryAlerts";
import { extractSKU, parseItemData } from "./useInventoryItemHelpers";
import { useRouteExecutionConfig } from "@/composables/useRouteExecutionConfig";
import axios from "axios";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface BatchCheckResult {
  sku: string;
  shopee: boolean;
  tiktok: boolean;
  lazada: boolean;
}

interface PriceUpdateData {
  sku: string;
  price: number;
}

/**
 * useInventoryPriceUpdate Composable
 * RESPONSIBILITY: Handle price update logic for inventory items
 * Uses single HARGA column for all platforms
 */
export function useInventoryPriceUpdate(state: any) {
  const alertMethods = useInventoryAlerts();
  const priceUpdate = usePriceUpdate();
  const routeConfig = useRouteExecutionConfig();

  /**
   * Check if item is selected for update
   */
  function isItemChecked(item: any): boolean {
    return (
      item.checkbox === true ||
      item.checkbox === "true" ||
      item.checkbox === 1 ||
      item.selected === true
    );
  }

  /**
   * Extract price from item (using HARGA column)
   * Must parse item data first since data is nested in item.data
   */
  function extractPrice(item: any): number {
    // Parse nested data first (same as extractSKU does)
    const itemData = parseItemData(item);

    // Try various field names for price
    const priceValue =
      itemData.HARGA || itemData.harga || itemData.price || itemData.Price || 0;

    const parsed = parseInt(String(priceValue).replace(/[^\d]/g, ""), 10);
    return isNaN(parsed) ? 0 : parsed;
  }

  /**
   * Build price update data for a single item
   */
  function buildUpdateData(
    item: any,
    _batchResults: BatchCheckResult[]
  ): PriceUpdateData | null {
    const sku = extractSKU(item);
    const price = extractPrice(item);

    if (!sku || price <= 0) {
      return null;
    }

    return { sku, price };
  }

  /**
   * Enqueue a job to the queue system
   */
  async function enqueueJob(
    type: string,
    data: Record<string, any>,
    priority: "low" | "normal" | "high" = "normal"
  ): Promise<string | null> {
    try {
      const response = await axios.post(
        `${getApiBaseUrl()}/jobs/enqueue`,
        { type, data, priority },
        { headers: getAuthHeaders() }
      );

      if (response.data.success) {
        console.log(`[Queue] ✅ Job enqueued: ${response.data.data.jobId}`);
        return response.data.data.jobId;
      }
      return null;
    } catch (error: any) {
      console.error(`[Queue] ❌ Failed to enqueue job: ${error.message}`);
      return null;
    }
  }

  /**
   * Handle batch price update - Supports Queue vs Direct mode
   */
  async function handleUpdatePrice() {
    try {
      if (!state.inventoryList || state.inventoryList.length === 0) {
        alertMethods.showAlertWarning(
          "⚠️ No Data",
          "Inventory list is empty. Please sync data first."
        );
        return;
      }

      const checkedItems = state.inventoryList.filter((item: any) => {
        const sku = extractSKU(item);
        const price = extractPrice(item);
        const isChecked = isItemChecked(item);
        return isChecked && sku && price > 0;
      });

      if (checkedItems.length === 0) {
        alertMethods.showAlertWarning(
          "⚠️ No Items Selected",
          "Please check/select items with valid HARGA to update price"
        );
        return;
      }

      const items = checkedItems
        .map((item: any) => buildUpdateData(item, state.batchCheckResults))
        .filter(
          (item: PriceUpdateData | null) => item !== null
        ) as PriceUpdateData[];

      if (items.length === 0) {
        alertMethods.showAlertWarning(
          "⚠️ No Valid Items",
          "Selected items have no valid SKU or HARGA value"
        );
        return;
      }

      state.updatingPrice = true;

      // Check execution mode for update_price
      const useQueue = await routeConfig.shouldUseQueue("update_price");
      console.log(
        `[Price Update] 🚀 Mode: ${useQueue ? "QUEUE" : "DIRECT"} for ${items.length} items`
      );

      if (useQueue) {
        // QUEUE MODE: Enqueue each item as a job
        await handleQueueModeUpdate(items);
      } else {
        // DIRECT MODE: Execute immediately
        await handleDirectModeUpdate(items);
      }
    } catch (error: any) {
      console.error(`[Price Update] ❌ Error: ${error.message}`);
      alertMethods.showAlertError(
        "❌ Update Failed",
        error.message || "Unknown error"
      );
    } finally {
      state.updatingPrice = false;
    }
  }

  /**
   * Handle price update in QUEUE mode
   */
  async function handleQueueModeUpdate(items: PriceUpdateData[]) {
    console.log(
      `[Price Update] 📋 QUEUE MODE: Enqueueing ${items.length} jobs`
    );
    alertMethods.showAlertInfo(
      "📋 Enqueueing Jobs",
      `Adding ${items.length} price updates to queue...`
    );

    let enqueuedCount = 0;

    for (const item of items) {
      const jobId = await enqueueJob(
        "price_update",
        { sku: item.sku, price: item.price },
        "normal"
      );

      if (jobId) {
        enqueuedCount++;
      }
    }

    console.log(
      `[Price Update] ✅ Enqueued ${enqueuedCount}/${items.length} jobs`
    );
    alertMethods.showAlertSuccess(
      "✅ Jobs Enqueued",
      `${enqueuedCount} price updates added to queue. Check Script Monitor for progress.`
    );
  }

  /**
   * Handle price update in DIRECT mode (immediate execution)
   */
  async function handleDirectModeUpdate(items: PriceUpdateData[]) {
    console.log(
      `[Price Update] ⚡ DIRECT MODE: Updating ${items.length} items`
    );
    alertMethods.showAlertInfo(
      "🔄 Updating Price",
      `Updating price for ${items.length} items to all platforms...`
    );

    const result = await priceUpdate.updateBatchPrice(items);

    if (result) {
      if (result.failed === 0) {
        console.log(
          `[Price Update] ✅ Success: ${result.successful} items updated`
        );
        alertMethods.showAlertSuccess(
          "✅ Price Updated",
          `Successfully updated ${result.successful} items`
        );
      } else {
        console.warn(
          `[Price Update] ⚠️ Partial: ${result.successful} succeeded, ${result.failed} failed`
        );
        alertMethods.showAlertWarning(
          "⚠️ Partial Success",
          `${result.successful} succeeded, ${result.failed} failed`
        );
      }
    } else {
      console.error(`[Price Update] ❌ Failed: ${priceUpdate.error.value}`);
      alertMethods.showAlertError(
        "❌ Update Failed",
        priceUpdate.error.value || "Unknown error"
      );
    }
  }

  return {
    handleUpdatePrice,
    isItemChecked,
    extractPrice,
    buildUpdateData,
  };
}

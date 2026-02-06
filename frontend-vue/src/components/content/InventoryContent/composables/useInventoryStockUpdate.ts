import { useStockUpdate } from "@/composables/useStockUpdate";
import { useInventoryAlerts } from "./useInventoryAlerts";
import { extractSKU } from "./useInventoryItemHelpers";
import { getAllocationForItem } from "./useMarketplaceAllocation";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";
import { useRouteExecutionConfig } from "@/composables/useRouteExecutionConfig";
import axios from "axios";

interface BatchCheckResult {
  sku: string;
  shopee: boolean;
  tiktok: boolean;
  lazada: boolean;
}

interface StockUpdateData {
  sku: string;
  shopeeStock?: number;
  tiktokStock?: number;
  lazadaStock?: number;
}

/**
 * useInventoryStockUpdate Composable
 * RESPONSIBILITY: Handle stock update logic for inventory items
 * Supports Queue vs Direct execution mode
 */
export function useInventoryStockUpdate(state: any) {
  const alertMethods = useInventoryAlerts();
  const stockUpdate = useStockUpdate();
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
   * Get platform availability from batch check result
   */
  function getPlatformAvailability(batchResult: BatchCheckResult | null) {
    return {
      shopee: batchResult?.shopee !== false,
      tiktok: batchResult?.tiktok !== false,
      lazada: batchResult?.lazada !== false,
    };
  }

  /**
   * Build stock update data for a single item
   */
  function buildUpdateData(
    item: any,
    batchResults: BatchCheckResult[]
  ): StockUpdateData {
    const sku = extractSKU(item);
    const batchResult = batchResults?.find((r) => r.sku === sku) || null;
    const platformAvailability = getPlatformAvailability(batchResult);

    const allocation = getAllocationForItem(
      item,
      "TOTAL",
      "AUTO",
      platformAvailability
    );

    const updateData: StockUpdateData = { sku };

    if (batchResult?.shopee) {
      updateData.shopeeStock = parseInt(String(allocation.shopee || "0"), 10);
    }
    if (batchResult?.tiktok) {
      updateData.tiktokStock = parseInt(String(allocation.tiktok || "0"), 10);
    }
    if (batchResult?.lazada) {
      updateData.lazadaStock = parseInt(String(allocation.lazada || "0"), 10);
    }

    return updateData;
  }

  /**
   * Fetch locked orders from API
   */
  async function fetchLockedOrders() {
    try {
      console.log("[Stock Update] 📋 Fetching locked orders...");
      const response = await axios.post(
        `${getApiBaseUrl()}/orders/locked-today`,
        { days: 7 },
        { headers: getAuthHeaders() }
      );

      if (response.data.success) {
        const lockedOrders = response.data.items || [];
        console.log(
          `[Stock Update] ✅ Fetched ${lockedOrders.length} locked orders`
        );
        return lockedOrders;
      } else {
        console.warn(
          "[Stock Update] ⚠️ Locked orders fetch returned unsuccessful"
        );
        return [];
      }
    } catch (error: any) {
      console.error(
        `[Stock Update] ❌ Failed to fetch locked orders: ${error.message}`
      );
      alertMethods.showAlertWarning(
        "⚠️ Locked Orders Fetch Failed",
        "Could not fetch locked orders, but proceeding with stock update..."
      );
      return [];
    }
  }

  /**
   * Update stock for a single item with per-platform logging
   */
  async function updateSingleItemStock(
    item: StockUpdateData
  ): Promise<boolean> {
    try {
      const response = await axios.post(
        `${getApiBaseUrl("/inventory")}/update-stock`,
        { sku: item.sku, stock: item.shopeeStock || 0 },
        { headers: getAuthHeaders() }
      );

      const result = response.data.data;

      // Per-platform logging
      if (result.platforms) {
        const platforms = Object.entries(result.platforms)
          .map(([platform, data]: any) => {
            if (data?.success) {
              console.log(
                `[Stock Update] ✅ ${item.sku} → ${platform.toUpperCase()} SUCCESS`
              );
              return `${platform.toUpperCase()} ✅`;
            } else {
              const error = data?.error || "Unknown error";
              console.error(
                `[Stock Update] ❌ ${item.sku} → ${platform.toUpperCase()} FAILED: ${error}`
              );
              return `${platform.toUpperCase()} ❌`;
            }
          })
          .join(" | ");

        console.log(`[Stock Update] 📊 ${item.sku}: ${platforms}`);
      }

      return result.success === true;
    } catch (error: any) {
      console.error(
        `[Stock Update] ❌ ${item.sku} → API Error: ${error.message}`
      );
      return false;
    }
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
   * Handle batch stock update with locked orders check
   * Supports both Queue and Direct execution modes
   */
  async function handleUpdateStock() {
    const checkedItems = state.inventoryList.filter((item: any) => {
      const sku = extractSKU(item);
      return isItemChecked(item) && sku;
    });

    if (checkedItems.length === 0) {
      alertMethods.showAlertWarning(
        "⚠️ No Items Selected",
        "Please check/select items to update stock"
      );
      return;
    }

    const items = checkedItems.map((item: any) =>
      buildUpdateData(item, state.batchCheckResults)
    );

    state.updatingStock = true;

    try {
      // Check execution mode for update_stock
      const useQueue = await routeConfig.shouldUseQueue("update_stock");

      console.log(
        `[Stock Update] 🚀 Mode: ${useQueue ? "QUEUE" : "DIRECT"} for ${items.length} items`
      );

      // Step 1: Fetch locked orders (always direct for speed)
      alertMethods.showAlertInfo(
        "🔄 Step 1/2: Fetching Locked Orders",
        `Retrieving locked orders data...`
      );

      const lockedOrders = await fetchLockedOrders();
      console.log(
        `[Stock Update] 📍 Locked orders info: ${lockedOrders.length} orders locked`
      );

      // Step 2: Update stock based on mode
      if (useQueue) {
        // QUEUE MODE: Enqueue each item as a job
        await handleQueueModeUpdate(items);
      } else {
        // DIRECT MODE: Execute immediately
        await handleDirectModeUpdate(items);
      }
    } catch (error: any) {
      console.error(`[Stock Update] ❌ Error: ${error.message}`);
      alertMethods.showAlertError(
        "❌ Update Failed",
        error.message || "Unknown error"
      );
    } finally {
      state.updatingStock = false;
    }
  }

  /**
   * Handle stock update in QUEUE mode
   */
  async function handleQueueModeUpdate(items: StockUpdateData[]) {
    console.log(
      `[Stock Update] 📋 QUEUE MODE: Enqueueing ${items.length} jobs`
    );
    alertMethods.showAlertInfo(
      "📋 Enqueueing Jobs",
      `Adding ${items.length} stock updates to queue...`
    );

    let enqueuedCount = 0;

    for (const item of items) {
      const jobId = await enqueueJob(
        "stock_update",
        { sku: item.sku, stock: item.shopeeStock || 0 },
        "normal"
      );

      if (jobId) {
        enqueuedCount++;
      }
    }

    console.log(
      `[Stock Update] ✅ Enqueued ${enqueuedCount}/${items.length} jobs`
    );
    alertMethods.showAlertSuccess(
      "✅ Jobs Enqueued",
      `${enqueuedCount} stock updates added to queue. Check Script Monitor for progress.`
    );
  }

  /**
   * Handle stock update in DIRECT mode (immediate execution)
   */
  async function handleDirectModeUpdate(items: StockUpdateData[]) {
    console.log(
      `[Stock Update] ⚡ DIRECT MODE: Updating ${items.length} items`
    );
    alertMethods.showAlertInfo(
      "🔄 Step 2/2: Updating Stock",
      `Updating ${items.length} items to all platforms (1 by 1)...`
    );

    let successCount = 0;
    let failedCount = 0;

    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      const itemNumber = i + 1;

      console.log(
        `[Stock Update] 📝 [${itemNumber}/${items.length}] Processing: ${item.sku}`
      );

      const success = await updateSingleItemStock(item);

      if (success) {
        successCount++;
      } else {
        failedCount++;
      }

      // Small delay to prevent server overload
      if (i < items.length - 1) {
        await new Promise((resolve) => setTimeout(resolve, 100));
      }
    }

    // Summary
    console.log(
      `[Stock Update] 📊 SUMMARY: ${successCount} succeeded, ${failedCount} failed`
    );

    if (failedCount === 0) {
      alertMethods.showAlertSuccess(
        "✅ Stock Updated Successfully",
        `All ${successCount} items updated to all platforms`
      );
    } else {
      alertMethods.showAlertWarning(
        "⚠️ Partial Success",
        `${successCount} succeeded, ${failedCount} failed. Check console for details.`
      );
    }
  }

  return {
    handleUpdateStock,
    isItemChecked,
    buildUpdateData,
    fetchLockedOrders,
    updateSingleItemStock,
    enqueueJob,
    handleQueueModeUpdate,
    handleDirectModeUpdate,
  };
}

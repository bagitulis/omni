import apiClient from "./client";
import { throwIfFailed } from "./inventoryCore";
import type {
  BatchPriceUpdateResult,
  PriceUpdateItem,
  PriceUpdateResult,
} from "./inventoryPriceTypes";

/**
 * Update price for a single item
 * Backend route: POST /api/inventory/update-price
 */
export async function updatePrice(
  sku: string,
  price: number,
  platforms?: string[],
): Promise<PriceUpdateResult> {
  const response = await apiClient.post<PriceUpdateResult>(
    "/inventory/update-price",
    {
      sku,
      price,
      platforms,
    },
  );
  throwIfFailed(response, "Failed to update price");
  if (!response.data) {
    throw new Error("Updated price response is empty");
  }
  return response.data;
}

/**
 * Batch update price for multiple items
 * Backend route: POST /api/inventory/update-price-batch
 */
export async function updatePriceBatch(
  items: PriceUpdateItem[],
): Promise<BatchPriceUpdateResult> {
  const response = await apiClient.post<
    | PriceUpdateResult[]
    | {
        total?: number;
        success?: number;
        successful?: number;
        failed?: number;
        skipped?: number;
        results?: PriceUpdateResult[];
        data?: PriceUpdateResult[];
      }
  >("/inventory/update-price-batch", {
    items,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update price");
  }

  const payload = response.data;

  // Legacy handler returns data as array directly
  if (Array.isArray(payload)) {
    const failed = payload.filter((result) => !result.success).length;
    const successful = payload.length - failed;
    return {
      total: payload.length,
      successful,
      failed,
      skipped: 0,
      results: payload,
    };
  }

  const results = payload?.results ?? payload?.data ?? [];
  const successfulFromResults = results.filter(
    (result) => result.success,
  ).length;

  return {
    total: payload?.total ?? results.length,
    successful:
      payload?.successful ?? payload?.success ?? successfulFromResults,
    failed:
      payload?.failed ?? Math.max(0, results.length - successfulFromResults),
    skipped: payload?.skipped ?? 0,
    results,
  };
}

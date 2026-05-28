import apiClient from "./client";
import {
  DEFAULT_SETTINGS,
  getMessage,
  normalizeWholesaleSettings,
  extractSettingsPayload,
  buildTiktokMpqProducts,
  type BatchUpdateItemInput,
} from "./wholesaleHelpers";

// Re-export types so existing consumers don't break
export type {
  WholesaleTier,
  WholesaleInfo,
  WholesaleResult,
  WholesaleSettings,
  WholesaleTierCalculated,
  SkuLookupResult,
  BatchDeleteBySkusResult,
  BatchUpdateBySkusResult,
  BatchMpqResult,
  TiktokBatchMpqResult,
  TierPreviewResult,
  BatchWholesaleResetResult,
} from "./wholesaleTypes";

import type {
  WholesaleTier,
  WholesaleInfo,
  WholesaleResult,
  WholesaleSettings,
  WholesaleTierCalculated,
  SkuLookupResult,
  BatchDeleteBySkusResult,
  BatchUpdateBySkusResult,
  BatchMpqResult,
  TiktokBatchMpqResult,
  TierPreviewResult,
  BatchWholesaleResetResult,
} from "./wholesaleTypes";

export type BatchUpdateItem = BatchUpdateItemInput;

// --- Empty-failure helpers (DRY) ---

function emptyBatchDeleteData(count: number): BatchDeleteBySkusResult["data"] {
  return {
    total_skus: count,
    unique_items: 0,
    processed: 0,
    failed: count,
    skipped: [],
    results: [],
  };
}

function emptyBatchUpdateData(
  count: number,
): BatchUpdateBySkusResult["data"] {
  return {
    ...emptyBatchDeleteData(count),
    settings_used: DEFAULT_SETTINGS,
  };
}

// --- Shopee CRUD ---

export async function deleteShopeeWholesale(
  itemId: number,
): Promise<WholesaleResult> {
  try {
    const response = await apiClient.delete<unknown>(
      `/wholesale/shopee/${itemId}`,
    );
    return response.success
      ? { success: true, data: response.data }
      : { success: false, error: response.error || "Failed to delete wholesale" };
  } catch (error) {
    return { success: false, error: getMessage(error, "Failed to delete wholesale") };
  }
}

export async function updateShopeeWholesale(
  itemId: number,
  tiers: WholesaleTier[],
): Promise<WholesaleResult> {
  try {
    const response = await apiClient.put<unknown>(
      `/wholesale/shopee/${itemId}`,
      { tiers },
    );
    return response.success
      ? { success: true, data: response.data }
      : { success: false, error: response.error || "Failed to update wholesale" };
  } catch (error) {
    return { success: false, error: getMessage(error, "Failed to update wholesale") };
  }
}

export async function getShopeeWholesale(
  itemId: number,
): Promise<WholesaleInfo | null> {
  try {
    const response = await apiClient.get<WholesaleInfo>(
      `/wholesale/shopee/${itemId}`,
    );
    return response.success ? (response.data ?? null) : null;
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return null;
  }
}

export async function lookupItemBySku(
  sku: string,
): Promise<SkuLookupResult | null> {
  try {
    const response = await apiClient.get<SkuLookupResult>(
      `/wholesale/shopee/lookup/${encodeURIComponent(sku)}`,
    );
    return response.success ? (response.data ?? null) : null;
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return null;
  }
}

// --- Batch Operations ---

export async function batchDeleteBySkus(
  skus: string[],
): Promise<BatchDeleteBySkusResult> {
  try {
    const response = await apiClient.post<BatchDeleteBySkusResult["data"]>(
      "/wholesale/shopee/batch-delete-skus",
      { skus },
    );
    if (!response.success || !response.data)
      return {
        success: false,
        message: response.error || response.message || "Batch delete failed",
        data: emptyBatchDeleteData(skus.length),
      };
    return { success: true, message: response.message || "Batch delete completed", data: response.data };
  } catch (error) {
    return {
      success: false,
      message: getMessage(error, "Batch delete failed"),
      data: emptyBatchDeleteData(skus.length),
    };
  }
}

export async function batchUpdateBySkus(
  items: BatchUpdateItem[],
): Promise<BatchUpdateBySkusResult> {
  try {
    const response = await apiClient.post<BatchUpdateBySkusResult["data"]>(
      "/wholesale/shopee/batch-update-skus",
      { items },
    );
    if (!response.success || !response.data)
      return {
        success: false,
        message: response.error || response.message || "Batch update failed",
        data: emptyBatchUpdateData(items.length),
      };
    return { success: true, message: response.message || "Batch update completed", data: response.data };
  } catch (error) {
    return {
      success: false,
      message: getMessage(error, "Batch update failed"),
      data: emptyBatchUpdateData(items.length),
    };
  }
}

// --- Settings ---

export async function getSettings(): Promise<WholesaleSettings | null> {
  try {
    const response = await apiClient.get<WholesaleSettings>(
      "/wholesale/settings",
    );
    if (!response.success) return null;
    return normalizeWholesaleSettings(extractSettingsPayload(response));
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return null;
  }
}

export async function updateSettings(
  settings: Partial<WholesaleSettings>,
): Promise<WholesaleResult> {
  try {
    const response = await apiClient.put<unknown>(
      "/wholesale/settings",
      settings,
    );
    return response.success
      ? { success: true, data: response.data }
      : { success: false, error: response.error || "Failed to update settings" };
  } catch (error) {
    return { success: false, error: getMessage(error, "Failed to update settings") };
  }
}

export async function previewTiers(
  basePrice: number,
  customSettings?: Partial<WholesaleSettings>,
): Promise<TierPreviewResult | null> {
  try {
    const response = await apiClient.post<TierPreviewResult>(
      "/wholesale/shopee/preview",
      { base_price: basePrice, settings: customSettings },
    );
    return response.success ? (response.data ?? null) : null;
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return null;
  }
}

// --- Local Calculations ---

export function calculateTiersLocal(
  basePrice: number,
  settings: WholesaleSettings,
): WholesaleTierCalculated[] {
  const min1 = settings.min_order_1;
  const max1 = settings.max_order_1;
  const min2 = max1 + 1;
  const max2 = min2 + 1;
  const min3 = max2 + 1;

  const calcPrice = (minQty: number) =>
    Math.round(basePrice - settings.admin_fee + settings.admin_fee / minQty);

  return [
    { tier: 1, min_count: min1, max_count: max1, unit_price: calcPrice(min1) },
    { tier: 2, min_count: min2, max_count: max2, unit_price: calcPrice(min2) },
    { tier: 3, min_count: min3, max_count: settings.max_order_tier_3, unit_price: calcPrice(min3) },
  ];
}

// --- MPQ Operations ---

export async function batchShopeeMpq(
  items: BatchUpdateItem[],
  mpq: number,
): Promise<BatchMpqResult> {
  const response = await apiClient.post<unknown>(
    "/wholesale/shopee/batch-mpq",
    { items, mpq },
  );
  if (!response.success)
    throw new Error(response.error || "Batch Shopee MPQ failed");
  return { success: true, data: response.data };
}

export async function batchTiktokMpq(
  items: BatchUpdateItem[],
  mpq: number,
): Promise<TiktokBatchMpqResult> {
  const products = buildTiktokMpqProducts(items, mpq);
  if (products.length === 0) {
    throw new Error("No valid products to update for TikTok MPQ");
  }
  const response = await apiClient.post<unknown>(
    "/wholesale/tiktok/batch-mpq",
    { products, items, mpq },
  );
  if (!response.success)
    throw new Error(response.error || "Batch TikTok MPQ failed");
  return { success: true, data: response.data };
}

export async function batchWholesaleWithReset(
  items: BatchUpdateItem[],
): Promise<BatchWholesaleResetResult> {
  const response = await apiClient.post<unknown>(
    "/wholesale/shopee/batch-wholesale-reset",
    { items },
  );
  if (!response.success)
    throw new Error(response.error || "Batch wholesale reset failed");
  return { success: true, data: response.data };
}

import apiClient from "./client";

export interface WholesaleTier {
  min_count: number;
  max_count: number;
  unit_price: number;
}
export interface WholesaleInfo {
  item_id: number;
  tiers: WholesaleTier[];
}
export interface WholesaleResult {
  success: boolean;
  error?: string;
  data?: unknown;
}
export interface WholesaleSettings {
  admin_fee: number;
  min_order_1: number;
  max_order_1: number;
  max_order_tier_3: number;
}
export interface WholesaleTierCalculated {
  tier: number;
  min_count: number;
  max_count: number;
  unit_price: number;
}
export interface BatchUpdateItem {
  sku: string;
  item_id?: number;
}
export interface SkuLookupResult {
  item_id: number;
  sku: string;
}
export interface BatchDeleteBySkusResult {
  success: boolean;
  message: string;
  data: {
    total_skus: number;
    unique_items: number;
    processed: number;
    failed: number;
    skipped: string[];
    results: unknown[];
  };
}
export interface BatchUpdateBySkusResult {
  success: boolean;
  message: string;
  data: {
    total_skus: number;
    unique_items: number;
    processed: number;
    failed: number;
    skipped: string[];
    results: unknown[];
    settings_used: WholesaleSettings;
  };
}
export interface BatchMpqResult {
  success: boolean;
  data?: unknown;
  error?: string;
}
export interface TiktokBatchMpqResult {
  success: boolean;
  data?: unknown;
  error?: string;
}
export interface TierPreviewResult {
  tiers: WholesaleTierCalculated[];
}
export interface BatchWholesaleResetResult {
  success: boolean;
  data?: unknown;
  error?: string;
}

const DEFAULT_SETTINGS: WholesaleSettings = {
  admin_fee: 0,
  min_order_1: 1,
  max_order_1: 1,
  max_order_tier_3: 3,
};
const getMessage = (error: unknown, fallback: string): string =>
  error instanceof Error && error.message ? error.message : fallback;

export async function deleteShopeeWholesale(
  itemId: number,
): Promise<WholesaleResult> {
  try {
    const response = await apiClient.delete<unknown>(
      `/wholesale/shopee/${itemId}`,
    );
    return response.success
      ? { success: true, data: response.data }
      : {
          success: false,
          error: response.error || "Failed to delete wholesale",
        };
  } catch (error) {
    return {
      success: false,
      error: getMessage(error, "Failed to delete wholesale"),
    };
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
      : {
          success: false,
          error: response.error || "Failed to update wholesale",
        };
  } catch (error) {
    return {
      success: false,
      error: getMessage(error, "Failed to update wholesale"),
    };
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
  } catch {
    return null;
  }
}
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
        data: {
          total_skus: skus.length,
          unique_items: 0,
          processed: 0,
          failed: skus.length,
          skipped: [],
          results: [],
        },
      };
    return {
      success: true,
      message: response.message || "Batch delete completed",
      data: response.data,
    };
  } catch (error) {
    return {
      success: false,
      message: getMessage(error, "Batch delete failed"),
      data: {
        total_skus: skus.length,
        unique_items: 0,
        processed: 0,
        failed: skus.length,
        skipped: [],
        results: [],
      },
    };
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
  } catch {
    return null;
  }
}
export async function getSettings(): Promise<WholesaleSettings | null> {
  try {
    const response = await apiClient.get<WholesaleSettings>(
      "/wholesale/settings",
    );
    return response.success ? (response.data ?? null) : null;
  } catch {
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
      : {
          success: false,
          error: response.error || "Failed to update settings",
        };
  } catch (error) {
    return {
      success: false,
      error: getMessage(error, "Failed to update settings"),
    };
  }
}
export async function previewTiers(
  basePrice: number,
  customSettings?: Partial<WholesaleSettings>,
): Promise<TierPreviewResult | null> {
  try {
    const response = await apiClient.post<TierPreviewResult>(
      "/wholesale/preview",
      { base_price: basePrice, settings: customSettings },
    );
    return response.success ? (response.data ?? null) : null;
  } catch {
    return null;
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
        data: {
          total_skus: items.length,
          unique_items: 0,
          processed: 0,
          failed: items.length,
          skipped: [],
          results: [],
          settings_used: DEFAULT_SETTINGS,
        },
      };
    return {
      success: true,
      message: response.message || "Batch update completed",
      data: response.data,
    };
  } catch (error) {
    return {
      success: false,
      message: getMessage(error, "Batch update failed"),
      data: {
        total_skus: items.length,
        unique_items: 0,
        processed: 0,
        failed: items.length,
        skipped: [],
        results: [],
        settings_used: DEFAULT_SETTINGS,
      },
    };
  }
}
export function calculateTiersLocal(
  basePrice: number,
  settings: WholesaleSettings,
): WholesaleTierCalculated[] {
  const min1 = settings.min_order_1;
  const max1 = settings.max_order_1;
  const min2 = max1 + 1;
  const max2 = min2 + 1;
  const min3 = max2 + 1;
  return [
    {
      tier: 1,
      min_count: min1,
      max_count: max1,
      unit_price: Math.round(
        basePrice - settings.admin_fee + settings.admin_fee / min1,
      ),
    },
    {
      tier: 2,
      min_count: min2,
      max_count: max2,
      unit_price: Math.round(
        basePrice - settings.admin_fee + settings.admin_fee / min2,
      ),
    },
    {
      tier: 3,
      min_count: min3,
      max_count: settings.max_order_tier_3,
      unit_price: Math.round(
        basePrice - settings.admin_fee + settings.admin_fee / min3,
      ),
    },
  ];
}
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
  const response = await apiClient.post<unknown>(
    "/wholesale/tiktok/batch-mpq",
    { items, mpq },
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

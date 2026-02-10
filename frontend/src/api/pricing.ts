import api from "@/api/client";

/**
 * Price Update Types
 */

export interface PriceUpdateItem {
  sku: string;
  price: number;
  platforms?: string[]; // Optional: ["shopee", "lazada", "tiktok"]
}

export interface PlatformPriceResult {
  success: boolean;
  item_id?: string;
  error?: string;
}

export interface PriceUpdateResult {
  sku: string;
  price: number;
  success: boolean;
  platforms: Record<string, PlatformPriceResult>;
  errors?: string[];
}

export interface BatchPriceUpdateResult {
  total: number;
  success: number;
  failed: number;
  results: PriceUpdateResult[];
}

/**
 * Update price for a single SKU (using batch endpoint for consistency)
 */
export async function updatePrice(
  item: PriceUpdateItem,
): Promise<BatchPriceUpdateResult> {
  return updatePriceBatch([item]);
}

/**
 * Batch update price for multiple SKUs
 */
export async function updatePriceBatch(
  items: PriceUpdateItem[],
): Promise<BatchPriceUpdateResult> {
  const response = await api.post<BatchPriceUpdateResult>(
    "/inventory/update-price-batch",
    { items },
  );

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to update prices");
  }

  return response.data;
}

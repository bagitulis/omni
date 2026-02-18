import apiClient from "./client";
import type {
  InventoryWholesaleTier,
  InventoryWholesaleInfo,
  InventoryWholesaleSettings,
  InventoryWholesaleBatchUpdateItem,
  InventoryWholesaleBatchResult,
  InventoryMpqSettings,
  InventoryMpqBatchItem,
  InventoryMpqBatchResult,
} from "@/types/wholesale";

// --- Inventory Module Wholesale/MPQ API Functions ---

export type {
  InventoryWholesaleTier,
  InventoryWholesaleInfo,
  InventoryWholesaleSettings,
  InventoryWholesaleBatchUpdateItem,
  InventoryWholesaleBatchResult,
  InventoryMpqSettings,
  InventoryMpqBatchItem,
  InventoryMpqBatchResult,
} from "@/types/wholesale";

/**
 * Get wholesale tiers for a specific SKU
 * Backend route: GET /api/inventory/wholesale/:sku
 * TODO: Backend endpoint verification pending
 */
export async function getInventoryWholesaleTiers(
  sku: string,
): Promise<InventoryWholesaleTier[]> {
  const response = await apiClient.get<InventoryWholesaleTier[]>(
    `/inventory/wholesale/${sku}`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch wholesale tiers");
  }
  return response.data || [];
}

/**
 * Update wholesale tiers for a specific SKU
 * Backend route: PUT /api/inventory/wholesale/:sku
 * TODO: Backend endpoint verification pending
 */
export async function updateInventoryWholesaleTiers(
  sku: string,
  tiers: InventoryWholesaleTier[],
): Promise<void> {
  const response = await apiClient.put(`/inventory/wholesale/${sku}`, {
    tiers,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to update wholesale tiers");
  }
}

/**
 * Get wholesale settings
 * Backend route: GET /api/inventory/wholesale/settings
 * TODO: Backend endpoint verification pending
 */
export async function getInventoryWholesaleSettings(): Promise<InventoryWholesaleSettings | null> {
  const response = await apiClient.get<InventoryWholesaleSettings>(
    "/inventory/wholesale/settings",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch wholesale settings");
  }
  return response.data || null;
}

/**
 * Update wholesale settings
 * Backend route: PUT /api/inventory/wholesale/settings
 * TODO: Backend endpoint verification pending
 */
export async function updateInventoryWholesaleSettings(
  settings: Partial<InventoryWholesaleSettings>,
): Promise<InventoryWholesaleSettings> {
  const response = await apiClient.put<InventoryWholesaleSettings>(
    "/inventory/wholesale/settings",
    settings,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to update wholesale settings");
  }
  if (response.data == null) {
    throw new Error("Failed to update wholesale settings: no data in response");
  }
  return response.data;
}

/**
 * Batch update wholesale tiers for multiple SKUs
 * Backend route: POST /api/inventory/wholesale/batch-update
 * TODO: Backend endpoint verification pending
 */
export async function batchUpdateInventoryWholesale(
  items: InventoryWholesaleBatchUpdateItem[],
): Promise<InventoryWholesaleBatchResult> {
  const response = await apiClient.post<InventoryWholesaleBatchResult>(
    "/inventory/wholesale/batch-update",
    { items },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update wholesale");
  }
  if (response.data == null) {
    throw new Error("Failed to batch update wholesale: no data in response");
  }
  return response.data;
}

/**
 * Batch delete wholesale tiers for multiple SKUs
 * Backend route: POST /api/wholesale/shopee/batch-delete-skus
 */
export async function batchDeleteInventoryWholesale(
  skus: string[],
): Promise<void> {
  const response = await apiClient.post("/wholesale/shopee/batch-delete-skus", {
    skus,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to batch delete wholesale");
  }
}

/**
 * Get detailed wholesale info for a specific SKU
 * Backend route: GET /api/inventory/wholesale/:sku/info
 * TODO: Backend endpoint verification pending
 */
export async function getInventoryWholesaleInfo(
  sku: string,
): Promise<InventoryWholesaleInfo> {
  const response = await apiClient.get<InventoryWholesaleInfo>(
    `/inventory/wholesale/${sku}/info`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch wholesale info");
  }
  if (response.data == null) {
    throw new Error("Failed to fetch wholesale info: no data in response");
  }
  return response.data;
}

/**
 * Get MPQ (Minimum Purchase Quantity) settings
 * Backend route: GET /api/inventory/mpq/settings
 * TODO: Backend endpoint verification pending
 */
export async function getInventoryMpqSettings(): Promise<
  InventoryMpqSettings[]
> {
  const response = await apiClient.get<InventoryMpqSettings[]>(
    "/inventory/mpq/settings",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch MPQ settings");
  }
  return response.data || [];
}

/**
 * Update MPQ settings for multiple SKUs
 * Backend route: PUT /api/inventory/mpq/settings
 * TODO: Backend endpoint verification pending
 */
export async function updateInventoryMpqSettings(
  settings: InventoryMpqSettings[],
): Promise<void> {
  const response = await apiClient.put("/inventory/mpq/settings", {
    settings,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to update MPQ settings");
  }
}

/**
 * Batch update MPQ for multiple items
 * Backend route: POST /api/inventory/mpq/batch-update
 * TODO: Backend endpoint verification pending
 */
export async function batchUpdateInventoryMpq(
  items: InventoryMpqBatchItem[],
): Promise<InventoryMpqBatchResult> {
  const response = await apiClient.post<InventoryMpqBatchResult>(
    "/inventory/mpq/batch-update",
    { items },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update MPQ");
  }
  if (response.data == null) {
    throw new Error("Failed to batch update MPQ: no data in response");
  }
  return response.data;
}

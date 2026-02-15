import apiClient from "./client";
import type { ApiResponse } from "./client";
import {
  InventoryRecord,
  InventoryListResult,
  InventoryConfig,
  InventoryStats,
  SyncHistoryEntry,
  BatchCheckResult,
  SkuCheckResult,
} from "@/types/inventory";

export type {
  InventoryRecord,
  InventoryListResult,
  InventoryConfig,
  InventoryStats,
  SyncHistoryEntry,
  BatchCheckResult,
  SkuCheckResult,
};

export interface GetInventoryParams {
  offset?: number;
  limit?: number;
  search?: string;
  sort_by?: string;
  sort_dir?: "asc" | "desc";
  sync_status?: string[];
  stock_status?: string;
  platform?: string[];
}

/** Backend response for /inventory/list includes pagination fields at top level */
interface InventoryListResponse extends ApiResponse<InventoryRecord[]> {
  total?: number;
  offset?: number;
  limit?: number;
}

/**
 * Fetch inventory data
 * Backend route: GET /api/inventory/list
 */
export async function getInventory(
  params?: GetInventoryParams,
): Promise<InventoryListResult> {
  const response = (await apiClient.get<InventoryRecord[]>("/inventory/list", {
    params,
  })) as InventoryListResponse;
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory");
  }
  return {
    records: response.data || [],
    total: response.total || 0,
    offset: response.offset || 0,
    limit: response.limit || 100,
  };
}

/**
 * Get inventory item by SKU
 * Backend route: GET /api/inventory/:keyValue
 */
export async function getInventoryBySku(sku: string): Promise<InventoryRecord> {
  const response = await apiClient.get<InventoryRecord>(`/inventory/${sku}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory item");
  }
  return response.data!;
}

/**
 * Update stock for a single item
 * Backend route: POST /api/inventory/update-stock
 * Backend reads stock from inventory_records (not from request payload)
 */
export async function updateStock(
  sku: string,
  platforms?: string[],
): Promise<void> {
  const response = await apiClient.post("/inventory/update-stock", {
    sku,
    platforms,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to update stock");
  }
}

/**
 * Batch update stock for multiple items
 * Backend route: POST /api/inventory/update-stock-batch
 * Backend reads stock from inventory_records for each SKU
 */
export async function updateStockBatch(
  skus: string[],
  platforms?: string[],
): Promise<void> {
  const response = await apiClient.post("/inventory/update-stock-batch", {
    skus,
    platforms,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update stock");
  }
}

/**
 * Get inventory configuration
 * Backend route: GET /api/inventory/config
 */
export async function getInventoryConfig(): Promise<InventoryConfig | null> {
  const response = await apiClient.get<InventoryConfig>("/inventory/config");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory config");
  }
  return response.data || null;
}

/**
 * Trigger inventory sync from Google Sheets
 * Backend route: POST /api/inventory/sync/from-sheets
 */
export async function syncInventory(
  spreadsheetId?: string,
  sheetName?: string,
): Promise<void> {
  const response = await apiClient.post("/inventory/sync/from-sheets", {
    spreadsheet_id: spreadsheetId,
    sheet_name: sheetName,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to sync inventory");
  }
}

/**
 * Get inventory stats
 * Backend route: GET /api/inventory/stats
 */
export async function getInventoryStats(): Promise<InventoryStats> {
  const response = await apiClient.get<InventoryStats>("/inventory/stats");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory stats");
  }
  return response.data!;
}

/**
 * Get sync history
 * Backend route: GET /api/inventory/sync/history
 */
export async function getSyncHistory(): Promise<SyncHistoryEntry[]> {
  const response = await apiClient.get<SyncHistoryEntry[]>(
    "/inventory/sync/history",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch sync history");
  }
  return response.data || [];
}

/**
 * Update inventory configuration
 * Backend route: PUT /api/inventory/config
 */
export async function updateInventoryConfig(
  config: Partial<InventoryConfig>,
): Promise<InventoryConfig> {
  const response = await apiClient.put<InventoryConfig>(
    "/inventory/config",
    config,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to update inventory config");
  }
  return response.data!;
}

/**
 * Trigger inventory export to Google Sheets
 * Backend route: POST /api/inventory/sync/to-sheets
 */
export async function syncToSheets(): Promise<void> {
  const response = await apiClient.post("/inventory/sync/to-sheets");
  if (!response.success) {
    throw new Error(response.error || "Failed to export to sheets");
  }
}

/**
 * Batch check platform status
 * Backend route: POST /api/inventory/batch-check-sku
 */
export async function checkPlatformStatus(): Promise<BatchCheckResult[]> {
  const response = await apiClient.post<BatchCheckResult[]>(
    "/inventory/batch-check-sku",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to check platform status");
  }
  return response.data || [];
}

/**
 * Get available columns for inventory table
 * Backend route: GET /api/inventory/columns/available
 */
export async function getAvailableColumns(): Promise<string[]> {
  const response = await apiClient.client.get("/inventory/columns/available");
  const data = response.data;
  if (!data.success) {
    throw new Error("Failed to fetch available columns");
  }
  // Backend returns { columns: [{name, spreadsheet_column, type, position}] }
  const columns = data.columns || [];
  return columns.map((col: { name: string }) => col.name);
}

/**
 * Get selected columns for inventory table
 * Backend route: GET /api/inventory/columns/selected
 */
export async function getSelectedColumns(): Promise<string[]> {
  const response = await apiClient.client.get("/inventory/columns/selected");
  const data = response.data;
  if (!data.success) {
    throw new Error("Failed to fetch selected columns");
  }
  // Backend returns { selected_columns: [...] }
  return data.selected_columns || [];
}

/**
 * Save selected columns
 * Backend route: POST /api/inventory/columns/selected
 */
export async function saveSelectedColumns(columns: string[]): Promise<void> {
  const response = await apiClient.post("/inventory/columns/selected", {
    columns,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to save column selection");
  }
}

export interface PriceUpdateItem {
  sku: string;
  price: number;
  platforms?: string[];
}

export interface PlatformPriceResult {
  success: boolean;
  item_id?: string;
  model_id?: string;
  sku_id?: string;
  product_id?: string;
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
  if (!response.success) {
    throw new Error(response.error || "Failed to update price");
  }
  return response.data!;
}

/**
 * Batch update price for multiple items
 * Backend route: POST /api/inventory/update-price-batch
 */
export async function updatePriceBatch(
  items: PriceUpdateItem[],
): Promise<BatchPriceUpdateResult> {
  const response = await apiClient.post<{
    total: number;
    success: number;
    failed: number;
    results: PriceUpdateResult[];
  }>("/inventory/update-price-batch", {
    items,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update price");
  }
  // Normalize backend response: backend returns "success" (count), we expect "successful"
  return {
    total: response.data?.total ?? 0,
    successful: response.data?.success ?? 0, // Map "success" to "successful"
    failed: response.data?.failed ?? 0,
    skipped: 0,
    results: response.data?.results ?? [],
  };
}

/**
 * Batch check SKU status across platforms
 * Backend route: POST /api/inventory/batch-check-sku
 */
export async function batchCheckSku(skus: string[]): Promise<SkuCheckResult[]> {
  const response = await apiClient.post<{
    total: number;
    checked: number;
    results: SkuCheckResult[];
  }>("/inventory/batch-check-sku", { skus });
  if (!response.success) {
    throw new Error(response.error || "Failed to check SKU status");
  }
  return response.data?.results || [];
}

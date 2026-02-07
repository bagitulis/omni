import apiClient from "./client";
import { InventoryItem, InventoryListResponse } from "@/types/product";

export interface GetInventoryParams {
  page?: number;
  limit?: number;
  search?: string;
  category?: string;
}

export interface InventoryConfig {
  id: string;
  tenant_id: string;
  spreadsheet_id: string;
  sheet_name: string;
  selected_columns: string;
  all_columns: string;
  header_row: number;
  data_start_row: number;
  key_column: string;
  auto_sync: boolean;
  sync_interval_seconds: number;
  last_sync_timestamp: string | null;
  last_headers_hash: string;
  last_sync_status: string;
  created_at: string;
  updated_at: string;
}

/**
 * Fetch inventory data
 * Backend route: GET /api/inventory/list
 */
export async function getInventory(
  params?: GetInventoryParams,
): Promise<InventoryListResponse> {
  const response = await apiClient.get<InventoryListResponse>(
    "/inventory/list",
    {
      params,
    },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory");
  }
  return response.data!;
}

/**
 * Get inventory item by SKU
 * Backend route: GET /api/inventory/:keyValue
 */
export async function getInventoryBySku(sku: string): Promise<InventoryItem> {
  const response = await apiClient.get<InventoryItem>(`/inventory/${sku}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory item");
  }
  return response.data!;
}

/**
 * Update stock for a single item
 * Backend route: POST /api/inventory/update-stock
 */
export async function updateStock(
  sku: string,
  newStock: number,
): Promise<void> {
  const response = await apiClient.post("/inventory/update-stock", {
    sku,
    stock: newStock,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to update stock");
  }
}

/**
 * Batch update stock for multiple items
 * Backend route: POST /api/inventory/update-stock-batch
 */
export async function updateStockBatch(
  updates: Array<{ sku: string; stock: number }>,
): Promise<void> {
  const response = await apiClient.post("/inventory/update-stock-batch", {
    updates,
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
 * Get available columns for inventory table
 * Backend route: GET /api/inventory/columns/available
 */
export async function getAvailableColumns(): Promise<string[]> {
  const response = await apiClient.get<string[]>(
    "/inventory/columns/available",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch available columns");
  }
  return response.data || [];
}

/**
 * Get selected columns for inventory table
 * Backend route: GET /api/inventory/columns/selected
 */
export async function getSelectedColumns(): Promise<string[]> {
  const response = await apiClient.get<string[]>("/inventory/columns/selected");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch selected columns");
  }
  return response.data || [];
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

import apiClient from "./client";
import { InventoryRecord, InventoryListResult } from "@/types/product";

export interface GetInventoryParams {
  offset?: number;
  limit?: number;
  search?: string;
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
): Promise<InventoryListResult> {
  const response = await apiClient.get<InventoryRecord[]>("/inventory/list", {
    params,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory");
  }
  const raw = response as unknown as Record<string, unknown>;
  return {
    records: response.data || [],
    total: (raw.total as number) || 0,
    offset: (raw.offset as number) || 0,
    limit: (raw.limit as number) || 100,
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

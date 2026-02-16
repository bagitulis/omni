import type {
  InventoryListResult,
  InventoryRecord,
  InventoryStats,
  SyncHistoryEntry,
} from "@/types/inventory";
import type { ApiResponse } from "./client";
import apiClient from "./client";

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
 * Shared helper to throw error if API response failed
 */
export function throwIfFailed<T>(
  response: ApiResponse<T>,
  message: string,
): void {
  if (!response.success) {
    throw new Error(response.error || message);
  }
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
  throwIfFailed(response, "Failed to fetch inventory item");
  if (!response.data) {
    throw new Error("Inventory item response is empty");
  }
  return response.data;
}

/**
 * Update inventory record data by SKU key value
 * Backend route: PUT /api/inventory/:keyValue
 */
export async function updateInventoryRecord(
  sku: string,
  data: Record<string, unknown>,
): Promise<InventoryRecord> {
  const response = await apiClient.put<InventoryRecord>(
    `/inventory/${sku}`,
    data,
  );
  throwIfFailed(response, "Failed to update inventory record");
  if (!response.data) {
    throw new Error("Updated inventory record response is empty");
  }
  return response.data;
}

/**
 * Get inventory stats
 * Backend route: GET /api/inventory/stats
 */
export async function getInventoryStats(): Promise<InventoryStats> {
  const response = await apiClient.get<InventoryStats>("/inventory/stats");
  throwIfFailed(response, "Failed to fetch inventory stats");
  if (!response.data) {
    throw new Error("Inventory stats response is empty");
  }
  return response.data;
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

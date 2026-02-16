import apiClient from "./client";
import type { ApiResponse } from "./client";
import type {
  BatchCheckResult,
  InventoryConfig,
  InventoryListResult,
  InventoryRecord,
  InventoryStats,
  SkuCheckResult,
  SyncHistoryEntry,
} from "@/types/inventory";
import type {
  BatchPriceUpdateResult,
  PriceUpdateItem,
  PriceUpdateResult,
} from "./inventoryPriceTypes";

export type {
  BatchCheckResult,
  BatchPriceUpdateResult,
  InventoryConfig,
  InventoryListResult,
  InventoryRecord,
  InventoryStats,
  PriceUpdateItem,
  PriceUpdateResult,
  SkuCheckResult,
  SyncHistoryEntry,
};

type RawInventoryConfig = Partial<InventoryConfig> & {
  selected_columns?: string | string[] | null;
  key_column?: string | null;
  key_column_name?: string | null;
};

function normalizeSelectedColumns(
  value: RawInventoryConfig["selected_columns"],
): string {
  if (Array.isArray(value)) {
    return JSON.stringify(
      value.filter((column): column is string => typeof column === "string"),
    );
  }

  if (typeof value === "string") {
    return value;
  }

  return "";
}

function normalizeInventoryConfig(
  config: RawInventoryConfig | null | undefined,
): InventoryConfig | null {
  if (!config) {
    return null;
  }

  const keyColumn =
    typeof config.key_column === "string"
      ? config.key_column
      : typeof config.key_column_name === "string"
        ? config.key_column_name
        : "";

  return {
    id: typeof config.id === "string" ? config.id : "",
    tenant_id: typeof config.tenant_id === "string" ? config.tenant_id : "",
    spreadsheet_id:
      typeof config.spreadsheet_id === "string" ? config.spreadsheet_id : "",
    sheet_name: typeof config.sheet_name === "string" ? config.sheet_name : "",
    selected_columns: normalizeSelectedColumns(config.selected_columns),
    all_columns:
      typeof config.all_columns === "string" ? config.all_columns : "",
    header_row: typeof config.header_row === "number" ? config.header_row : 1,
    data_start_row:
      typeof config.data_start_row === "number" ? config.data_start_row : 2,
    key_column: keyColumn,
    auto_sync: Boolean(config.auto_sync),
    sync_interval_seconds:
      typeof config.sync_interval_seconds === "number"
        ? config.sync_interval_seconds
        : 300,
    last_sync_timestamp: config.last_sync_timestamp ?? null,
    last_headers_hash:
      typeof config.last_headers_hash === "string"
        ? config.last_headers_hash
        : "",
    last_sync_status:
      typeof config.last_sync_status === "string"
        ? config.last_sync_status
        : "",
    low_stock_threshold:
      typeof config.low_stock_threshold === "number"
        ? config.low_stock_threshold
        : undefined,
    created_at: typeof config.created_at === "string" ? config.created_at : "",
    updated_at: typeof config.updated_at === "string" ? config.updated_at : "",
  };
}

function throwIfFailed<T>(response: ApiResponse<T>, message: string): void {
  if (!response.success) {
    throw new Error(response.error || message);
  }
}

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
    items: skus.map((sku) => ({ sku, platforms })),
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
  const response = await apiClient.get<RawInventoryConfig>("/inventory/config");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch inventory config");
  }
  return normalizeInventoryConfig(response.data);
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
  throwIfFailed(response, "Failed to update inventory config");
  if (!response.data) {
    throw new Error("Updated inventory config response is empty");
  }
  return response.data;
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
export async function checkPlatformStatus(
  skus: string[],
): Promise<SkuCheckResult[]> {
  return batchCheckSku(skus);
}

/**
 * Get available columns for inventory table
 * Backend route: GET /api/inventory/columns/available
 */
export async function getAvailableColumns(): Promise<string[]> {
  const response = await apiClient.client.get("/inventory/columns/available");
  const data = response.data as {
    success?: boolean;
    error?: string;
    columns?: Array<{ name?: string; key?: string; label?: string }>;
    data?: Array<{ name?: string; key?: string; label?: string }>;
  };
  if (!data.success) {
    throw new Error(data.error || "Failed to fetch available columns");
  }

  const columns = Array.isArray(data.columns)
    ? data.columns
    : Array.isArray(data.data)
      ? data.data
      : [];

  const normalizedColumns = columns
    .map((col) => {
      if (typeof col.name === "string" && col.name.trim().length > 0) {
        return col.name;
      }
      if (typeof col.label === "string" && col.label.trim().length > 0) {
        return col.label;
      }
      if (typeof col.key === "string" && col.key.trim().length > 0) {
        return col.key;
      }
      return "";
    })
    .filter((column): column is string => column.length > 0);

  return Array.from(new Set(normalizedColumns));
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

  if (Array.isArray(data.selected_columns)) {
    return data.selected_columns;
  }

  // Modular handler returns { data: [...] }
  if (Array.isArray(data.data)) {
    return data.data;
  }

  return [];
}

/**
 * Save selected columns
 * Backend route: POST /api/inventory/columns/selected
 */
export async function saveSelectedColumns(columns: string[]): Promise<void> {
  const response = await apiClient.post("/inventory/columns/selected", {
    selected_columns: columns,
    columns,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to save column selection");
  }
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

import apiClient from "./client";

export interface StockBatchSyncItem {
  sku: string;
  stock?: number;
  platforms?: string[];
}

export interface SyncResult {
  status: string;
  message?: string;
  total_records: number;
  synced_records: number;
  new_records: number;
  updated_records: number;
  unchanged_records: number;
  failed_records: number;
  headers_changed?: boolean;
  duration?: number;
  timestamp?: string;
}

interface PlatformStockSyncResult {
  success?: boolean;
  error?: string;
}

interface StockSyncResult {
  success?: boolean;
  error?: string;
  errors?: string[];
  platforms?: Record<string, PlatformStockSyncResult>;
}

export interface StockBatchResult {
  total: number;
  succeeded: number;
  failed: number;
  results: StockSyncResult[];
}

function collectStockSyncErrors(result: StockSyncResult | undefined): string[] {
  if (!result) {
    return [];
  }

  const errors: string[] = [];
  if (result.error) {
    errors.push(result.error);
  }
  if (Array.isArray(result.errors)) {
    errors.push(...result.errors.filter(Boolean));
  }

  for (const platformResult of Object.values(result.platforms ?? {})) {
    if (platformResult?.error) {
      errors.push(platformResult.error);
    }
  }

  return errors;
}

function firstStockSyncError(
  result: StockSyncResult | undefined,
): string | undefined {
  return collectStockSyncErrors(result)[0];
}

/**
 * Update stock for a single item
 * Backend route: POST /api/inventory/update-stock
 */
export async function updateStock(
  sku: string,
  platforms?: string[],
  stock?: number,
): Promise<void> {
  const response = await apiClient.post("/inventory/update-stock", {
    sku,
    platforms,
    ...(stock !== undefined ? { stock } : {}),
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to update stock");
  }

  const syncResult = response.data as StockSyncResult | undefined;
  if (syncResult?.success === false) {
    throw new Error(firstStockSyncError(syncResult) || "Stock sync failed");
  }
}

/**
 * Batch update stock for multiple items
 * Backend route: POST /api/inventory/update-stock-batch
 */
export async function updateStockBatch(
  itemsOrSkus: StockBatchSyncItem[] | string[],
  platforms?: string[],
): Promise<StockBatchResult> {
  const useLegacySkuList = itemsOrSkus.every(
    (item) => typeof item === "string",
  );

  const items: StockBatchSyncItem[] = useLegacySkuList
    ? (itemsOrSkus as string[]).map((sku) => ({ sku, platforms }))
    : (itemsOrSkus as StockBatchSyncItem[]).map((item) => ({
        sku: item.sku,
        ...(item.stock !== undefined ? { stock: item.stock } : {}),
        ...(item.platforms ? { platforms: item.platforms } : {}),
      }));

  const response = await apiClient.post("/inventory/update-stock-batch", {
    skus: items.map((item) => item.sku),
    platforms,
    items,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to batch update stock");
  }

  const payload = response.data as
    | unknown[]
    | { results?: unknown[]; data?: unknown[] }
    | undefined;

  const rawResults = Array.isArray(payload)
    ? payload
    : Array.isArray(payload?.results)
      ? payload.results
      : Array.isArray(payload?.data)
        ? payload.data
        : [];

  const parsedResults = rawResults as StockSyncResult[];
  const succeeded = parsedResults.filter((r) => r.success !== false).length;
  const failed = parsedResults.filter((r) => r.success === false).length;

  return {
    total: parsedResults.length,
    succeeded,
    failed,
    results: parsedResults,
  };
}

/**
 * Trigger inventory sync from Google Sheets
 * Backend route: POST /api/inventory/sync/from-sheets
 */
export async function syncInventory(
  spreadsheetId?: string,
  sheetName?: string,
): Promise<SyncResult> {
  const response = await apiClient.post("/inventory/sync/from-sheets", {
    spreadsheet_id: spreadsheetId,
    sheet_name: sheetName,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to sync inventory");
  }
  return (response.data ?? {}) as SyncResult;
}

/**
 * Trigger inventory export to Google Sheets
 * Backend route: POST /api/inventory/sync/to-sheets
 */
export async function syncToSheets(): Promise<SyncResult> {
  const response = await apiClient.post("/inventory/sync/to-sheets");
  if (!response.success) {
    throw new Error(response.error || "Failed to export to sheets");
  }
  return (response.data ?? {}) as SyncResult;
}

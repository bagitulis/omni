import apiClient from "./client";

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
 * Trigger inventory export to Google Sheets
 * Backend route: POST /api/inventory/sync/to-sheets
 */
export async function syncToSheets(): Promise<void> {
  const response = await apiClient.post("/inventory/sync/to-sheets");
  if (!response.success) {
    throw new Error(response.error || "Failed to export to sheets");
  }
}

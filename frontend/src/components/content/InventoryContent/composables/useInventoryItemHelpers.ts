/**
 * useInventoryItemHelpers Composable
 * RESPONSIBILITY: Helper functions for inventory item operations
 * - Parse item data from JSON
 * - Extract SKU from item
 * - Get cell values
 * - Calculate stock values
 */

export interface ItemHelperOptions {
  lockStockMap?: Record<string, number>;
}

/**
 * Parse item data from inventory row
 * Handles both JSON string and object formats from backend
 * Go backend returns data as object, Node.js might return as string
 */
export function parseItemData(item: any): any {
  if (!item) return {};

  // If item has a data property
  if (item.data !== undefined && item.data !== null) {
    // If data is a string, parse it as JSON
    if (typeof item.data === "string") {
      try {
        return JSON.parse(item.data);
      } catch {
        return item;
      }
    }
    // If data is already an object (from Go backend), return it directly
    if (typeof item.data === "object") {
      return item.data;
    }
  }

  // Fallback: return item itself (for flat structure)
  return item;
}

/**
 * Extract SKU from inventory item
 * Searches common SKU field names
 */
export function extractSKU(item: any): string {
  const itemData = parseItemData(item);
  const skuFields = [
    "sku",
    "SKU",
    "Sku",
    "sku_id",
    "SKU_ID",
    "Sku_Id",
    "seller_sku",
    "sellerSku",
    "SellerSku",
    "product_sku",
    "productSku",
    "ProductSku",
  ];
  for (const field of skuFields) {
    const value = itemData[field];
    if (value) {
      const skuValue = String(value).trim();
      if (skuValue) return skuValue;
    }
  }
  return "";
}

/**
 * Get raw cell value from item
 */
export function getCellValueRaw(item: any, column: string): any {
  return parseItemData(item)[column];
}

/**
 * Get formatted cell value for display
 */
export function getCellValue(
  item: any,
  column: string,
  isBooleanLike: (val: any) => boolean,
): string {
  const value = parseItemData(item)[column];
  if (value === null || value === undefined) return "-";
  if (isBooleanLike(value)) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value).substring(0, 100);
}

/**
 * Get locked quantity for a SKU
 */
export function getLockedQty(
  sku: string,
  lockStockMap?: Record<string, number>,
): number {
  if (!sku) return 0;
  return lockStockMap?.[sku] || 0;
}

/**
 * Calculate available stock: Total - Locked Qty
 */
export function calculateAvailableStock(
  sku: string,
  total: number,
  lockStockMap?: Record<string, number>,
): number {
  if (!sku || total === null || total === undefined) return 0;
  const lockedQty = lockStockMap?.[sku] || 0;
  return Math.max(0, total - lockedQty);
}

/**
 * Get total quantity from item
 * Searches standard column names
 */
export function getTotalQty(item: any): number {
  const itemData = parseItemData(item);

  const standardFields = [
    "total",
    "Total",
    "TOTAL",
    "qty",
    "quantity",
    "Qty",
    "QUANTITY",
  ];

  for (const field of standardFields) {
    const value = itemData[field];
    if (value !== null && value !== undefined) {
      const num = parseInt(value, 10);
      if (!isNaN(num)) return num;
    }
  }

  return 0;
}

export function useInventoryItemHelpers(options: ItemHelperOptions = {}) {
  const { lockStockMap } = options;

  return {
    parseItemData,
    extractSKU,
    getCellValueRaw,
    getCellValue: (
      item: any,
      column: string,
      isBooleanLike: (v: any) => boolean,
    ) => getCellValue(item, column, isBooleanLike),
    getLockedQty: (sku: string) => getLockedQty(sku, lockStockMap),
    calculateAvailableStock: (sku: string, total: number) =>
      calculateAvailableStock(sku, total, lockStockMap),
    getTotalQty,
  };
}

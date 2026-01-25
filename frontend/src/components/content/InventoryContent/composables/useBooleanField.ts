/**
 * Boolean/checkbox field detection and handling
 * Only recognizes true/false values from Google Sheets checkbox
 */

/**
 * Get raw cell value for validation
 */
export const getCellValueRaw = (item: any, column: string): any => {
  let itemData = item.data;

  // If data is a JSON string, parse it
  if (typeof itemData === "string") {
    try {
      itemData = JSON.parse(itemData);
    } catch {
      itemData = {};
    }
  }

  // If no data field, use the item itself
  if (!itemData) {
    itemData = item;
  }

  return itemData[column];
};

/**
 * Check if cell checkbox is checked
 * Only based on true/false boolean values (from Google Sheets checkbox)
 */
export const isCellChecked = (item: any, column: string): boolean => {
  const value = getCellValueRaw(item, column);
  if (typeof value === "boolean") return value;
  if (value === "" || value === null || value === undefined) return false;
  // Only check exact "true" and "false" strings from Google Sheets
  const str = String(value).toLowerCase();
  return str === "true";
};

/**
 * Check if value is boolean field (based on true/false only)
 */
export const isBooleanLike = (value: any): boolean => {
  if (typeof value === "boolean") return true;
  if (value === "" || value === null || value === undefined) return true;
  const str = String(value).toLowerCase();
  return str === "true" || str === "false";
};

/**
 * Create toggle checkbox handler
 */
export const createToggleCellCheckbox =
  (filteredList: any[], emitCallback: (data: any) => void) =>
  (itemIndex: number, column: string) => {
    const item = filteredList[itemIndex];
    const isCurrentlyChecked = isCellChecked(item, column);
    const newValue = !isCurrentlyChecked;
    emitCallback({ itemIndex, column, value: newValue.toString() });
  };

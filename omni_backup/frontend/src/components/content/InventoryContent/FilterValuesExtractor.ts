/**
 * FilterValuesExtractor.ts
 * Extract unique values from column for filter dropdown
 */

export class FilterValuesExtractor {
  /**
   * Extract unique values from inventory data column
   */
  static extractColumnValues(
    inventoryList: Record<string, any>[],
    columnName: string,
  ): string[] {
    const values = new Set<string>();

    for (const item of inventoryList) {
      // Handle nested data field - supports both string (Node.js) and object (Go) formats
      let itemData = item;
      if (item.data !== undefined && item.data !== null) {
        if (typeof item.data === "string") {
          try {
            itemData = JSON.parse(item.data);
          } catch {
            itemData = item;
          }
        } else if (typeof item.data === "object") {
          itemData = item.data;
        }
      }

      const value = itemData[columnName];
      if (value !== null && value !== undefined) {
        values.add(String(value).trim());
      }
    }

    // Sort values
    return Array.from(values).sort();
  }

  /**
   * Extract values dari multiple items
   */
  static extractMultipleColumnValues(
    inventoryList: Record<string, any>[],
    columnNames: string[],
  ): Record<string, string[]> {
    const result: Record<string, string[]> = {};

    for (const columnName of columnNames) {
      result[columnName] = this.extractColumnValues(inventoryList, columnName);
    }

    return result;
  }
}

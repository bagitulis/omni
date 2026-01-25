/**
 * RESPONSIBILITY: Excel-like filtering with robust data extraction
 * - Parse nested data structures
 * - Case-insensitive substring matching
 * - AND logic for multiple filters
 */

class FilterEngine {
  /**
   * Extract data from item (handle nested JSON strings or objects)
   * Supports both string (Node.js) and object (Go) formats
   */
  private extractItemData(item: any): Record<string, any> {
    if (!item) return {};

    if (item.data !== undefined && item.data !== null) {
      if (typeof item.data === "string") {
        try {
          return JSON.parse(item.data);
        } catch {
          return item;
        }
      } else if (typeof item.data === "object") {
        return item.data;
      }
    }
    return item;
  }

  /**
   * Get cell value (handle nested structures)
   */
  private getCellValue(item: any, column: string): string {
    const itemData = this.extractItemData(item);
    const value = itemData[column];

    if (value === null || value === undefined) return "";
    if (typeof value === "object") {
      try {
        return JSON.stringify(value);
      } catch {
        return String(value);
      }
    }
    return String(value).trim();
  }

  /**
   * Apply search filter (all columns)
   */
  applySearchFilter(items: any[], query: string): any[] {
    if (!query || !query.trim()) return items;

    const searchTerm = query.toLowerCase().trim();
    return items.filter((item) => {
      const itemData = this.extractItemData(item);
      // Search across all column values
      return Object.keys(itemData).some((key) => {
        const cellValue = this.getCellValue(item, key);
        return cellValue.toLowerCase().includes(searchTerm);
      });
    });
  }

  /**
   * Apply column filters with AND logic (Excel-like)
   * Support both single string and multi-select array values
   */
  applyColumnFilters(
    items: any[],
    columnFilters: Record<string, string>,
  ): any[] {
    const activeFilters = Object.entries(columnFilters)
      .filter(([_, value]) => value && String(value).trim())
      .map(([column, value]) => {
        // Parse values - could be JSON array or simple string
        let filterValues: string[] = [];
        try {
          const parsed = JSON.parse(String(value));
          filterValues = Array.isArray(parsed) ? parsed : [String(value)];
        } catch {
          filterValues = [String(value)];
        }
        return { column, filterValues };
      });

    if (activeFilters.length === 0) return items;

    return items.filter((item) => {
      // ALL filters must match (AND logic) - like Excel
      return activeFilters.every(({ column, filterValues }) => {
        const cellValue = this.getCellValue(item, column).toLowerCase();
        // Check if cell value matches ANY of the filter values (OR within same column)
        return filterValues.some((term) =>
          cellValue.includes(term.toLowerCase()),
        );
      });
    });
  }

  /**
   * Apply all filters in sequence
   */
  applyAll(
    items: any[],
    searchQuery: string,
    columnFilters: Record<string, string>,
  ): any[] {
    let result = [...items];
    result = this.applySearchFilter(result, searchQuery);
    result = this.applyColumnFilters(result, columnFilters);
    return result;
  }
}

export function useInventoryFilterEngine() {
  const engine = new FilterEngine();

  return {
    applySearchFilter: (items: any[], query: string) =>
      engine.applySearchFilter(items, query),

    applyColumnFilters: (items: any[], filters: Record<string, string>) =>
      engine.applyColumnFilters(items, filters),

    applyAll: (items: any[], search: string, filters: Record<string, string>) =>
      engine.applyAll(items, search, filters),
  };
}

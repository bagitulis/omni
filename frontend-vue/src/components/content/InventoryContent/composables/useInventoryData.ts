import {
  fetchWithRetry,
  isSuccessResponse,
  DEFAULT_STATS,
} from "./useInventoryFetchHelpers";
import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

const API_BASE_URL = getApiBaseUrl("/inventory");

/**
 * Extract column schema from inventory data items
 * Fallback when stats don't return columns
 */
function extractColumnsFromData(items: any[]): any[] {
  if (!items || items.length === 0) return [];

  const columnsSet = new Set<string>();
  const skuFields = ["sku", "SKU", "Sku", "sku_id", "seller_sku", "sellerSku"];

  // Sample first few items to extract column names
  const sampleSize = Math.min(items.length, 5);
  for (let i = 0; i < sampleSize; i++) {
    const item = items[i];
    let itemData = item;

    // Parse JSON data - supports both string (Node.js) and object (Go) formats
    if (item?.data !== undefined && item?.data !== null) {
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

    // Extract all keys from item data
    if (typeof itemData === "object" && itemData !== null) {
      Object.keys(itemData).forEach((key) => columnsSet.add(key));
    }
  }

  // Convert to column schema format
  const columns: any[] = [];
  let position = 1;
  columnsSet.forEach((colName) => {
    const isKey = skuFields.some(
      (f) => f.toLowerCase() === colName.toLowerCase(),
    );
    columns.push({
      column_name: colName,
      column_type: "text",
      is_key: isKey,
      spreadsheet_column: String.fromCharCode(64 + position),
      column_position: position,
    });
    position++;
  });

  return columns;
}

export function useInventoryData(state: any) {
  // Track pending requests to prevent duplicates
  let pendingLoadStats: Promise<void> | null = null;
  let pendingLoadInventory: Promise<void> | null = null;
  let pendingLoadConfiguration: Promise<void> | null = null;

  /**
   * Load inventory list data
   */
  async function loadInventoryData(retries = 3): Promise<void> {
    if (pendingLoadInventory) return pendingLoadInventory;

    pendingLoadInventory = (async (): Promise<void> => {
      try {
        state.loading = true;

        const result = await fetchWithRetry("/list?offset=0&limit=10000", {
          timeout: 30000,
          retryTimeout: 60000,
        });

        if (isSuccessResponse(result)) {
          state.inventoryList = result.data || [];
          state.totalRecords = result.meta?.total ?? result.total ?? 0;
          console.log("✅ Loaded", state.inventoryList.length, "items");
        } else {
          throw new Error(result.message || "Failed to load inventory");
        }
      } catch (error) {
        console.error("❌ Error loading inventory:", error);

        if (retries > 0) {
          console.log(`🔄 Retrying... (${retries} attempts left)`);
          await new Promise((resolve) => setTimeout(resolve, 1000));
          pendingLoadInventory = null;
          return loadInventoryData(retries - 1);
        }

        state.inventoryList = [];
        state.totalRecords = 0;
        throw error;
      } finally {
        state.loading = false;
        pendingLoadInventory = null;
      }
    })();

    return pendingLoadInventory;
  }

  /**
   * Load inventory statistics and schema columns
   */
  async function loadStats(): Promise<void> {
    if (pendingLoadStats) return pendingLoadStats;

    pendingLoadStats = (async (): Promise<void> => {
      try {
        const result = await fetchWithRetry("/stats", {
          timeout: 30000,
          retryTimeout: 60000,
        });

        if (isSuccessResponse(result)) {
          const statsData = result.data || result;
          state.stats = { ...DEFAULT_STATS, ...statsData };

          // Process columns with visibility filtering
          let columns = statsData.columns || [];
          columns = filterColumnsByVisibility(columns, state);
          state.schemaColumns = columns;

          // Fallback: extract columns from inventory data if needed
          if (
            state.schemaColumns.length === 0 &&
            state.inventoryList?.length > 0
          ) {
            const extracted = extractColumnsFromData(state.inventoryList);
            if (extracted.length > 0) {
              state.schemaColumns = extracted;
              console.log(`✅ Extracted ${extracted.length} columns from data`);
            }
          }

          // Set key column
          const keyCol = state.schemaColumns.find((c: any) => c.is_key);
          state.keyColumn =
            keyCol?.column_name ||
            keyCol?.name ||
            state.schemaColumns[0]?.column_name ||
            state.schemaColumns[0]?.name ||
            "";
        } else {
          state.stats = { ...DEFAULT_STATS };
          state.schemaColumns = [];
        }
      } catch (error) {
        console.error("❌ Error loading stats:", error);
        state.stats = { ...DEFAULT_STATS };
        state.schemaColumns = [];
      } finally {
        pendingLoadStats = null;
      }
    })();

    return pendingLoadStats;
  }

  /**
   * Load inventory configuration
   */
  async function loadConfiguration(): Promise<void> {
    if (pendingLoadConfiguration) return pendingLoadConfiguration;

    pendingLoadConfiguration = (async (): Promise<void> => {
      try {
        // Fetch config and selected columns in parallel
        const [configResult, columnsResult] = await Promise.all([
          fetchWithRetry("/config", { timeout: 30000, retryTimeout: 60000 }),
          fetchWithRetry("/columns/selected", {
            timeout: 30000,
            retryTimeout: 60000,
          }),
        ]);

        if (isSuccessResponse(configResult)) {
          state.configData = configResult.data;
        }

        // Merge selected columns if available
        if (
          isSuccessResponse(columnsResult) &&
          columnsResult.selected_columns?.length > 0
        ) {
          state.configData = state.configData || {};
          state.configData.selected_columns = columnsResult.selected_columns;
        }
      } catch (error) {
        console.error("❌ Error loading configuration:", error);
      } finally {
        pendingLoadConfiguration = null;
      }
    })();

    return pendingLoadConfiguration;
  }

  return {
    loadInventoryData,
    loadStats,
    loadConfiguration,
  };
}

/**
 * Filter columns by visibility settings
 */
function filterColumnsByVisibility(columns: any[], state: any): any[] {
  // Priority 1: Use visibleColumns from filter-preferences
  const visibleFromState = state.columnVisibility
    ? Object.keys(state.columnVisibility).filter(
        (colName: string) => state.columnVisibility[colName] === true,
      )
    : [];

  if (visibleFromState.length > 0) {
    return columns.filter((col: any) => {
      const colName = col.column_name || col.name;
      return visibleFromState.includes(colName);
    });
  }

  // Priority 2: Fallback to selected_columns from config
  if (state.configData?.selected_columns?.length > 0) {
    return columns.filter((col: any) => {
      const colName = col.column_name || col.name;
      return state.configData.selected_columns.includes(colName);
    });
  }

  return columns;
}

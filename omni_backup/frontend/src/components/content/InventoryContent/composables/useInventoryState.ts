import { reactive } from "vue";

/**
 * Default inventory state factory
 * Creates initial state for InventoryContent component
 */
export function createInventoryState() {
  return reactive({
    // Inventory data
    inventoryList: [] as any[],
    schemaColumns: [] as any[],
    stats: {
      total_records: 0,
      total_columns: 0,
      columns: [] as string[],
      last_sync: null as string | null,
      db_size_kb: 0,
    },
    syncHistory: [] as any[],

    // UI state
    loading: false,
    syncing: false,
    showAddForm: false,
    showSyncHistory: false,
    searchQuery: "",
    currentOffset: 0,
    pageSize: 50,
    totalRecords: 0,

    // Lock stock data
    lockStock: {
      loading: false,
      error: null as string | null,
      map: {} as Record<string, number>,
    },

    // Filter state
    showFilterPanel: false,
    columnVisibility: {} as Record<string, boolean>,
    columnFilters: {} as Record<string, string>,
    isInitialized: false,
    sortColumn: null as string | null,
    sortDirection: null as "asc" | "desc" | null,

    // Configuration
    configData: {
      spreadsheet_id: "",
      sheet_name: "",
      selected_columns: [] as string[],
      key_column: "",
      header_row: 1,
      data_start_row: 2,
      auto_sync: false,
      sync_interval_seconds: 300,
      last_sync_timestamp: null as string | null,
    },

    // Form data
    formData: {} as Record<string, any>,
    editingItem: null as any,
    editingCell: {} as Record<string, any>,
    keyColumn: "",

    // Alerts
    syncAlert: null as any,
    syncStatus: null as any,

    // Batch SKU check
    batchCheckResults: [] as any[],
    batchCheckLoading: false,

    // Stock update
    updatingStock: false,

    // Marketplace settings modal
    showMarketplaceSettings: false,
  });
}

export type InventoryState = ReturnType<typeof createInventoryState>;

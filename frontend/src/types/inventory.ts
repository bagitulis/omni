export interface InventoryRecord {
  id: string;
  key_value: string;
  key_column_name: string;
  data: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface InventoryListResult {
  records: InventoryRecord[];
  total: number;
  offset: number;
  limit: number;
}

export interface InventoryConfig {
  id: string;
  tenant_id: string;
  spreadsheet_id: string;
  sheet_name: string;
  selected_columns: string;
  all_columns: string;
  header_row: number;
  data_start_row: number;
  key_column: string;
  auto_sync: boolean;
  sync_interval_seconds: number;
  last_sync_timestamp: string | null;
  last_headers_hash: string;
  last_sync_status: string;
  low_stock_threshold?: number;
  created_at: string;
  updated_at: string;
}

export interface InventoryStats {
  total_records: number;
  total_columns: number;
  last_sync?: string;
  db_size_kb?: number;
}

export interface SyncHistoryEntry {
  id: number;
  tenant_id: string;
  sync_type: "import" | "export";
  status: "success" | "failed" | "partial";
  started_at: string;
  completed_at: string;
  records_processed: number;
  errors?: string;
}

export interface BatchCheckResult {
  sku: string;
  platform: string;
  status: "synced" | "mismatch" | "missing";
  remote_stock?: number;
  remote_price?: number;
  local_stock?: number;
  local_price?: number;
  message?: string;
}

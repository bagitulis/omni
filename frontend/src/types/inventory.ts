export interface PlatformStatusItem {
  platform: string;
  platform_product_id: string;
  platform_item_id?: string;
  platform_sku?: string;
  status: string;
  stock: number;
  price: number;
}

export interface InventoryRecord {
  id: string;
  key_value: string;
  key_column_name: string;
  data: Record<string, unknown>;
  sync_status?: string;
  platform_status?: PlatformStatusItem[];
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
  total_column?: string;
  raw_total_column?: string;
  auto_sync: boolean;
  sync_interval_seconds: number;
  last_sync_timestamp: string | null;
  last_headers_hash: string;
  last_sync_status: string;
  low_stock_threshold?: number;
  price_column?: string;
  price_column_shopee?: string;
  price_column_tiktok?: string;
  price_column_lazada?: string;
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
  id: string;
  tenant_id: string;
  status: "SUCCESS" | "ERROR" | "PARTIAL";
  total_records: number;
  new_records: number;
  updated_records: number;
  unchanged_records: number;
  failed_records: number;
  duration_ms: number;
  error_message?: string;
  headers_changed: boolean;
  synced_at: string;
  created_at: string;
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

export interface SkuCheckResult {
  sku: string;
  shopee: boolean;
  lazada: boolean;
  tiktok: boolean;
}

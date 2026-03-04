/**
 * Shared Type Definitions
 * Used across unified product management components
 * API types use snake_case to match backend JSON response
 */

// === Platform Indicator Types ===
export type Platform = "shopee" | "tiktok" | "lazada";
export type PlatformSyncState = "idle" | "syncing" | "success" | "error";
export type PlatformLinkStatus = "linked" | "not_linked" | "pending" | "error";

export interface PlatformIndicatorData {
  platform: Platform;
  linked: boolean;
  has_update: boolean;
  sync_state: PlatformSyncState;
  platform_product_id?: string;
  platform_item_id?: string;
  platform_sku_id?: string;
  last_synced_at?: string;
  error_message?: string;
}

// === Inline Edit Types ===
export type InlineEditMode = "price" | "stock";

export interface InlineEditCellProps {
  value: number;
  mode: InlineEditMode;
  onSave: (newValue: number) => Promise<void>;
  disabled?: boolean;
  min?: number;
  prefix?: string; // "Rp" for price, "" for stock
}

// === Batch Action Types ===
export type BatchActionType =
  | "update_price"
  | "sync_stock"
  | "sync_marketplace"
  | "bulk_pricing"
  | "clone"
  | "delete_products";

export interface BatchActionItem {
  key: BatchActionType;
  label: string;
  icon: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
  platformRestriction?: Platform[]; // e.g., wholesale = ["shopee"] only
}

// === Column Manager Types ===
export interface ColumnConfig {
  key: string;
  title: string;
  visible: boolean;
  locked?: boolean; // Cannot be hidden (e.g., Name, SKU)
  order: number;
  width?: number;
  resizable?: boolean;
}

export interface ColumnManagerProps {
  columns: ColumnConfig[];
  onChange: (columns: ColumnConfig[]) => void;
  onReset: () => void;
}

// === Filter Types ===
export interface ProductFilterValues {
  search: string;
  platform: Platform | "all";
  status: "active" | "archived" | "draft" | "all";
  category: string | "all";
  mapping: "all" | "mapped" | "unmapped"; // GAP-16: filter by platform link status
}

// === Stock Sync Types ===
export type StockSyncMode = "uniform" | "per_platform";

export interface StockSyncItem {
  sku: string;
  platforms?: Platform[]; // Only for per_platform mode
}

// === Platform Price (real marketplace price from staging tables) ===
export interface PlatformPrice {
  platform: Platform;
  platform_price: number;
  platform_stock: number;
}

// === Unified Product Row (for main table) ===
export interface UnifiedProductRow {
  id: number;
  title: string;
  description: string;
  images: string[];
  status: "active" | "archived" | "draft";
  skus: Array<{
    id: number;
    seller_sku: string;
    variant_name: string;
    price: number;
    stock: number;
    platform_links: Array<{
      platform: Platform;
      platform_product_id?: string;
      platform_item_id?: string;
      platform_sku_id?: string;
      sync_status: "pending" | "synced" | "error" | "outdated";
      last_synced_at?: string;
      error_message?: string;
    }>;
    // Opsi A: Per-platform real prices from staging tables
    platform_prices?: PlatformPrice[];
    // Inventory reference price/stock from Google Sheets
    inventory_price?: number;
    inventory_stock?: number;
  }>;
  // Computed fields for table display
  primary_sku: string; // First SKU's seller_sku
  primary_price: number; // First SKU's price
  primary_stock: number; // First SKU's stock
  platform_summary: {
    // Aggregated from all SKUs' platform_links
    shopee: PlatformLinkStatus;
    tiktok: PlatformLinkStatus;
    lazada: PlatformLinkStatus;
  };
  created_at: string;
  updated_at: string;
}

// === Image Gallery Types ===
// Backend Image model returns: id, tenant_id, filename, content_hash, original_url,
// local_path, mime_type, category, product_id, width, height, file_size, ref_count,
// created_at, updated_at. URLs are constructed from local_path:
//   thumb:    `/${local_path}/thumb.webp`
//   medium:   `/${local_path}/medium.webp`
//   original: `/${local_path}/original.webp`
export interface GalleryImage {
  id: number;
  filename: string;
  content_hash: string;
  local_path: string; // Base path — construct URLs as `/${local_path}/{size}.webp`
  mime_type: string;
  width: number;
  height: number;
  file_size: number;
  ref_count: number;
  created_at: string;
}

// Helper to construct image URLs from local_path
// Usage: getImageUrl(image.local_path, "thumb") → "/uploads/tenant/images/hash/thumb.webp"
export const getImageUrl = (
  localPath: string,
  size: "thumb" | "medium" | "original",
): string => `/${localPath}/${size}.webp`;

export interface GalleryPaginationMeta {
  page: number;
  pages: number;
  total: number;
  page_size: number;
}

// === Marketplace Sync History Types ===
export type SyncOperation =
  | "stock_update"
  | "price_update"
  | "wholesale_update"
  | "mpq_update"
  | "clone";
export type SyncResultStatus = "success" | "failed" | "partial";

export interface MarketplaceSyncHistoryEntry {
  id: string;
  tenant_id: string;
  sku: string;
  platform: Platform;
  operation: SyncOperation;
  status: SyncResultStatus;
  request_data?: string;
  response_data?: string;
  error_message?: string;
  created_at: string;
}

export interface MarketplaceSyncHistoryListResult {
  entries: MarketplaceSyncHistoryEntry[];
  total: number;
  page: number;
  page_size: number;
}

export interface MarketplaceSyncHistoryFilter {
  page?: number;
  page_size?: number;
  platform?: Platform;
  operation?: SyncOperation;
  status?: SyncResultStatus;
  sku_search?: string;
  date_from?: string;
  date_to?: string;
}

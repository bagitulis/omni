/**
 * Wholesale Types
 * Interfaces for wholesale operations across platforms
 */

export interface WholesaleTier {
  min_qty: number;
  max_qty: number;
  price: number;
}

export interface WholesaleInfo {
  item_id: number;
  product_name: string;
  has_wholesale: boolean;
  tiers: WholesaleTier[];
}

export interface WholesaleResult {
  success: boolean;
  item_id?: number;
  message?: string;
  error?: string;
  data?: Record<string, unknown>;
}

export interface BatchDeleteBySkusResult {
  success: boolean;
  message: string;
  data: BatchDeleteData;
}

export interface BatchDeleteData {
  total_skus: number;
  unique_items: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: WholesaleResult[];
}

export interface SkuLookupResult {
  item_id: string;
  model_id: string | null;
  seller_sku: string | null;
  product_name?: string;
}

export interface WholesaleSettings {
  tenant_id: string;
  platform: string;
  admin_fee: number;
  min_order_1: number;
  max_order_1: number;
  max_order_tier_3: number;
}

export interface WholesaleTierCalculated {
  tier: number;
  min_count: number;
  max_count: number;
  unit_price: number;
}

export interface BatchUpdateItem {
  sku: string;
  price: number;
}

export interface BatchUpdateBySkusResult {
  success: boolean;
  message: string;
  data: BatchUpdateData;
}

export interface BatchUpdateData {
  total_skus: number;
  unique_items: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: WholesaleResult[];
  settings_used: WholesaleSettings;
}

export interface BatchMpqResult {
  success: boolean;
  message: string;
  data: BatchMpqData;
}

export interface BatchMpqData {
  total_skus: number;
  unique_items: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: MpqItemResult[];
}

export interface MpqItemResult {
  success: boolean;
  item_id: number;
  error?: string;
  message?: string;
}

export interface TiktokBatchMpqResult {
  success: boolean;
  message: string;
  data: TiktokBatchMpqData;
}

export interface TiktokBatchMpqData {
  total_skus: number;
  unique_products: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: TiktokMpqItemResult[];
}

export interface TiktokMpqItemResult {
  success: boolean;
  product_id: string;
  error?: string;
  message?: string;
}

export interface TierPreviewResult {
  base_price: number;
  settings: WholesaleSettings;
  tiers: WholesaleTierCalculated[];
}

export interface BatchWholesaleResetData {
  total_skus: number;
  unique_items: number;
  processed: number;
  failed: number;
  skipped: string[];
  results: WholesaleResult[];
  settings_used: WholesaleSettings;
  success: boolean;
  message: string;
}

export interface BatchWholesaleResetResult {
  success: boolean;
  data: BatchWholesaleResetData;
  message?: string;
}

// --- Inventory Module Wholesale/MPQ Types ---

export interface InventoryWholesaleTier {
  id?: string;
  sku: string;
  min_qty: number;
  price: number;
  discount_percent?: number;
}

export interface InventoryWholesaleInfo {
  sku: string;
  tiers: InventoryWholesaleTier[];
  created_at?: string;
  updated_at?: string;
}

export interface InventoryWholesaleSettings {
  enabled: boolean;
  default_tiers?: InventoryWholesaleTier[];
  apply_to_all?: boolean;
}

export interface InventoryWholesaleBatchUpdateItem {
  sku: string;
  tiers: InventoryWholesaleTier[];
}

export interface InventoryWholesaleBatchResult {
  total: number;
  successful: number;
  failed: number;
  errors?: string[];
}

export interface InventoryMpqSettings {
  sku: string;
  min_purchase_qty: number;
  enabled: boolean;
}

export interface InventoryMpqBatchItem {
  sku: string;
  min_purchase_qty: number;
  enabled?: boolean;
}

export interface InventoryMpqBatchResult {
  total: number;
  successful: number;
  failed: number;
  errors?: string[];
}

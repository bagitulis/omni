/**
 * Analytics/Report Types
 *
 * Matches backend DTOs in backend/internal/dto/analytics_dto.go
 * All fields use snake_case to match backend JSON responses.
 */

/** Report settings for a platform (price_column, formula_deduction, formula_multiplier) */
export interface ReportSettings {
  price_column: string;
  formula_deduction: number;
  formula_multiplier: number;
}

/** Sync status for a given month/year */
export interface SyncStatus {
  synced: boolean;
  total_orders: number;
  failed_orders: number;
  synced_at: string | null; // ISO datetime or null
}

/** Request body to trigger sync */
export interface SyncRequest {
  month: number;
  year: number;
  force_resync: boolean;
}

/** Result returned after triggering a sync */
export interface SyncResult {
  job_id: string;
}

/** Summary statistics for reconciliation */
export interface ReconciliationSummary {
  total_sku: number;
  total_transactions: number;
  sku_ok: number;
  sku_with_price_diff: number;
  sku_no_inventory: number;
}

/** A single SKU group entry in reconciliation details */
export interface SkuGroup {
  sku: string;
  item_name: string;
  total_quantity: number;
  total_amount: number;
  system_amount: number;
  price_diff: number;
  price_diff_percent: number;
  order_count: number;
}

/** Full reconciliation result (summary + details) */
export interface ReconciliationResult {
  summary: ReconciliationSummary;
  details: SkuGroup[];
}

/** Summary statistics for shipping fee analysis */
export interface ShippingFeeSummary {
  total_orders: number;
  orders_with_difference: number;
  total_profit: number;
  total_loss: number;
  net_impact: number;
}

/** Shopee shipping order entry */
export interface ShopeeShippingOrder {
  order_sn: string;
  platform_fee: number;
  actual_fee: number;
  difference: number;
  status: string;
  order_date: string;
}

/** Shopee shipping fee analysis result */
export interface ShopeeShippingFeeResult {
  summary: ShippingFeeSummary;
  details: ShopeeShippingOrder[];
}

/** TikTok shipping order entry */
export interface TiktokShippingOrder {
  order_sn: string;
  shipping_fee: number;
  actual_fee: number;
  difference: number;
  status: string;
  order_date: string;
}

/** TikTok shipping fee analysis result */
export interface TiktokShippingFeeResult {
  summary: ShippingFeeSummary;
  details: TiktokShippingOrder[];
}

/** Supported report platforms */
export type ReportPlatform = "shopee" | "tiktok";

/** Report tab options */
export type ReportTab = "reconciliation" | "shipping_fee";

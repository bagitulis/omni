/**
 * Analytics Types
 * API types use snake_case to match backend JSON response
 */

// Unified Analytics Types

export interface KPIData {
  total_products: number;
  avg_roas: number;
  actions: {
    scale_up: number;
    maintain: number;
    reduce: number;
    stop: number;
  };
}

export interface PlatformSummary {
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  avg_roas: number;
  products_count: number;
}

export interface UnifiedSummary {
  combined: PlatformSummary;
  tiktok: PlatformSummary;
  shopee: PlatformSummary;
}

// Platform type for parameterized analytics
export type Platform = "shopee" | "tiktok";

// Shared types

export interface SyncStatus {
  synced: boolean;
  total_orders: number;
  synced_at: string | null;
}

export interface AnalyticsSettings {
  price_column: string;
  formula_deduction: number;
  formula_multiplier: number;
}

export interface ReconciliationSummary {
  total_sku: number;
  total_transactions: number;
  sku_ok: number;
  sku_with_price_diff: number;
  sku_no_inventory: number;
}

export interface ShippingFeeSummary {
  total_orders: number;
  orders_with_difference: number;
  total_profit: number;
  total_loss: number;
  net_impact: number;
}

export interface JobProgress {
  id: string;
  type: string;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  progress_percent: number;
  progress_message: string;
  total_items: number;
  processed_items: number;
  error_message?: string;
  result_data?: string;
  started_at?: string;
  completed_at?: string;
}

// Shopee-specific types

export interface TransactionDetail {
  order_sn: string;
  order_date: string;
  price: number;
  escrow_amount: number;
  quantity: number;
}

export interface PriceVariant {
  price: number;
  count: number;
  total_escrow: number;
  avg_escrow: number;
  transactions: TransactionDetail[];
}

export interface SkuGroup {
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  inventory_price: number | null;
  expected_income: number | null;
  total_transactions: number;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  price_variants: PriceVariant[];
  has_multiple_prices: boolean;
  has_price_difference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface ReconciliationResult {
  summary: ReconciliationSummary;
  sku_groups: SkuGroup[];
}

export interface ShopeeShippingFeeOrder {
  order_sn: string;
  order_date: string | null;
  buyer_paid: number;
  actual_fee: number;
  shopee_rebate: number;
  difference: number;
  buyer_name: string | null;
  payment_method: string | null;
}

export interface ShopeeShippingFeeResult {
  summary: ShippingFeeSummary;
  orders: ShopeeShippingFeeOrder[];
}

// TikTok-specific types

export interface TiktokSkuGroup {
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  inventory_price: number | null;
  expected_income: number | null;
  total_transactions: number;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  has_multiple_prices: boolean;
  has_price_difference: boolean;
  status: "OK" | "PRICE_DIFF" | "NO_INVENTORY";
}

export interface TiktokReconciliationResult {
  summary: ReconciliationSummary;
  sku_groups: TiktokSkuGroup[];
}

export interface TiktokShippingFeeOrder {
  order_id: string;
  order_date: string | null;
  buyer_paid: number;
  actual_fee: number;
  platform_discount: number;
  difference: number;
  order_status: string | null;
  currency: string;
}

export interface TiktokShippingFeeResult {
  summary: ShippingFeeSummary;
  orders: TiktokShippingFeeOrder[];
}

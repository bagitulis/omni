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

/** Shopee SKU group entry in reconciliation details */
export interface SkuGroup {
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  variant_name: string;
  inventory_price: number | null;
  expected_income: number | null;
  total_transactions: number;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  has_multiple_prices: boolean;
  has_price_difference: boolean;
  status: string;
}

/** TikTok SKU group entry in reconciliation details */
export interface TiktokSkuGroup {
  sku: string;
  seller_sku: string;
  product_name: string;
  variant_name: string;
  inventory_price: number | null;
  expected_income: number | null;
  total_transactions: number;
  unique_unit_prices: number[];
  unique_actual_incomes: number[];
  has_multiple_prices: boolean;
  has_price_difference: boolean;
  status: string;
}

/** Full reconciliation result (summary + details) */
export interface ReconciliationResult {
  summary: ReconciliationSummary;
  sku_groups: SkuGroup[];
}

export interface TiktokReconciliationResult {
  summary: ReconciliationSummary;
  sku_groups: TiktokSkuGroup[];
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
  buyer_paid: number;
  actual_fee: number;
  shopee_rebate: number;
  difference: number;
  status: string;
  order_date: string;
  buyer_name: string;
  payment_method: string;
}

/** Shopee shipping fee analysis result */
export interface ShopeeShippingFeeResult {
  summary: ShippingFeeSummary;
  details: ShopeeShippingOrder[];
}

/** TikTok shipping order entry */
export interface TiktokShippingOrder {
  order_sn: string;
  customer_paid: number;
  actual_fee: number;
  platform_discount: number;
  difference: number;
  status: string;
  order_date: string;
  order_status: string;
  currency: string;
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

export interface AnalyticsDetailParams {
  sku?: string;
  order_sn?: string;
  month: number;
  year: number;
}

export interface ShopeeSkuOrder {
  id: string;
  order_sn: string;
  escrow_amount: number;
  commission_fee: number;
  service_fee: number;
  seller_processing_fee: number;
  buyer_paid_shipping_fee: number;
  actual_shipping_fee: number;
  shopee_shipping_rebate: number;
  estimated_shipping_fee: number;
  buyer_total_amount: number;
  buyer_name: string;
  payment_method: string;
  order_date: string;
  item_name: string;
  model_name: string;
  sku: string;
  model_sku: string;
  quantity: number;
  original_price: number;
}

export interface ShopeeSkuOrdersResult {
  orders: ShopeeSkuOrder[];
}

export interface ShopeeOrderItem {
  id: string;
  escrow_order_id: string;
  item_id?: number | null;
  model_id?: number | null;
  sku: string;
  model_sku: string;
  item_name: string;
  model_name: string;
  quantity: number;
  original_price: number;
  selling_price: number;
  discounted_price: number;
  seller_discount: number;
  shopee_discount: number;
  discount_from_coin: number;
  discount_from_voucher_seller: number;
  discount_from_voucher_shopee: number;
  ams_commission_fee: number;
  seller_order_processing_fee: number;
}

export interface ShopeeOrderItemsResult {
  items: ShopeeOrderItem[];
}

export interface TiktokSkuOrder {
  id: string;
  order_id: string;
  order_status: string;
  total_settlement_amount: number;
  product_revenue: number;
  platform_commission: number;
  transaction_fee: number;
  shipping_fee_customer_paid: number;
  shipping_fee_actual: number;
  shipping_fee_platform_discount: number;
  seller_shipping_discount: number;
  refund_amount: number;
  currency: string;
  buyer_name: string;
  order_date: string;
  product_name: string;
  seller_sku: string;
  quantity: number;
  sale_price: number;
  original_price: number;
}

export interface TiktokSkuOrdersResult {
  orders: TiktokSkuOrder[];
}

export interface TiktokOrderItem {
  id: string;
  escrow_order_id: string;
  product_name: string;
  sku_id: string;
  seller_sku: string;
  quantity: number;
  sale_price: number;
  original_price: number;
  subtotal_after_seller_discount: number;
  platform_discount: number;
  seller_discount: number;
  commission: number;
  transaction_fee_item: number;
  settlement_amount: number;
}

export interface TiktokOrderItemsResult {
  items: TiktokOrderItem[];
}

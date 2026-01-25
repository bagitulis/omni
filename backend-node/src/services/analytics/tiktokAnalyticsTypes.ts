/**
 * TikTok Analytics Types
 * Shared types for TikTok analytics services
 * Single Responsibility: Type definitions
 */

export interface SyncResult {
  success: boolean;
  month: number;
  year: number;
  totalOrders: number;
  totalItems: number;
  failedOrders: number;
  message: string;
}

export interface TiktokOrder {
  id: string; // TikTok API returns 'id', not 'order_id'
  order_status: string;
  create_time: number;
  buyer_message?: string;
  payment_info?: {
    total_amount?: number;
    currency?: string;
  };
  line_items?: TiktokLineItem[];
}

export interface TiktokLineItem {
  product_id: string;
  product_name: string;
  sku_id: string;
  seller_sku?: string;
  quantity: number;
  sale_price: string;
  original_price?: string;
}

export interface TiktokTransaction {
  transaction_id: string;
  transaction_type: string;
  order_id: string;
  statement_time: number;
  total_settlement_amount: string;
  revenue_details?: RevenueDetails;
  fee_details?: FeeDetails;
  order_line_item_list?: TransactionItem[];
}

export interface RevenueDetails {
  product_refund_amount?: string;
  customer_order_refund?: string;
  customer_shipping_fee_refund?: string;
  tiktok_shipping_incentive_refund?: string;
  other_refund?: string;
}

export interface FeeDetails {
  platform_commission?: string;
  affiliate_commission?: string;
  affiliate_partner_commission?: string;
  transaction_fee?: string;
  shipping_fee?: string;
  customer_paid_shipping_fee_after_discount?: string;
  fbt_fulfillment_fee?: string;
  customer_shipping_fee?: string;
  shipping_cost?: string;
  shipping_fee_subsidy?: string;
  return_shipping_fee?: string;
  adjustment_amount?: string;
}

export interface TransactionItem {
  product_id: string;
  product_name: string;
  sku_id: string;
  seller_sku?: string;
  quantity: number;
  sale_price: string;
  original_price?: string;
  subtotal_after_seller_discount?: string;
  platform_discount?: string;
  seller_discount?: string;
  sku_settlement_amount?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type PrismaClientAny = any;

/**
 * Parse TikTok amount string to number
 */
export function parseAmount(val: string | undefined): number {
  return val ? parseFloat(val) : 0;
}

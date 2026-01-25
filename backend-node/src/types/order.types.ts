/**
 * Order Data Types
 * Mirrors Python backend's order structures
 */

/**
 * Generic order data structure
 * Shared properties across all platforms
 */
export interface OrderData {
  order_sn: string; // Order number/ID
  order_id?: string; // Alternative order ID
  order_status: string; // Status like UNPAID, READY_TO_SHIP, etc
  shop_id?: number; // Shop ID
  order_timestamp?: number; // Order creation time (unix timestamp)
  items?: OrderItem[]; // Order line items
  created_at?: string; // ISO datetime when created
  updated_at?: string; // ISO datetime when last updated
}

/**
 * Generic order item/line item
 */
export interface OrderItem {
  item_id?: string | number;
  item_name?: string;
  quantity?: number;
  price?: number;
  sku?: string;
}

/**
 * Shopee-specific order
 */
export interface ShopeeOrderData extends OrderData {
  order_timestamp?: number; // Shopee uses timestamp
  model_id?: number;
  item_sku?: string;
  model_sku?: string;
}

/**
 * Lazada-specific order
 */
export interface LazadaOrderData extends OrderData {
  sku_id?: string;
  seller_sku?: string;
}

/**
 * TikTok-specific order
 */
export interface TiktokOrderData extends OrderData {
  line_item_id?: string;
  product_id?: number;
  sku_id?: string;
  seller_sku?: string;
}

/**
 * Order sync result
 */
export interface OrderSyncResult {
  platform: string;
  category: string;
  success: boolean;
  count: number;
  message?: string;
  error?: string;
}

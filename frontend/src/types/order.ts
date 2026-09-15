/**
 * Order Types
 * API types use snake_case to match backend JSON response
 */

export interface Order {
  id: string;
  order_sn: string; // Mapped from order_no
  order_no: string; // Backend field name
  order_status: string; // Mapped from status
  status: string; // Backend field name
  platform: string;
  category: string;
  buyer_username: string;
  total_amount: number;
  currency: string;
  payment_method: string;
  shipping_carrier: string;
  tracking_number?: string;
  ship_by_date: number;
  countdown?: string; // Pre-computed by backend, e.g. "2d 5h"
  buyer_message?: string;
  sku: string;
  product_name: string;
  variation_name: string;
  qty: number;
  price: number;
  product_image: string;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  item_id: string;
  item_name: string;
  item_sku: string;
  quantity: number;
  price: number;
  total: number;
  order_item_id?: string; // Lazada-specific
}

export interface OrderDetail extends Order {
  items: OrderItem[];
  shipping_address?: ShippingAddress;
  buyer_email?: string;
  buyer_phone?: string;
  /**
   * Lazada PII masking metadata (rolled out 2026-07-01, docId=145548).
   * When present and `any_masked` is true, the listed `masked_fields` were
   * returned by Lazada as e.g. "J***s" / "+62812***56". Downstream UI should
   * render the raw masked string and surface a "Apply for DBS unmasking"
   * affordance; DO NOT try to parse or synthesize the underlying value.
   *
   * Only populated for orders originating from Lazada. Shopee, TikTok, and
   * other platforms leave this undefined.
   */
  pii_mask?: {
    any_masked: boolean;
    masked_fields: string[];
  };
}

export interface ShippingAddress {
  receiver_name: string;
  receiver_phone: string;
  address: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;
}

/**
 * Platform counts for order breakdown
 * e.g., { shopee: 13, tiktok: 5, lazada: 2 }
 */
export interface PlatformCounts {
  shopee?: number;
  tiktok?: number;
  lazada?: number;
  [key: string]: number | undefined;
}

/**
 * Backend response format for order lists
 * Backend returns: { success, count, data: [...], items: [...], platform_counts: {...} }
 */
export interface BackendOrderResponse {
  count: number;
  data: Order[];
  items: Order[];
  platform_counts?: PlatformCounts;
}

/**
 * Frontend-friendly format for order lists
 */
export interface OrderListResponse {
  orders: Order[];
  total: number;
  page: number;
  page_size: number;
  platform_counts?: PlatformCounts;
}

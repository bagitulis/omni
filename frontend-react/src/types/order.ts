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
  ship_by_date: number;
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
}

export interface OrderDetail extends Order {
  items: OrderItem[];
  shipping_address?: ShippingAddress;
  buyer_email?: string;
  buyer_phone?: string;
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
 * Backend response format for order lists
 * Backend returns: { success, count, data: [...], items: [...] }
 */
export interface BackendOrderResponse {
  count: number;
  data: Order[];
  items: Order[];
}

/**
 * Frontend-friendly format for order lists
 */
export interface OrderListResponse {
  orders: Order[];
  total: number;
  page: number;
  page_size: number;
}

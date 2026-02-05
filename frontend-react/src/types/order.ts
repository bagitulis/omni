/**
 * Order Types
 * API types use snake_case to match backend JSON response
 */

export interface Order {
  order_sn: string;
  order_status: string;
  platform: string;
  buyer_username: string;
  total_amount: number;
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

export interface OrderListResponse {
  orders: Order[];
  total: number;
  page: number;
  page_size: number;
}

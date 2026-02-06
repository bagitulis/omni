export interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  price?: number;
  product_image?: string;
}

export interface GroupedOrder {
  order_no: string;
  order_sn: string; // Ensure we have a unique ID for keys
  platform: string;
  status: string;
  buyer_username: string;
  total_amount: number;
  currency: string;
  payment_method?: string;
  shipping_carrier?: string;
  ship_by_date?: number;
  items: OrderItem[];
  // Include other fields from Order if needed for actions
  [key: string]: any;
}

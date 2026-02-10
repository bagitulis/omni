import { Order } from "@/types/order";

export interface OrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  price: number;
  product_image?: string;
}

export interface GroupedOrder {
  key: string;
  order_no: string;
  order_sn?: string;
  buyer_username: string;
  platform: string;
  status: string;
  total_amount: number;
  currency: string;
  payment_method?: string;
  shipping_carrier?: string;
  tracking_number?: string;
  ship_by_date?: number;
  countdown?: string;
  buyer_message?: string;
  items: OrderItem[];
}

export interface OrderTableProps {
  orders: Order[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number, pageSize: number) => void;
  };
  selectedRowKeys: React.Key[];
  onSelectionChange: (selectedRowKeys: React.Key[]) => void;
  onShip: (order: GroupedOrder) => void;
  onPrint: (order: GroupedOrder) => void;
  onCancel?: (order: GroupedOrder) => void;
  onViewDetail?: (order: GroupedOrder) => void;
}

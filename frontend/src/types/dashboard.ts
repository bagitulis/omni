export interface DashboardMetrics {
  orders_pending: number;
  stock_warning: number;
  ready_to_ship: number;
  platform_status: number; // Percentage 0-100 or status code
}

export interface RecentOrder {
  order_sn: string;
  status: string;
  platform: "shopee" | "lazada" | "tiktok" | "manual";
  amount: number;
  buyer_username: string;
  created_at: string;
}

export interface DashboardData {
  metrics: DashboardMetrics;
  recent_orders: RecentOrder[];
}

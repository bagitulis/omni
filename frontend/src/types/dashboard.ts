/**
 * Dashboard Types - matches backend SalesAnalytics response
 */

/**
 * Backend SalesAnalytics response from /api/analytics/dashboard
 */
export interface SalesAnalytics {
  total_sales: number;
  total_orders: number;
  average_order: number;
  by_platform: Record<string, number>;
  period_start: string;
  period_end: string;
}

/**
 * Recent order for dashboard display
 * Mapped from Order type
 */
export interface RecentOrder {
  order_sn: string;
  order_no: string;
  status: string;
  platform: string;
  total_amount: number;
  buyer_username: string;
  created_at: string;
}

/**
 * Combined dashboard data with analytics + recent orders
 */
export interface DashboardData {
  analytics: SalesAnalytics;
  recent_orders: RecentOrder[];
  // Computed metrics for dashboard cards
  orders_pending: number;
  ready_to_ship: number;
}

export interface WalletData {
  total_balance: number;
  pending_balance: number;
  available_balance: number;
  currency: string;
  updated_at: string;
}

export interface ShippingFeeData {
  total_orders: number;
  orders_with_difference: number;
  total_profit: number;
  total_loss: number;
  net_impact: number;
}

export interface SyncStatusData {
  platform: string;
  status: "connected" | "disconnected" | "syncing" | "error";
  last_sync: string;
  details?: string;
}

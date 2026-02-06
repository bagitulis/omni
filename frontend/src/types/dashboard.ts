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

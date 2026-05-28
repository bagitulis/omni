import apiClient from "./client";
import {
  SalesAnalytics,
  DashboardData,
  RecentOrder,
  WalletData,
  ShippingFeeData,
} from "../types/dashboard";
import { Order } from "../types/order";
import { logger } from "@/lib/logger";

/**
 * Fetch dashboard summary data from analytics endpoint

 * Backend route: GET /api/analytics/dashboard
 */
async function getAnalytics(): Promise<SalesAnalytics> {
  // Dashboard shows last 30 days data
  const endDate = new Date().toISOString().split("T")[0];
  const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
    .toISOString()
    .split("T")[0];
  const response = await apiClient.get<SalesAnalytics>("/analytics/dashboard", {
    params: { startDate, endDate },
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch analytics");
  }
  return response.data!;
}

/**
 * Fetch recent orders for dashboard
 * Backend route: GET /api/orders/unprocess (ready to ship)
 */
async function getRecentOrders(): Promise<RecentOrder[]> {
  const response = await apiClient.client.get("/orders/unprocess", {
    params: { page: 1, pageSize: 5 },
  });

  const data = response.data;
  if (!data.success) {
    return [];
  }

  // Map backend order format to RecentOrder
  const orders = (data.data || data.items || []) as Order[];
  return orders.slice(0, 5).map((order) => ({
    order_sn: order.order_no || order.order_sn,
    order_no: order.order_no,
    status: order.status || order.order_status,
    platform: order.platform.toLowerCase(),
    total_amount: order.total_amount,
    buyer_username: order.buyer_username,
    created_at: order.created_at,
  }));
}

/**
 * Fetch pending orders count
 * Backend route: GET /api/orders/unpaid
 */
async function getPendingCount(): Promise<number> {
  try {
    const response = await apiClient.client.get("/orders/unpaid", {
      params: { page: 1, pageSize: 1 },
    });
    return response.data?.count || 0;
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return 0;
  }
}

/**
 * Fetch ready to ship count
 * Backend route: GET /api/orders/unprocess
 */
async function getReadyToShipCount(): Promise<number> {
  try {
    const response = await apiClient.client.get("/orders/unprocess", {
      params: { page: 1, pageSize: 1 },
    });
    return response.data?.count || 0;
  } catch (err) { logger.warn("Operation failed:", { err: err });
    return 0;
  }
}

/**
 * Fetch combined dashboard data
 * Combines analytics + recent orders + counts
 */
export async function getDashboardData(): Promise<DashboardData> {
  const [analytics, recentOrders, ordersPending, readyToShip] =
    await Promise.all([
      getAnalytics(),
      getRecentOrders(),
      getPendingCount(),
      getReadyToShipCount(),
    ]);

  return {
    analytics,
    recent_orders: recentOrders,
    orders_pending: ordersPending,
    ready_to_ship: readyToShip,
  };
}

/**
 * Fetch order analytics
 * Backend route: GET /api/analytics/orders
 */
export async function getOrderAnalytics(params?: {
  start_date?: string;
  end_date?: string;
}): Promise<unknown> {
  const response = await apiClient.get("/analytics/orders", { params });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch order analytics");
  }
  return response.data;
}

/**
 * Fetch revenue analytics
 * Backend route: GET /api/analytics/revenue
 */
export async function getRevenueAnalytics(params?: {
  start_date?: string;
  end_date?: string;
}): Promise<unknown> {
  const response = await apiClient.get("/analytics/revenue", { params });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch revenue analytics");
  }
  return response.data;
}

/**
 * Fetch wallet data for a platform
 * Backend route: GET /api/{platform}/wallet/balance
 */
export async function getWalletData(platform: string): Promise<WalletData> {
  if (platform === "shopee") {
    const response = await apiClient.get<WalletData>(`/shopee/wallet/balance`);
    if (!response.success) {
      throw new Error(response.error || "Failed to fetch wallet data");
    }
    return response.data!;
  }

  // Return default for other platforms until implemented
  return {
    total_balance: 0,
    pending_balance: 0,
    available_balance: 0,
    currency: "IDR",
    updated_at: new Date().toISOString(),
  };
}

/**
 * Fetch shipping fee analysis
 * Backend route: GET /api/analytics/{platform}/shipping-fee
 */
export async function getShippingFeeData(
  platform: string,
): Promise<ShippingFeeData> {
  const response = await apiClient.get(`/analytics/${platform}/shipping-fee`);
  if (!response.success) {
    // Return empty/safe default on error
    return {
      total_orders: 0,
      orders_with_difference: 0,
      total_profit: 0,
      total_loss: 0,
      net_impact: 0,
    };
  }
  return response.data as ShippingFeeData;
}


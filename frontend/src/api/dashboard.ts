import apiClient from "./client";
import { DashboardData } from "../types/dashboard";

/**
 * Fetch dashboard summary data from analytics endpoint
 * Backend route: GET /api/analytics/dashboard
 */
export async function getDashboardData(): Promise<DashboardData> {
  const response = await apiClient.get<DashboardData>("/analytics/dashboard");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch dashboard data");
  }
  return response.data!;
}

/**
 * Fetch order analytics
 * Backend route: GET /api/analytics/orders
 */
export async function getOrderAnalytics(params?: {
  start_date?: string;
  end_date?: string;
}): Promise<any> {
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
}): Promise<any> {
  const response = await apiClient.get("/analytics/revenue", { params });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch revenue analytics");
  }
  return response.data;
}

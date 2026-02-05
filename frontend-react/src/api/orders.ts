import apiClient from "./client";
import { OrderListResponse, Order } from "@/types/order";

/**
 * Order tab types matching Vue frontend
 */
export type OrderTab =
  | "unpaid"
  | "unprocess"
  | "processed"
  | "locked"
  | "today";

export interface GetOrdersParams {
  page?: number;
  pageSize?: number;
  status?: string;
  platform?: string;
  search?: string;
  startDate?: string;
  endDate?: string;
}

/**
 * Map status to backend endpoint
 * Backend routes:
 * - /orders/unpaid
 * - /orders/unprocess
 * - /orders/processed
 * - /orders/locked-today
 * - /orders/today
 */
function getOrderEndpoint(status: string): string {
  const endpointMap: Record<string, string> = {
    unpaid: "/orders/unpaid",
    unprocess: "/orders/unprocess",
    processed: "/orders/processed",
    locked: "/orders/locked-today",
    today: "/orders/today",
    ALL: "/orders/unpaid", // Default to unpaid
  };
  return endpointMap[status] || "/orders/unpaid";
}

/**
 * Fetch orders by tab/status
 */
export async function getOrders(
  params: GetOrdersParams = {},
): Promise<OrderListResponse> {
  const status = params.status || "unpaid";
  const endpoint = getOrderEndpoint(status);

  const response = await apiClient.get<OrderListResponse>(endpoint, {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      platform: params.platform,
      search: params.search,
      start_date: params.startDate,
      end_date: params.endDate,
    },
  });

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch orders");
  }

  return response.data!;
}

/**
 * Fetch orders by specific tab
 */
export async function getOrdersByTab(tab: OrderTab): Promise<Order[]> {
  const endpoint = getOrderEndpoint(tab);
  const response = await apiClient.get<Order[]>(endpoint);

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch orders");
  }

  return response.data || [];
}

/**
 * Sync orders for today (POST to /orders/today)
 */
export async function syncOrdersToday(): Promise<Order[]> {
  const response = await apiClient.post<Order[]>("/orders/today");
  if (!response.success) {
    throw new Error(response.error || "Failed to sync today's orders");
  }
  return response.data || [];
}

/**
 * Lock orders for today (POST to /orders/locked-today)
 */
export async function lockOrdersToday(): Promise<Order[]> {
  const response = await apiClient.post<Order[]>("/orders/locked-today");
  if (!response.success) {
    throw new Error(response.error || "Failed to lock today's orders");
  }
  return response.data || [];
}

/**
 * Sync all orders from all platforms
 */
export async function syncAllOrders(): Promise<void> {
  const response = await apiClient.post("/orders/sync-all");
  if (!response.success) {
    throw new Error(response.error || "Failed to sync orders");
  }
}

/**
 * Bulk ship orders
 */
export async function bulkShipOrders(orderSns: string[]): Promise<void> {
  const response = await apiClient.post("/orders/bulk-ship", {
    order_sns: orderSns,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to ship orders");
  }
}

/**
 * Bulk print labels
 */
export async function bulkPrintLabels(orderSns: string[]): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  console.log("Printing labels for:", orderSns);
}

import apiClient from "./client";
import { OrderListResponse, Order, BackendOrderResponse } from "@/types/order";

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
 * Transform backend order to frontend Order type
 * Backend uses: order_no, status
 * Frontend expects: order_sn, order_status (plus original fields)
 */
function transformOrder(backendOrder: Order): Order {
  return {
    ...backendOrder,
    order_sn: backendOrder.order_no || backendOrder.order_sn,
    order_status: backendOrder.status || backendOrder.order_status,
  };
}

/**
 * Fetch orders by tab/status
 */
export async function getOrders(
  params: GetOrdersParams = {},
): Promise<OrderListResponse> {
  const status = params.status || "unpaid";
  const endpoint = getOrderEndpoint(status);

  // Backend returns flat response: { success, count, data: [...], items: [...] }
  // NOT nested: { success, data: { count, data, items } }
  const response = await apiClient.client.get(endpoint, {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      platform: params.platform,
      search: params.search,
      start_date: params.startDate,
      end_date: params.endDate,
    },
  });

  const backendData = response.data as BackendOrderResponse & {
    success: boolean;
  };

  if (!backendData.success) {
    throw new Error("Failed to fetch orders");
  }

  // Transform backend response to frontend format
  const orders = (backendData.data || backendData.items || []).map(
    transformOrder,
  );

  return {
    orders,
    total: backendData.count || orders.length,
    page: params.page || 1,
    page_size: params.pageSize || 10,
  };
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
 * Bulk print labels response
 */
export interface BulkPrintLabelsResponse {
  labels: Array<{
    order_sn: string;
    file_data: string; // Base64 encoded PDF
    status: string;
  }>;
  failed: Array<{
    order_sn: string;
    error: string;
  }>;
  count: number;
}

/**
 * Bulk print labels - calls real backend API
 */
export async function bulkPrintLabels(
  orderSns: string[],
): Promise<BulkPrintLabelsResponse> {
  const response = await apiClient.post<BulkPrintLabelsResponse>(
    "/orders/bulk-print-labels",
    {
      order_sns: orderSns,
    },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to print labels");
  }
  return response.data!;
}

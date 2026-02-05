import apiClient from "./client";
import { OrderListResponse } from "@/types/order";

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
 * Fetch orders with filters
 */
export async function getOrders(
  params: GetOrdersParams = {},
): Promise<OrderListResponse> {
  const response = await apiClient.get<OrderListResponse>("/orders", {
    params,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch orders");
  }
  return response.data!;
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

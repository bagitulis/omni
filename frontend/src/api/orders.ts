import apiClient from "./client";
import {
  OrderListResponse,
  Order,
  BackendOrderResponse,
  OrderDetail,
} from "@/types/order";

function computeUniquePlatformCounts(orders: Order[]): Record<string, number> {
  const orderIdsByPlatform = new Map<string, Set<string>>();

  for (const order of orders) {
    const platform = order.platform?.toLowerCase();
    const orderId = order.order_no || order.order_sn;
    if (!platform || !orderId) continue;

    const existing = orderIdsByPlatform.get(platform);
    if (existing) {
      existing.add(orderId);
      continue;
    }
    orderIdsByPlatform.set(platform, new Set([orderId]));
  }

  const result: Record<string, number> = {};
  for (const [platform, ids] of orderIdsByPlatform.entries()) {
    result[platform] = ids.size;
  }
  return result;
}

/**
 * Order tab types matching Vue frontend
 */
export type OrderTab =
  | "unpaid"
  | "unprocess"
  | "processed"
  | "locked"
  | "today";

const SYNCABLE_TABS = ["unpaid", "unprocess", "processed"] as const;
export type SyncableOrderTab = (typeof SYNCABLE_TABS)[number];

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

  // Use axios AxiosResponse type for direct client calls
  interface AxiosResponse<T> {
    data: T;
  }

  let axiosResponse: AxiosResponse<BackendOrderResponse & { success: boolean }>;

  // Special handling for locked and today tabs which require POST
  if (status === "locked" || status === "today") {
    axiosResponse = await apiClient.client.post<
      BackendOrderResponse & { success: boolean }
    >(endpoint, {
      days: 7, // Default to 7 days like Vue
      page: params.page,
      pageSize: params.pageSize,
      platform: params.platform,
      search: params.search,
    });
  } else {
    // Standard GET for other tabs
    axiosResponse = await apiClient.client.get<
      BackendOrderResponse & { success: boolean }
    >(endpoint, {
      params: {
        page: params.page,
        pageSize: params.pageSize,
        platform: params.platform,
        search: params.search,
        start_date: params.startDate,
        end_date: params.endDate,
      },
    });
  }

  const backendData = axiosResponse.data;

  if (!backendData.success) {
    throw new Error("Failed to fetch orders");
  }

  // Transform backend response to frontend format
  const orders = (backendData.data || backendData.items || []).map(
    transformOrder,
  );

  // Calculate platform counts from orders if not provided by backend
  let platformCounts = backendData.platform_counts;
  if (!platformCounts) {
    platformCounts = computeUniquePlatformCounts(orders);
  }

  return {
    orders,
    total: backendData.count || orders.length,
    page: params.page || 1,
    page_size: params.pageSize || 10,
    platform_counts: platformCounts,
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

export function isSyncableOrderTab(status: string): status is SyncableOrderTab {
  return (SYNCABLE_TABS as readonly string[]).includes(status);
}

/**
 * Sync orders by category (unpaid/unprocess/processed)
 */
export async function syncOrdersByCategory(
  category: SyncableOrderTab,
): Promise<void> {
  // Processed tab often needs a longer window to match marketplace reality.
  // Also, syncing ALL platforms can be slow; for processed we prioritize Shopee.
  if (category === "processed") {
    const response = await apiClient.client.post(
      "/orders/sync/processed?days=30&platforms=shopee",
      {},
      { timeout: 120_000 },
    );
    if (!response.data?.success) {
      throw new Error(response.data?.error || "Failed to sync orders");
    }
    return;
  }

  const response = await apiClient.post(`/orders/sync/${category}`, {
    days: 7,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to sync orders");
  }
}

/**
 * Bulk ship orders
 * @param orderSns - Array of order serial numbers
 * @param platform - Platform name: "shopee" | "tiktok" | "lazada" (defaults to "shopee" on backend)
 */
export async function bulkShipOrders(
  orderSns: string[],
  platform?: string,
): Promise<void> {
  const response = await apiClient.post("/orders/bulk-ship", {
    order_sns: orderSns,
    platform: platform || "",
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
    file_data: string; // Base64 encoded PDF or URL document
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

/**
 * Fetch detailed order information by order_sn
 */
export async function getOrderById(orderSn: string): Promise<OrderDetail> {
  const response = await apiClient.get<OrderDetail>(`/orders/${orderSn}`);

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch order details");
  }

  return response.data!;
}

/**
 * Cancel order parameters
 */
export interface CancelOrderParams {
  order_no: string;
  platform: string;
  cancel_reason: string;
  reason_detail?: string;
  order_item_id?: string; // For Lazada
}

/**
 * Cancel an order
 */
export async function cancelOrder(params: CancelOrderParams): Promise<void> {
  const response = await apiClient.post("/orders/cancel", params);
  if (!response.success) {
    throw new Error(response.error || "Failed to cancel order");
  }
}

/**
 * Ship order parameters
 */
export interface ShipOrderParams {
  order_no: string;
  platform: string;
  shipping_provider: string;
  tracking_number?: string;
  address_id?: number;
  pickup_time_id?: string;
  branch_id?: number;
  package_id?: string;
  order_item_ids?: string[]; // For Lazada
}

/**
 * Ship a single order
 */
export async function shipOrder(params: ShipOrderParams): Promise<void> {
  const response = await apiClient.post("/orders/ship", params);
  if (!response.success) {
    throw new Error(response.error || "Failed to ship order");
  }
}

export interface LazadaDocumentResponse {
  document?: {
    file?: string;
    url?: string;
    mime_type?: string;
  };
}

/**
 * Get Lazada document (shipping label or invoice)
 * @param orderItemIds - Array of Lazada order item IDs
 * @param docType - "shippingLabel" | "invoice"
 */
export async function getLazadaDocument(
  orderItemIds: string[],
  docType: string,
): Promise<LazadaDocumentResponse> {
  const response = await apiClient.post<LazadaDocumentResponse>(
    "/lazada/orders/document",
    {
      order_item_ids: orderItemIds,
      doc_type: docType,
    },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to get Lazada document");
  }
  return response.data!;
}

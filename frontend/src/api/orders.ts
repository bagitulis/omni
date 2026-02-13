import apiClient from "./client";
import type {
  OrderListResponse,
  Order,
  BackendOrderResponse,
  OrderDetail,
} from "@/types/order";
import {
  getOrderEndpointFromTab,
  getSyncCategoryFromTab,
} from "./orderTabMapping";

type RawOrder = Partial<Order> & {
  tracking_no?: string;
  courier?: string;
  seller_sku?: string;
  quantity?: number;
  synced_at?: string;
};

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
  | "shipped"
  | "completed"
  | "cancelled"
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
 * Transform backend order to frontend Order type
 * Backend uses: order_no, status
 * Frontend expects: order_sn, order_status (plus original fields)
 */
function transformOrder(backendOrder: RawOrder): Order {
  const orderNo = backendOrder.order_no || backendOrder.order_sn || "";
  const status = backendOrder.status || backendOrder.order_status || "";
  const platform = (backendOrder.platform || "").toLowerCase();

  return {
    ...backendOrder,
    id: backendOrder.id || orderNo,
    order_no: orderNo,
    order_sn: orderNo,
    order_status: status,
    status,
    platform,
    category: backendOrder.category || "",
    buyer_username: backendOrder.buyer_username || "",
    total_amount: backendOrder.total_amount || 0,
    currency: backendOrder.currency || "IDR",
    payment_method: backendOrder.payment_method || "",
    shipping_carrier:
      backendOrder.shipping_carrier || backendOrder.courier || "",
    tracking_number:
      backendOrder.tracking_number || backendOrder.tracking_no || "",
    ship_by_date: backendOrder.ship_by_date || 0,
    buyer_message: backendOrder.buyer_message || "",
    sku: backendOrder.sku || backendOrder.seller_sku || "",
    product_name: backendOrder.product_name || "",
    variation_name: backendOrder.variation_name || "",
    qty: backendOrder.qty || backendOrder.quantity || 1,
    price: backendOrder.price || 0,
    product_image: backendOrder.product_image || "",
    created_at: backendOrder.created_at || backendOrder.synced_at || "",
    updated_at:
      backendOrder.updated_at ||
      backendOrder.created_at ||
      backendOrder.synced_at ||
      "",
  };
}

/**
 * Fetch orders by tab/status
 */
export async function getOrders(
  params: GetOrdersParams = {},
): Promise<OrderListResponse> {
  const status = params.status || "unpaid";
  const endpoint = getOrderEndpointFromTab(status);

  // Use axios AxiosResponse type for direct client calls
  interface AxiosResponse<T> {
    data: T;
  }

  let axiosResponse: AxiosResponse<
    BackendOrderResponse & { success: boolean; error?: string }
  >;

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
    throw new Error(backendData.error || "Failed to fetch orders");
  }

  // Transform backend response to frontend format
  const orders = (backendData.data || backendData.items || []).map((order) =>
    transformOrder(order as RawOrder),
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
  const endpoint = getOrderEndpointFromTab(tab);
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

export { isSyncableOrderTab } from "./orderTabMapping";

/**
 * Sync orders by category (unpaid/unprocess/processed)
 */
export async function syncOrdersByCategory(
  tabKey: string,
  platform?: string,
): Promise<void> {
  const category = getSyncCategoryFromTab(tabKey);
  if (!category) {
    return;
  }

  const normalizedPlatform = (platform || "").toLowerCase();
  const isSpecificPlatform =
    normalizedPlatform === "shopee" ||
    normalizedPlatform === "lazada" ||
    normalizedPlatform === "tiktok";
  const platformQuery = isSpecificPlatform
    ? `?platforms=${encodeURIComponent(normalizedPlatform)}`
    : "";

  const endpoint = `/orders/sync/${category}${platformQuery}`;
  const days = category === "processed" ? 30 : 7;

  const response = await apiClient.client.post(
    endpoint,
    { days },
    { timeout: 120_000 },
  );
  if (!response.data?.success) {
    throw new Error(response.data?.error || "Failed to sync orders");
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

export interface BulkPrintLabelsOptions {
  platform?: string;
  include_products?: boolean;
  tiktok_document_type?:
    | "SHIPPING_LABEL"
    | "PACKING_SLIP"
    | "SHIPPING_LABEL_AND_PACKING_SLIP";
}

/**
 * Bulk print labels - calls real backend API
 */
export async function bulkPrintLabels(
  orderSns: string[],
  options?: BulkPrintLabelsOptions,
): Promise<BulkPrintLabelsResponse> {
  const response = await apiClient.post<BulkPrintLabelsResponse>(
    "/orders/bulk-print-labels",
    {
      order_sns: orderSns,
      platform: options?.platform,
      include_products: options?.include_products,
      tiktok_document_type: options?.tiktok_document_type,
    },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to print labels");
  }
  if (!response.data) {
    throw new Error("Failed to print labels");
  }
  return response.data;
}

/**
 * Fetch detailed order information by order_sn
 */
export async function getOrderById(orderSn: string): Promise<OrderDetail> {
  const response = await apiClient.get<OrderDetail>(`/orders/${orderSn}`);

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch order details");
  }
  if (!response.data) {
    throw new Error("Failed to fetch order details");
  }
  return response.data;
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
  if (!response.data) {
    throw new Error("Failed to get Lazada document");
  }
  return response.data;
}

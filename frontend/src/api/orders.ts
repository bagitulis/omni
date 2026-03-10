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
  normalizeOrderTabKey,
} from "./orderTabMapping";
import {
  transformOrder,
  computeUniquePlatformCounts,
  type RawOrder,
} from "./orderTransforms";

// Re-export types so existing consumers don't break
export type {
  OrderTab,
  GetOrdersParams,
  BulkPrintLabelsResponse,
  BulkPrintLabelsOptions,
  CancelOrderParams,
  ShipOrderParams,
  LazadaDocumentResponse,
} from "./orderTypes";

import type {
  GetOrdersParams,
  BulkPrintLabelsOptions,
  BulkPrintLabelsResponse,
  CancelOrderParams,
  ShipOrderParams,
  LazadaDocumentResponse,
  OrderTab,
} from "./orderTypes";

// --- Fetch Operations ---

export async function getOrders(
  params: GetOrdersParams = {},
): Promise<OrderListResponse> {
  const status = params.status || "unprocess";
  const normalizedStatus = normalizeOrderTabKey(status);
  const endpoint = getOrderEndpointFromTab(normalizedStatus);

  interface AxiosResponse<T> {
    data: T;
  }

  let axiosResponse: AxiosResponse<
    BackendOrderResponse & { success: boolean; error?: string }
  >;

  if (normalizedStatus === "locked" || normalizedStatus === "today") {
    axiosResponse = await apiClient.client.post<
      BackendOrderResponse & { success: boolean }
    >(endpoint, {
      days: 7,
      page: params.page,
      pageSize: params.pageSize,
      platform: params.platform,
      search: params.search,
    });
  } else {
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

  const orders = (backendData.data || backendData.items || []).map((order) =>
    transformOrder(order as RawOrder),
  );

  const platformCounts =
    backendData.platform_counts ?? computeUniquePlatformCounts(orders);

  return {
    orders,
    total: backendData.count || orders.length,
    page: params.page || 1,
    page_size: params.pageSize || 10,
    platform_counts: platformCounts,
  };
}

export async function getOrdersByTab(tab: OrderTab): Promise<Order[]> {
  const endpoint = getOrderEndpointFromTab(tab);
  const response = await apiClient.get<Order[]>(endpoint);
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch orders");
  }
  return response.data || [];
}

export async function getOrderById(orderSn: string): Promise<OrderDetail> {
  const response = await apiClient.get<OrderDetail>(`/orders/${orderSn}`);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch order details");
  }
  return response.data;
}

// --- Sync Operations ---

export async function syncOrdersToday(): Promise<Order[]> {
  const response = await apiClient.post<Order[]>("/orders/today");
  if (!response.success) {
    throw new Error(response.error || "Failed to sync today's orders");
  }
  return response.data || [];
}

export async function lockOrdersToday(): Promise<Order[]> {
  const response = await apiClient.post<Order[]>("/orders/locked-today");
  if (!response.success) {
    throw new Error(response.error || "Failed to lock today's orders");
  }
  return response.data || [];
}

export async function syncAllOrders(): Promise<void> {
  const response = await apiClient.post("/orders/sync-all");
  if (!response.success) {
    throw new Error(response.error || "Failed to sync orders");
  }
}

export { isSyncableOrderTab } from "./orderTabMapping";

export async function syncOrdersByCategory(
  tabKey: string,
  platform?: string,
): Promise<void> {
  const category = getSyncCategoryFromTab(tabKey);
  if (!category) return;

  const normalizedPlatform = (platform || "").toLowerCase();
  const isSpecificPlatform = ["shopee", "lazada", "tiktok"].includes(
    normalizedPlatform,
  );
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

export interface BulkShipItemResult {
  order_sn: string;
  status: "shipped" | "failed" | "shipped_but_local_failed";
  error?: string;
  message?: string;
  marketplace_ok?: boolean;
}

export interface BulkShipResult {
  partial: boolean;
  summary: { total: number; shipped: number; failed: number };
  results: BulkShipItemResult[];
}

export async function bulkShipOrders(
  orderSns: string[],
  platform?: string,
): Promise<BulkShipResult> {
  const response = await apiClient.post<{
    summary: { total: number; shipped: number; failed: number };
    results: BulkShipItemResult[];
  }>("/orders/bulk-ship", {
    order_sns: orderSns,
    platform: platform || "",
  });

  if (!response.success) {
    throw new Error(response.error || "Failed to ship orders");
  }

  const data = response.data;
  const summary = data?.summary ?? {
    total: orderSns.length,
    shipped: 0,
    failed: orderSns.length,
  };

  return {
    partial: response.message === "partial_success",
    summary,
    results: data?.results ?? [],
  };
}

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
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to print labels");
  }
  return response.data;
}

// --- Single Order Actions ---

export async function cancelOrder(params: CancelOrderParams): Promise<void> {
  const response = await apiClient.post("/orders/cancel", params);
  if (!response.success) {
    throw new Error(response.error || "Failed to cancel order");
  }
}

export async function shipOrder(params: ShipOrderParams): Promise<void> {
  const response = await apiClient.post("/orders/ship", params);
  if (!response.success) {
    throw new Error(response.error || "Failed to ship order");
  }
}

export async function getLazadaDocument(
  orderItemIds: string[],
  docType: string,
): Promise<LazadaDocumentResponse> {
  const response = await apiClient.post<LazadaDocumentResponse>(
    "/lazada/orders/document",
    { order_item_ids: orderItemIds, doc_type: docType },
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to get Lazada document");
  }
  return response.data;
}

import { Order, OrderListResponse } from "@/types/order";

// Mock data for development
const MOCK_ORDERS: Order[] = [
  {
    order_sn: "SHP-20231025-0001",
    order_status: "PENDING",
    platform: "shopee",
    buyer_username: "john_doe",
    total_amount: 150000,
    created_at: "2023-10-25T10:00:00Z",
    updated_at: "2023-10-25T10:00:00Z",
  },
  {
    order_sn: "TT-20231025-0002",
    order_status: "READY_TO_SHIP",
    platform: "tiktok",
    buyer_username: "jane_smith",
    total_amount: 250000,
    created_at: "2023-10-25T11:30:00Z",
    updated_at: "2023-10-25T11:30:00Z",
  },
  {
    order_sn: "LZD-20231024-0003",
    order_status: "SHIPPED",
    platform: "lazada",
    buyer_username: "bob_wilson",
    total_amount: 75000,
    created_at: "2023-10-24T09:15:00Z",
    updated_at: "2023-10-25T14:20:00Z",
  },
  {
    order_sn: "SHP-20231024-0004",
    order_status: "COMPLETED",
    platform: "shopee",
    buyer_username: "alice_wong",
    total_amount: 320000,
    created_at: "2023-10-24T08:00:00Z",
    updated_at: "2023-10-26T10:00:00Z",
  },
  {
    order_sn: "TKP-20231023-0005",
    order_status: "CANCELLED",
    platform: "tokopedia",
    buyer_username: "charlie_brown",
    total_amount: 120000,
    created_at: "2023-10-23T15:45:00Z",
    updated_at: "2023-10-23T16:00:00Z",
  },
];

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
  // Simulate API call with delay
  await new Promise((resolve) => setTimeout(resolve, 800));

  // Filter mock data
  let filteredOrders = [...MOCK_ORDERS];

  if (params.status && params.status !== "ALL") {
    filteredOrders = filteredOrders.filter(
      (o) => o.order_status === params.status,
    );
  }

  if (params.platform && params.platform !== "all") {
    filteredOrders = filteredOrders.filter(
      (o) => o.platform === params.platform,
    );
  }

  if (params.search) {
    const searchLower = params.search.toLowerCase();
    filteredOrders = filteredOrders.filter(
      (o) =>
        o.order_sn.toLowerCase().includes(searchLower) ||
        o.buyer_username.toLowerCase().includes(searchLower),
    );
  }

  // Pagination logic
  const page = params.page || 1;
  const pageSize = params.pageSize || 10;
  const total = filteredOrders.length;
  const start = (page - 1) * pageSize;
  const end = start + pageSize;

  return {
    orders: filteredOrders.slice(start, end),
    total,
    page,
    page_size: pageSize,
  };

  /* 
  // Real implementation
  const response = await apiClient.get<OrderListResponse>("/orders", { params });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch orders");
  }
  return response.data!;
  */
}

/**
 * Bulk ship orders
 */
export async function bulkShipOrders(orderSns: string[]): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  console.log("Bulk shipping orders:", orderSns);

  /*
  const response = await apiClient.post("/orders/bulk-ship", { order_sns: orderSns });
  if (!response.success) {
    throw new Error(response.error || "Failed to ship orders");
  }
  */
}

/**
 * Bulk print labels
 */
export async function bulkPrintLabels(orderSns: string[]): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  console.log("Printing labels for:", orderSns);
}

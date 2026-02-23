import apiClient from "./client";

// --- Types (snake_case to match backend) ---

export interface LockedOrderItem {
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
}

interface LockedOrdersData {
  items?: LockedOrderItem[];
  count?: number;
}

// --- API Functions ---

/**
 * Trigger sync + fetch locked orders for today.
 * POST /api/orders/locked-today
 */
export async function syncLockedToday(
  days: number = 7,
): Promise<LockedOrderItem[]> {
  const response = await apiClient.post<LockedOrdersData>(
    "/orders/locked-today",
    { days },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to sync locked orders");
  }

  // Backend returns items at top level (not inside data wrapper)
  const raw = response as unknown as Record<string, unknown>;
  const items = (response.data?.items ?? raw.items ?? []) as LockedOrderItem[];
  return items;
}

/**
 * Get cached locked orders.
 * GET /api/orders/locked-today
 */
export async function getLockedOrders(): Promise<LockedOrderItem[]> {
  const response =
    await apiClient.get<LockedOrdersData>("/orders/locked-today");

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch locked orders");
  }

  // Backend returns items at top level (not inside data wrapper)
  const raw = response as unknown as Record<string, unknown>;
  const items = (response.data?.items ?? raw.items ?? []) as LockedOrderItem[];
  return items;
}

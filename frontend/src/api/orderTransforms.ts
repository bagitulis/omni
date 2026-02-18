import type { Order } from "@/types/order";

export type RawOrder = Partial<Order> & {
  tracking_no?: string;
  courier?: string;
  seller_sku?: string;
  quantity?: number;
  synced_at?: string;
};

export function computeUniquePlatformCounts(
  orders: Order[],
): Record<string, number> {
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
 * Transform backend order to frontend Order type
 * Backend uses: order_no, status
 * Frontend expects: order_sn, order_status (plus original fields)
 */
export function transformOrder(backendOrder: RawOrder): Order {
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

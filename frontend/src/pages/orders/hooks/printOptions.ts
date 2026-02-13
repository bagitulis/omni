import type { BulkPrintLabelsOptions } from "@/api/orders";
import type { OrderListResponse } from "@/types/order";

function normalizePlatform(platform: string): string {
  return platform.trim().toLowerCase();
}

function selectedOrderSet(selectedOrderSns: string[]): Set<string> {
  return new Set(selectedOrderSns.map((orderSn) => orderSn.trim()));
}

function resolveOrderSN(order: {
  order_sn?: string;
  order_no?: string;
}): string {
  return (order.order_sn || order.order_no || "").trim();
}

export function shouldPromptTikTokPackingSlip(
  platform: string,
  data: OrderListResponse | undefined,
  selectedOrderSns: string[],
): boolean {
  const normalizedPlatform = normalizePlatform(platform);
  if (normalizedPlatform === "tiktok") {
    return true;
  }

  if (normalizedPlatform !== "all" || !data?.orders?.length) {
    return false;
  }

  const selectedOrders = selectedOrderSet(selectedOrderSns);

  return data.orders.some((order) => {
    const orderSN = resolveOrderSN(order);
    return (
      selectedOrders.has(orderSN) &&
      normalizePlatform(order.platform || "") === "tiktok"
    );
  });
}

export function buildBulkPrintOptions(
  platform: string,
  includeProducts?: boolean,
): BulkPrintLabelsOptions {
  const options: BulkPrintLabelsOptions = {};

  const normalizedPlatform = normalizePlatform(platform);
  if (normalizedPlatform !== "all" && normalizedPlatform !== "") {
    options.platform = normalizedPlatform;
  }

  if (typeof includeProducts === "boolean") {
    options.include_products = includeProducts;
  }

  return options;
}

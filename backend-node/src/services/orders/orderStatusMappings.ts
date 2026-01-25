/**
 * Order Status Mappings
 * SRP: Platform-specific order status configurations
 *
 * Maps internal categories to platform-specific status values
 * MUST match Python backend exactly!
 *
 * From backend/routes/order_routes.py:
 * - Shopee: UNPAID, READY_TO_SHIP, PROCESSED
 * - Lazada: unpaid, topack, toship
 * - TikTok: UNPAID, AWAITING_SHIPMENT, AWAITING_COLLECTION
 */

export type PlatformType = "shopee" | "lazada" | "tiktok";
export type OrderStatusCategory = "unpaid" | "unprocess" | "processed";

export const ORDER_STATUS_MAPPINGS: Record<
  PlatformType,
  Record<OrderStatusCategory, string>
> = {
  shopee: {
    unpaid: "UNPAID",
    unprocess: "READY_TO_SHIP",
    processed: "PROCESSED",
  },
  lazada: {
    unpaid: "unpaid",
    unprocess: "topack",
    processed: "toship",
  },
  tiktok: {
    unpaid: "UNPAID",
    unprocess: "AWAITING_SHIPMENT",
    processed: "AWAITING_COLLECTION",
  },
};

/**
 * Get platform-specific status for a category
 */
export function getPlatformStatus(
  platform: PlatformType,
  category: OrderStatusCategory
): string {
  return ORDER_STATUS_MAPPINGS[platform][category];
}

/**
 * Get all statuses for a platform
 */
export function getAllPlatformStatuses(platform: PlatformType): string[] {
  return Object.values(ORDER_STATUS_MAPPINGS[platform]);
}

export type OrderCategory =
  | "unpaid"
  | "unprocess"
  | "processed"
  | "shipped"
  | "completed"
  | "cancelled";

export type OrderTabKey = OrderCategory | "locked" | "today";

const TAB_TO_CATEGORY: Record<string, OrderCategory> = {
  UNPAID: "unpaid",
  UNPROCESS: "unprocess",
  READY_TO_SHIP: "unprocess",
  TOPACK: "unprocess",
  AWAITING_SHIPMENT: "unprocess",
  PROCESSED: "processed",
  TOSHIP: "processed",
  AWAITING_COLLECTION: "processed",
  SHIPPED: "shipped",
  IN_TRANSIT: "shipped",
  DELIVERED: "completed",
  COMPLETED: "completed",
  IN_CANCEL: "cancelled",
  CANCELLED: "cancelled",
  CANCELED: "cancelled",
  FAILED: "cancelled",
  RETURNED: "cancelled",
};

const SYNCABLE_CATEGORIES: readonly OrderCategory[] = [
  "unpaid",
  "unprocess",
  "processed",
  "shipped",
  "completed",
  "cancelled",
];

export function normalizeOrderTabKey(tabKey: string): OrderTabKey {
  const normalized = tabKey.trim().toUpperCase();
  if (normalized === "LOCKED") {
    return "locked";
  }
  if (normalized === "TODAY") {
    return "today";
  }
  return TAB_TO_CATEGORY[normalized] ?? "unpaid";
}

export function getOrderEndpointFromTab(tabKey: string): string {
  const normalized = normalizeOrderTabKey(tabKey);

  switch (normalized) {
    case "unpaid":
      return "/orders/unpaid";
    case "unprocess":
      return "/orders/unprocess";
    case "processed":
      return "/orders/processed";
    case "locked":
      return "/orders/locked-today";
    case "today":
      return "/orders/today";
    case "shipped":
    case "completed":
    case "cancelled":
      return `/orders/category/${normalized}`;
    default:
      return "/orders/unpaid";
  }
}

export function getSyncCategoryFromTab(tabKey: string): OrderCategory | null {
  const normalized = normalizeOrderTabKey(tabKey);
  if (normalized === "locked" || normalized === "today") {
    return null;
  }
  return normalized;
}

export function isSyncableOrderTab(tabKey: string): boolean {
  const category = getSyncCategoryFromTab(tabKey);
  if (!category) {
    return false;
  }
  return SYNCABLE_CATEGORIES.includes(category);
}

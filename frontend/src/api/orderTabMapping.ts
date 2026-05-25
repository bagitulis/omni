export type OrderCategory =
  | "unpaid"
  | "unprocess"
  | "processed"
  | "shipped"
  | "completed"
  | "cancelled";

export type OrderTabKey = OrderCategory | "locked" | "today" | "booking";

const TAB_TO_CATEGORY: Record<string, OrderCategory> = {
  UNPAID: "unpaid",
  UNPROCESS: "unprocess",
  UNPROCESSED: "unprocess",
  READY_TO_SHIP: "unprocess",
  READY_TO_PACK: "unprocess",
  TO_PACK: "unprocess",
  TOPACK: "unprocess",
  AWAITING_SHIPMENT: "unprocess",
  PROCESSED: "processed",
  TO_SHIP: "processed",
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

const TAB_TO_SPECIAL: Record<
  string,
  Extract<OrderTabKey, "locked" | "today" | "booking">
> = {
  LOCKED: "locked",
  LOCKED_TODAY: "locked",
  LOCKEDTODAY: "locked",
  TODAY: "today",
  TODAYS_ORDERS: "today",
  TODAYSORDERS: "today",
  BOOKING: "booking",
};

const TAB_TO_CATEGORY_COMPACT: Record<string, OrderCategory> =
  Object.fromEntries(
    Object.entries(TAB_TO_CATEGORY).map(([key, value]) => [
      key.replace(/[^A-Z0-9]/g, ""),
      value,
    ]),
  );

const TAB_TO_SPECIAL_COMPACT: Record<
  string,
  Extract<OrderTabKey, "locked" | "today" | "booking">
> = Object.fromEntries(
  Object.entries(TAB_TO_SPECIAL).map(([key, value]) => [
    key.replace(/[^A-Z0-9]/g, ""),
    value,
  ]),
);

const SYNCABLE_CATEGORIES: readonly OrderCategory[] = [
  "unpaid",
  "unprocess",
  "processed",
  "shipped",
  "completed",
  "cancelled",
];

export function normalizeOrderTabKey(tabKey: string): OrderTabKey {
  const normalized = tabKey.trim().toUpperCase().replace(/\s+/g, "_");
  const compact = normalized.replace(/[^A-Z0-9]/g, "");

  const special = TAB_TO_SPECIAL[normalized] ?? TAB_TO_SPECIAL_COMPACT[compact];
  if (special) {
    return special;
  }

  return (
    TAB_TO_CATEGORY[normalized] ??
    TAB_TO_CATEGORY_COMPACT[compact] ??
    "unprocess"
  );
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
    case "booking":
      return "/orders/booking";
    case "today":
      return "/orders/today";
    case "shipped":
    case "completed":
    case "cancelled":
      return `/orders/category/${normalized}`;
    default:
      return "/orders/unprocess";
  }
}

export function getSyncCategoryFromTab(tabKey: string): OrderCategory | "booking" | null {
  const normalized = normalizeOrderTabKey(tabKey);
  if (normalized === "locked" || normalized === "today") {
    return null;
  }
  if (normalized === "booking") {
    return "booking";
  }
  return normalized;
}

export function isSyncableOrderTab(tabKey: string): boolean {
  const category = getSyncCategoryFromTab(tabKey);
  if (!category) {
    return false;
  }
  if (category === "booking") {
    return true;
  }
  return SYNCABLE_CATEGORIES.includes(category);
}

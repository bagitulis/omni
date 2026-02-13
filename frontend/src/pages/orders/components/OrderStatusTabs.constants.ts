/**
 * Order Status Tab Configurations
 * Separated from component for better Fast Refresh support
 */

export const ALL_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Shipped" },
  { key: "locked", label: "Locked Today" },
  { key: "today", label: "Today's Orders" },
];

export const SHOPEE_TABS = [
  { key: "UNPAID", label: "Unpaid" },
  { key: "READY_TO_SHIP", label: "To Ship" },
  { key: "PROCESSED", label: "Processed" },
  { key: "SHIPPED", label: "Shipped" },
  { key: "COMPLETED", label: "Completed" },
  { key: "IN_CANCEL", label: "In Cancel" },
  { key: "CANCELLED", label: "Cancelled" },
];

export const LAZADA_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "topack", label: "To Pack" },
  { key: "toship", label: "To Ship" },
  { key: "shipped", label: "Shipped" },
  { key: "delivered", label: "Delivered" },
  { key: "failed", label: "Failed" },
  { key: "returned", label: "Returned" },
];

export const TIKTOK_TABS = [
  { key: "AWAITING_SHIPMENT", label: "To Ship" },
  { key: "AWAITING_COLLECTION", label: "To Collect" },
  { key: "IN_TRANSIT", label: "In Transit" },
  { key: "DELIVERED", label: "Delivered" },
  { key: "COMPLETED", label: "Completed" },
  { key: "CANCELLED", label: "Cancelled" },
];

export const ORDER_TABS = ALL_TABS;

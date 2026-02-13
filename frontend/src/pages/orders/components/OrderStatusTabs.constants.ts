/**
 * Order Status Tab Configurations
 * Separated from component for better Fast Refresh support
 */

export const ALL_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Processed" },
  { key: "locked", label: "Locked Today" },
  { key: "today", label: "Today's Orders" },
];

export const SHOPEE_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Processed" },
  { key: "shipped", label: "Shipped" },
  { key: "completed", label: "Completed" },
  { key: "cancelled", label: "Cancelled" },
];

export const LAZADA_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Pack" },
  { key: "processed", label: "To Ship" },
  { key: "shipped", label: "Shipped" },
  { key: "completed", label: "Delivered" },
  { key: "cancelled", label: "Cancelled" },
];

export const TIKTOK_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "To Collect" },
  { key: "shipped", label: "In Transit" },
  { key: "completed", label: "Completed" },
  { key: "cancelled", label: "Cancelled" },
];

export const ORDER_TABS = ALL_TABS;

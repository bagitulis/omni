/**
 * Order Status Tab Configurations
 * Separated from component for better Fast Refresh support
 */

export const ALL_TABS = [
  { key: "booking", label: "Booking" },
  { key: "unprocess", label: "To Ship" },
  { key: "processed", label: "Processed" },
  { key: "shipped", label: "Shipped" },
  { key: "completed", label: "Completed" },
  { key: "cancelled", label: "Cancelled" },
  { key: "locked", label: "Locked Today" },
  { key: "today", label: "Today's Orders" },
];

export const ORDER_TABS = ALL_TABS;

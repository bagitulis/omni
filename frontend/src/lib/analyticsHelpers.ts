/**
 * Analytics Helper Functions
 * Shared utilities for analytics pages
 */

import type {
  ReconciliationResult,
  ShopeeShippingFeeResult,
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
} from "@/types/analytics";

// ============================================================================
// Constants
// ============================================================================

export const MONTHS = [
  { value: 0, label: "January" },
  { value: 1, label: "February" },
  { value: 2, label: "March" },
  { value: 3, label: "April" },
  { value: 4, label: "May" },
  { value: 5, label: "June" },
  { value: 6, label: "July" },
  { value: 7, label: "August" },
  { value: 8, label: "September" },
  { value: 9, label: "October" },
  { value: 10, label: "November" },
  { value: 11, label: "December" },
];

// ============================================================================
// Date & Year Functions
// ============================================================================

export function getAvailableYears(): number[] {
  const currentYear = new Date().getFullYear();
  return [currentYear, currentYear - 1, currentYear - 2];
}

export function formatAnalyticsDate(dateStr: string | null): string {
  if (!dateStr) return "—";
  try {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US");
  } catch (err) { console.warn("Operation failed:", err);
    return "—";
  }
}

// ============================================================================
// Currency & Formatting
// ============================================================================

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);
}

// ============================================================================
// CSV Export Helpers
// ============================================================================

function escapeCSVValue(value: string | number | null | undefined): string {
  if (value === null || value === undefined) return '""';
  const stringValue = String(value);
  if (
    stringValue.includes(",") ||
    stringValue.includes('"') ||
    stringValue.includes("\n")
  ) {
    return `"${stringValue.replace(/"/g, '""')}"`;
  }
  return stringValue;
}

export function downloadCSV(rows: string[][], filename: string): void {
  const csvContent = rows
    .map((row) => row.map(escapeCSVValue).join(","))
    .join("\n");
  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const link = document.createElement("a");
  const url = URL.createObjectURL(blob);
  link.setAttribute("href", url);
  link.setAttribute("download", filename);
  link.style.visibility = "hidden";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function getCurrentDateForFilename(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function exportShopeeReconciliationCSV(
  data: ReconciliationResult,
): void {
  const headers = [
    "Status",
    "SKU",
    "Model SKU",
    "Item Name",
    "Model Name",
    "Inventory Price",
    "Expected Income",
    "Transactions",
    "Unit Prices",
    "Actual Incomes",
  ];

  const rows = data.sku_groups.map((group) => [
    getStatusLabel(group.status),
    group.sku,
    group.model_sku,
    group.item_name,
    group.model_name,
    group.inventory_price !== null ? String(group.inventory_price) : "",
    group.expected_income !== null ? String(group.expected_income) : "",
    String(group.total_transactions),
    group.unique_unit_prices.join(";"),
    group.unique_actual_incomes.join(";"),
  ]);

  const csvRows = [headers, ...rows];
  const filename = `shopee-reconciliation-${getCurrentDateForFilename()}.csv`;
  downloadCSV(csvRows, filename);
}

export function exportShopeeShippingCSV(data: ShopeeShippingFeeResult): void {
  const headers = [
    "Order Date",
    "Order SN",
    "Buyer Paid",
    "Actual Fee",
    "Rebate",
    "Difference",
    "Buyer Name",
    "Payment Method",
  ];

  const rows = data.orders.map((order) => [
    order.order_date || "",
    order.order_sn,
    String(order.buyer_paid),
    String(order.actual_fee),
    String(order.shopee_rebate),
    String(order.difference),
    order.buyer_name || "",
    order.payment_method || "",
  ]);

  const csvRows = [headers, ...rows];
  const filename = `shopee-shipping-fee-${getCurrentDateForFilename()}.csv`;
  downloadCSV(csvRows, filename);
}

export function exportTiktokReconciliationCSV(
  data: TiktokReconciliationResult,
): void {
  const headers = [
    "Status",
    "SKU",
    "Model SKU",
    "Item Name",
    "Inventory Price",
    "Expected Income",
    "Transactions",
    "Unit Prices",
    "Actual Incomes",
  ];

  const rows = data.sku_groups.map((group) => [
    getStatusLabel(group.status),
    group.sku,
    group.model_sku,
    group.item_name,
    group.inventory_price !== null ? String(group.inventory_price) : "",
    group.expected_income !== null ? String(group.expected_income) : "",
    String(group.total_transactions),
    group.unique_unit_prices.join(";"),
    group.unique_actual_incomes.join(";"),
  ]);

  const csvRows = [headers, ...rows];
  const filename = `tiktok-reconciliation-${getCurrentDateForFilename()}.csv`;
  downloadCSV(csvRows, filename);
}

export function exportTiktokShippingCSV(data: TiktokShippingFeeResult): void {
  const headers = [
    "Order Date",
    "Order ID",
    "Buyer Paid",
    "Actual Fee",
    "Platform Discount",
    "Difference",
    "Order Status",
    "Currency",
  ];

  const rows = data.orders.map((order) => [
    order.order_date || "",
    order.order_id,
    String(order.buyer_paid),
    String(order.actual_fee),
    String(order.platform_discount),
    String(order.difference),
    order.order_status || "",
    order.currency,
  ]);

  const csvRows = [headers, ...rows];
  const filename = `tiktok-shipping-fee-${getCurrentDateForFilename()}.csv`;
  downloadCSV(csvRows, filename);
}

// ============================================================================
// Status Helpers
// ============================================================================

export function getStatusColor(status: string): string {
  switch (status) {
    case "OK":
      return "success";
    case "PRICE_DIFF":
      return "warning";
    case "NO_INVENTORY":
      return "error";
    default:
      return "default";
  }
}

export function getStatusLabel(status: string): string {
  switch (status) {
    case "OK":
      return "OK";
    case "PRICE_DIFF":
      return "Price Difference";
    case "NO_INVENTORY":
      return "No Inventory";
    default:
      return status;
  }
}

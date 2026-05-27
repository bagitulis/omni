/**
 * Analytics & Report Utility Helpers
 *
 * Formatting, status determination, and CSV export utilities
 * for the analytics report pages.
 */
import { formatCurrency } from "@/lib/helpers";
import type { ReconciliationSummary, ReportPlatform } from "@/types/analytics";

// Re-export for convenience
export { formatCurrency };

/**
 * Format month/year pair into a readable string.
 * Example: "January 2026"
 */
export function formatMonthYear(month: number, year: number): string {
  const date = new Date(year, month - 1, 1);
  return date.toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  });
}

/**
 * Format ISO date string to a short readable format.
 * Example: "Jan 15, 2026"
 * Returns "-" for null/undefined/falsy input.
 */
export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return "-";
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return "-";
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  } catch {
    return "-";
  }
}

/**
 * Determine status color based on price difference value.
 * - positive → "success" (profit)
 * - negative → "danger" (loss)
 * - zero     → "default"
 */
export function getPriceDiffStatus(
  diff: number
): "success" | "warning" | "danger" {
  if (diff > 0) return "success";
  if (diff < 0) return "danger";
  return "warning";
}

/**
 * Format price difference with sign and currency.
 * Example: "+Rp 5,000" or "-Rp 2,000"
 */
export function formatPriceDiff(diff: number): string {
  const abs = formatCurrency(Math.abs(diff));
  if (diff > 0) return `+${abs}`;
  if (diff < 0) return `-${abs}`;
  return formatCurrency(0);
}

/**
 * Determine overall reconciliation health.
 * - "good"    → all SKUs are OK
 * - "warning" → some SKUs have price differences
 * - "error"   → some SKUs have no inventory
 */
export function getReconciliationHealth(
  summary: ReconciliationSummary
): "good" | "warning" | "error" {
  if (summary.total_sku === 0) return "warning";
  if (summary.sku_no_inventory > 0) return "error";
  if (summary.sku_with_price_diff > 0) return "warning";
  if (summary.sku_ok === summary.total_sku) return "good";
  return "warning";
}

export const shopeeReconciliationExportHeaders = {
  sku: "SKU",
  model_sku: "Model SKU",
  item_name: "Item Name",
  model_name: "Model Name",
  variant_name: "Variant Name",
  inventory_price: "Inventory Price",
  expected_income: "Expected Income",
  total_transactions: "Total Transactions",
  unique_unit_prices: "Unique Unit Prices",
  unique_actual_incomes: "Unique Actual Incomes",
  has_multiple_prices: "Has Multiple Prices",
  has_price_difference: "Has Price Difference",
  status: "Status",
} as const;

export const tiktokReconciliationExportHeaders = {
  sku: "SKU ID",
  seller_sku: "Seller SKU",
  product_name: "Product Name",
  variant_name: "Variant Name",
  inventory_price: "Inventory Price",
  expected_income: "Expected Income",
  total_transactions: "Total Transactions",
  unique_unit_prices: "Unique Unit Prices",
  unique_actual_incomes: "Unique Actual Incomes",
  has_multiple_prices: "Has Multiple Prices",
  has_price_difference: "Has Price Difference",
  status: "Status",
} as const;

export const shopeeShippingExportHeaders = {
  order_sn: "Order SN",
  buyer_paid: "Buyer Paid Shipping Fee",
  actual_fee: "Actual Shipping Fee",
  shopee_rebate: "Shopee Shipping Rebate",
  difference: "Difference",
  status: "Status",
  order_date: "Order Date",
  buyer_name: "Buyer Name",
  payment_method: "Payment Method",
} as const;

export const tiktokShippingExportHeaders = {
  order_sn: "Order SN",
  customer_paid: "Shipping Fee Customer Paid",
  actual_fee: "Shipping Fee Actual",
  platform_discount: "Shipping Fee Platform Discount",
  difference: "Difference",
  status: "Status",
  order_date: "Order Date",
  order_status: "Order Status",
  currency: "Currency",
} as const;

export function getAnalyticsExportHeaders(
  platform: ReportPlatform,
  tab: "reconciliation" | "shipping_fee"
): Record<string, string> {
  if (tab === "reconciliation") {
    return platform === "shopee"
      ? shopeeReconciliationExportHeaders
      : tiktokReconciliationExportHeaders;
  }
  return platform === "shopee"
    ? shopeeShippingExportHeaders
    : tiktokShippingExportHeaders;
}

export function createCSV<T extends Record<string, unknown>>(
  data: T[],
  headers?: Record<string, string>
): string {
  if (!data || data.length === 0 || data.some((row) => row == null)) return "";

  const keys = headers ? Object.keys(headers) : Object.keys(data[0]);
  const headerRow = headers ? keys.map((k) => headers[k] || k) : keys;

  const csvRows = [
    headerRow.join(","),
    ...data.map((row) =>
      keys
        .map((key) => {
          const val = row[key];
          const str = Array.isArray(val)
            ? val.join(" | ")
            : val == null
              ? ""
              : String(val);
          if (str.includes(",") || str.includes('"') || str.includes("\n")) {
            return `"${str.replace(/"/g, '""')}"`;
          }
          return str;
        })
        .join(",")
    ),
  ];

  return csvRows.join("\r\n");
}

/**
 * Generate and download a CSV file from tabular data.
 *
 * @param data - Array of objects to export
 * @param filename - Output filename (without extension)
 * @param headers - Optional column header mapping (key → display name).
 *                  Defaults to object keys if not provided.
 */
export function exportToCSV<T extends Record<string, unknown>>(
  data: T[],
  filename: string,
  headers?: Record<string, string>
): void {
  if (!data || data.length === 0) return;
  const csvString = createCSV(data, headers);
  const blob = new Blob([csvString], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
  link.href = url;
  link.download = `${filename}.csv`;
  link.click();

  URL.revokeObjectURL(url);
}

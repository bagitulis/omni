/**
 * TikTok Analytics CSV Export Utils
 * CSV export functionality for TikTok analytics
 * API types use snake_case to match backend JSON response
 */

import type {
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
} from "@/composables/useTiktokAnalytics";

export function downloadCSV(rows: string[][], filename: string) {
  const blob = new Blob([rows.map((r) => r.join(",")).join("\n")], {
    type: "text/csv;charset=utf-8;",
  });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  link.click();
}

export function exportPriceToCSV(
  result: TiktokReconciliationResult,
  year: number,
  month: number,
) {
  const rows: string[][] = [
    [
      "Status",
      "SKU",
      "Product Name",
      "Inventory Price",
      "Expected Income",
      "Qty",
    ],
  ];
  result.sku_groups.forEach((sku) =>
    rows.push([
      sku.status,
      sku.seller_sku || sku.sku,
      `"${sku.product_name.replace(/"/g, '""')}"`,
      String(sku.inventory_price || 0),
      String(sku.expected_income || 0),
      String(sku.total_transactions),
    ]),
  );
  downloadCSV(rows, `tiktok-price-${year}-${month + 1}.csv`);
}

export function exportShippingToCSV(
  result: TiktokShippingFeeResult,
  year: number,
  month: number,
) {
  const rows: string[][] = [
    [
      "Order Date",
      "Order ID",
      "Buyer Paid",
      "Actual Fee",
      "Platform Discount",
      "Difference",
      "Status",
    ],
  ];
  result.orders.forEach((o) =>
    rows.push([
      o.order_date || "",
      o.order_id,
      String(o.buyer_paid),
      String(o.actual_fee),
      String(o.platform_discount),
      String(o.difference),
      o.order_status || "",
    ]),
  );
  downloadCSV(rows, `tiktok-shipping-${year}-${month + 1}.csv`);
}

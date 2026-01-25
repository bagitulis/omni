/**
 * TikTok Analytics CSV Export Utils
 * CSV export functionality for TikTok analytics
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
  month: number
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
  result.skuGroups.forEach((sku) =>
    rows.push([
      sku.status,
      sku.sellerSku || sku.sku,
      `"${sku.productName.replace(/"/g, '""')}"`,
      String(sku.inventoryPrice || 0),
      String(sku.expectedIncome || 0),
      String(sku.totalTransactions),
    ])
  );
  downloadCSV(rows, `tiktok-price-${year}-${month + 1}.csv`);
}

export function exportShippingToCSV(
  result: TiktokShippingFeeResult,
  year: number,
  month: number
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
      o.orderDate || "",
      o.orderId,
      String(o.buyerPaid),
      String(o.actualFee),
      String(o.platformDiscount),
      String(o.difference),
      o.orderStatus || "",
    ])
  );
  downloadCSV(rows, `tiktok-shipping-${year}-${month + 1}.csv`);
}

/**
 * TikTok Analytics CSV Export Utils
 * CSV export functionality for TikTok analytics
 * API types use snake_case to match backend JSON response
 */

// Types defined inline since the analytics composable is not ported yet.
interface TiktokSkuGroup {
  status: string;
  seller_sku: string;
  sku: string;
  product_name: string;
  inventory_price: number;
  expected_income: number;
  total_transactions: number;
}

interface TiktokReconciliationResult {
  sku_groups: TiktokSkuGroup[];
}

interface TiktokShippingOrder {
  order_date: string;
  order_id: string;
  buyer_paid: number;
  actual_fee: number;
  platform_discount: number;
  difference: number;
  order_status: string;
}

interface TiktokShippingFeeResult {
  orders: TiktokShippingOrder[];
}

export function downloadCSV(rows: string[][], filename: string): void {
  const csvContent = rows.map((row) => row.join(",")).join("\n");
  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const link = document.createElement("a");
  const url = URL.createObjectURL(blob);

  link.href = url;
  link.download = filename;
  link.click();

  URL.revokeObjectURL(url);
}

export function exportPriceToCSV(
  result: TiktokReconciliationResult,
  year: number,
  month: number,
): void {
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

  result.sku_groups.forEach((sku) => {
    rows.push([
      sku.status,
      sku.seller_sku || sku.sku,
      `"${sku.product_name.replace(/"/g, '""')}"`,
      String(sku.inventory_price || 0),
      String(sku.expected_income || 0),
      String(sku.total_transactions),
    ]);
  });

  downloadCSV(rows, `tiktok-price-${year}-${month + 1}.csv`);
}

export function exportShippingToCSV(
  result: TiktokShippingFeeResult,
  year: number,
  month: number,
): void {
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

  result.orders.forEach((order) => {
    rows.push([
      order.order_date || "",
      order.order_id,
      String(order.buyer_paid),
      String(order.actual_fee),
      String(order.platform_discount),
      String(order.difference),
      order.order_status || "",
    ]);
  });

  downloadCSV(rows, `tiktok-shipping-${year}-${month + 1}.csv`);
}

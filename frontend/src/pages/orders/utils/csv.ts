import { Order } from "@/types/order";

export function generateOrdersCSV(orders: Order[]): string {
  if (!orders || orders.length === 0) {
    return ""; // Empty CSV headers
  }

  // CSV headers
  const headers = [
    "Order No",
    "Platform",
    "Status",
    "Customer",
    "Total",
    "Date",
  ];

  // Convert orders to CSV rows
  const rows = orders.map((order) => [
    `"${order.order_sn || ""}"`, // Order No - quoted to preserve numbers
    `"${order.platform || ""}"`, // Platform
    `"${order.status || ""}"`, // Status
    `"${order.buyer_username || ""}"`, // Customer
    order.total_amount?.toFixed(2) || "0.00", // Total - no quotes for numbers
    `"${order.created_at || ""}"`, // Date - quoted for timestamp
  ]);

  // Combine headers and rows
  const csv = [headers.join(","), ...rows.map((row) => row.join(","))].join(
    "\n",
  );

  return csv;
}

export function downloadCSV(csv: string, filename: string) {
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
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

/**
 * Order Export Utilities
 * Handles CSV export and n8n webhook functionality
 */

import { useToast } from "@/composables/useToast";
import api from "@/services/api";
import type { Order } from "./useOrderManager";

export function useOrderExport() {
  const toast = useToast();

  /**
   * Prepare order data with consistent field mapping
   */
  function prepareOrderData(orders: Order[]) {
    return orders.map((order) => ({
      platform: order.platform || "",
      orderNo: order.order_no || order.orderSn || "",
      trackingNo: order.tracking_no || order.trackingNo || "",
      courier: order.shipping_carrier || order.courier || "",
      sku: order.sku || order.sellerSku || "",
      productName: order.product_name || order.productName || "",
      variationName: order.variation_name || order.variationName || "",
      qty: order.qty || order.quantity || 0,
    }));
  }

  /**
   * Export orders to CSV file locally only
   */
  async function exportToCSVOnly(
    orders: Order[],
    activeTab: string
  ): Promise<void> {
    try {
      if (orders.length === 0) {
        toast.add({
          severity: "warning",
          summary: "Warning",
          detail: "No data to export",
          life: 3000,
        });
        return;
      }

      const headers = [
        "Platform",
        "No. Pesanan",
        "No. Resi",
        "Ekspedisi",
        "SKU Seller",
        "Nama Produk",
        "Nama Variasi",
        "Qty",
      ];

      const preparedData = prepareOrderData(orders);
      const rows = preparedData.map((order) => [
        order.platform,
        order.orderNo,
        order.trackingNo,
        order.courier,
        order.sku,
        order.productName,
        order.variationName,
        order.qty,
      ]);

      const csvData = [headers, ...rows]
        .map((row) =>
          row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(",")
        )
        .join("\n");

      downloadCSV(csvData, activeTab);

      toast.add({
        severity: "success",
        summary: "Export Berhasil",
        detail: `${orders.length} order berhasil di-export ke CSV`,
        life: 4000,
      });
    } catch (err: any) {
      console.error("Error exporting CSV:", err);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to export CSV: " + err.message,
        life: 3000,
      });
    }
  }

  /**
   * Send orders to N8N webhook only
   */
  async function exportToN8NOnly(
    orders: Order[],
    activeTab: string
  ): Promise<void> {
    try {
      if (orders.length === 0) {
        toast.add({
          severity: "warning",
          summary: "Warning",
          detail: "No data to export",
          life: 3000,
        });
        return;
      }

      await sendToN8n(orders, activeTab);

      toast.add({
        severity: "success",
        summary: "Export ke N8N Berhasil",
        detail: `${orders.length} order berhasil dikirim ke N8N`,
        life: 4000,
      });
    } catch (err: any) {
      console.error("Error sending to n8n:", err);
      const isAuthError = err?.response?.status === 401;

      toast.add({
        severity: "error",
        summary: isAuthError ? "Login Required" : "Export Gagal",
        detail: isAuthError
          ? "Login diperlukan untuk kirim ke N8N"
          : "Gagal kirim ke N8N: " + err.message,
        life: 5000,
      });
    }
  }

  /**
   * Export orders to CSV file locally AND send to n8n webhook (legacy)
   */
  async function exportToCSV(
    orders: Order[],
    activeTab: string
  ): Promise<void> {
    try {
      if (orders.length === 0) {
        toast.add({
          severity: "warning",
          summary: "Warning",
          detail: "No data to export",
          life: 3000,
        });
        return;
      }

      // 1. Download CSV locally
      const headers = [
        "Platform",
        "No. Pesanan",
        "No. Resi",
        "Ekspedisi",
        "SKU Seller",
        "Nama Produk",
        "Nama Variasi",
        "Qty",
      ];

      // Support both snake_case (from other tabs) and camelCase (from today tab)
      const rows = orders.map((order) => [
        order.platform || "",
        order.order_no || order.orderSn || "",
        order.tracking_no || order.trackingNo || "",
        order.shipping_carrier || order.courier || "",
        order.sku || order.sellerSku || "",
        order.product_name || order.productName || "",
        order.variation_name || order.variationName || "",
        order.qty || order.quantity || 0,
      ]);

      const csvData = [headers, ...rows]
        .map((row) =>
          row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(",")
        )
        .join("\n");

      downloadCSV(csvData, activeTab);

      // 2. Send to n8n webhook
      try {
        await sendToN8n(orders, activeTab);
        toast.add({
          severity: "success",
          summary: "Export Berhasil",
          detail: `${orders.length} order di-export ke CSV dan dikirim ke n8n`,
          life: 4000,
        });
      } catch (n8nError: any) {
        // CSV download succeeded, but n8n failed
        console.warn("[Export] n8n webhook failed:", n8nError);

        // Check if it's an auth error
        const isAuthError = n8nError?.response?.status === 401;

        toast.add({
          severity: "warning",
          summary: isAuthError ? "Login Required" : "Partial Success",
          detail: isAuthError
            ? "CSV downloaded. Login diperlukan untuk kirim ke n8n."
            : `CSV downloaded, tapi n8n gagal: ${n8nError.message}`,
          life: 5000,
        });
      }
    } catch (err: any) {
      console.error("Error exporting CSV:", err);
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to export CSV: " + err.message,
        life: 3000,
      });
    }
  }

  /**
   * Send order data to n8n webhook via backend API
   * Field names match Google Sheets headers (Indonesian)
   */
  async function sendToN8n(orders: Order[], exportType: string): Promise<void> {
    // Support both snake_case (from other tabs) and camelCase (from today tab)
    // Field names MUST match Google Sheets column headers exactly
    const response = await api.client.post("/n8n/export-orders", {
      orders: orders.map((order) => ({
        Platform: order.platform || "",
        "No. Pesanan": order.order_no || order.orderSn || "",
        "No. Resi": order.tracking_no || order.trackingNo || "",
        Ekspedisi: order.shipping_carrier || order.courier || "",
        "SKU Seller": order.sku || order.sellerSku || "",
        "Nama Produk": order.product_name || order.productName || "",
        "Nama Variasi": order.variation_name || order.variationName || "",
        Qty: order.qty || order.quantity || 0,
      })),
      exportType,
      metadata: {
        exportedAt: new Date().toISOString(),
        source: "order-manager",
      },
    });

    if (!response.data.success) {
      throw new Error(response.data.error || "n8n export failed");
    }

    console.log("[Export] n8n webhook response:", response.data);
  }

  function downloadCSV(csvData: string, activeTab: string): void {
    const blob = new Blob([csvData], { type: "text/csv;charset=utf-8;" });
    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);

    link.setAttribute("href", url);
    link.setAttribute(
      "download",
      `orders_${activeTab}_${new Date().toISOString().split("T")[0]}.csv`
    );
    link.style.visibility = "hidden";

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  return { exportToCSV, exportToCSVOnly, exportToN8NOnly, sendToN8n };
}

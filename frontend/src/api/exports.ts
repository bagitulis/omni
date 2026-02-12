import apiClient from "@/api/client";
import { logger } from "@/lib/logger";

/**
 * Export orders request parameters
 */
export interface ExportOrdersParams {
  platform?: string[]; // Multi-select support
  status?: string[]; // Multi-select support
  date_from: string; // YYYY-MM-DD
  date_to: string; // YYYY-MM-DD
  format: "csv" | "excel";
}

/**
 * Export orders to a file (CSV/Excel)
 * Triggers a file download in the browser
 */
export async function exportOrders(
  params: ExportOrdersParams,
): Promise<boolean> {
  // We need to use the raw axios client to handle blob response
  // specific for file downloads
  try {
    const response = await apiClient.client.post("/orders/export", params, {
      responseType: "blob", // Important: response is a binary file
      timeout: 60000, // Longer timeout for exports
    });

    // Create a URL for the blob
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement("a");
    link.href = url;

    // Extract filename from content-disposition header if available, or default
    const contentDisposition = response.headers["content-disposition"];
    let filename = `orders_export_${new Date().toISOString().split("T")[0]}.${params.format === "excel" ? "xlsx" : "csv"}`;

    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename="?([^"]+)"?/);
      if (filenameMatch && filenameMatch.length === 2) {
        filename = filenameMatch[1];
      }
    }

    link.setAttribute("download", filename);
    document.body.appendChild(link);
    link.click();

    // Cleanup
    link.parentNode?.removeChild(link);
    window.URL.revokeObjectURL(url);

    return true;
  } catch (error) {
    logger.error("Export failed:", { error });
    throw error;
  }
}

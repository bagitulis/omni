import apiService from "./api";

export interface ShippingLabelResponse {
  success: boolean;
  data: {
    doc_url?: string;
    tracking_number?: string;
    file_data?: string; // Base64 PDF for some platforms
    file_path?: string; // Local file path for download endpoints
    file_size?: number;
    status?: string;
    message?: string;
  };
  error?: string;
}

/**
 * Download shipping label for an order
 * Uses backend download endpoints that save to /app/uploads/labels
 * Then triggers browser download via doc_url or file_data
 */
export async function downloadShippingLabel(
  platform: string,
  orderId: string,
  packageId?: string,
  options?: { includeProducts?: boolean },
): Promise<void> {
  // Normalize platform name
  const p = platform.toLowerCase();

  // Build query params
  const queryParams = new URLSearchParams();
  if (options?.includeProducts) {
    queryParams.set("include_products", "true");
  }
  const queryString = queryParams.toString();

  // Use download endpoints that save to local file
  const endpoints: Record<string, string> = {
    shopee: `/shopee/shipping/download/${orderId}`,
    tiktok: `/tiktok/shipping/download/order/${orderId}${queryString ? `?${queryString}` : ""}`,
    lazada: `/lazada/shipping/label/${orderId}`,
  };

  if (!endpoints[p]) {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  try {
    const response = await apiService.get<ShippingLabelResponse>(endpoints[p]);

    if (!response.success) {
      throw new Error(response.error || "Failed to get shipping label");
    }

    // Handle different response types
    if (response.data.doc_url) {
      // Open URL in new tab (for signed URLs)
      window.open(response.data.doc_url, "_blank");
    } else if (response.data.file_data) {
      // Download base64 PDF
      downloadBase64PDF(
        response.data.file_data,
        `shipping-label-${orderId}.pdf`,
      );
    } else if (response.data.file_path) {
      // File saved on server - show success message
      // The file is at /app/uploads/labels which maps to backend/uploads/labels
      const fileName =
        response.data.file_path.split("/").pop() || `label-${orderId}.pdf`;
      console.log(`Label saved to server: ${response.data.file_path}`);

      // Try to fetch the file via serve endpoint if available
      const serveUrl = `/api/uploads/labels/${fileName}`;
      try {
        const blob = await fetch(serveUrl, {
          headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
        }).then((r) => (r.ok ? r.blob() : null));

        if (blob) {
          const url = URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = url;
          a.download = fileName;
          a.click();
          URL.revokeObjectURL(url);
          return;
        }
      } catch {
        // Fallback: just log the path
      }

      // Alert user about file location if we can't serve it
      alert(
        `Label saved to: ${response.data.file_path}\nFile size: ${response.data.file_size} bytes`,
      );
    } else {
      throw new Error("No shipping label available");
    }
  } catch (error) {
    console.error("Failed to download shipping label:", error);
    throw error;
  }
}

/**
 * Helper to download base64 PDF
 */
function downloadBase64PDF(base64: string, filename: string): void {
  const link = document.createElement("a");
  link.href = `data:application/pdf;base64,${base64}`;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

export default {
  downloadShippingLabel,
  downloadShippingLabelsBatch,
};

/**
 * Batch download shipping labels for multiple orders (TikTok only)
 * Returns results for each order
 */
export async function downloadShippingLabelsBatch(
  platform: string,
  orderIds: string[],
  options?: { includeProducts?: boolean },
): Promise<{
  total: number;
  success: number;
  failed: number;
  results: Array<{
    order_id: string;
    status: string;
    file_path?: string;
    file_size?: number;
    error?: string;
  }>;
}> {
  const p = platform.toLowerCase();

  if (p !== "tiktok") {
    throw new Error("Batch download only supported for TikTok");
  }

  const response = await apiService.post<{
    success: boolean;
    data: {
      total: number;
      success: number;
      failed: number;
      results: Array<{
        order_id: string;
        status: string;
        file_path?: string;
        file_size?: number;
        file_data?: string;
        error?: string;
      }>;
    };
    error?: string;
  }>("/tiktok/shipping/download/batch", {
    order_ids: orderIds,
    include_products: options?.includeProducts || false,
  });

  if (!response.success) {
    throw new Error(response.error || "Batch download failed");
  }

  // For successful downloads, trigger browser downloads for each file_data
  const downloadedResults: typeof response.data.results = [];
  for (const result of response.data.results) {
    if (result.status === "SUCCESS" && result.file_data) {
      downloadBase64PDF(result.file_data, `shipping-label-${result.order_id}.pdf`);
      // Don't include file_data in returned results (too large)
      const { file_data, ...rest } = result;
      downloadedResults.push(rest);
    } else {
      downloadedResults.push(result);
    }
  }

  return {
    total: response.data.total,
    success: response.data.success,
    failed: response.data.failed,
    results: downloadedResults,
  };
}

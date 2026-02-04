import apiService from "./api";

export interface ShippingLabelResponse {
  success: boolean;
  data: {
    doc_url?: string;
    tracking_number?: string;
    file_data?: string; // Base64 PDF for some platforms
  };
}

/**
 * Download shipping label for an order
 * Opens PDF in new tab or triggers download
 */
export async function downloadShippingLabel(
  platform: string,
  orderId: string,
  packageId?: string,
): Promise<void> {
  // Normalize platform name
  const p = platform.toLowerCase();

  const endpoints: Record<string, string> = {
    shopee: `/shopee/shipping/label/${orderId}`,
    tiktok: `/tiktok/shipping/document/${packageId || orderId}`,
    lazada: `/lazada/shipping/label/${orderId}`,
  };

  if (!endpoints[p]) {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  try {
    const response = await apiService.get<ShippingLabelResponse>(endpoints[p]);

    if (!response.success) {
      throw new Error("Failed to get shipping label");
    }

    if (response.data.doc_url) {
      // Open URL in new tab
      window.open(response.data.doc_url, "_blank");
    } else if (response.data.file_data) {
      // Download base64 PDF
      downloadBase64PDF(
        response.data.file_data,
        `shipping-label-${orderId}.pdf`,
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
};

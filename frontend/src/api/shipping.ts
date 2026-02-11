import apiClient from "./client";

export interface ShippingLabelData {
  doc_url?: string;
  tracking_number?: string;
  file_data?: string;
  file_path?: string;
  file_size?: number;
  status?: string;
  message?: string;
}

export interface DownloadShippingLabelOptions {
  include_products?: boolean;
}

export interface BatchShippingLabelResult {
  order_id: string;
  status: string;
  file_path?: string;
  file_size?: number;
  file_data?: string;
  error?: string;
}

export interface BatchShippingLabelResponse {
  total: number;
  success: number;
  failed: number;
  results: BatchShippingLabelResult[];
}

function getShippingDownloadEndpoint(
  platform: string,
  order_id: string,
): string {
  const normalized_platform = platform.toLowerCase();

  if (normalized_platform === "shopee") {
    return `/shopee/shipping/download/${order_id}`;
  }

  if (normalized_platform === "tiktok") {
    return `/tiktok/shipping/download/order/${order_id}`;
  }

  if (normalized_platform === "lazada") {
    return `/lazada/shipping/label/${order_id}`;
  }

  throw new Error(`Unsupported platform: ${platform}`);
}

function triggerBlobDownload(blob: Blob, filename: string): void {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

export function downloadBase64PDF(base64: string, filename: string): void {
  const pdf_blob = new Blob(
    [Uint8Array.from(atob(base64), (char) => char.charCodeAt(0))],
    {
      type: "application/pdf",
    },
  );
  triggerBlobDownload(pdf_blob, filename);
}

async function downloadServerFile(
  file_path: string,
  fallback_filename: string,
) {
  const file_name = file_path.split("/").pop() || fallback_filename;
  const response = await apiClient.client.get(`/uploads/labels/${file_name}`, {
    responseType: "blob",
  });
  const blob = response.data as Blob;
  triggerBlobDownload(blob, file_name);
}

export async function downloadShippingLabel(
  platform: string,
  order_id: string,
  package_id?: string,
  options?: DownloadShippingLabelOptions,
): Promise<void> {
  const endpoint = getShippingDownloadEndpoint(platform, order_id);
  const response = await apiClient.get<ShippingLabelData>(endpoint, {
    params: {
      package_id,
      include_products: options?.include_products,
    },
  });

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to download shipping label");
  }

  if (response.data.doc_url) {
    window.open(response.data.doc_url, "_blank", "noopener,noreferrer");
    return;
  }

  if (response.data.file_data) {
    downloadBase64PDF(
      response.data.file_data,
      `shipping-label-${order_id}.pdf`,
    );
    return;
  }

  if (response.data.file_path) {
    await downloadServerFile(response.data.file_path, `label-${order_id}.pdf`);
    return;
  }

  throw new Error("No shipping label available");
}

export async function downloadShippingLabelsBatch(
  platform: string,
  order_ids: string[],
  options?: DownloadShippingLabelOptions,
): Promise<BatchShippingLabelResponse> {
  if (platform.toLowerCase() !== "tiktok") {
    throw new Error("Batch download only supported for TikTok");
  }

  const response = await apiClient.post<BatchShippingLabelResponse>(
    "/tiktok/shipping/download/batch",
    {
      order_ids,
      include_products: options?.include_products || false,
    },
  );

  if (!response.success || !response.data) {
    throw new Error(response.error || "Batch download failed");
  }

  const download_results: BatchShippingLabelResult[] = [];
  for (const result of response.data.results) {
    if (result.status === "SUCCESS" && result.file_data) {
      downloadBase64PDF(
        result.file_data,
        `shipping-label-${result.order_id}.pdf`,
      );
    }

    // Exclude file_data from results
    const { file_data, ...without_file_data } = result;
    void file_data; // Mark as intentionally unused
    download_results.push(without_file_data);
  }

  return {
    total: response.data.total,
    success: response.data.success,
    failed: response.data.failed,
    results: download_results,
  };
}

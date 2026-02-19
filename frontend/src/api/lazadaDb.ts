import apiClient from "./client";

export interface LazadaSyncResult {
  message?: string;
  status?: string;
  processed?: number;
  queued?: number;
  total?: number;
  detail_saved?: number;
  [key: string]: unknown;
}

export async function syncProductsToDb(): Promise<LazadaSyncResult> {
  const response = await apiClient.get<LazadaSyncResult>("/lazada/products");

  if (!response.success) {
    throw new Error(response.error || "Failed to sync Lazada products");
  }

  const payload = response as unknown as LazadaSyncResult;
  return {
    ...payload,
    processed:
      typeof payload.detail_saved === "number"
        ? payload.detail_saved
        : payload.processed,
  };
}

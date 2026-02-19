import apiClient from "./client";

export interface LazadaSyncResult {
  message?: string;
  status?: string;
  processed?: number;
  queued?: number;
  [key: string]: unknown;
}

export async function syncProductsToDb(): Promise<LazadaSyncResult> {
  const response = await apiClient.post<LazadaSyncResult>(
    "/lazada/sync/products",
    {},
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to sync Lazada products");
  }

  return response.data ?? {};
}

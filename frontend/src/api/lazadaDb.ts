import { API_TIMEOUT } from "@/lib/constants";

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
  const response = await apiClient.post<LazadaSyncResult>(
    "/lazada/sync/products",
    undefined,
    { timeout: API_TIMEOUT.EXTRA_LONG },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to sync Lazada products");
  }

  const syncResult: LazadaSyncResult = response.data ?? {};
  return {
    ...syncResult,
    processed:
      typeof syncResult.detail_saved === "number"
        ? syncResult.detail_saved
        : syncResult.processed,
  };
}

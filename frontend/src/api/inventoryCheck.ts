import type { SkuCheckResult } from "@/types/inventory";
import apiClient from "./client";

/**
 * Batch check SKU status across platforms
 * Backend route: POST /api/inventory/batch-check-sku
 */
export async function batchCheckSku(skus: string[]): Promise<SkuCheckResult[]> {
  const response = await apiClient.post<{
    total: number;
    checked: number;
    results: SkuCheckResult[];
  }>("/inventory/batch-check-sku", { skus });
  if (!response.success) {
    throw new Error(response.error || "Failed to check SKU status");
  }
  return response.data?.results || [];
}

/**
 * Batch check platform status (alias for batchCheckSku)
 * Backend route: POST /api/inventory/batch-check-sku
 */
export async function checkPlatformStatus(
  skus: string[],
): Promise<SkuCheckResult[]> {
  return batchCheckSku(skus);
}

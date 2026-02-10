import apiClient from "./client";
import {
  CloneRequest,
  BatchCloneRequest,
  CloneResult,
  BatchCloneResult,
  ProductData,
  CloneTargetsResult,
  ConflictResult,
} from "@/types/clone";

/**
 * Clone a single product to another platform
 * POST /api/products/clone
 */
export async function cloneProduct(req: CloneRequest): Promise<CloneResult> {
  const response = await apiClient.post<CloneResult>("/products/clone", req);
  if (!response.success) {
    throw new Error(response.error || "Failed to clone product");
  }
  return response.data!;
}

/**
 * Clone multiple products in batch
 * POST /api/products/clone/batch
 */
export async function batchClone(
  req: BatchCloneRequest,
): Promise<BatchCloneResult> {
  const response = await apiClient.post<BatchCloneResult>(
    "/products/clone/batch",
    req,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to batch clone products");
  }
  return response.data!;
}

/**
 * Get clone status by ID
 * GET /api/products/clone/status/:id
 */
export async function getCloneStatus(id: string): Promise<CloneResult> {
  const response = await apiClient.get<CloneResult>(
    `/products/clone/status/${id}`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to get clone status");
  }
  return response.data!;
}

/**
 * Get product data for cloning
 * GET /api/clone/product-data?platform=shopee&sku=SKU123
 */
export async function getProductData(
  platform: string,
  sku: string,
): Promise<ProductData> {
  const response = await apiClient.get<{ product: ProductData }>(
    "/clone/product-data",
    { params: { platform, sku } },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to get product data");
  }
  return response.data!.product;
}

/**
 * Get available clone targets for a SKU
 * GET /api/clone/available-targets?sku=SKU123
 */
export async function getAvailableTargets(
  sku: string,
): Promise<CloneTargetsResult> {
  const response = await apiClient.get<CloneTargetsResult>(
    "/clone/available-targets",
    { params: { sku } },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to get available targets");
  }
  // Map response fields to match interface
  return {
    sku: response.data?.sku || sku,
    status: response.data?.status || {
      shopee: false,
      lazada: false,
      tiktok: false,
    },
    sources: response.data?.sources || [],
    targets: response.data?.targets || [],
  };
}

/**
 * Get clone preview with conflict detection
 * GET /api/clone/preview?source_platform=shopee&target_platform=lazada&source_item_id=123&sku=SKU123
 */
export async function getClonePreview(params: {
  source_platform: string;
  target_platform: string;
  source_item_id: string;
  sku?: string;
}): Promise<ConflictResult> {
  const response = await apiClient.get<ConflictResult>("/clone/preview", {
    params,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to get clone preview");
  }
  return response.data!;
}

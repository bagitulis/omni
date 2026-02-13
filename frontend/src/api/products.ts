import apiClient from "./client";
import type {
  MasterProduct,
  ProductListResponse,
  ProductListFilter,
  CreateMasterProductInput,
  UpdateMasterProductInput,
  ImportRow,
  ImportPreviewData,
  AutoMapResult,
  MappingStatus,
  LinkSkuData,
  UnlinkSkuData,
  BatchSkuUpdateItem,
  BatchSkuUpdateResult,
} from "@/types/product";

const BASE_PATH = "/master-products";

export async function getProducts(
  params: ProductListFilter = {},
): Promise<ProductListResponse> {
  const cleanParams: Record<string, unknown> = {};
  if (params.page) cleanParams.page = params.page;
  if (params.limit) cleanParams.limit = params.limit;
  if (params.search) cleanParams.search = params.search;
  if (params.status && params.status !== "all")
    cleanParams.status = params.status;
  if (params.platform && params.platform !== "all")
    cleanParams.platform = params.platform;

  // Cast to any to handle the response shape correctly
  // Backend returns { success, data: [...], meta: {...} } which IS ProductListResponse
  const response = await apiClient.get<unknown>(BASE_PATH, {
    params: cleanParams,
  });

  if (!response.success) {
    throw new Error(response.error || "Failed to fetch products");
  }

  return response as ProductListResponse;
}

export async function getProductById(
  id: string | number,
): Promise<MasterProduct> {
  const response = await apiClient.get<unknown>(`${BASE_PATH}/${id}`);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch product");
  }
  return response.data;
}

export async function createProduct(
  data: CreateMasterProductInput,
): Promise<MasterProduct> {
  const response = await apiClient.post<unknown>(BASE_PATH, data);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to create product");
  }
  return response.data;
}

export async function updateProduct(
  id: string | number,
  data: UpdateMasterProductInput,
): Promise<MasterProduct> {
  const response = await apiClient.put<unknown>(`${BASE_PATH}/${id}`, data);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to update product");
  }
  return response.data;
}

export async function deleteProduct(id: string | number): Promise<void> {
  const response = await apiClient.delete(`${BASE_PATH}/${id}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to delete product");
  }
}

export async function syncProduct(
  id: string | number,
  platform?: string,
): Promise<{ skus_synced: number; platform: string }> {
  const response = await apiClient.post<unknown>(`${BASE_PATH}/${id}/sync`, {
    target_platform: platform,
  });
  if (!response.success || !response.data) {
    throw new Error("Failed to sync product");
  }
  return response.data;
}

export async function importProducts(
  rows: ImportRow[],
): Promise<{ imported: number }> {
  const response = await apiClient.post<unknown>(`${BASE_PATH}/import`, { rows });
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to import products");
  }
  return response.data;
}

// Import preview — sends file to backend for parsing and validation
export async function getImportPreview(file: File): Promise<ImportPreviewData> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await apiClient.client.post<unknown>(
    `${BASE_PATH}/import/preview`,
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );

  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error || "Failed to preview import file");
  }

  return response.data.data;
}

// Auto-map SKUs to platform products
export async function autoMapSkus(skus: string[]): Promise<AutoMapResult> {
  const response = await apiClient.post<unknown>(`${BASE_PATH}/mapping/auto-link`, {
    skus,
  });

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to auto-map SKUs");
  }

  return response.data;
}

// Get mapping status
export async function getMappingStatus(): Promise<MappingStatus> {
  const response = await apiClient.get<unknown>(`${BASE_PATH}/mapping/status`);

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch mapping status");
  }

  return response.data;
}

// Link a master SKU to a platform product
export async function linkSkuToPlatform(data: LinkSkuData): Promise<void> {
  const response = await apiClient.post(`${BASE_PATH}/mapping/link`, data);

  if (!response.success) {
    throw new Error(response.error || "Failed to link SKU to platform");
  }
}

// Unlink a master SKU from a platform product
export async function unlinkSkuFromPlatform(
  data: UnlinkSkuData,
): Promise<void> {
  const response = await apiClient.delete(`${BASE_PATH}/mapping/link`, {
    data,
  });

  if (!response.success) {
    throw new Error(response.error || "Failed to unlink SKU from platform");
  }
}

// Batch update SKUs (price, stock, seller_sku)
export async function batchUpdateSkus(
  items: BatchSkuUpdateItem[],
): Promise<BatchSkuUpdateResult> {
  const response = await apiClient.put<unknown>(`${BASE_PATH}/skus/batch`, {
    items,
  });

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to batch update SKUs");
  }

  return response.data;
}

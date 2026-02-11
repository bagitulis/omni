import apiClient from "./client";
import type {
  MasterProduct,
  ProductListResponse,
  ProductListFilter,
  CreateMasterProductInput,
  UpdateMasterProductInput,
  ImportRow,
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
  const response = await apiClient.get<any>(BASE_PATH, {
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
  const response = await apiClient.get<any>(`${BASE_PATH}/${id}`);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to fetch product");
  }
  return response.data;
}

export async function createProduct(
  data: CreateMasterProductInput,
): Promise<MasterProduct> {
  const response = await apiClient.post<any>(BASE_PATH, data);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to create product");
  }
  return response.data;
}

export async function updateProduct(
  id: string | number,
  data: UpdateMasterProductInput,
): Promise<MasterProduct> {
  const response = await apiClient.put<any>(`${BASE_PATH}/${id}`, data);
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
  const response = await apiClient.post<any>(`${BASE_PATH}/${id}/sync`, {
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
  const response = await apiClient.post<any>(`${BASE_PATH}/import`, { rows });
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to import products");
  }
  return response.data;
}

import apiClient from "./client";
import {
  ProductListResponse,
  Product,
  BackendProduct,
  ImportRow,
} from "@/types/product";

export interface GetProductsParams {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}

/**
 * Backend response format for product lists
 */
interface BackendProductResponse {
  success: boolean;
  data: BackendProduct[];
  meta: {
    total: number;
    page: number;
    page_size: number;
  };
}

/**
 * Transform backend product to frontend format
 * Sets all required mapped fields for UI components
 */
function transformProduct(backendProduct: BackendProduct): Product {
  return {
    ...backendProduct,
    item_id: String(backendProduct.id),
    item_name: backendProduct.title,
    item_sku: `SKU-${backendProduct.id}`,
    price: 0, // Master products don't have price - set from platform variants
    stock: 0, // Master products don't have stock - set from platform variants
    platform: "master", // Master products are platform-agnostic
    image_url: backendProduct.images?.[0] || "",
  };
}

/**
 * Fetch master products list
 * Backend route: GET /api/master-products
 */
export async function getProducts(
  params?: GetProductsParams,
): Promise<ProductListResponse> {
  // Clean params - don't send "all" values to backend
  const cleanParams: Record<string, unknown> = {};
  if (params) {
    if (params.page) cleanParams.page = params.page;
    if (params.limit) cleanParams.limit = params.limit;
    if (params.search) cleanParams.search = params.search;
    if (params.status && params.status !== "all")
      cleanParams.status = params.status;
    if (params.platform && params.platform !== "all")
      cleanParams.platform = params.platform;
  }

  // Backend returns flat response: { success, data: [...], meta: {...} }
  const response = await apiClient.client.get("/master-products", {
    params: cleanParams,
  });

  const backendData = response.data as BackendProductResponse;

  if (!backendData.success) {
    throw new Error("Failed to fetch products");
  }

  // Transform backend response to frontend format
  const products = (backendData.data || []).map(transformProduct);

  return {
    products,
    total: backendData.meta?.total || products.length,
    page: backendData.meta?.page || 1,
    page_size: backendData.meta?.page_size || 20,
  };
}

/**
 * Get product by ID
 * Backend route: GET /api/master-products/:id
 */
export async function getProductById(id: string): Promise<Product> {
  const response = await apiClient.get<Product>(`/master-products/${id}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch product");
  }
  return response.data!;
}

/**
 * Create a new master product
 * Backend route: POST /api/master-products
 */
export async function createProduct(data: Partial<Product>): Promise<Product> {
  const response = await apiClient.post<Product>("/master-products", data);
  if (!response.success) {
    throw new Error(response.error || "Failed to create product");
  }
  return response.data!;
}

/**
 * Update a master product
 * Backend route: PUT /api/master-products/:id
 */
export async function updateProduct(
  id: string,
  data: Partial<Product>,
): Promise<Product> {
  const response = await apiClient.put<Product>(`/master-products/${id}`, data);
  if (!response.success) {
    throw new Error(response.error || "Failed to update product");
  }
  return response.data!;
}

/**
 * Delete a master product
 * Backend route: DELETE /api/master-products/:id
 */
export async function deleteProduct(id: string): Promise<boolean> {
  const response = await apiClient.delete(`/master-products/${id}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to delete product");
  }
  return true;
}

/**
 * Get product mapping status
 * Backend route: GET /api/master-products/:id/mapping
 */
export async function getProductMappingStatus(id: string): Promise<unknown> {
  const response = await apiClient.get<unknown>(
    `/master-products/${id}/mapping`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch mapping status");
  }
  return response.data;
}

/**
 * Sync product to platforms
 * Backend route: POST /api/master-products/:id/sync
 */
export async function syncProduct(id: string): Promise<unknown> {
  const response = await apiClient.post<unknown>(`/master-products/${id}/sync`);
  if (!response.success) {
    throw new Error(response.error || "Failed to sync product");
  }
  return response.data;
}

/**
 * Import products in bulk
 * Backend route: POST /api/master-products/import
 */
export async function importProducts(
  rows: ImportRow[],
): Promise<{ imported: number }> {
  const response = await apiClient.post<{ imported: number }>(
    "/master-products/import",
    { rows },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to import products");
  }
  return response.data!;
}

import apiClient from "./client";
import { ProductListResponse, Product } from "@/types/product";

export interface GetProductsParams {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}

/**
 * Fetch master products list
 * Backend route: GET /api/master-products
 */
export async function getProducts(
  params?: GetProductsParams,
): Promise<ProductListResponse> {
  const response = await apiClient.get<ProductListResponse>(
    "/master-products",
    {
      params,
    },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch products");
  }
  return response.data!;
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
export async function getProductMappingStatus(id: string): Promise<any> {
  const response = await apiClient.get(`/master-products/${id}/mapping`);
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch mapping status");
  }
  return response.data;
}

/**
 * Sync product to platforms
 * Backend route: POST /api/master-products/:id/sync
 */
export async function syncProduct(id: string): Promise<any> {
  const response = await apiClient.post(`/master-products/${id}/sync`);
  if (!response.success) {
    throw new Error(response.error || "Failed to sync product");
  }
  return response.data;
}

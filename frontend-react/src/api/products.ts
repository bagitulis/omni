import apiClient from "./client";
import { ProductListResponse } from "@/types/product";

export async function getProducts(params?: {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}): Promise<ProductListResponse> {
  const response = await apiClient.get<ProductListResponse>("/products", {
    params,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch products");
  }
  return response.data!;
}

export async function deleteProduct(id: string): Promise<boolean> {
  const response = await apiClient.delete(`/products/${id}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to delete product");
  }
  return true;
}

import { API_TIMEOUT } from "@/lib/constants";

import apiClient from "./client";

export interface TiktokDbProduct {
  id?: number;
  product_id: string;
  title?: string;
  status?: string;
  skus?: Array<Record<string, unknown>>;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

export interface TiktokDbStatistics {
  total_products?: number;
  active_products?: number;
  inactive_products?: number;
  [key: string]: unknown;
}

export interface TiktokProductSearchParams {
  status?: string;
  page_size?: number;
  limit?: number;
  sync_to_db: boolean;
}

export async function searchProducts(
  status: string = "ACTIVATE",
  page_size: number = 100,
  limit?: number,
): Promise<TiktokDbProduct[]> {
  const payload: TiktokProductSearchParams = {
    status,
    page_size,
    limit,
    sync_to_db: true,
  };

  const response = await apiClient.post<TiktokDbProduct[]>(
    "/tiktok/products/search",
    payload,
    { timeout: API_TIMEOUT.EXTRA_LONG },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to search products");
  }

  return response.data ?? [];
}

export async function getAllProducts(
  status: string = "ACTIVATE",
  page_size: number = 100,
): Promise<TiktokDbProduct[]> {
  const response = await apiClient.post<TiktokDbProduct[]>(
    "/tiktok/products/get-all",
    {
      status,
      page_size,
      sync_to_db: true,
    },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get all products");
  }

  return response.data ?? [];
}

export async function getProductDetail(
  product_id: string | number,
): Promise<TiktokDbProduct | null> {
  const response = await apiClient.get<TiktokDbProduct>(
    `/tiktok/products/${product_id}`,
    {
      params: { sync_to_db: true },
    },
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get product detail");
  }

  return response.data ?? null;
}

export async function getProductListFromDB(): Promise<TiktokDbProduct[]> {
  const response = await apiClient.get<TiktokDbProduct[]>(
    "/tiktok/db/products/list",
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get product list from DB");
  }

  return response.data ?? [];
}

export async function getMasterProductsFromDB(): Promise<TiktokDbProduct[]> {
  const response = await apiClient.get<TiktokDbProduct[]>(
    "/tiktok/db/products/master",
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get master products from DB");
  }

  return response.data ?? [];
}

export async function searchProductsInDB(
  query: string,
  field: string = "name",
): Promise<TiktokDbProduct[]> {
  const response = await apiClient.get<TiktokDbProduct[]>("/tiktok/db/search", {
    params: {
      q: query,
      field,
    },
  });

  if (!response.success) {
    throw new Error(response.error || "Failed to search products in DB");
  }

  return response.data ?? [];
}

export async function getProductsByStatus(
  status: string,
): Promise<TiktokDbProduct[]> {
  const response = await apiClient.get<TiktokDbProduct[]>(
    `/tiktok/db/products/status/${status}`,
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get products by status");
  }

  return response.data ?? [];
}

export async function getProductById(
  product_id: string,
): Promise<TiktokDbProduct | null> {
  const response = await apiClient.get<TiktokDbProduct>(
    `/tiktok/db/products/${product_id}`,
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get product by ID");
  }

  return response.data ?? null;
}

export async function getStatistics(): Promise<TiktokDbStatistics> {
  const response = await apiClient.get<TiktokDbStatistics>(
    "/tiktok/db/statistics",
  );

  if (!response.success) {
    throw new Error(response.error || "Failed to get statistics");
  }

  return response.data ?? {};
}

export async function deleteProduct(product_id: string): Promise<void> {
  const response = await apiClient.delete(`/tiktok/db/products/${product_id}`);

  if (!response.success) {
    throw new Error(response.error || "Failed to delete product");
  }
}

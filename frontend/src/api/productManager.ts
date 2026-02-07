import apiClient from "./client";
import {
  DbProductsResponse,
  ProductManagerPlatform,
  SyncProductsResponse,
} from "@/types/product_manager";

export interface GetDbProductsParams {
  offset?: number;
  limit?: number;
}

export async function getDbProducts(
  platform: ProductManagerPlatform,
  params: GetDbProductsParams,
): Promise<DbProductsResponse> {
  const response = await apiClient.client.get(`/${platform}/db/products`, {
    params,
  });

  const data = response.data as DbProductsResponse;
  if (!data.success) {
    throw new Error("Failed to fetch products");
  }
  return data;
}

export async function syncPlatformProducts(
  platform: ProductManagerPlatform,
): Promise<SyncProductsResponse> {
  const response = await apiClient.client.post(`/${platform}/sync/products`);
  const data = response.data as SyncProductsResponse;
  if (!data.success) {
    throw new Error(data.error || "Failed to sync products");
  }
  return data;
}

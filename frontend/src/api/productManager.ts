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

export interface FilterPreferences {
  search: string;
  column_visibility: Record<string, boolean>;
  page_size: number;
  sort_by?: string;
  sort_dir?: "asc" | "desc";
}

export interface SaveFilterPreferencesRequest {
  platform: string;
  page: string;
  preferences: FilterPreferences;
}

export async function getFilterPreferences(
  platform: string,
  page: string,
): Promise<FilterPreferences | null> {
  const response = await apiClient.get<FilterPreferences>(
    "/filter-preferences",
    {
      params: { platform, page },
    },
  );

  if (!response.success) {
    return null;
  }

  return response.data || null;
}

export async function saveFilterPreferences(
  request: SaveFilterPreferencesRequest,
): Promise<void> {
  const response = await apiClient.post("/filter-preferences", request);

  if (!response.success) {
    throw new Error(response.error || "Failed to save filter preferences");
  }
}

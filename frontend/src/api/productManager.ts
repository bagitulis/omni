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

export interface SyncSelectedResult {
  synced: number;
  failed: number;
  details: string[];
}

/**
 * Sync selected products from marketplace APIs to refresh price/stock
 */
export async function syncSelectedProducts(
  productIds: number[],
): Promise<SyncSelectedResult> {
  const response = await apiClient.post<SyncSelectedResult>(
    "/products/master/sync-selected",
    { product_ids: productIds },
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to sync selected products");
  }
  return response.data!;
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
  const response = await apiClient.get<{
    platform: string;
    tab: string;
    visible_columns: string[];
    column_filters: Record<string, unknown>;
    search_query: string;
    locked_columns: string[];
  }>("/filter-preferences", {
    params: { platform, page },
  });

  if (!response.success || !response.data) {
    return null;
  }

  // Normalize backend response shape to frontend expectations
  const backendData = response.data;
  const normalized: FilterPreferences = {
    search: backendData.search_query || "",
    column_visibility:
      backendData.visible_columns?.reduce(
        (acc, col) => {
          acc[col] = true;
          return acc;
        },
        {} as Record<string, boolean>,
      ) || {},
    page_size: 50, // Default page size
    sort_by: undefined,
    sort_dir: undefined,
  };

  return normalized;
}

export async function saveFilterPreferences(
  request: SaveFilterPreferencesRequest,
): Promise<void> {
  // Transform frontend shape to backend expectations
  const visibleColumns = Object.keys(
    request.preferences.column_visibility,
  ).filter((key) => request.preferences.column_visibility[key]);

  const backendPayload = {
    platform: request.platform,
    page: request.page,
    visible_columns: visibleColumns,
    column_filters: {},
    search_query: request.preferences.search || "",
    locked_columns: [],
  };

  const response = await apiClient.post("/filter-preferences", backendPayload);

  if (!response.success) {
    throw new Error(response.error || "Failed to save filter preferences");
  }
}

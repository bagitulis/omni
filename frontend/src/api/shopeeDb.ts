import { API_TIMEOUT } from "@/lib/constants";
import apiClient from "./client";
import type { ApiResponse } from "./client";

export interface ShopeeDbProduct {
  item_id: number;
  item_name?: string;
  item_sku?: string;
  current_price?: number;
  price?: number;
  original_price?: number;
  seller_stock?: number;
  stock?: number;
  item_status?: string;
  has_model?: boolean;
  category_id?: number;
  brand?: string;
  description?: string;
  image_urls?: string[] | string;
  updated_at?: string;
}

export interface ShopeeDbStats {
  total_products?: number;
  total_models?: number;
  total_variations?: number;
  [key: string]: unknown;
}

export interface ShopeeSyncLog {
  id?: number;
  status?: string;
  message?: string;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

export interface ShopeeBatchOperationResult {
  message?: string;
  status?: string;
  processed?: number;
  queued?: number;
  [key: string]: unknown;
}

export interface DisplayShopeeProduct {
  item_id: number;
  name: string;
  sku: string;
  price: number;
  original_price: number;
  stock: number;
  status: string;
  has_models: boolean;
  category_id: number | null;
  brand: string;
  description: string;
  images: string[];
  updated_at: string;
}

function ensureSuccess<T>(response: ApiResponse<T>, fallback_error: string): T {
  if (!response.success) {
    throw new Error(response.error || fallback_error);
  }
  return (response.data ?? null) as T;
}

async function getList(
  path: string,
  fallback_error: string,
  params?: Record<string, unknown>,
): Promise<ShopeeDbProduct[]> {
  const response = await apiClient.get<ShopeeDbProduct[]>(path, { params });
  return ensureSuccess(response, fallback_error) ?? [];
}

export async function getProductList(
  limit: number = 100,
  offset: number = 0,
  status?: string,
): Promise<ShopeeDbProduct[]> {
  const params: Record<string, unknown> = { limit, offset };
  if (status) params.status = status;
  return getList(
    "/shopee/db/products/list",
    "Failed to fetch product list",
    params,
  );
}

export async function getProductBaseList(
  limit: number = 100,
  offset: number = 0,
): Promise<ShopeeDbProduct[]> {
  return getList(
    "/shopee/db/products/base",
    "Failed to fetch product base list",
    { limit, offset },
  );
}

export async function getProductModelList(
  limit: number = 100,
  offset: number = 0,
): Promise<ShopeeDbProduct[]> {
  return getList(
    "/shopee/db/products/model",
    "Failed to fetch product model list",
    { limit, offset },
  );
}

export async function getProductBase(
  item_id: number | string,
): Promise<ShopeeDbProduct | null> {
  const response = await apiClient.get<ShopeeDbProduct>(
    `/shopee/db/products/base/${item_id}`,
  );
  return ensureSuccess(response, "Failed to fetch product base") ?? null;
}

export async function getProductModels(
  item_id: number | string,
): Promise<ShopeeDbProduct[]> {
  return getList(
    `/shopee/db/products/models/${item_id}`,
    "Failed to fetch product models",
  );
}

export async function getProductVariations(
  item_id: number | string,
): Promise<ShopeeDbProduct[]> {
  return getList(
    `/shopee/db/products/variations/${item_id}`,
    "Failed to fetch product variations",
  );
}

export async function getProductFull(
  item_id: number | string,
): Promise<ShopeeDbProduct | null> {
  const response = await apiClient.get<ShopeeDbProduct>(
    `/shopee/db/products/full/${item_id}`,
  );
  return ensureSuccess(response, "Failed to fetch product details") ?? null;
}

export async function searchBySku(sku: string): Promise<ShopeeDbProduct[]> {
  const normalized_sku = sku.trim();
  if (!normalized_sku) throw new Error("SKU cannot be empty");
  return getList("/shopee/db/products/search", "Failed to search products", {
    sku: normalized_sku,
  });
}

export async function getProductsByStatus(
  status: string,
): Promise<ShopeeDbProduct[]> {
  if (!status) throw new Error("Status is required");
  return getList(
    `/shopee/db/products/status/${status}`,
    "Failed to fetch products by status",
  );
}

export async function getDatabaseStats(): Promise<ShopeeDbStats> {
  const response = await apiClient.get<ShopeeDbStats>("/shopee/db/stats");
  return ensureSuccess(response, "Failed to fetch database stats") ?? {};
}

export async function getUnprocessedItems(): Promise<ShopeeDbProduct[]> {
  return getList(
    "/shopee/db/sync/unprocessed",
    "Failed to fetch unprocessed items",
  );
}

export async function getItemsWithoutModels(): Promise<ShopeeDbProduct[]> {
  return getList(
    "/shopee/db/sync/no-models",
    "Failed to fetch items without models",
  );
}

export async function getSyncLogs(
  limit: number = 100,
): Promise<ShopeeSyncLog[]> {
  const response = await apiClient.get<ShopeeSyncLog[]>(
    "/shopee/db/sync/logs",
    {
      params: { limit },
    },
  );
  return ensureSuccess(response, "Failed to fetch sync logs") ?? [];
}

export async function syncShopeeProducts(): Promise<ShopeeBatchOperationResult> {
  const response = await apiClient.post<ShopeeBatchOperationResult>(
    "/shopee/sync/products",
    undefined,
    { timeout: API_TIMEOUT.EXTRA_LONG },
  );
  return ensureSuccess(response, "Failed to sync Shopee products") ?? {};
}

export function formatProductForDisplay(
  product: ShopeeDbProduct,
): DisplayShopeeProduct {
  const images = Array.isArray(product.image_urls)
    ? product.image_urls
    : typeof product.image_urls === "string" && product.image_urls.length > 0
      ? product.image_urls.split(",")
      : [];
  return {
    item_id: product.item_id,
    name: product.item_name || "N/A",
    sku: product.item_sku || "N/A",
    price: product.current_price ?? product.price ?? 0,
    original_price: product.original_price ?? 0,
    stock: product.seller_stock ?? product.stock ?? 0,
    status: product.item_status || "UNKNOWN",
    has_models: product.has_model ?? false,
    category_id: product.category_id ?? null,
    brand: product.brand || "N/A",
    description: product.description || "",
    images,
    updated_at: product.updated_at || new Date().toISOString(),
  };
}

export function formatPrice(price: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
  }).format(price || 0);
}

export function formatNumber(num: number): string {
  return new Intl.NumberFormat("id-ID").format(num || 0);
}

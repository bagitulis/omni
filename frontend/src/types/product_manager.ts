export type ProductManagerPlatform = "shopee" | "lazada" | "tiktok";

export type DbProductRow = Record<string, unknown>;

export interface DbProductsResponse {
  success: boolean;
  products: DbProductRow[];
  total: number;
  offset: number;
  limit: number;
  count: number;
}

export interface SyncProductsResponse {
  success: boolean;
  data?: {
    message?: string;
    synced?: number;
  };
  error?: string;
}

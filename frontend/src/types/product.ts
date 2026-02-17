/**
 * Product Types
 * API types use snake_case to match backend JSON response
 */

import type { Platform } from "./shared";

export interface MasterProduct {
  id: number;
  tenant_id: string;
  title: string;
  description: string;
  images: string[];
  status: "active" | "inactive" | "draft";
  created_at: string;
  updated_at: string;
  skus?: MasterProductSku[];
}

export interface MasterProductSku {
  id: number;
  tenant_id: string;
  master_product_id: number;
  seller_sku: string;
  variant_name: string;
  variant_data: Record<string, unknown>;
  price: number;
  stock: number;
  created_at: string;
  updated_at: string;
  platform_links?: MasterProductPlatformLink[];
}

export interface MasterProductPlatformLink {
  id: number;
  master_product_id: number;
  master_sku_id?: number;
  platform: Platform;
  platform_product_id?: string;
  platform_sku_id?: string;
  platform_item_id?: number;
  sync_status: "synced" | "pending" | "failed" | "not_synced";
  last_synced_at?: string;
  error_message?: string;
}

export interface CreateMasterProductInput {
  title: string;
  description: string;
  images?: string[];
  status?: "active" | "inactive" | "draft";
  skus?: CreateSkuInput[];
}

export interface CreateSkuInput {
  seller_sku: string;
  variant_name?: string;
  variant_data?: Record<string, unknown>;
  price: number;
  stock: number;
}

export interface UpdateMasterProductInput {
  title?: string;
  description?: string;
  images?: string[];
  status?: "active" | "inactive" | "draft";
  skus?: UpdateSkuInput[];
}

export interface UpdateSkuInput {
  id?: number;
  seller_sku?: string;
  price?: number;
  stock?: number;
}

// Legacy Product interface for table compatibility (can be refactored later)
export interface Product extends MasterProduct {
  // Mapped fields for table compatibility
  item_id: string;
  item_name: string;
  item_sku: string;
  price: number;
  stock: number;
  platform: string;
  image_url: string;
}

export interface ProductDetail extends Product {
  category?: string;
  attributes?: Record<string, string>;
}

export interface ProductListResponse {
  success: boolean;
  data: MasterProduct[];
  meta?: {
    total: number;
    page: number;
    page_size: number;
  };
}

export interface SingleProductResponse {
  success: boolean;
  data: MasterProduct;
}

export interface ProductListFilter {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
  platform?: string;
}

/** Matches backend InventoryListItem - dynamic JSONB data from Google Sheets */
export interface InventoryRecord {
  id: string;
  key_value: string;
  key_column_name: string;
  data: Record<string, string | number | null>;
  created_at: string;
  updated_at: string;
}

/** Result from getInventory() - includes pagination metadata */
export interface InventoryListResult {
  records: InventoryRecord[];
  total: number;
  offset: number;
  limit: number;
}

export interface ImportRow {
  row_number: number;
  item_name: string;
  item_sku: string;
  price: number;
  stock: number;
  valid: boolean;
  errors: string[];
}

export interface ImportPreviewData {
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
  rows: ImportRow[];
}

export interface ImportResponse {
  success: boolean;
  imported_count: number;
  failed_count: number;
  message: string;
}

export interface AutoMapResult {
  success: boolean;
  mapped_count: number;
  skipped_count: number;
  mappings: {
    master_sku: string;
    platform: string;
    platform_product_id: string;
    platform_sku_id: string;
  }[];
  errors?: string[];
}

export interface MappingStatus {
  total_master_skus: number;
  mapped_skus: number;
  unmapped_skus: number;
  platforms: {
    platform: string;
    mapped_count: number;
  }[];
}

export interface LinkSkuData {
  master_sku_id: number;
  platform: string;
  platform_product_id: string;
  platform_sku_id: string;
}

export interface UnlinkSkuData {
  master_sku_id: number;
  platform: string;
}

export interface BatchSkuUpdateItem {
  id: number;
  seller_sku?: string;
  price?: number;
  stock?: number;
}

export interface BatchSkuUpdateResult {
  updated: number;
  failed: number;
  errors?: string[];
}

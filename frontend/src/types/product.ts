/**
 * Product Types
 * API types use snake_case to match backend JSON response
 */

/**
 * Backend product format - raw response from API
 */
export interface BackendProduct {
  id: number;
  tenant_id: string;
  title: string;
  description: string;
  images: string[];
  status: "active" | "inactive" | "draft";
  created_at: string;
  updated_at: string;
}

/**
 * Frontend product format - with mapped fields for UI components
 * transformProduct() always sets these fields, so they are required
 */
export interface Product {
  id: number;
  tenant_id: string;
  title: string;
  description: string;
  images: string[];
  status: "active" | "inactive" | "draft";
  created_at: string;
  updated_at: string;
  // Mapped fields for table compatibility - REQUIRED (set by transformProduct)
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

/**
 * Frontend-friendly format for product lists
 */
export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
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

/**
 * Product Types
 * API types use snake_case to match backend JSON response
 */

export interface Product {
  item_id: string;
  item_name: string;
  item_sku: string;
  price: number;
  stock: number;
  platform: string;
  status: "active" | "inactive" | "draft";
  image_url?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ProductDetail extends Product {
  description?: string;
  category?: string;
  images?: string[];
  attributes?: Record<string, string>;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
}

export interface InventoryItem {
  item_id: string;
  item_sku: string;
  item_name: string;
  current_stock: number;
  reserved_stock: number;
  available_stock: number;
  warehouse?: string;
  last_updated: string;
}

export interface InventoryListResponse {
  items: InventoryItem[];
  total: number;
  page: number;
  page_size: number;
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

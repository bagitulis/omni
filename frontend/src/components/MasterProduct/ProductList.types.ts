export interface PlatformLink {
  id: number;
  master_product_id: number;
  master_sku_id?: number;
  platform: string;
  platform_product_id?: string;
  platform_sku_id?: string;
  sync_status: string;
}

export interface ProductSku {
  id: number;
  master_product_id: number;
  seller_sku: string;
  variant_name?: string;
  variant_data?: Record<string, any>;
  price: number;
  stock: number;
  platform_links?: PlatformLink[];
}

export interface MasterProduct {
  id: number;
  tenant_id: string;
  title: string;
  description?: string;
  images?: string[];
  status: string;
  skus?: ProductSku[];
  created_at: string;
  updated_at: string;
}

export interface BatchUpdateInput {
  sku_ids: number[];
  price?: number;
  stock?: number;
}

export interface ProductSku {
  key: string;
  id?: number;
  seller_sku: string;
  variant_name: string;
  variant_data?: Record<string, unknown>;
  stock: number;
  price: number;
}

export interface ProductPlatform {
  platform: string;
  status: string;
  last_sync: string;
}

export interface ProductData {
  id: number;
  title: string;
  description: string;
  images: string[];
  status: string;
  skus?: ProductSku[];
  platforms?: ProductPlatform[];
}

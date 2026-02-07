export interface ProductSku {
  key: string;
  seller_sku: string;
  variant_name: string;
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

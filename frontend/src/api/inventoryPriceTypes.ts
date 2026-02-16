export interface PriceUpdateItem {
  sku: string;
  price: number;
  platforms?: string[];
}

export interface PlatformPriceResult {
  success: boolean;
  item_id?: string;
  model_id?: string;
  sku_id?: string;
  product_id?: string;
  error?: string;
}

export interface PriceUpdateResult {
  sku: string;
  success: boolean;
  platforms: {
    shopee?: PlatformPriceResult;
    lazada?: PlatformPriceResult;
    tiktok?: PlatformPriceResult;
  };
  errors: string[];
  skipped: string[];
}

export interface BatchPriceUpdateResult {
  total: number;
  successful: number;
  failed: number;
  skipped: number;
  results: PriceUpdateResult[];
}

export interface ShopeeProduct {
  id: number;
  item_id: number;
  name: string;
  description?: string;
  images?: string[];
  price?: number;
  price_info?: {
    current_price?: number;
  };
  models?: Array<{
    model_id: number;
    model_sku: string;
    tier_index?: number[];
  }>;
}

// Re-export or define ImportPreviewResponse if needed,
// but it's likely better to keep using the service import if it's shared.
// For now, we will just use this for the local type defined in the original file.

export interface ProductSummary {
  platform: string;
  item_id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
  images: string[];
}

export interface Difference {
  field: string;
  source_value: string;
  target_value: string;
}

export interface AdjustmentInfo {
  title_will_truncate: boolean;
  desc_will_truncate: boolean;
  original_title?: string;
  adjusted_title?: string;
  title_limit: number;
  desc_limit: number;
}

export interface PreviewData {
  has_conflict: boolean;
  source_product?: ProductSummary;
  target_product?: ProductSummary;
  differences?: Difference[];
  adjustments?: AdjustmentInfo;
}

export interface AdsPagination {
  total: number;
  limit: number;
  offset: number;
}

export interface DateRange {
  start_date: string;
  end_date: string;
}

// Shopee Ads Analytics Types

export interface ShopeeAdsDashboardTopProduct {
  product_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
}

export interface ShopeeAdsBiddingModeStat {
  bidding_mode: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
  cost_per_order: number;
}

export interface ShopeeAdsDashboardData {
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  avg_roas: number;
  total_impressions: number;
  total_clicks: number;
  avg_ctr: number;
  avg_conversion_rate: number;
  top_products: ShopeeAdsDashboardTopProduct[];
  bidding_mode_stats: ShopeeAdsBiddingModeStat[];
}

export interface ShopeeAdsProductData {
  id: number;
  product_id: string;
  product_name: string;
  bidding_mode: string;
  cost: number;
  revenue: number;
  direct_revenue: number;
  conversions: number;
  roas: number;
  direct_roas: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversion_rate: number;
  period_start: string;
  period_end: string;
  period_label: string;
}

export interface ShopeeAdsUploadBatch {
  id: number;
  filename: string;
  period_label: string;
  record_count: number;
  uploaded_at: string;
}

// TikTok Ads Analytics Types

export interface TiktokAdsDashboardTopProduct {
  product_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number; // Note: Handler calls it 'roas' in struct but 'avg_roi' in response JSON, let's follow JSON
}

export interface TiktokAdsCreativeTypeStat {
  creative_type: string;
  cost: number;
  revenue: number;
  orders: number;
  roas: number;
  cost_per_order: number;
}

export interface TiktokAdsDashboardData {
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  avg_roi: number;
  total_impressions: number;
  total_clicks: number;
  avg_ctr: number;
  avg_conversion_rate: number;
  top_products: TiktokAdsDashboardTopProduct[];
  creative_type_comparison: TiktokAdsCreativeTypeStat[];
}

export interface TiktokAdsCreativeData {
  id: number;
  campaign_id: string;
  campaign_name: string;
  product_id: string;
  creative_type: string;
  video_title: string;
  cost: number;
  orders_sku: number;
  gross_revenue: number;
  roi: number;
  impressions: number;
  clicks: number;
  ctr: number;
  conversion_rate: number;
  period_start: string;
  period_end: string;
}

export interface TiktokAdsUploadBatch {
  id: number;
  file_name: string;
  period_start: string;
  period_end: string;
  total_rows: number;
  inserted_rows: number;
  skipped_rows: number;
  status: string;
  created_at: string;
}

// Ads Management Types (from ads_handler_shopee.go and ads_handler_tiktok.go)

export interface ShopeeAdsSummary {
  period: string;
  total_spend: number;
  total_gmv: number;
  total_orders: number;
  roas: number;
  cir: number;
}

export interface ShopeeAdsTrend {
  date: string;
  spend: number;
  gmv: number;
  orders: number;
  roas: number;
}

export interface ShopeeAdsProductPerformance {
  item_id: string;
  item_name: string;
  spend: number;
  gmv: number;
  orders: number;
  roas: number;
  clicks: number;
  impressions: number;
  ctr: number;
  cr: number;
}

export interface TiktokAdsSummary {
  period: string;
  total_cost: number;
  total_revenue: number; // Gross Revenue
  total_conversions: number; // Orders SKU
  roi: number;
  cpa: number;
}

export interface TiktokAdsTrend {
  date: string;
  cost: number;
  revenue: number;
  conversions: number;
  roi: number;
}

export interface TiktokAdsProductPerformance {
  sku_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  conversions: number;
  roi: number;
  clicks: number;
  impressions: number;
  ctr: number;
  cvr: number;
}

export interface TiktokAdsPrediction {
  date: string;
  predicted_revenue: number;
  predicted_cost: number;
  confidence_lower: number;
  confidence_upper: number;
}

// Reports Types

export interface AdsReport {
  generated_at: string;
  report_url: string;
  date_range: {
    start: string;
    end: string;
  };
  summary: {
    total_spend: number;
    total_sales: number;
    roas: number;
  };
}

export interface AdsReportLatest {
  last_updated: string;
  status: "success" | "pending" | "failed";
  data: AdsReport;
}

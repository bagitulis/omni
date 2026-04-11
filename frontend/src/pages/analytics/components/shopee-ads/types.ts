export interface AdsData {
  product_id: string;
  product_name: string;
  bidding_mode?: string;
  cost: number;
  revenue: number;
  clicks: number;
  impressions: number;
  ctr: number;
  cpc: number;
  roas: number;
  acos?: number;
  conversions: number;
  date: string;
  period_label?: string;
}

export const SHOPEE_ORANGE = "#ee4d2d";

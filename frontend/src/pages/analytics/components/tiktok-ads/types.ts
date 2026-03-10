export interface TikTokAdsData {
  creative_id: string;
  creative_name: string;
  campaign_name?: string;
  product_id?: string;
  creative_type?: string;
  cost: number;
  revenue: number;
  views: number;
  clicks: number;
  ctr: number;
  cpc: number;
  roi: number;
  conversions: number;
  video_plays: number;
  engagement_rate: number;
  date: string;
}

export const TIKTOK_BLACK = "#010101";
export const TIKTOK_ACCENT = "#fe2c55";

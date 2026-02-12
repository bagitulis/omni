import apiClient from "./client";

// Types matching backend response
export interface PortfolioHealth {
  tenant_id: string;
  total_products: number;
  total_cost: number;
  total_revenue: number;
  total_profit: number;
  overall_roas: number;
  health_score: number;
  health_label: string;
  star_count: number;
  growth_count: number;
  stable_count: number;
  watch_count: number;
  problem_count: number;
  scale_up_count: number;
  maintain_count: number;
  reduce_count: number;
  stop_count: number;
  fatigue_warnings: number;
  churn_risks: number;
  active_alerts: number;
  last_updated: string;
}

export interface MLProduct {
  product_id: string;
  product_name: string;
  sku: string;
  unified_score: number;
  category: string;
  action: string;
  recommendation: string;
  total_cost: number;
  total_revenue: number;
  total_profit: number;
  roas: number;
  ctr: number;
  has_fatigue_warning: boolean;
  has_churn_risk: boolean;
  fatigue_status: string;
  churn_risk_score: number;
  last_updated: string;
  // Detailed scores
  roas_score: number;
  trend_score: number;
  volatility_score: number;
  momentum_score: number;
  // Action details
  action_label: string;
  budget_change_pct: number;
  confidence_level: string;
  success_probability: number;
  // Trend details
  trend_direction: string;
  trend_strength: number;
}

export interface MLAlert {
  id: string;
  tenant_id: string;
  product_id: string;
  product_name: string;
  alert_type: string;
  severity: string;
  message: string;
  created_at: string;
  status: string;
}

export interface ScoreDistribution {
  category: string;
  count: number;
  percentage: number;
}

export interface PaginationMeta {
  total: number;
  limit: number;
  has_more: boolean;
  next_cursor: string | null;
}

export interface AlertsMeta {
  total_active: number;
  high_priority: number;
  medium_priority: number;
  low_priority: number;
}

// Query params
export interface MLProductsParams {
  platform?: string;
  limit?: number;
  cursor?: string;
  sort_by?: string;
  sort_dir?: string;
  category?: string;
  action?: string;
}

// API functions
export async function getPortfolioHealth(
  platform: string = "tiktok",
): Promise<PortfolioHealth> {
  const response = await apiClient.get<PortfolioHealth>(
    `/analytics/ml/portfolio-health`,
    { params: { platform } },
  );
  if (!response.data) {
    throw new Error("No data received from server");
  }
  return response.data;
}

export async function getMLProducts(
  params: MLProductsParams = {},
): Promise<{ products: MLProduct[]; meta: PaginationMeta }> {
  const response = await apiClient.client.get(`/analytics/ml/products`, {
    params,
  });
  const data = response.data;
  if (!data.success) {
    throw new Error("No data received from server");
  }
  return {
    products: data.data || [],
    meta: data.meta,
  };
}

export async function getMLAlerts(): Promise<{
  alerts: MLAlert[];
  meta: AlertsMeta;
}> {
  const response = await apiClient.client.get(`/analytics/ml/alerts`);
  const data = response.data;
  if (!data.success) {
    throw new Error("No data received from server");
  }
  return {
    alerts: data.data || [],
    meta: data.meta,
  };
}

export async function getScoreDistribution(): Promise<ScoreDistribution[]> {
  const response = await apiClient.get<ScoreDistribution[]>(
    `/analytics/ml/distribution`,
  );
  if (!response.data) {
    throw new Error("No data received from server");
  }
  return response.data || [];
}

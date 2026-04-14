import api from "./client";

export interface AnalyticsKPI {
  total_products: number;
  total_revenue: number;
  total_profit: number;
  avg_roas: number;
  total_orders: number;
}

export interface PlatformSummary {
  total_revenue: number;
  total_cost: number;
  avg_roas: number;
  total_orders: number;
}

export interface AnalyticsSummary {
  combined: PlatformSummary;
  shopee: PlatformSummary;
  tiktok: PlatformSummary;
}

export interface ActionCounts {
  scale_up: number;
  maintain: number;
  reduce: number;
  stop: number;
}

export interface UnifiedAnalyticsResponse {
  summary: AnalyticsSummary;
  kpi: AnalyticsKPI;
  action_counts: ActionCounts;
}

export interface ClassifiedProduct {
  product_id: string;
  product_name: string;
  total_cost: number;
  total_revenue: number;
  total_orders: number;
  roas: number;
  action: "SCALE_UP" | "MAINTAIN" | "REDUCE" | "STOP";
  action_label: string;
  recommendation: string;
  source_platform?: "tiktok" | "shopee";
}

export interface ClassifiedProductsResponse {
  scale_up: ClassifiedProduct[];
  maintain: ClassifiedProduct[];
  reduce: ClassifiedProduct[];
  stop: ClassifiedProduct[];
}

export interface ProductFromAds {
  product_id: string;
  product_name: string;
  avg_roas: number;
  source: "tiktok" | "shopee";
  current_daily_spend: number;
  ml_category?: string;
}

export interface SimulationRequest {
  product_id: string;
  target_roas: number;
  budget_per_day: number;
  period_days: number;
}

export interface FunnelBreakdown {
  projected_clicks: number;
  projected_orders: number;
  projected_revenue: number;
  projected_cpc: number;
  projected_cpo: number;
  projected_cvr: number;
  projected_aov: number;
  projected_impressions: number;
  projected_ctr: number;
}

export interface CalendarEvent {
  date: string;
  type: string;
  description: string;
  multiplier: number;
}

export interface SimulationResult {
  // Core metrics
  feasibility: "ACHIEVABLE" | "DIFFICULT" | "NOT_ACHIEVABLE";
  confidence_percent: number;
  current_roas: number;
  projected_roas: number;
  trend_prediction: "UP" | "DOWN" | "STAGNANT";
  optimal_budget: number;
  recommendation: string;
  simulation_mode: "FUNNEL" | "YIELD";
  ml_category: string;
  current_daily_spend: number;
  alternatives: {
    target_roas: number;
    required_budget: number;
    expected_roas: number;
  }[];

  // Period projection
  total_budget?: number;
  total_projected_revenue?: number;
  total_projected_orders?: number;

  // Funnel breakdown per day
  funnel?: FunnelBreakdown;

  // Calendar context
  calendar_multiplier?: number;
  calendar_events?: CalendarEvent[];

  // Scale-up analysis
  max_safe_budget?: number;
  scale_factor?: number;
  optimal_label?: string;
}



export async function getUnifiedAnalytics(): Promise<UnifiedAnalyticsResponse> {
  const response = await api.get<UnifiedAnalyticsResponse>(
    "/analytics/unified/summary",
  );
  if (!response.success || !response.data)
    throw new Error(response.error || "Failed to fetch unified analytics");
  return response.data;
}

export async function getClassifiedProducts(): Promise<ClassifiedProductsResponse> {
  const response = await api.get<ClassifiedProductsResponse>(
    "/analytics/products/classified",
  );
  if (!response.success || !response.data)
    throw new Error(response.error || "Failed to fetch classified products");
  return response.data;
}

// --- Budget Simulator ---

export async function getProductsFromAds(): Promise<ProductFromAds[]> {
  const response = await api.get<ProductFromAds[]>(
    "/analytics/products/from-ads",
  );
  if (!response.success || !response.data)
    throw new Error(response.error || "Failed to fetch products from ads");
  return response.data;
}

export async function runSimulation(
  request: SimulationRequest,
): Promise<SimulationResult> {
  const response = await api.post<SimulationResult>(
    "/analytics/simulation/calculate",
    request,
  );
  if (!response.success || !response.data)
    throw new Error(response.error || "Failed to run simulation");
  return response.data;
}



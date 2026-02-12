import api from "./client";
import { logger } from "@/lib/logger";

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
}

export interface SimulationRequest {
  product_id: string;
  target_roas: number;
  budget_per_day: number;
  period_days: number;
}

export interface SimulationResult {
  feasibility: "ACHIEVABLE" | "DIFFICULT" | "NOT_ACHIEVABLE";
  confidence_percent: number;
  current_roas: number;
  projected_roas: number;
  trend_prediction: "UP" | "DOWN" | "STAGNANT";
  optimal_budget: number;
  recommendation: string;
  alternatives: {
    target_roas?: number;
    required_budget?: number;
    budget?: number;
    expected_roas?: number;
  }[];
}

export interface MLReport {
  id: number;
  tenant_id: string;
  platform: "shopee" | "tiktok";
  report_type: "full" | "executive" | "quick";
  period_start: string;
  period_end: string;
  period_label: string;
  file_path: string;
  file_name: string;
  file_size: number;
  status: "pending" | "completed" | "failed";
  error_msg?: string;
  created_at: string;
  updated_at: string;
}

export interface MLJob {
  id: number;
  tenant_id: string;
  job_type: string;
  platform: "shopee" | "tiktok";
  status: "pending" | "running" | "completed" | "failed";
  progress: number;
  result_id?: number;
  error_msg?: string;
  created_at: string;
}

export interface GenerateReportRequest {
  platform: "shopee" | "tiktok";
  report_type?: "full" | "executive" | "quick";
  period_label?: string;
}

// --- Unified Analytics ---

export async function getUnifiedAnalytics(): Promise<UnifiedAnalyticsResponse> {
  const response = await api.get("/analytics/unified/summary");
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data;
}

export async function getClassifiedProducts(): Promise<ClassifiedProductsResponse> {
  const response = await api.get("/analytics/products/classified");
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data;
}

// --- Budget Simulator ---

export async function getProductsFromAds(): Promise<ProductFromAds[]> {
  const response = await api.get("/analytics/products/from-ads");
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data;
}

export async function runSimulation(
  request: SimulationRequest,
): Promise<SimulationResult> {
  const response = await api.post("/analytics/simulation/calculate", request);
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data;
}

// --- AI Reports ---

export async function generateReport(
  request: GenerateReportRequest,
): Promise<MLJob> {
  const response = await api.post("/analytics/ml/reports/generate", request);
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data;
}

export async function getReports(
  platform: "shopee" | "tiktok",
  page: number = 1,
  limit: number = 20,
): Promise<{ reports: MLReport[]; total: number }> {
  try {
    const response = await api.get(`/analytics/ml/reports/${platform}/list`, {
      params: { page, limit },
    });
    if (!response.data.success) throw new Error(response.data.error);
    return response.data.data;
  } catch (error: any) {
    // Handle 404 gracefully (e.g., if backend feature is not enabled or reachable)
    if (error.response && error.response.status === 404) {
      logger.warn("ML Reports endpoint not found (404), returning empty list");
      return { reports: [], total: 0 };
    }
    throw error;
  }
}

export async function getReportHTML(
  platform: "shopee" | "tiktok",
  filename: string,
): Promise<string> {
  const response = await api.get(
    `/analytics/ml/reports/${platform}/${filename}`,
  );
  if (!response.data.success) throw new Error(response.data.error);
  return response.data.data.html;
}

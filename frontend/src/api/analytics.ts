import apiClient from "./client";
import type {
  Platform,
  AnalyticsSettings,
  SyncStatus,
  ReconciliationResult,
  TiktokReconciliationResult,
  ShopeeShippingFeeResult,
  TiktokShippingFeeResult,
  JobProgress,
  KPIData,
  UnifiedSummary,
} from "@/types/analytics";

export async function getUnifiedKPI(): Promise<KPIData> {
  const response = await apiClient.get<KPIData>("/analytics/unified/kpi");
  if (!response.data) throw new Error("Failed to fetch unified KPI");
  return response.data;
}

export async function getUnifiedSummary(): Promise<UnifiedSummary> {
  const response = await apiClient.get<UnifiedSummary>(
    "/analytics/unified/summary",
  );
  if (!response.data) throw new Error("Failed to fetch unified summary");
  return response.data;
}

export async function getSettings(
  platform: Platform,
): Promise<AnalyticsSettings> {
  const response = await apiClient.get<AnalyticsSettings>(
    `/analytics/${platform}/settings`,
  );
  if (!response.data) throw new Error("Failed to fetch analytics settings");
  return response.data;
}

export async function saveSettings(
  platform: Platform,
  settings: AnalyticsSettings,
): Promise<void> {
  const response = await apiClient.post(
    `/analytics/${platform}/settings`,
    settings,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to save analytics settings");
  }
}

export async function getSyncStatus(
  platform: Platform,
  month: number,
  year: number,
): Promise<SyncStatus> {
  const response = await apiClient.get<SyncStatus>(
    `/analytics/${platform}/sync-status`,
    { params: { month, year } },
  );
  if (!response.data) throw new Error("Failed to fetch sync status");
  return response.data;
}

export async function syncEscrow(
  platform: Platform,
  month: number,
  year: number,
  forceResync: boolean,
): Promise<string> {
  const response = await apiClient.client.post(`/analytics/${platform}/sync`, {
    month,
    year,
    force_resync: forceResync,
  });
  if (!response.data.success) {
    throw new Error(response.data.error || "Failed to start sync");
  }
  return response.data.data.job_id;
}

export async function deleteSyncData(
  platform: Platform,
  month: number,
  year: number,
): Promise<void> {
  const response = await apiClient.delete(`/analytics/${platform}/sync`, {
    params: { month, year },
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to delete sync data");
  }
}

export async function getReconciliation(
  platform: Platform,
  month: number,
  year: number,
): Promise<ReconciliationResult | TiktokReconciliationResult> {
  const response = await apiClient.get<
    ReconciliationResult | TiktokReconciliationResult
  >(`/analytics/${platform}/reconciliation`, { params: { month, year } });
  if (!response.data) throw new Error("Failed to fetch reconciliation data");
  return response.data;
}

export async function getShippingFee(
  platform: Platform,
  month: number,
  year: number,
): Promise<ShopeeShippingFeeResult | TiktokShippingFeeResult> {
  const response = await apiClient.get<
    ShopeeShippingFeeResult | TiktokShippingFeeResult
  >(`/analytics/${platform}/shipping-fee`, { params: { month, year } });
  if (!response.data) throw new Error("Failed to fetch shipping fee data");
  return response.data;
}

export async function getJobStatus(jobId: string): Promise<JobProgress> {
  const response = await apiClient.client.get(`/jobs/${jobId}`);
  if (!response.data.success) {
    throw new Error(response.data.error || "Failed to fetch job status");
  }
  return response.data.job;
}

/**
 * Analytics & Report API Functions
 *
 * All 16 endpoints for Shopee and TikTok analytics.
 * Backend routes: /api/analytics/shopee/* and /api/analytics/tiktok/*
 */
import apiClient from "./client";
import type {
  ReconciliationResult,
  ReportSettings,
  ShopeeShippingFeeResult,
  SyncRequest,
  SyncResult,
  SyncStatus,
  TiktokShippingFeeResult,
} from "../types/analytics";

// ──────────────────────────────────────────────
//  Shopee Analytics (8 functions)
// ──────────────────────────────────────────────

/**
 * Get Shopee report settings
 * GET /api/analytics/shopee/settings
 */
export async function getShopeeSettings(): Promise<ReportSettings> {
  const response = await apiClient.get<ReportSettings>(
    "/analytics/shopee/settings"
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch Shopee settings");
  }
  return response.data!;
}

/**
 * Save Shopee report settings
 * POST /api/analytics/shopee/settings
 */
export async function saveShopeeSettings(
  settings: Partial<ReportSettings>
): Promise<ReportSettings> {
  const response = await apiClient.post<ReportSettings>(
    "/analytics/shopee/settings",
    settings
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to save Shopee settings");
  }
  return response.data!;
}

/**
 * Get Shopee sync status for a given month/year
 * GET /api/analytics/shopee/sync-status?month=&year=
 */
export async function getShopeeSyncStatus(
  month: number,
  year: number
): Promise<SyncStatus> {
  const response = await apiClient.get<SyncStatus>(
    "/analytics/shopee/sync-status",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch Shopee sync status");
  }
  return response.data!;
}

/**
 * Trigger Shopee sync
 * POST /api/analytics/shopee/sync
 */
export async function triggerShopeeSync(
  req: SyncRequest
): Promise<SyncResult> {
  const response = await apiClient.post<SyncResult>(
    "/analytics/shopee/sync",
    req
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to trigger Shopee sync");
  }
  return response.data!;
}

/**
 * Delete Shopee sync data for a given month/year
 * DELETE /api/analytics/shopee/sync?month=&year=
 */
export async function deleteShopeeSync(
  month: number,
  year: number
): Promise<void> {
  const response = await apiClient.delete("/analytics/shopee/sync", {
    params: { month, year },
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to delete Shopee sync data");
  }
}

/**
 * Get Shopee reconciliation data
 * GET /api/analytics/shopee/reconciliation?month=&year=
 */
export async function getShopeeReconciliation(
  month: number,
  year: number
): Promise<ReconciliationResult> {
  const response = await apiClient.get<ReconciliationResult>(
    "/analytics/shopee/reconciliation",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(
      response.error || "Failed to fetch Shopee reconciliation"
    );
  }
  return response.data!;
}

/**
 * Get Shopee shipping fee analysis
 * GET /api/analytics/shopee/shipping-fee?month=&year=
 */
export async function getShopeeShippingFee(
  month: number,
  year: number
): Promise<ShopeeShippingFeeResult> {
  const response = await apiClient.get<ShopeeShippingFeeResult>(
    "/analytics/shopee/shipping-fee",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(
      response.error || "Failed to fetch Shopee shipping fee"
    );
  }
  return response.data!;
}

/**
 * Repopulate Shopee items
 * POST /api/analytics/shopee/repopulate-items?period=
 */
export async function repopulateShopeeItems(
  period?: string
): Promise<void> {
  const params: Record<string, string> = {};
  if (period) {
    params.period = period;
  }
  const response = await apiClient.post(
    "/analytics/shopee/repopulate-items",
    null,
    { params }
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to repopulate Shopee items");
  }
}

// ──────────────────────────────────────────────
//  TikTok Analytics (8 functions)
// ──────────────────────────────────────────────

/**
 * Get TikTok report settings
 * GET /api/analytics/tiktok/settings
 */
export async function getTiktokSettings(): Promise<ReportSettings> {
  const response = await apiClient.get<ReportSettings>(
    "/analytics/tiktok/settings"
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch TikTok settings");
  }
  return response.data!;
}

/**
 * Save TikTok report settings
 * POST /api/analytics/tiktok/settings
 */
export async function saveTiktokSettings(
  settings: Partial<ReportSettings>
): Promise<ReportSettings> {
  const response = await apiClient.post<ReportSettings>(
    "/analytics/tiktok/settings",
    settings
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to save TikTok settings");
  }
  return response.data!;
}

/**
 * Get TikTok sync status for a given month/year
 * GET /api/analytics/tiktok/sync-status?month=&year=
 */
export async function getTiktokSyncStatus(
  month: number,
  year: number
): Promise<SyncStatus> {
  const response = await apiClient.get<SyncStatus>(
    "/analytics/tiktok/sync-status",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch TikTok sync status");
  }
  return response.data!;
}

/**
 * Trigger TikTok sync
 * POST /api/analytics/tiktok/sync
 */
export async function triggerTiktokSync(
  req: SyncRequest
): Promise<SyncResult> {
  const response = await apiClient.post<SyncResult>(
    "/analytics/tiktok/sync",
    req
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to trigger TikTok sync");
  }
  return response.data!;
}

/**
 * Delete TikTok sync data for a given month/year
 * DELETE /api/analytics/tiktok/sync?month=&year=
 */
export async function deleteTiktokSync(
  month: number,
  year: number
): Promise<void> {
  const response = await apiClient.delete("/analytics/tiktok/sync", {
    params: { month, year },
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to delete TikTok sync data");
  }
}

/**
 * Get TikTok reconciliation data
 * GET /api/analytics/tiktok/reconciliation?month=&year=
 */
export async function getTiktokReconciliation(
  month: number,
  year: number
): Promise<ReconciliationResult> {
  const response = await apiClient.get<ReconciliationResult>(
    "/analytics/tiktok/reconciliation",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(
      response.error || "Failed to fetch TikTok reconciliation"
    );
  }
  return response.data!;
}

/**
 * Get TikTok shipping fee analysis
 * GET /api/analytics/tiktok/shipping-fee?month=&year=
 */
export async function getTiktokShippingFee(
  month: number,
  year: number
): Promise<TiktokShippingFeeResult> {
  const response = await apiClient.get<TiktokShippingFeeResult>(
    "/analytics/tiktok/shipping-fee",
    { params: { month, year } }
  );
  if (!response.success) {
    throw new Error(
      response.error || "Failed to fetch TikTok shipping fee"
    );
  }
  return response.data!;
}

/**
 * Repopulate TikTok items
 * POST /api/analytics/tiktok/repopulate-items?period=
 */
export async function repopulateTiktokItems(
  period?: string
): Promise<void> {
  const params: Record<string, string> = {};
  if (period) {
    params.period = period;
  }
  const response = await apiClient.post(
    "/analytics/tiktok/repopulate-items",
    null,
    { params }
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to repopulate TikTok items");
  }
}

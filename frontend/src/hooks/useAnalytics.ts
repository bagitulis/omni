/**
 * Analytics & Report TanStack Query Hooks
 *
 * Provides typed hooks for all analytics API endpoints.
 * Uses platform-agnostic pattern: pass platform ("shopee" | "tiktok")
 * and the hook selects the correct API function.
 */
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getShopeeSettings,
  saveShopeeSettings,
  getShopeeSyncStatus,
  triggerShopeeSync,
  deleteShopeeSync,
  getShopeeReconciliation,
  getShopeeShippingFee,
  repopulateShopeeItems,
  getTiktokSettings,
  saveTiktokSettings,
  getTiktokSyncStatus,
  triggerTiktokSync,
  deleteTiktokSync,
  getTiktokReconciliation,
  getTiktokShippingFee,
  repopulateTiktokItems,
} from "@/api/analytics";
import type {
  ReportSettings,
  SyncStatus,
  SyncRequest,
  SyncResult,
  ReconciliationResult,
  ShopeeShippingFeeResult,
  TiktokShippingFeeResult,
  ReportPlatform,
} from "@/types/analytics";

// ──────────────────────────────────────────────
//  Query Keys
// ──────────────────────────────────────────────

export const analyticsKeys = {
  settings: (platform: string) => ["analytics", platform, "settings"] as const,
  syncStatus: (platform: string, month: number, year: number) =>
    ["analytics", platform, "sync-status", month, year] as const,
  reconciliation: (platform: string, month: number, year: number) =>
    ["analytics", platform, "reconciliation", month, year] as const,
  shippingFee: (platform: string, month: number, year: number) =>
    ["analytics", platform, "shipping-fee", month, year] as const,
};

// ──────────────────────────────────────────────
//  Settings Hooks
// ──────────────────────────────────────────────

/**
 * Fetch report settings for the given platform.
 */
export function useReportSettings(platform: ReportPlatform) {
  const fn = platform === "shopee" ? getShopeeSettings : getTiktokSettings;
  return useQuery<ReportSettings>({
    queryKey: analyticsKeys.settings(platform),
    queryFn: fn,
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Save (update) report settings for the given platform.
 * Invalidates settings query on success.
 */
export function useSaveReportSettings(platform: ReportPlatform) {
  const queryClient = useQueryClient();
  const fn = platform === "shopee" ? saveShopeeSettings : saveTiktokSettings;
  return useMutation({
    mutationFn: (settings: Partial<ReportSettings>) => fn(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: analyticsKeys.settings(platform),
      });
    },
  });
}

// ──────────────────────────────────────────────
//  Sync Status Hook
// ──────────────────────────────────────────────

/**
 * Poll sync status for the given platform/month/year.
 * Stale time of 30s for polling during active sync.
 */
export function useSyncStatus(
  platform: ReportPlatform,
  month: number,
  year: number
) {
  const fn =
    platform === "shopee" ? getShopeeSyncStatus : getTiktokSyncStatus;
  return useQuery<SyncStatus>({
    queryKey: analyticsKeys.syncStatus(platform, month, year),
    queryFn: () => fn(month, year),
    staleTime: 30 * 1000,
  });
}

// ──────────────────────────────────────────────
//  Sync Trigger Mutation
// ──────────────────────────────────────────────

/**
 * Trigger a sync for the given platform.
 * Invalidates sync status on success.
 */
export function useTriggerSync(platform: ReportPlatform) {
  const queryClient = useQueryClient();
  const fn = platform === "shopee" ? triggerShopeeSync : triggerTiktokSync;
  return useMutation<SyncResult, Error, SyncRequest>({
    mutationFn: (req: SyncRequest) => fn(req),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: analyticsKeys.syncStatus(
          platform,
          variables.month,
          variables.year
        ),
      });
    },
  });
}

// ──────────────────────────────────────────────
//  Delete Sync Mutation
// ──────────────────────────────────────────────

/**
 * Delete sync data for the given platform/month/year.
 * Invalidates all analytics queries for the platform on success.
 */
export function useDeleteSync(platform: ReportPlatform) {
  const queryClient = useQueryClient();
  const fn = platform === "shopee" ? deleteShopeeSync : deleteTiktokSync;
  return useMutation<void, Error, { month: number; year: number }>({
    mutationFn: ({ month, year }) => fn(month, year),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["analytics", platform],
      });
    },
  });
}

// ──────────────────────────────────────────────
//  Reconciliation Hook
// ──────────────────────────────────────────────

/**
 * Fetch reconciliation data for the given platform/month/year.
 * Reconciliation data is relatively static — 5 min stale time.
 */
export function useReconciliation(
  platform: ReportPlatform,
  month: number,
  year: number
) {
  const fn =
    platform === "shopee"
      ? getShopeeReconciliation
      : getTiktokReconciliation;
  return useQuery<ReconciliationResult>({
    queryKey: analyticsKeys.reconciliation(platform, month, year),
    queryFn: () => fn(month, year),
    staleTime: 5 * 60 * 1000,
  });
}

// ──────────────────────────────────────────────
//  Shipping Fee Hook
// ──────────────────────────────────────────────

/**
 * Fetch shipping fee analysis for the given platform/month/year.
 * Returns ShopeeShippingFeeResult or TiktokShippingFeeResult
 * depending on platform.
 */
export function useShippingFee(
  platform: ReportPlatform,
  month: number,
  year: number
) {
  const fn =
    platform === "shopee" ? getShopeeShippingFee : getTiktokShippingFee;
  return useQuery<ShopeeShippingFeeResult | TiktokShippingFeeResult>({
    queryKey: analyticsKeys.shippingFee(platform, month, year),
    queryFn: () => fn(month, year),
    staleTime: 5 * 60 * 1000,
  });
}

// ──────────────────────────────────────────────
//  Repopulate Items Mutation
// ──────────────────────────────────────────────

/**
 * Trigger item repopulation for the given platform.
 * No automatic cache invalidation — caller handles it.
 */
export function useRepopulateItems(platform: ReportPlatform) {
  const fn =
    platform === "shopee" ? repopulateShopeeItems : repopulateTiktokItems;
  return useMutation<void, Error, string | undefined>({
    mutationFn: (period?: string) => fn(period),
  });
}

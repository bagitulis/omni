import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type {
  Platform,
  AnalyticsSettings,
  JobProgress,
} from "@/types/analytics";
import {
  getSettings,
  saveSettings,
  getSyncStatus,
  syncEscrow,
  deleteSyncData,
  getReconciliation,
  getShippingFee,
  getJobStatus,
} from "@/api/analytics";

/**
 * Hook 1: UI State Management
 */
export function useAnalyticsParams(_platform: Platform) {
  const now = new Date();
  // Default to previous month
  const defaultMonth = now.getMonth() === 0 ? 11 : now.getMonth() - 1;
  const defaultYear =
    now.getMonth() === 0 ? now.getFullYear() - 1 : now.getFullYear();

  const [searchParams, setSearchParams] = useSearchParams();
  const [settingsOpen, setSettingsOpen] = useState(false);

  // Read from URL search params, fallback to defaults
  const tabParam = searchParams.get("tab");
  const activeTab: "price" | "shipping" =
    tabParam === "shipping" ? "shipping" : "price";

  const monthParam = searchParams.get("month");
  const selectedMonth =
    monthParam !== null && !isNaN(Number(monthParam))
      ? Math.max(0, Math.min(11, Number(monthParam)))
      : defaultMonth;

  const yearParam = searchParams.get("year");
  const selectedYear =
    yearParam !== null && !isNaN(Number(yearParam))
      ? Number(yearParam)
      : defaultYear;

  // Helper to update search params while preserving other params
  const updateSearchParams = useCallback(
    (updates: Record<string, string>) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          for (const [key, value] of Object.entries(updates)) {
            next.set(key, value);
          }
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const setActiveTab = useCallback(
    (tab: "price" | "shipping") => {
      updateSearchParams({ tab });
    },
    [updateSearchParams],
  );

  const setSelectedMonth = useCallback(
    (month: number) => {
      updateSearchParams({ month: String(month) });
    },
    [updateSearchParams],
  );

  const setSelectedYear = useCallback(
    (year: number) => {
      updateSearchParams({ year: String(year) });
    },
    [updateSearchParams],
  );

  // API uses 1-indexed month
  const apiMonth = selectedMonth + 1;

  // Can only sync if not current month
  const isCurrentMonth =
    selectedMonth === now.getMonth() && selectedYear === now.getFullYear();
  const canSync = !isCurrentMonth;

  return {
    activeTab,
    selectedMonth,
    selectedYear,
    settingsOpen,
    setActiveTab,
    setSelectedMonth,
    setSelectedYear,
    setSettingsOpen,
    apiMonth,
    canSync,
  };
}

/**
 * Hook 2: Data Fetching
 */
export function useAnalyticsQueries(
  platform: Platform,
  month: number,
  year: number,
) {
  const syncStatusQuery = useQuery({
    queryKey: ["analytics", platform, "sync-status", month, year],
    queryFn: () => getSyncStatus(platform, month, year),
    staleTime: 30 * 1000, // 30s
  });

  const settingsQuery = useQuery({
    queryKey: ["analytics", platform, "settings"],
    queryFn: () => getSettings(platform),
    staleTime: 5 * 60 * 1000, // 5 mins
  });

  const isSynced = syncStatusQuery.data?.synced === true;

  const reconciliationQuery = useQuery({
    queryKey: ["analytics", platform, "reconciliation", month, year],
    queryFn: () => getReconciliation(platform, month, year),
    enabled: isSynced,
    staleTime: 60 * 1000,
  });

  const shippingFeeQuery = useQuery({
    queryKey: ["analytics", platform, "shipping-fee", month, year],
    queryFn: () => getShippingFee(platform, month, year),
    enabled: isSynced,
    staleTime: 60 * 1000,
  });

  return {
    syncStatusQuery,
    settingsQuery,
    reconciliationQuery,
    shippingFeeQuery,
  };
}

/**
 * Hook 3: Sync & Mutations
 */
export function useAnalyticsSync(platform: Platform) {
  const queryClient = useQueryClient();
  const [currentJobId, setCurrentJobId] = useState<string | null>(null);
  const [jobProgress, setJobProgress] = useState<JobProgress | null>(null);

  const jobStatusQuery = useQuery({
    queryKey: ["analytics", "job", currentJobId],
    queryFn: () => getJobStatus(currentJobId!),
    enabled: !!currentJobId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (
        status === "completed" ||
        status === "failed" ||
        status === "cancelled"
      ) {
        return false;
      }
      return 2000;
    },
  });

  // Watch for job completion
  useEffect(() => {
    if (!jobStatusQuery.data) return;

    setJobProgress(jobStatusQuery.data);

    const status = jobStatusQuery.data.status;
    if (
      status === "completed" ||
      status === "failed" ||
      status === "cancelled"
    ) {
      setCurrentJobId(null);

      if (status === "completed") {
        // Invalidate relevant queries to refresh data
        queryClient.invalidateQueries({
          queryKey: ["analytics", platform, "sync-status"],
        });
        queryClient.invalidateQueries({
          queryKey: ["analytics", platform, "reconciliation"],
        });
        queryClient.invalidateQueries({
          queryKey: ["analytics", platform, "shipping-fee"],
        });
      }

      // Clear progress bar after a brief delay so user can see the completion
      const timer = setTimeout(() => {
        setJobProgress(null);
      }, 1500);
      return () => clearTimeout(timer);
    }
  }, [jobStatusQuery.data, platform, queryClient]);

  const syncMutation = useMutation({
    mutationFn: async ({
      month,
      year,
      forceResync,
    }: {
      month: number;
      year: number;
      forceResync: boolean;
    }) => {
      return syncEscrow(platform, month, year, forceResync);
    },
    onSuccess: (jobId) => {
      setCurrentJobId(jobId);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async ({ month, year }: { month: number; year: number }) => {
      return deleteSyncData(platform, month, year);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["analytics", platform, "sync-status"],
      });
      queryClient.invalidateQueries({
        queryKey: ["analytics", platform, "reconciliation"],
      });
      queryClient.invalidateQueries({
        queryKey: ["analytics", platform, "shipping-fee"],
      });
    },
  });

  const saveSettingsMutation = useMutation({
    mutationFn: async (settings: AnalyticsSettings) => {
      return saveSettings(platform, settings);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["analytics", platform, "settings"],
      });
    },
  });

  return {
    currentJobId,
    jobProgress,
    jobStatusQuery,
    syncMutation,
    deleteMutation,
    saveSettingsMutation,
  };
}

/**
 * Main Hook: Composer
 */
export function useAnalytics(platform: Platform) {
  const params = useAnalyticsParams(platform);

  // Pass apiMonth (1-indexed) to queries
  const queries = useAnalyticsQueries(
    platform,
    params.apiMonth,
    params.selectedYear,
  );

  const sync = useAnalyticsSync(platform);

  return {
    ...params,
    ...queries,
    ...sync,
  };
}

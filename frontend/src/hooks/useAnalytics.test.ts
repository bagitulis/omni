import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();
const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
    setQueryData: vi.fn(),
  }),
}));

vi.mock("@/api/analytics", () => ({
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
  getSyncStatus: vi.fn(),
  syncEscrow: vi.fn(),
  deleteSyncData: vi.fn(),
  getReconciliation: vi.fn(),
  getShippingFee: vi.fn(),
  getJobStatus: vi.fn(),
  getUnifiedKPI: vi.fn(),
  getUnifiedSummary: vi.fn(),
}));

vi.mock("react-router-dom", () => ({
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

import { useAnalyticsQueries, useAnalyticsSync } from "./useAnalytics";
import * as analyticsApi from "@/api/analytics";

describe("useAnalyticsQueries", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery four times for sync-status, settings, reconciliation, and shipping-fee", () => {
    useAnalyticsQueries("shopee", 3, 2025);
    expect(useQueryMock).toHaveBeenCalledTimes(4);
  });

  it("calls useQuery with correct queryKey for sync-status", () => {
    useAnalyticsQueries("shopee", 3, 2025);
    const firstCall = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(firstCall.queryKey).toEqual([
      "analytics",
      "shopee",
      "sync-status",
      3,
      2025,
    ]);
  });

  it("calls useQuery with correct queryKey for settings", () => {
    useAnalyticsQueries("shopee", 3, 2025);
    const secondCall = useQueryMock.mock.calls[1][0] as {
      queryKey: unknown[];
      queryFn: () => unknown;
    };
    expect(secondCall.queryKey).toEqual(["analytics", "shopee", "settings"]);
    secondCall.queryFn();
    expect(analyticsApi.getSettings).toHaveBeenCalledWith("shopee");
  });

  it("calls useQuery with correct queryFn for sync-status", () => {
    useAnalyticsQueries("tiktok", 5, 2024);
    const firstCall = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    firstCall.queryFn();
    expect(analyticsApi.getSyncStatus).toHaveBeenCalledWith("tiktok", 5, 2024);
  });
});

describe("useAnalyticsSync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation for sync with correct mutationFn", () => {
    useAnalyticsSync("shopee");
    const calls = useMutationMock.mock.calls;
    // syncMutation is first mutation
    const syncOptions = calls[0][0] as {
      mutationFn: (args: {
        month: number;
        year: number;
        forceResync: boolean;
      }) => unknown;
    };
    syncOptions.mutationFn({ month: 3, year: 2025, forceResync: false });
    expect(analyticsApi.syncEscrow).toHaveBeenCalledWith(
      "shopee",
      3,
      2025,
      false,
    );
  });

  it("calls useMutation for delete with correct mutationFn", () => {
    useAnalyticsSync("tiktok");
    const calls = useMutationMock.mock.calls;
    // deleteMutation is second mutation
    const deleteOptions = calls[1][0] as {
      mutationFn: (args: { month: number; year: number }) => unknown;
    };
    deleteOptions.mutationFn({ month: 4, year: 2024 });
    expect(analyticsApi.deleteSyncData).toHaveBeenCalledWith("tiktok", 4, 2024);
  });

  it("calls useMutation for saveSettings with correct mutationFn", () => {
    useAnalyticsSync("shopee");
    const calls = useMutationMock.mock.calls;
    // saveSettingsMutation is third mutation
    const saveOptions = calls[2][0] as {
      mutationFn: (settings: unknown) => unknown;
    };
    const settings = { fee_rate: 5 };
    saveOptions.mutationFn(settings);
    expect(analyticsApi.saveSettings).toHaveBeenCalledWith("shopee", settings);
  });

  it("invalidates queries on delete success", () => {
    useAnalyticsSync("shopee");
    const calls = useMutationMock.mock.calls;
    const deleteOptions = calls[1][0] as {
      onSuccess: () => void;
    };
    deleteOptions.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledTimes(3);
  });
});

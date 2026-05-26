import { beforeEach, describe, expect, it, vi } from "vitest";

// ──────────────────────────────────────────────
//  Hoisted mocks (run before vi.mock factories)
// ──────────────────────────────────────────────
const useQueryMock = vi.hoisted(
  () => vi.fn(() => ({ data: undefined, isLoading: false })),
);
const useMutationMock = vi.hoisted(
  () => vi.fn((options: Record<string, unknown>) => options),
);
const invalidateQueriesMock = vi.hoisted(() => vi.fn());

const analyticsApiMock = vi.hoisted(() => ({
  getShopeeSettings: vi.fn(),
  saveShopeeSettings: vi.fn(),
  getShopeeSyncStatus: vi.fn(),
  triggerShopeeSync: vi.fn(),
  deleteShopeeSync: vi.fn(),
  getShopeeReconciliation: vi.fn(),
  getShopeeShippingFee: vi.fn(),
  repopulateShopeeItems: vi.fn(),
  getTiktokSettings: vi.fn(),
  saveTiktokSettings: vi.fn(),
  getTiktokSyncStatus: vi.fn(),
  triggerTiktokSync: vi.fn(),
  deleteTiktokSync: vi.fn(),
  getTiktokReconciliation: vi.fn(),
  getTiktokShippingFee: vi.fn(),
  repopulateTiktokItems: vi.fn(),
}));

// ──────────────────────────────────────────────
//  Module mocks
// ──────────────────────────────────────────────
vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("@/api/analytics", () => analyticsApiMock);

// ──────────────────────────────────────────────
//  Imports (after mocks)
// ──────────────────────────────────────────────
import {
  useReportSettings,
  useSaveReportSettings,
  useSyncStatus,
  useTriggerSync,
  useDeleteSync,
  useReconciliation,
  useShippingFee,
  useRepopulateItems,
  analyticsKeys,
} from "./useAnalytics";

// ──────────────────────────────────────────────
//  Helpers
// ──────────────────────────────────────────────
function getUseQueryOptions() {
  return (useQueryMock.mock.calls[0] as unknown[])[0] as {
    queryKey: unknown[];
    queryFn: () => unknown;
    staleTime?: number;
  };
}

function getUseMutationOptions() {
  return (useMutationMock.mock.calls[0] as unknown[])[0] as {
    mutationFn: (...args: unknown[]) => unknown;
    onSuccess?: (...args: unknown[]) => void;
  };
}

// ──────────────────────────────────────────────
//  useReportSettings
// ──────────────────────────────────────────────
describe("useReportSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with shopee settings key when platform is 'shopee'", () => {
    useReportSettings("shopee");

    const opts = getUseQueryOptions();
    expect(opts.queryKey).toEqual(["analytics", "shopee", "settings"]);
  });

  it("queryFn calls getShopeeSettings for shopee platform", () => {
    useReportSettings("shopee");

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getShopeeSettings).toHaveBeenCalledOnce();
  });

  it("calls useQuery with tiktok settings key when platform is 'tiktok'", () => {
    useReportSettings("tiktok");

    const opts = getUseQueryOptions();
    expect(opts.queryKey).toEqual(["analytics", "tiktok", "settings"]);
  });

  it("queryFn calls getTiktokSettings for tiktok platform", () => {
    useReportSettings("tiktok");

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getTiktokSettings).toHaveBeenCalledOnce();
  });

  it("sets staleTime to 5 minutes", () => {
    useReportSettings("shopee");

    const opts = getUseQueryOptions();
    expect(opts.staleTime).toBe(5 * 60 * 1000);
  });
});

// ──────────────────────────────────────────────
//  useSaveReportSettings
// ──────────────────────────────────────────────
describe("useSaveReportSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("mutationFn calls saveShopeeSettings for shopee platform", () => {
    const result = useSaveReportSettings("shopee") as unknown as {
      mutationFn: (settings: Record<string, unknown>) => unknown;
    };

    result.mutationFn({ price_column: "cost" });

    expect(analyticsApiMock.saveShopeeSettings).toHaveBeenCalledWith({
      price_column: "cost",
    });
  });

  it("mutationFn calls saveTiktokSettings for tiktok platform", () => {
    const result = useSaveReportSettings("tiktok") as unknown as {
      mutationFn: (settings: Record<string, unknown>) => unknown;
    };

    result.mutationFn({ formula_multiplier: 2 });

    expect(analyticsApiMock.saveTiktokSettings).toHaveBeenCalledWith({
      formula_multiplier: 2,
    });
  });

  it("onSuccess invalidates the settings query for the platform", () => {
    const result = useSaveReportSettings("shopee") as unknown as {
      onSuccess: () => void;
    };

    result.onSuccess();

    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["analytics", "shopee", "settings"],
    });
  });

  it("invalidates tiktok settings on success for tiktok platform", () => {
    const result = useSaveReportSettings("tiktok") as unknown as {
      onSuccess: () => void;
    };

    result.onSuccess();

    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["analytics", "tiktok", "settings"],
    });
  });
});

// ──────────────────────────────────────────────
//  useSyncStatus
// ──────────────────────────────────────────────
describe("useSyncStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct query key for shopee", () => {
    useSyncStatus("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    expect(opts.queryKey).toEqual(["analytics", "shopee", "sync-status", 3, 2026]);
  });

  it("queryFn calls getShopeeSyncStatus with month and year", () => {
    useSyncStatus("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getShopeeSyncStatus).toHaveBeenCalledWith(3, 2026);
  });

  it("queryFn calls getTiktokSyncStatus for tiktok platform", () => {
    useSyncStatus("tiktok", 4, 2025);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getTiktokSyncStatus).toHaveBeenCalledWith(4, 2025);
  });

  it("sets staleTime to 30 seconds", () => {
    useSyncStatus("shopee", 1, 2026);

    const opts = getUseQueryOptions();
    expect(opts.staleTime).toBe(30 * 1000);
  });
});

// ──────────────────────────────────────────────
//  useTriggerSync
// ──────────────────────────────────────────────
describe("useTriggerSync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("mutationFn calls triggerShopeeSync with request", () => {
    const result = useTriggerSync("shopee") as unknown as {
      mutationFn: (req: Record<string, unknown>) => unknown;
    };

    const req = { month: 3, year: 2026, force_resync: true };
    result.mutationFn(req);

    expect(analyticsApiMock.triggerShopeeSync).toHaveBeenCalledWith(req);
  });

  it("mutationFn calls triggerTiktokSync for tiktok", () => {
    const result = useTriggerSync("tiktok") as unknown as {
      mutationFn: (req: Record<string, unknown>) => unknown;
    };

    const req = { month: 4, year: 2025, force_resync: false };
    result.mutationFn(req);

    expect(analyticsApiMock.triggerTiktokSync).toHaveBeenCalledWith(req);
  });

  it("onSuccess invalidates sync-status query for the given month/year", () => {
    const result = useTriggerSync("shopee") as unknown as {
      onSuccess: (...args: unknown[]) => void;
    };

    result.onSuccess(null, { month: 3, year: 2026, force_resync: false });

    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["analytics", "shopee", "sync-status", 3, 2026],
    });
  });
});

// ──────────────────────────────────────────────
//  useDeleteSync
// ──────────────────────────────────────────────
describe("useDeleteSync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("mutationFn calls deleteShopeeSync with month and year", () => {
    const result = useDeleteSync("shopee") as unknown as {
      mutationFn: (args: { month: number; year: number }) => unknown;
    };

    result.mutationFn({ month: 3, year: 2026 });

    expect(analyticsApiMock.deleteShopeeSync).toHaveBeenCalledWith(3, 2026);
  });

  it("mutationFn calls deleteTiktokSync for tiktok", () => {
    const result = useDeleteSync("tiktok") as unknown as {
      mutationFn: (args: { month: number; year: number }) => unknown;
    };

    result.mutationFn({ month: 1, year: 2025 });

    expect(analyticsApiMock.deleteTiktokSync).toHaveBeenCalledWith(1, 2025);
  });

  it("onSuccess invalidates all analytics queries for the platform", () => {
    const result = useDeleteSync("shopee") as unknown as {
      onSuccess: () => void;
    };

    result.onSuccess();

    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["analytics", "shopee"],
    });
  });
});

// ──────────────────────────────────────────────
//  useReconciliation
// ──────────────────────────────────────────────
describe("useReconciliation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct reconciliation key for shopee", () => {
    useReconciliation("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    expect(opts.queryKey).toEqual([
      "analytics",
      "shopee",
      "reconciliation",
      3,
      2026,
    ]);
  });

  it("queryFn calls getShopeeReconciliation with month and year", () => {
    useReconciliation("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getShopeeReconciliation).toHaveBeenCalledWith(
      3,
      2026,
    );
  });

  it("queryFn calls getTiktokReconciliation for tiktok platform", () => {
    useReconciliation("tiktok", 1, 2025);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getTiktokReconciliation).toHaveBeenCalledWith(
      1,
      2025,
    );
  });

  it("sets staleTime to 5 minutes", () => {
    useReconciliation("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    expect(opts.staleTime).toBe(5 * 60 * 1000);
  });
});

// ──────────────────────────────────────────────
//  useShippingFee
// ──────────────────────────────────────────────
describe("useShippingFee", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct shipping-fee key for shopee", () => {
    useShippingFee("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    expect(opts.queryKey).toEqual([
      "analytics",
      "shopee",
      "shipping-fee",
      3,
      2026,
    ]);
  });

  it("queryFn calls getShopeeShippingFee with month and year", () => {
    useShippingFee("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getShopeeShippingFee).toHaveBeenCalledWith(3, 2026);
  });

  it("queryFn calls getTiktokShippingFee for tiktok platform", () => {
    useShippingFee("tiktok", 4, 2025);

    const opts = getUseQueryOptions();
    opts.queryFn();

    expect(analyticsApiMock.getTiktokShippingFee).toHaveBeenCalledWith(
      4,
      2025,
    );
  });

  it("sets staleTime to 5 minutes", () => {
    useShippingFee("shopee", 3, 2026);

    const opts = getUseQueryOptions();
    expect(opts.staleTime).toBe(5 * 60 * 1000);
  });
});

// ──────────────────────────────────────────────
//  useRepopulateItems
// ──────────────────────────────────────────────
describe("useRepopulateItems", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("mutationFn calls repopulateShopeeItems for shopee platform", () => {
    const result = useRepopulateItems("shopee") as unknown as {
      mutationFn: (period?: string) => unknown;
    };

    result.mutationFn("2026-03");

    expect(analyticsApiMock.repopulateShopeeItems).toHaveBeenCalledWith(
      "2026-03",
    );
  });

  it("mutationFn calls repopulateTiktokItems for tiktok platform", () => {
    const result = useRepopulateItems("tiktok") as unknown as {
      mutationFn: (period?: string) => unknown;
    };

    result.mutationFn(undefined);

    expect(analyticsApiMock.repopulateTiktokItems).toHaveBeenCalledWith(
      undefined,
    );
  });

  it("mutationFn works without period argument", () => {
    const result = useRepopulateItems("shopee") as unknown as {
      mutationFn: (period?: string) => unknown;
    };

    result.mutationFn();

    expect(analyticsApiMock.repopulateShopeeItems).toHaveBeenCalledWith(
      undefined,
    );
  });
});

// ──────────────────────────────────────────────
//  analyticsKeys
// ──────────────────────────────────────────────
describe("analyticsKeys", () => {
  it("settings returns correct key tuple", () => {
    expect(analyticsKeys.settings("shopee")).toEqual([
      "analytics",
      "shopee",
      "settings",
    ]);
    expect(analyticsKeys.settings("tiktok")).toEqual([
      "analytics",
      "tiktok",
      "settings",
    ]);
  });

  it("syncStatus returns correct key tuple", () => {
    expect(analyticsKeys.syncStatus("shopee", 3, 2026)).toEqual([
      "analytics",
      "shopee",
      "sync-status",
      3,
      2026,
    ]);
  });

  it("reconciliation returns correct key tuple", () => {
    expect(analyticsKeys.reconciliation("shopee", 3, 2026)).toEqual([
      "analytics",
      "shopee",
      "reconciliation",
      3,
      2026,
    ]);
  });

  it("shippingFee returns correct key tuple", () => {
    expect(analyticsKeys.shippingFee("shopee", 3, 2026)).toEqual([
      "analytics",
      "shopee",
      "shipping-fee",
      3,
      2026,
    ]);
  });
});

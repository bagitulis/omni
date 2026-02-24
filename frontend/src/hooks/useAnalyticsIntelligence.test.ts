import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("../api/analyticsIntelligence", () => ({
  getUnifiedAnalytics: vi.fn(),
  getClassifiedProducts: vi.fn(),
  getProductsFromAds: vi.fn(),
  runSimulation: vi.fn(),
  generateReport: vi.fn(),
  getReports: vi.fn(),
  getReportHTML: vi.fn(),
}));

import {
  useUnifiedAnalytics,
  useClassifiedProducts,
  useProductsFromAds,
  useBudgetSimulation,
  useGenerateReport,
  useReports,
  useReportHTML,
} from "./useAnalyticsIntelligence";
import * as intelligenceApi from "../api/analyticsIntelligence";

describe("useUnifiedAnalytics", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useUnifiedAnalytics();
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["analytics", "unified"]);
  });

  it("queryFn calls getUnifiedAnalytics", () => {
    useUnifiedAnalytics();
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(intelligenceApi.getUnifiedAnalytics).toHaveBeenCalledOnce();
  });
});

describe("useClassifiedProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useClassifiedProducts();
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["analytics", "classification"]);
  });

  it("queryFn calls getClassifiedProducts", () => {
    useClassifiedProducts();
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(intelligenceApi.getClassifiedProducts).toHaveBeenCalledOnce();
  });
});

describe("useProductsFromAds", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useProductsFromAds();
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["analytics", "products-ads"]);
  });

  it("queryFn calls getProductsFromAds", () => {
    useProductsFromAds();
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(intelligenceApi.getProductsFromAds).toHaveBeenCalledOnce();
  });
});

describe("useBudgetSimulation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation", () => {
    useBudgetSimulation();
    expect(useMutationMock).toHaveBeenCalledOnce();
  });

  it("mutationFn calls runSimulation with request", () => {
    useBudgetSimulation();
    const opts = useMutationMock.mock.calls[0][0] as {
      mutationFn: (req: unknown) => unknown;
    };
    const req = { budget: 1000 };
    opts.mutationFn(req);
    expect(intelligenceApi.runSimulation).toHaveBeenCalledWith(req);
  });
});

describe("useGenerateReport", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation", () => {
    useGenerateReport();
    expect(useMutationMock).toHaveBeenCalledOnce();
  });

  it("mutationFn calls generateReport with request", () => {
    useGenerateReport();
    const opts = useMutationMock.mock.calls[0][0] as {
      mutationFn: (req: unknown) => unknown;
    };
    const req = { platform: "shopee" as const };
    opts.mutationFn(req);
    expect(intelligenceApi.generateReport).toHaveBeenCalledWith(req);
  });

  it("onSuccess invalidates reports query for platform", () => {
    useGenerateReport();
    const opts = useMutationMock.mock.calls[0][0] as {
      onSuccess: (data: unknown, variables: { platform: string }) => void;
    };
    opts.onSuccess(undefined, { platform: "shopee" });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["analytics", "reports", "shopee"],
    });
  });
});

describe("useReports", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useReports("shopee", 1, 20);
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["analytics", "reports", "shopee", 1, 20]);
  });

  it("queryFn calls getReports with platform, page, limit", () => {
    useReports("tiktok", 2, 10);
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(intelligenceApi.getReports).toHaveBeenCalledWith("tiktok", 2, 10);
  });
});

describe("useReportHTML", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("is disabled when filename is null", () => {
    useReportHTML("shopee", null);
    const opts = useQueryMock.mock.calls[0][0] as { enabled: boolean };
    expect(opts.enabled).toBe(false);
  });

  it("is enabled when filename is provided", () => {
    useReportHTML("shopee", "report.html");
    const opts = useQueryMock.mock.calls[0][0] as { enabled: boolean };
    expect(opts.enabled).toBe(true);
  });

  it("queryFn calls getReportHTML with platform and filename", () => {
    useReportHTML("tiktok", "test.html");
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(intelligenceApi.getReportHTML).toHaveBeenCalledWith(
      "tiktok",
      "test.html",
    );
  });
});

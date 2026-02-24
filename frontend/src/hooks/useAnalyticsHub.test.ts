import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/analytics", () => ({
  getUnifiedKPI: vi.fn(),
  getUnifiedSummary: vi.fn(),
}));

import { useAnalyticsHub } from "./useAnalyticsHub";
import * as analyticsApi from "@/api/analytics";

describe("useAnalyticsHub", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
  });

  it("calls useQuery twice (kpi + summary)", () => {
    useAnalyticsHub();
    expect(useQueryMock).toHaveBeenCalledTimes(2);
  });

  it("kpi query uses correct queryKey", () => {
    useAnalyticsHub();
    const kpiOptions = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(kpiOptions.queryKey).toEqual(["analytics", "unified", "kpi"]);
  });

  it("summary query uses correct queryKey", () => {
    useAnalyticsHub();
    const summaryOptions = useQueryMock.mock.calls[1][0] as {
      queryKey: unknown[];
    };
    expect(summaryOptions.queryKey).toEqual([
      "analytics",
      "unified",
      "summary",
    ]);
  });

  it("kpi queryFn calls getUnifiedKPI", () => {
    useAnalyticsHub();
    const kpiOptions = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    kpiOptions.queryFn();
    expect(analyticsApi.getUnifiedKPI).toHaveBeenCalledOnce();
  });

  it("summary queryFn calls getUnifiedSummary", () => {
    useAnalyticsHub();
    const summaryOptions = useQueryMock.mock.calls[1][0] as {
      queryFn: () => unknown;
    };
    summaryOptions.queryFn();
    expect(analyticsApi.getUnifiedSummary).toHaveBeenCalledOnce();
  });

  it("returns isLoading as combined loading state", () => {
    useQueryMock
      .mockReturnValueOnce({ isLoading: true, error: null, refetch: vi.fn() })
      .mockReturnValueOnce({
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
    const result = useAnalyticsHub();
    expect(result.isLoading).toBe(true);
  });

  it("returns isLoading false when both queries done", () => {
    useQueryMock
      .mockReturnValueOnce({
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      })
      .mockReturnValueOnce({
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
    const result = useAnalyticsHub();
    expect(result.isLoading).toBe(false);
  });

  it("returns error from kpi query if present", () => {
    const mockError = new Error("kpi error");
    useQueryMock
      .mockReturnValueOnce({
        isLoading: false,
        error: mockError,
        refetch: vi.fn(),
      })
      .mockReturnValueOnce({
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
    const result = useAnalyticsHub();
    expect(result.error).toBe(mockError);
  });
});

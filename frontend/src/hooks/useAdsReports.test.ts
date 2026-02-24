import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/ads", () => ({
  getShopeeAdsReports: vi.fn(),
  getTiktokAdsReports: vi.fn(),
}));

import { useAdsReports } from "./useAdsReports";
import * as adsApi from "@/api/ads";

describe("useAdsReports", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
  });

  it("calls useQuery twice (shopee + tiktok)", () => {
    useAdsReports("shopee");
    expect(useQueryMock).toHaveBeenCalledTimes(2);
  });

  it("passes shopee queryKey for shopee query", () => {
    useAdsReports("shopee");
    const firstCall = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(firstCall.queryKey).toEqual(["shopee-ads-reports"]);
  });

  it("passes tiktok queryKey for tiktok query", () => {
    useAdsReports("shopee");
    const secondCall = useQueryMock.mock.calls[1][0] as { queryKey: unknown[] };
    expect(secondCall.queryKey).toEqual(["tiktok-ads-reports"]);
  });

  it("enables shopee query when platform is shopee", () => {
    useAdsReports("shopee");
    const shopeeOptions = useQueryMock.mock.calls[0][0] as { enabled: boolean };
    expect(shopeeOptions.enabled).toBe(true);
  });

  it("disables shopee query when platform is tiktok", () => {
    useAdsReports("tiktok");
    const shopeeOptions = useQueryMock.mock.calls[0][0] as { enabled: boolean };
    expect(shopeeOptions.enabled).toBe(false);
  });

  it("enables tiktok query when platform is tiktok", () => {
    useAdsReports("tiktok");
    const tiktokOptions = useQueryMock.mock.calls[1][0] as { enabled: boolean };
    expect(tiktokOptions.enabled).toBe(true);
  });

  it("shopee queryFn calls getShopeeAdsReports", () => {
    useAdsReports("shopee");
    const shopeeOptions = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    shopeeOptions.queryFn();
    expect(adsApi.getShopeeAdsReports).toHaveBeenCalledOnce();
  });

  it("tiktok queryFn calls getTiktokAdsReports", () => {
    useAdsReports("tiktok");
    const tiktokOptions = useQueryMock.mock.calls[1][0] as {
      queryFn: () => unknown;
    };
    tiktokOptions.queryFn();
    expect(adsApi.getTiktokAdsReports).toHaveBeenCalledOnce();
  });

  it("returns getReportFileUrl that builds correct path", () => {
    const result = useAdsReports("shopee");
    const url = result.getReportFileUrl("report-2024.html");
    expect(url).toBe("/api/reports/shopee/report-2024.html");
  });

  it("returns empty reports array when data is undefined", () => {
    const result = useAdsReports("shopee");
    expect(result.reports).toEqual([]);
  });

  it("returns reports from data when success is true", () => {
    const mockReports = [{ name: "report1" }];
    useQueryMock.mockReturnValue({
      data: { success: true, data: mockReports },
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    const result = useAdsReports("shopee");
    expect(result.reports).toEqual(mockReports);
  });
});

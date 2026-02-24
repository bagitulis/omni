import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/ads", () => ({
  getShopeeAdsDashboard: vi.fn(),
  getShopeeAdsData: vi.fn(),
  getTiktokAdsDashboard: vi.fn(),
  getTiktokAdsData: vi.fn(),
  getShopeeAdsReports: vi.fn(),
  getTiktokAdsReports: vi.fn(),
}));

import {
  useShopeeAdsDashboard,
  useShopeeAdsData,
  useTiktokAdsDashboard,
  useTiktokAdsData,
} from "./useAds";
import * as adsApi from "@/api/ads";

describe("useShopeeAdsDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey", () => {
    useShopeeAdsDashboard();
    expect(useQueryMock).toHaveBeenCalledOnce();
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["shopee-ads-dashboard"]);
  });

  it("calls useQuery with a queryFn that invokes getShopeeAdsDashboard", () => {
    useShopeeAdsDashboard();
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    options.queryFn();
    expect(adsApi.getShopeeAdsDashboard).toHaveBeenCalledOnce();
  });
});

describe("useShopeeAdsData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey including params", () => {
    const params = { limit: 10, offset: 0 };
    useShopeeAdsData(params);
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["shopee-ads-data", params]);
  });

  it("calls useQuery with a queryFn that invokes getShopeeAdsData with params", () => {
    const params = { limit: 10, offset: 0 };
    useShopeeAdsData(params);
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    options.queryFn();
    expect(adsApi.getShopeeAdsData).toHaveBeenCalledWith(params);
  });

  it("calls useQuery with correct queryKey when no params provided", () => {
    useShopeeAdsData();
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["shopee-ads-data", undefined]);
  });
});

describe("useTiktokAdsDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey", () => {
    useTiktokAdsDashboard();
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["tiktok-ads-dashboard"]);
  });

  it("calls useQuery with a queryFn that invokes getTiktokAdsDashboard", () => {
    useTiktokAdsDashboard();
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    options.queryFn();
    expect(adsApi.getTiktokAdsDashboard).toHaveBeenCalledOnce();
  });
});

describe("useTiktokAdsData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey including params", () => {
    const params = { limit: 5, offset: 5 };
    useTiktokAdsData(params);
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["tiktok-ads-data", params]);
  });

  it("calls useQuery with a queryFn that invokes getTiktokAdsData with params", () => {
    const params = { limit: 5, offset: 5 };
    useTiktokAdsData(params);
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    options.queryFn();
    expect(adsApi.getTiktokAdsData).toHaveBeenCalledWith(params);
  });
});

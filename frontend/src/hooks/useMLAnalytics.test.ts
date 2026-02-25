import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/mlAnalytics", () => ({
  getPortfolioHealth: vi.fn(),
  getMLProducts: vi.fn(),
  getMLAlerts: vi.fn(),
  getScoreDistribution: vi.fn(),
}));

import {
  usePortfolioHealth,
  useMLProducts,
} from "./useMLAnalytics";
import * as mlApi from "@/api/mlAnalytics";

describe("usePortfolioHealth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey for default platform", () => {
    usePortfolioHealth();
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["ml", "portfolio-health", "tiktok"]);
  });

  it("calls useQuery with correct queryKey for custom platform", () => {
    usePortfolioHealth("shopee");
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["ml", "portfolio-health", "shopee"]);
  });

  it("queryFn calls getPortfolioHealth with platform", () => {
    usePortfolioHealth("shopee");
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(mlApi.getPortfolioHealth).toHaveBeenCalledWith("shopee");
  });
});

describe("useMLProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey for empty params", () => {
    useMLProducts({});
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["ml", "products", {}]);
  });

  it("calls useQuery with params in queryKey", () => {
    const params = { platform: "shopee", page: 1 };
    useMLProducts(params);
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["ml", "products", params]);
  });

  it("queryFn calls getMLProducts with params", () => {
    const params = { platform: "tiktok" };
    useMLProducts(params);
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(mlApi.getMLProducts).toHaveBeenCalledWith(params);
  });
});


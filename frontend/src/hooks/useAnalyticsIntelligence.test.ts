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
}));

import {
  useUnifiedAnalytics,
  useClassifiedProducts,
  useProductsFromAds,
  useBudgetSimulation,
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



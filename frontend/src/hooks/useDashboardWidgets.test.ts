import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/dashboard", () => ({
  getWalletData: vi.fn(),
  getShippingFeeData: vi.fn(),
}));

import {
  useWalletData,
  useShippingFeeData,
} from "./useDashboardWidgets";
import * as dashboardApi from "@/api/dashboard";

describe("useWalletData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useWalletData("shopee");
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["wallet", "shopee"]);
  });

  it("queryFn calls getWalletData with platform", () => {
    useWalletData("tiktok");
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(dashboardApi.getWalletData).toHaveBeenCalledWith("tiktok");
  });
});

describe("useShippingFeeData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with correct queryKey", () => {
    useShippingFeeData("lazada");
    const opts = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(opts.queryKey).toEqual(["shipping-fee", "lazada"]);
  });

  it("queryFn calls getShippingFeeData with platform", () => {
    useShippingFeeData("shopee");
    const opts = useQueryMock.mock.calls[0][0] as { queryFn: () => unknown };
    opts.queryFn();
    expect(dashboardApi.getShippingFeeData).toHaveBeenCalledWith("shopee");
  });
});


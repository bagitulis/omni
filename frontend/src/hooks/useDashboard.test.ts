import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/dashboard", () => ({
  getDashboardData: vi.fn(),
}));

import { useDashboard } from "./useDashboard";
import { getDashboardData } from "@/api/dashboard";

describe("useDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with dashboard queryKey", () => {
    useDashboard();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["dashboard"]);
  });

  it("uses getDashboardData as queryFn", () => {
    useDashboard();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryFn: () => unknown };
    args.queryFn();
    expect(getDashboardData).toHaveBeenCalled();
  });

  it("uses staleTime of 60 seconds", () => {
    useDashboard();
    const args = useQueryMock.mock.calls[0]?.[0] as { staleTime: number };
    expect(args.staleTime).toBe(60 * 1000);
  });

  it("uses refetchInterval of 60 seconds", () => {
    useDashboard();
    const args = useQueryMock.mock.calls[0]?.[0] as { refetchInterval: number };
    expect(args.refetchInterval).toBe(60 * 1000);
  });
});

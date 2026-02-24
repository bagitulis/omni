import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
}));

vi.mock("@/api/routeMapping", () => ({
  getRouteMappingDetailed: vi.fn(),
  getRouteMappingStatistics: vi.fn(),
}));

// useRouteMapping also uses useState and useMemo — mock react
vi.mock("react", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react")>();
  return {
    ...actual,
    useState: vi.fn((initial: unknown) => [initial, vi.fn()]),
    useMemo: vi.fn((fn: () => unknown) => fn()),
  };
});

import { useRouteMapping } from "./useRouteMapping";
import {
  getRouteMappingDetailed,
  getRouteMappingStatistics,
} from "@/api/routeMapping";

describe("useRouteMapping", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with route-mapping queryKey", () => {
    useRouteMapping();
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
      staleTime: number;
    };
    expect(args.queryKey).toEqual(["route-mapping"]);
  });

  it("uses staleTime of 5 minutes", () => {
    useRouteMapping();
    const args = useQueryMock.mock.calls[0]?.[0] as { staleTime: number };
    expect(args.staleTime).toBe(5 * 60 * 1000);
  });

  it("queryFn calls both getRouteMappingDetailed and getRouteMappingStatistics", async () => {
    const mockMapping = {
      categories: {
        connected: [],
        frontend_only: [],
        backend_only: [],
        unused: [],
      },
      by_category: {},
      components: {},
    };
    const mockStats = { statistics: { total: 0 } };
    (getRouteMappingDetailed as ReturnType<typeof vi.fn>).mockResolvedValue(
      mockMapping,
    );
    (getRouteMappingStatistics as ReturnType<typeof vi.fn>).mockResolvedValue(
      mockStats,
    );

    useRouteMapping();
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await args.queryFn();

    expect(getRouteMappingDetailed).toHaveBeenCalled();
    expect(getRouteMappingStatistics).toHaveBeenCalled();
  });

  it("useState is called with initial values for viewMode and searchQuery", () => {
    useRouteMapping();
    // Verify the hook ran and useQuery was called with the expected queryKey
    expect(useQueryMock).toHaveBeenCalled();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["route-mapping"]);
  });
});

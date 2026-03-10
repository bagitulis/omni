import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getRouteMappingDetailed,
  getRouteMappingStatistics,
} from "./routeMapping";

const { mockGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
  },
}));

const baseRouteMapping = {
  total_routes: 1,
  total_components: 1,
  total_categories: 4,
  total_dynamic_routes: 0,
  total_called_routes: 1,
  total_disconnected_routes: 0,
  total_unused_routes: 0,
  connection_rate: "100.0%",
  by_category: {
    connected: [{ endpoint: "/api/orders", method: "GET", category: "orders" }],
    frontend_only: [],
    backend_only: [],
    unused: [],
  },
  categories: {
    connected: [{ endpoint: "/api/orders", method: "GET", category: "orders" }],
    frontend_only: [],
    backend_only: [],
    unused: [],
  },
  category_labels: {},
  category_stats: {
    connected: 1,
    frontend_only: 0,
    backend_only: 0,
    unused: 0,
  },
  components: {
    orders: {
      path: "src/api/orders.ts",
      routes_called: ["GET /api/orders"],
    },
  },
  disconnected_routes: {},
  backend_only_routes: {},
  unused_routes: {},
  statistics: {},
  timestamp: "2026-02-17T00:00:00Z",
};

describe("routeMapping api adapters", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns detailed route mapping payload from backend", async () => {
    mockGet.mockResolvedValue({ success: true, data: baseRouteMapping });

    await expect(getRouteMappingDetailed()).resolves.toEqual(baseRouteMapping);
    expect(mockGet).toHaveBeenCalledWith("/routes/mapping");
  });

  it("throws when detailed route mapping response is unsuccessful", async () => {
    mockGet.mockResolvedValue({ success: false, error: "backend failed" });

    await expect(getRouteMappingDetailed()).rejects.toThrow("backend failed");
  });

  it("returns statistics object from stats endpoint", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {
        statistics: {
          connected_route_total: 10,
          method_mismatch_total: 2,
        },
      },
    });

    await expect(getRouteMappingStatistics()).resolves.toEqual({
      statistics: {
        connected_route_total: 10,
        method_mismatch_total: 2,
      },
    });
    expect(mockGet).toHaveBeenCalledWith("/routes/mapping/stats");
  });
});

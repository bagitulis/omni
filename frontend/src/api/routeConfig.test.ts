import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getRouteConfigs,
  patchRouteConfig,
  bulkUpdateRouteConfigs,
} from "./routeConfig";

const { mockGet, mockPost, mockPatch } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPatch: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: vi.fn(),
    patch: mockPatch,
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

describe("getRouteConfigs", () => {
  it("returns array of route configs", async () => {
    const routes = [
      {
        id: 1,
        route_path: "/api/orders",
        route_method: "GET",
        route_name: "orders.list",
        description: "Get orders",
        category: "orders",
        enabled: true,
        caching_enabled: false,
        cache_ttl: 0,
        cache_strategy: "none",
        queue_enabled: false,
        queue_max_size: 0,
        queue_priority: 0,
        max_concurrent: 0,
        rate_limit_enabled: false,
        rate_limit_window: 0,
        rate_limit_max: 0,
        min_interval_ms: 0,
        timeout: 30000,
        retry_enabled: false,
        max_retries: 0,
        retry_delay_ms: 0,
        custom_config: null,
      },
    ];
    mockGet.mockResolvedValue({ success: true, data: routes });

    const result = await getRouteConfigs();

    expect(result).toEqual(routes);
  });

  it("returns empty array when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    const result = await getRouteConfigs();

    expect(result).toEqual([]);
  });

  it("calls GET /routes-config", async () => {
    mockGet.mockResolvedValue({ success: true, data: [] });

    await getRouteConfigs();

    expect(mockGet).toHaveBeenCalledWith("/routes-config");
  });

  it("throws when response.success is false", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not authorized" });

    await expect(getRouteConfigs()).rejects.toThrow("Not authorized");
  });

  it("throws default message when error is absent", async () => {
    mockGet.mockResolvedValue({ success: false });

    await expect(getRouteConfigs()).rejects.toThrow(
      "Failed to fetch route configs",
    );
  });
});

describe("patchRouteConfig", () => {
  it("calls PATCH with correct id and data", async () => {
    mockPatch.mockResolvedValue({ success: true });

    await patchRouteConfig(5, { enabled: false });

    expect(mockPatch).toHaveBeenCalledWith("/routes-config/5", {
      enabled: false,
    });
  });

  it("resolves without error on success", async () => {
    mockPatch.mockResolvedValue({ success: true });

    await expect(
      patchRouteConfig(1, { timeout: 60000 }),
    ).resolves.toBeUndefined();
  });

  it("throws when response.success is false", async () => {
    mockPatch.mockResolvedValue({
      success: false,
      error: "Route not found",
    });

    await expect(patchRouteConfig(99, {})).rejects.toThrow("Route not found");
  });
});

describe("bulkUpdateRouteConfigs", () => {
  it("posts ids and data to bulk-update endpoint", async () => {
    mockPost.mockResolvedValue({ success: true });

    await bulkUpdateRouteConfigs([1, 2, 3], { enabled: true });

    expect(mockPost).toHaveBeenCalledWith("/routes-config/bulk-update", {
      ids: [1, 2, 3],
      enabled: true,
    });
  });

  it("resolves without error on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      bulkUpdateRouteConfigs([1], { caching_enabled: true }),
    ).resolves.toBeUndefined();
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Bulk update failed",
    });

    await expect(bulkUpdateRouteConfigs([1], {})).rejects.toThrow(
      "Bulk update failed",
    );
  });
});

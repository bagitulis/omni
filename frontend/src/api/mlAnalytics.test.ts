import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getPortfolioHealth,
  getMLProducts,
  getMLAlerts,
  getScoreDistribution,
} from "./mlAnalytics";

const { mockGet, mockClientGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockClientGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: mockClientGet, post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

describe("getPortfolioHealth", () => {
  const mockHealth = {
    tenant_id: "test-tenant",
    total_products: 100,
    health_score: 85,
    health_label: "Good",
    total_cost: 5000,
    total_revenue: 10000,
    total_profit: 5000,
    overall_roas: 2.0,
    star_count: 10,
    growth_count: 20,
    stable_count: 30,
    watch_count: 15,
    problem_count: 5,
    scale_up_count: 8,
    maintain_count: 12,
    reduce_count: 3,
    stop_count: 2,
    fatigue_warnings: 1,
    churn_risks: 2,
    active_alerts: 3,
    last_updated: "2026-01-01T00:00:00Z",
  };

  it("returns portfolio health data", async () => {
    mockGet.mockResolvedValue({ data: mockHealth });

    const result = await getPortfolioHealth();

    expect(result).toEqual(mockHealth);
  });

  it("uses default platform=tiktok", async () => {
    mockGet.mockResolvedValue({ data: mockHealth });

    await getPortfolioHealth();

    expect(mockGet).toHaveBeenCalledWith("/analytics/ml/portfolio-health", {
      params: { platform: "tiktok" },
    });
  });

  it("uses custom platform parameter", async () => {
    mockGet.mockResolvedValue({ data: mockHealth });

    await getPortfolioHealth("shopee");

    expect(mockGet).toHaveBeenCalledWith("/analytics/ml/portfolio-health", {
      params: { platform: "shopee" },
    });
  });

  it("throws when response.data is null/undefined", async () => {
    mockGet.mockResolvedValue({ data: null });

    await expect(getPortfolioHealth()).rejects.toThrow(
      "No data received from server",
    );
  });
});

describe("getMLProducts", () => {
  const mockProducts = [
    {
      product_id: "P1",
      product_name: "Product 1",
      sku: "SKU-1",
      unified_score: 75,
    },
  ];
  const mockMeta = { total: 1, limit: 20, has_more: false, next_cursor: null };

  it("returns products and meta", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: mockProducts, meta: mockMeta },
    });

    const result = await getMLProducts();

    expect(result.products).toEqual(mockProducts);
    expect(result.meta).toEqual(mockMeta);
  });

  it("passes query params to client.get", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: [], meta: mockMeta },
    });

    await getMLProducts({ platform: "tiktok", limit: 10, sort_by: "score" });

    expect(mockClientGet).toHaveBeenCalledWith("/analytics/ml/products", {
      params: { platform: "tiktok", limit: 10, sort_by: "score" },
    });
  });

  it("returns empty products array when data.data is absent", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: undefined, meta: mockMeta },
    });

    const result = await getMLProducts();

    expect(result.products).toEqual([]);
  });

  it("throws when data.success is false", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: false },
    });

    await expect(getMLProducts()).rejects.toThrow(
      "No data received from server",
    );
  });
});

describe("getMLAlerts", () => {
  const mockAlerts = [
    {
      id: "A1",
      tenant_id: "test-tenant",
      product_id: "P1",
      product_name: "Product 1",
      alert_type: "fatigue",
      severity: "high",
      message: "Ad fatigue detected",
      created_at: "2026-01-01T00:00:00Z",
      status: "active",
    },
  ];
  const mockMeta = {
    total_active: 1,
    high_priority: 1,
    medium_priority: 0,
    low_priority: 0,
  };

  it("returns alerts and meta", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: mockAlerts, meta: mockMeta },
    });

    const result = await getMLAlerts();

    expect(result.alerts).toEqual(mockAlerts);
    expect(result.meta).toEqual(mockMeta);
  });

  it("calls /analytics/ml/alerts endpoint", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: [], meta: mockMeta },
    });

    await getMLAlerts();

    expect(mockClientGet).toHaveBeenCalledWith("/analytics/ml/alerts");
  });

  it("returns empty alerts array when data.data is absent", async () => {
    mockClientGet.mockResolvedValue({
      data: { success: true, data: undefined, meta: mockMeta },
    });

    const result = await getMLAlerts();

    expect(result.alerts).toEqual([]);
  });

  it("throws when data.success is false", async () => {
    mockClientGet.mockResolvedValue({ data: { success: false } });

    await expect(getMLAlerts()).rejects.toThrow("No data received from server");
  });
});

describe("getScoreDistribution", () => {
  it("returns score distribution array", async () => {
    const distribution = [
      { category: "star", count: 10, percentage: 20 },
      { category: "growth", count: 15, percentage: 30 },
    ];
    mockGet.mockResolvedValue({ data: distribution });

    const result = await getScoreDistribution();

    expect(result).toEqual(distribution);
  });

  it("calls /analytics/ml/distribution endpoint", async () => {
    mockGet.mockResolvedValue({ data: [] });

    await getScoreDistribution();

    expect(mockGet).toHaveBeenCalledWith("/analytics/ml/distribution");
  });

  it("throws when response.data is null", async () => {
    mockGet.mockResolvedValue({ data: null });

    await expect(getScoreDistribution()).rejects.toThrow(
      "No data received from server",
    );
  });
});

import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getUnifiedAnalytics,
  getClassifiedProducts,
  getProductsFromAds,
  runSimulation,
} from "./analyticsIntelligence";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
  },
}));

vi.mock("@/lib/logger", () => ({
  logger: { warn: vi.fn(), error: vi.fn() },
}));

describe("analyticsIntelligence API", () => {
  beforeEach(() => vi.clearAllMocks());

  describe("getUnifiedAnalytics", () => {
    it("returns data on success", async () => {
      const data = { summary: {}, kpi: {}, action_counts: {} };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getUnifiedAnalytics();
      expect(mockGet).toHaveBeenCalledWith("/analytics/unified/summary");
      expect(result).toEqual(data);
    });

    it("throws when success is false", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Not authorized" });
      await expect(getUnifiedAnalytics()).rejects.toThrow("Not authorized");
    });

    it("throws when data is null", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getUnifiedAnalytics()).rejects.toThrow(
        "Failed to fetch unified analytics",
      );
    });
  });

  describe("getClassifiedProducts", () => {
    it("returns data on success", async () => {
      const data = { scale_up: [], maintain: [], reduce: [], stop: [] };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getClassifiedProducts();
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Server error" });
      await expect(getClassifiedProducts()).rejects.toThrow("Server error");
    });
  });

  describe("getProductsFromAds", () => {
    it("returns product list on success", async () => {
      const data = [
        {
          product_id: "p1",
          product_name: "Test",
          avg_roas: 2.5,
          source: "shopee",
        },
      ];
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getProductsFromAds();
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Failed" });
      await expect(getProductsFromAds()).rejects.toThrow("Failed");
    });
  });

  describe("runSimulation", () => {
    it("returns simulation result on success", async () => {
      const data = {
        feasibility: "ACHIEVABLE",
        confidence_percent: 80,
        current_roas: 2.0,
        projected_roas: 3.0,
        trend_prediction: "UP",
        optimal_budget: 1000,
        recommendation: "Increase budget",
        alternatives: [],
      };
      mockPost.mockResolvedValue({ success: true, data });
      const req = {
        product_id: "p1",
        target_roas: 3.0,
        budget_per_day: 100,
        period_days: 30,
      };
      const result = await runSimulation(req);
      expect(mockPost).toHaveBeenCalledWith(
        "/analytics/simulation/calculate",
        req,
      );
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({
        success: false,
        error: "Simulation failed",
      });
      await expect(
        runSimulation({
          product_id: "p1",
          target_roas: 3,
          budget_per_day: 100,
          period_days: 30,
        }),
      ).rejects.toThrow("Simulation failed");
    });
  });

});

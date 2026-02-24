import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getUnifiedAnalytics,
  getClassifiedProducts,
  getProductsFromAds,
  runSimulation,
  generateReport,
  getReports,
  getReportHTML,
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

  describe("generateReport", () => {
    it("returns job on success", async () => {
      const data = {
        id: 1,
        tenant_id: "t1",
        job_type: "report",
        platform: "shopee",
        status: "pending",
        progress: 0,
        created_at: "2024-01-01",
      };
      mockPost.mockResolvedValue({ success: true, data });
      const result = await generateReport({
        platform: "shopee",
        report_type: "full",
      });
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({ success: false, error: "Queue full" });
      await expect(generateReport({ platform: "tiktok" })).rejects.toThrow(
        "Queue full",
      );
    });
  });

  describe("getReports", () => {
    it("returns reports on success", async () => {
      const data = { reports: [{ id: 1 }], total: 1 };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getReports("shopee", 1, 20);
      expect(mockGet).toHaveBeenCalledWith(
        "/analytics/ml/reports/shopee/list",
        { params: { page: 1, limit: 20 } },
      );
      expect(result).toEqual(data);
    });

    it("returns empty list on 404", async () => {
      const axiosError = Object.assign(new Error("Not Found"), {
        response: { status: 404 },
      });
      mockGet.mockRejectedValue(axiosError);
      const result = await getReports("tiktok");
      expect(result).toEqual({ reports: [], total: 0 });
    });

    it("returns null-safe default when data is null", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      const result = await getReports("shopee");
      expect(result).toEqual({ reports: [], total: 0 });
    });

    it("re-throws non-404 errors", async () => {
      const axiosError = Object.assign(new Error("Server Error"), {
        response: { status: 500 },
      });
      mockGet.mockRejectedValue(axiosError);
      await expect(getReports("shopee")).rejects.toThrow("Server Error");
    });

    it("throws when success is false", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Unauthorized" });
      await expect(getReports("shopee")).rejects.toThrow("Unauthorized");
    });

    it("uses default params", async () => {
      mockGet.mockResolvedValue({
        success: true,
        data: { reports: [], total: 0 },
      });
      await getReports("shopee");
      expect(mockGet).toHaveBeenCalledWith(
        "/analytics/ml/reports/shopee/list",
        { params: { page: 1, limit: 20 } },
      );
    });
  });

  describe("getReportHTML", () => {
    it("returns html on success", async () => {
      mockGet.mockResolvedValue({
        success: true,
        data: { html: "<html>test</html>" },
      });
      const result = await getReportHTML("shopee", "report-2024.html");
      expect(mockGet).toHaveBeenCalledWith(
        "/analytics/ml/reports/shopee/report-2024.html",
      );
      expect(result).toBe("<html>test</html>");
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Not found" });
      await expect(getReportHTML("tiktok", "file.html")).rejects.toThrow(
        "Not found",
      );
    });
  });
});

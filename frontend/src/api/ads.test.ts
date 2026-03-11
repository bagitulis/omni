import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getShopeeAdsDashboard,
  getShopeeAdsData,
  getTiktokAdsDashboard,
  getTiktokAdsData,
  getShopeeAdsReports,
  getTiktokAdsReports,
} from "./ads";

const { mockGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

describe("ads API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getShopeeAdsDashboard", () => {
    it("returns dashboard data on success", async () => {
      const mockData = { success: true, data: { total_spend: 100 } };
      mockGet.mockResolvedValueOnce(mockData);
      const result = await getShopeeAdsDashboard();
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee-ads/dashboard");
      expect(result).toEqual(mockData);
    });

    it("propagates errors from api client", async () => {
      mockGet.mockRejectedValueOnce(new Error("Network error"));
      await expect(getShopeeAdsDashboard()).rejects.toThrow("Network error");
    });
  });

  describe("getShopeeAdsData", () => {
    it("returns items and pagination from nested response", async () => {
      const items = [{ product_id: "P1" }];
      const pagination = { total: 1, limit: 50, offset: 0 };
      mockGet.mockResolvedValueOnce({
        success: true,
        data: { items, pagination },
      });
      const result = await getShopeeAdsData({ limit: 50 });
      expect(result.success).toBe(true);
      expect(result.data).toEqual(items);
      expect(result.pagination).toEqual(pagination);
    });

    it("returns empty data and default pagination when data is null", async () => {
      mockGet.mockResolvedValueOnce({ success: true, data: null });
      const result = await getShopeeAdsData();
      expect(result.data).toEqual([]);
      expect(result.pagination).toEqual({ total: 0, limit: 50, offset: 0 });
    });

    it("passes params to api client", async () => {
      mockGet.mockResolvedValueOnce({ success: true, data: null });
      const params = {
        limit: 10,
        offset: 20,
        orderBy: "spend",
        orderDir: "desc" as const,
      };
      await getShopeeAdsData(params);
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee-ads/data", {
        params,
      });
    });

    it("handles flat array response (legacy)", async () => {
      const flatData = [{ product_id: "P2" }];
      mockGet.mockResolvedValueOnce({ success: true, data: flatData });
      const result = await getShopeeAdsData();
      expect(result.data).toEqual(flatData);
    });
  });

  describe("getTiktokAdsDashboard", () => {
    it("returns dashboard data on success", async () => {
      const mockData = { success: true, data: { total_spend: 200 } };
      mockGet.mockResolvedValueOnce(mockData);
      const result = await getTiktokAdsDashboard();
      expect(mockGet).toHaveBeenCalledWith("/analytics/tiktok-ads/dashboard");
      expect(result).toEqual(mockData);
    });

    it("propagates errors", async () => {
      mockGet.mockRejectedValueOnce(new Error("Tiktok error"));
      await expect(getTiktokAdsDashboard()).rejects.toThrow("Tiktok error");
    });
  });

  describe("getTiktokAdsData", () => {
    it("returns items and pagination from nested response", async () => {
      const items = [{ creative_id: "C1" }];
      const pagination = { total: 1, limit: 50, offset: 0 };
      mockGet.mockResolvedValueOnce({
        success: true,
        data: { items, pagination },
      });
      const result = await getTiktokAdsData({ limit: 50 });
      expect(result.data).toEqual(items);
      expect(result.pagination).toEqual(pagination);
    });

    it("returns default pagination when data missing", async () => {
      mockGet.mockResolvedValueOnce({ success: true, data: null });
      const result = await getTiktokAdsData();
      expect(result.pagination).toEqual({ total: 0, limit: 50, offset: 0 });
    });
  });

  describe("getShopeeAdsReports", () => {
    it("returns reports list", async () => {
      const mockData = { success: true, data: [{ filename: "report.csv" }] };
      mockGet.mockResolvedValueOnce(mockData);
      const result = await getShopeeAdsReports();
      expect(mockGet).toHaveBeenCalledWith("/analytics/ml/reports/shopee/list");
      expect(result).toEqual(mockData);
    });
  });

  describe("getTiktokAdsReports", () => {
    it("returns reports list", async () => {
      const mockData = {
        success: true,
        data: [{ filename: "tiktok-report.csv" }],
      };
      mockGet.mockResolvedValueOnce(mockData);
      const result = await getTiktokAdsReports();
      expect(mockGet).toHaveBeenCalledWith("/analytics/ml/reports/tiktok/list");
      expect(result).toEqual(mockData);
    });
  });
});

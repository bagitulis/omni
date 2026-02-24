import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getUnifiedKPI,
  getUnifiedSummary,
  getSettings,
  saveSettings,
  getSyncStatus,
  syncEscrow,
  deleteSyncData,
  getReconciliation,
  getShippingFee,
  getJobStatus,
} from "./analytics";

const { mockGet, mockPost, mockDelete, mockClientPost, mockClientGet } =
  vi.hoisted(() => ({
    mockGet: vi.fn(),
    mockPost: vi.fn(),
    mockDelete: vi.fn(),
    mockClientPost: vi.fn(),
    mockClientGet: vi.fn(),
  }));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    delete: mockDelete,
    client: {
      post: mockClientPost,
      get: mockClientGet,
    },
  },
}));

describe("analytics API", () => {
  beforeEach(() => vi.clearAllMocks());

  describe("getUnifiedKPI", () => {
    it("returns KPI data", async () => {
      const kpiData = { total_revenue: 100 };
      mockGet.mockResolvedValue({ success: true, data: kpiData });

      const result = await getUnifiedKPI();
      expect(mockGet).toHaveBeenCalledWith("/analytics/unified/kpi");
      expect(result).toEqual(kpiData);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getUnifiedKPI()).rejects.toThrow(
        "Failed to fetch unified KPI",
      );
    });
  });

  describe("getUnifiedSummary", () => {
    it("returns summary data", async () => {
      const summary = { total_orders: 5 };
      mockGet.mockResolvedValue({ success: true, data: summary });

      const result = await getUnifiedSummary();
      expect(result).toEqual(summary);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: undefined });
      await expect(getUnifiedSummary()).rejects.toThrow(
        "Failed to fetch unified summary",
      );
    });
  });

  describe("getSettings", () => {
    it("returns settings for platform", async () => {
      const settings = { cost_per_click: 100 };
      mockGet.mockResolvedValue({ success: true, data: settings });

      const result = await getSettings("shopee");
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee/settings");
      expect(result).toEqual(settings);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getSettings("tiktok")).rejects.toThrow(
        "Failed to fetch analytics settings",
      );
    });
  });

  describe("saveSettings", () => {
    it("resolves on success", async () => {
      mockPost.mockResolvedValue({ success: true });

      const settings = { cost_per_click: 200 };
      await expect(
        saveSettings("shopee", settings as never),
      ).resolves.toBeUndefined();
      expect(mockPost).toHaveBeenCalledWith(
        "/analytics/shopee/settings",
        settings,
      );
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({ success: false, error: "Server error" });
      await expect(saveSettings("shopee", {} as never)).rejects.toThrow(
        "Server error",
      );
    });

    it("throws default message when error is missing", async () => {
      mockPost.mockResolvedValue({ success: false });
      await expect(saveSettings("shopee", {} as never)).rejects.toThrow(
        "Failed to save analytics settings",
      );
    });
  });

  describe("getSyncStatus", () => {
    it("returns sync status", async () => {
      const status = { synced: true };
      mockGet.mockResolvedValue({ success: true, data: status });

      const result = await getSyncStatus("shopee", 1, 2024);
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee/sync-status", {
        params: { month: 1, year: 2024 },
      });
      expect(result).toEqual(status);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getSyncStatus("tiktok", 2, 2024)).rejects.toThrow(
        "Failed to fetch sync status",
      );
    });
  });

  describe("syncEscrow", () => {
    it("returns job_id on success", async () => {
      mockClientPost.mockResolvedValue({
        data: { success: true, data: { job_id: "job-123" } },
      });

      const result = await syncEscrow("shopee", 1, 2024, false);
      expect(mockClientPost).toHaveBeenCalledWith("/analytics/shopee/sync", {
        month: 1,
        year: 2024,
        force_resync: false,
      });
      expect(result).toBe("job-123");
    });

    it("throws on failure", async () => {
      mockClientPost.mockResolvedValue({
        data: { success: false, error: "Sync error" },
      });
      await expect(syncEscrow("shopee", 1, 2024, true)).rejects.toThrow(
        "Sync error",
      );
    });

    it("throws default message when error is missing", async () => {
      mockClientPost.mockResolvedValue({ data: { success: false } });
      await expect(syncEscrow("tiktok", 1, 2024, false)).rejects.toThrow(
        "Failed to start sync",
      );
    });
  });

  describe("deleteSyncData", () => {
    it("resolves on success", async () => {
      mockDelete.mockResolvedValue({ success: true });

      await expect(deleteSyncData("shopee", 1, 2024)).resolves.toBeUndefined();
      expect(mockDelete).toHaveBeenCalledWith("/analytics/shopee/sync", {
        params: { month: 1, year: 2024 },
      });
    });

    it("throws on failure", async () => {
      mockDelete.mockResolvedValue({ success: false, error: "Delete error" });
      await expect(deleteSyncData("shopee", 1, 2024)).rejects.toThrow(
        "Delete error",
      );
    });
  });

  describe("getReconciliation", () => {
    it("returns reconciliation data", async () => {
      const data = { reconciled: 10 };
      mockGet.mockResolvedValue({ success: true, data });

      const result = await getReconciliation("shopee", 1, 2024);
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee/reconciliation", {
        params: { month: 1, year: 2024 },
      });
      expect(result).toEqual(data);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getReconciliation("tiktok", 1, 2024)).rejects.toThrow(
        "Failed to fetch reconciliation data",
      );
    });
  });

  describe("getShippingFee", () => {
    it("returns shipping fee data", async () => {
      const data = { total_fee: 50000 };
      mockGet.mockResolvedValue({ success: true, data });

      const result = await getShippingFee("shopee", 1, 2024);
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee/shipping-fee", {
        params: { month: 1, year: 2024 },
      });
      expect(result).toEqual(data);
    });

    it("throws when data is missing", async () => {
      mockGet.mockResolvedValue({ success: true, data: null });
      await expect(getShippingFee("tiktok", 1, 2024)).rejects.toThrow(
        "Failed to fetch shipping fee data",
      );
    });
  });

  describe("getJobStatus", () => {
    it("returns job progress", async () => {
      const job = { id: "job-1", progress: 80 };
      mockClientGet.mockResolvedValue({
        data: { success: true, job },
      });

      const result = await getJobStatus("job-1");
      expect(mockClientGet).toHaveBeenCalledWith("/jobs/job-1");
      expect(result).toEqual(job);
    });

    it("throws on failure", async () => {
      mockClientGet.mockResolvedValue({
        data: { success: false, error: "Job not found" },
      });
      await expect(getJobStatus("invalid")).rejects.toThrow("Job not found");
    });

    it("throws default message when error is missing", async () => {
      mockClientGet.mockResolvedValue({ data: { success: false } });
      await expect(getJobStatus("job-1")).rejects.toThrow(
        "Failed to fetch job status",
      );
    });
  });
});

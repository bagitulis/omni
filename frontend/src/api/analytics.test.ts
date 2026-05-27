import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getShopeeSettings,
  saveShopeeSettings,
  getShopeeSyncStatus,
  triggerShopeeSync,
  deleteShopeeSync,
  getShopeeReconciliation,
  getShopeeShippingFee,
  getShopeeSkuOrders,
  getShopeeOrderItems,
  repopulateShopeeItems,
  getTiktokSettings,
  saveTiktokSettings,
  getTiktokSyncStatus,
  triggerTiktokSync,
  deleteTiktokSync,
  getTiktokReconciliation,
  getTiktokShippingFee,
  getTiktokSkuOrders,
  getTiktokOrderItems,
  repopulateTiktokItems,
} from "./analytics";

const { mockClientGet, mockClientPost, mockClientDelete } = vi.hoisted(() => ({
  mockClientGet: vi.fn(),
  mockClientPost: vi.fn(),
  mockClientDelete: vi.fn(),
}));

// Mock at the ApiClient method level (analytics.ts calls apiClient.get/post/delete, not .client.get)
vi.mock("./client", () => ({
  default: {
    get: mockClientGet,
    post: mockClientPost,
    delete: mockClientDelete,
  },
}));

// ──────────────────────────────────────────────
//  Shared mock data
// ──────────────────────────────────────────────
const mockReportSettings = {
  price_column: "price",
  formula_deduction: 0,
  formula_multiplier: 1,
};

const mockSyncStatus = {
  synced: true,
  total_orders: 10,
  failed_orders: 0,
  synced_at: "2026-05-01T00:00:00Z",
};

const mockSyncResult = { job_id: "job-123" };

const mockReconciliationResult = {
  summary: {
    total_sku: 10,
    total_transactions: 50,
    sku_ok: 10,
    sku_with_price_diff: 0,
    sku_no_inventory: 0,
  },
  sku_groups: [],
};

const mockShopeeShippingFeeResult = {
  summary: {
    total_orders: 10,
    orders_with_difference: 2,
    total_profit: 5000,
    total_loss: 1000,
    net_impact: 4000,
  },
  details: [],
};

const mockTiktokShippingFeeResult = {
  summary: {
    total_orders: 5,
    orders_with_difference: 1,
    total_profit: 3000,
    total_loss: 500,
    net_impact: 2500,
  },
  details: [],
};

// ──────────────────────────────────────────────
//  Shopee Analytics (8 functions)
// ──────────────────────────────────────────────
describe("Shopee Analytics API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getShopeeSettings", () => {
    it("calls GET /analytics/shopee/settings and returns data", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockReportSettings,
      });

      const result = await getShopeeSettings();

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/settings",
      );
      expect(result).toEqual(mockReportSettings);
    });

    it("throws error when response has success: false", async () => {
      mockClientGet.mockResolvedValue({
        success: false,
        error: "Fetch failed",
      });

      await expect(getShopeeSettings()).rejects.toThrow("Fetch failed");
    });
  });

  describe("saveShopeeSettings", () => {
    it("calls POST /analytics/shopee/settings with settings body", async () => {
      mockClientPost.mockResolvedValue({
        success: true,
        data: mockReportSettings,
      });

      const settings = { price_column: "cost" };
      const result = await saveShopeeSettings(settings);

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/shopee/settings",
        settings,
      );
      expect(result).toEqual(mockReportSettings);
    });

    it("throws error on failed save", async () => {
      mockClientPost.mockResolvedValue({
        success: false,
        error: "Save failed",
      });

      await expect(saveShopeeSettings({})).rejects.toThrow("Save failed");
    });
  });

  describe("getShopeeSyncStatus", () => {
    it("calls GET /analytics/shopee/sync-status with month/year params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockSyncStatus,
      });

      const result = await getShopeeSyncStatus(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/sync-status",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockSyncStatus);
    });
  });

  describe("triggerShopeeSync", () => {
    it("calls POST /analytics/shopee/sync with request body", async () => {
      mockClientPost.mockResolvedValue({
        success: true,
        data: mockSyncResult,
      });

      const req = { month: 3, year: 2026, force_resync: true };
      const result = await triggerShopeeSync(req);

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/shopee/sync",
        req,
      );
      expect(result).toEqual(mockSyncResult);
    });

    it("throws error on failed trigger", async () => {
      mockClientPost.mockResolvedValue({
        success: false,
        error: "Trigger failed",
      });

      await expect(
        triggerShopeeSync({ month: 3, year: 2026, force_resync: false }),
      ).rejects.toThrow("Trigger failed");
    });
  });

  describe("deleteShopeeSync", () => {
    it("calls DELETE /analytics/shopee/sync with month/year params", async () => {
      mockClientDelete.mockResolvedValue({ success: true });

      await deleteShopeeSync(4, 2026);

      expect(mockClientDelete).toHaveBeenCalledWith(
        "/analytics/shopee/sync",
        expect.objectContaining({ params: { month: 4, year: 2026 } }),
      );
    });

    it("throws on failed delete", async () => {
      mockClientDelete.mockResolvedValue({
        success: false,
        error: "Delete failed",
      });

      await expect(deleteShopeeSync(1, 2026)).rejects.toThrow("Delete failed");
    });
  });

  describe("getShopeeReconciliation", () => {
    it("calls GET /analytics/shopee/reconciliation with params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockReconciliationResult,
      });

      const result = await getShopeeReconciliation(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/reconciliation",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockReconciliationResult);
    });
  });

  describe("getShopeeShippingFee", () => {
    it("calls GET /analytics/shopee/shipping-fee with params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockShopeeShippingFeeResult,
      });

      const result = await getShopeeShippingFee(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/shipping-fee",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockShopeeShippingFeeResult);
    });
  });

  describe("Shopee detail endpoints", () => {
    it("calls SKU orders endpoint with sku/month/year params", async () => {
      const data = { orders: [] };
      mockClientGet.mockResolvedValue({ success: true, data });

      const result = await getShopeeSkuOrders({ sku: "SKU-001", month: 3, year: 2026 });

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/sku-orders",
        expect.objectContaining({ params: { sku: "SKU-001", month: 3, year: 2026 } }),
      );
      expect(result).toEqual(data);
    });

    it("calls order items endpoint with order_sn/month/year params", async () => {
      const data = { items: [] };
      mockClientGet.mockResolvedValue({ success: true, data });

      const result = await getShopeeOrderItems({ order_sn: "ORD-001", month: 3, year: 2026 });

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/shopee/order-items",
        expect.objectContaining({ params: { order_sn: "ORD-001", month: 3, year: 2026 } }),
      );
      expect(result).toEqual(data);
    });
  });

  describe("repopulateShopeeItems", () => {
    it("calls POST /analytics/shopee/repopulate-items without period", async () => {
      mockClientPost.mockResolvedValue({ success: true });

      await repopulateShopeeItems();

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/shopee/repopulate-items",
        null,
        expect.any(Object),
      );
    });

    it("calls POST /analytics/shopee/repopulate-items with period param", async () => {
      mockClientPost.mockResolvedValue({ success: true });

      await repopulateShopeeItems("2026-03");

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/shopee/repopulate-items",
        null,
        expect.objectContaining({ params: { period: "2026-03" } }),
      );
    });

    it("throws on failure", async () => {
      mockClientPost.mockResolvedValue({
        success: false,
        error: "Repopulate failed",
      });

      await expect(repopulateShopeeItems()).rejects.toThrow("Repopulate failed");
    });
  });
});

// ──────────────────────────────────────────────
//  TikTok Analytics (8 functions)
// ──────────────────────────────────────────────
describe("TikTok Analytics API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getTiktokSettings", () => {
    it("calls GET /analytics/tiktok/settings and returns data", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockReportSettings,
      });

      const result = await getTiktokSettings();

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/settings",
      );
      expect(result).toEqual(mockReportSettings);
    });

    it("throws error on failed response", async () => {
      mockClientGet.mockResolvedValue({
        success: false,
        error: "TikTok fetch failed",
      });

      await expect(getTiktokSettings()).rejects.toThrow("TikTok fetch failed");
    });
  });

  describe("saveTiktokSettings", () => {
    it("calls POST /analytics/tiktok/settings with settings body", async () => {
      mockClientPost.mockResolvedValue({
        success: true,
        data: mockReportSettings,
      });

      const settings = { formula_multiplier: 2 };
      const result = await saveTiktokSettings(settings);

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/tiktok/settings",
        settings,
      );
      expect(result).toEqual(mockReportSettings);
    });
  });

  describe("getTiktokSyncStatus", () => {
    it("calls GET /analytics/tiktok/sync-status with month/year params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockSyncStatus,
      });

      const result = await getTiktokSyncStatus(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/sync-status",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockSyncStatus);
    });
  });

  describe("triggerTiktokSync", () => {
    it("calls POST /analytics/tiktok/sync with request body", async () => {
      mockClientPost.mockResolvedValue({
        success: true,
        data: mockSyncResult,
      });

      const req = { month: 3, year: 2026, force_resync: false };
      const result = await triggerTiktokSync(req);

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/tiktok/sync",
        req,
      );
      expect(result).toEqual(mockSyncResult);
    });
  });

  describe("deleteTiktokSync", () => {
    it("calls DELETE /analytics/tiktok/sync with month/year params", async () => {
      mockClientDelete.mockResolvedValue({ success: true });

      await deleteTiktokSync(4, 2026);

      expect(mockClientDelete).toHaveBeenCalledWith(
        "/analytics/tiktok/sync",
        expect.objectContaining({ params: { month: 4, year: 2026 } }),
      );
    });
  });

  describe("getTiktokReconciliation", () => {
    it("calls GET /analytics/tiktok/reconciliation with params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockReconciliationResult,
      });

      const result = await getTiktokReconciliation(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/reconciliation",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockReconciliationResult);
    });
  });

  describe("getTiktokShippingFee", () => {
    it("calls GET /analytics/tiktok/shipping-fee with params", async () => {
      mockClientGet.mockResolvedValue({
        success: true,
        data: mockTiktokShippingFeeResult,
      });

      const result = await getTiktokShippingFee(3, 2026);

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/shipping-fee",
        expect.objectContaining({ params: { month: 3, year: 2026 } }),
      );
      expect(result).toEqual(mockTiktokShippingFeeResult);
    });
  });

  describe("TikTok detail endpoints", () => {
    it("calls SKU orders endpoint with sku/month/year params", async () => {
      const data = { orders: [] };
      mockClientGet.mockResolvedValue({ success: true, data });

      const result = await getTiktokSkuOrders({ sku: "T-SKU-001", month: 4, year: 2025 });

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/sku-orders",
        expect.objectContaining({ params: { sku: "T-SKU-001", month: 4, year: 2025 } }),
      );
      expect(result).toEqual(data);
    });

    it("calls order items endpoint with order_sn/month/year params", async () => {
      const data = { items: [] };
      mockClientGet.mockResolvedValue({ success: true, data });

      const result = await getTiktokOrderItems({ order_sn: "TK-001", month: 4, year: 2025 });

      expect(mockClientGet).toHaveBeenCalledWith(
        "/analytics/tiktok/order-items",
        expect.objectContaining({ params: { order_sn: "TK-001", month: 4, year: 2025 } }),
      );
      expect(result).toEqual(data);
    });
  });

  describe("repopulateTiktokItems", () => {
    it("calls POST /analytics/tiktok/repopulate-items without period", async () => {
      mockClientPost.mockResolvedValue({ success: true });

      await repopulateTiktokItems();

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/tiktok/repopulate-items",
        null,
        expect.any(Object),
      );
    });

    it("calls POST /analytics/tiktok/repopulate-items with period param", async () => {
      mockClientPost.mockResolvedValue({ success: true });

      await repopulateTiktokItems("2026-03");

      expect(mockClientPost).toHaveBeenCalledWith(
        "/analytics/tiktok/repopulate-items",
        null,
        expect.objectContaining({ params: { period: "2026-03" } }),
      );
    });

    it("throws on failure", async () => {
      mockClientPost.mockResolvedValue({
        success: false,
        error: "TikTok repopulate failed",
      });

      await expect(repopulateTiktokItems()).rejects.toThrow(
        "TikTok repopulate failed",
      );
    });
  });
});

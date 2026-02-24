import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getDashboardData,
  getOrderAnalytics,
  getRevenueAnalytics,
  getWalletData,
  getShippingFeeData,
  getSyncStatus,
} from "./dashboard";

const { mockGet, mockClientGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockClientGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    client: { get: mockClientGet },
  },
}));

describe("dashboard API", () => {
  beforeEach(() => vi.clearAllMocks());

  describe("getDashboardData", () => {
    it("returns combined dashboard data on success", async () => {
      // getAnalytics uses apiClient.get
      const analyticsData = { total_revenue: 1000, total_orders: 50 };
      mockGet.mockImplementation((url: string) => {
        if (url === "/analytics/dashboard")
          return Promise.resolve({ success: true, data: analyticsData });
        return Promise.resolve({ success: true, data: {} });
      });

      // getRecentOrders, getPendingCount, getReadyToShipCount use apiClient.client.get
      const ordersResponse = {
        data: {
          success: true,
          data: [
            {
              order_no: "ORDER1",
              order_sn: "ORDER1",
              status: "READY_TO_SHIP",
              order_status: "READY_TO_SHIP",
              platform: "Shopee",
              total_amount: 100,
              buyer_username: "buyer1",
              created_at: "2024-01-01",
            },
          ],
        },
      };
      mockClientGet.mockImplementation((url: string) => {
        if (url === "/orders/unprocess") return Promise.resolve(ordersResponse);
        if (url === "/orders/unpaid")
          return Promise.resolve({ data: { count: 5 } });
        return Promise.resolve({ data: { count: 3 } });
      });

      const result = await getDashboardData();
      expect(result).toHaveProperty("analytics");
      expect(result).toHaveProperty("recent_orders");
      expect(result).toHaveProperty("orders_pending");
      expect(result).toHaveProperty("ready_to_ship");
      expect(result.orders_pending).toBe(5);
    });

    it("returns 0 counts for pending/readyToShip when only those client.get calls throw", async () => {
      const analyticsData = { total_revenue: 500 };
      mockGet.mockResolvedValue({ success: true, data: analyticsData });
      // getRecentOrders has no try/catch — mock to return success with empty data
      // getPendingCount and getReadyToShipCount have try/catch and return 0 on error
      mockClientGet.mockImplementation((url: string) => {
        if (url === "/orders/unprocess") {
          return Promise.resolve({ data: { success: true, data: [] } });
        }
        // /orders/unpaid throws — getPendingCount catches and returns 0
        return Promise.reject(new Error("Network error"));
      });
      const result = await getDashboardData();
      expect(result.orders_pending).toBe(0);
      expect(result.ready_to_ship).toBe(0);
    });

    it("returns empty recent_orders when data.success is false", async () => {
      mockGet.mockResolvedValue({ success: true, data: {} });
      mockClientGet.mockResolvedValue({ data: { success: false } });
      const result = await getDashboardData();
      expect(result.recent_orders).toEqual([]);
    });
  });
  describe("getOrderAnalytics", () => {
    it("returns analytics data on success", async () => {
      const data = { orders_by_day: [] };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getOrderAnalytics({
        start_date: "2024-01-01",
        end_date: "2024-01-31",
      });
      expect(mockGet).toHaveBeenCalledWith("/analytics/orders", {
        params: { start_date: "2024-01-01", end_date: "2024-01-31" },
      });
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Failed" });
      await expect(getOrderAnalytics()).rejects.toThrow("Failed");
    });
  });

  describe("getRevenueAnalytics", () => {
    it("returns revenue data on success", async () => {
      const data = { revenue_by_day: [] };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getRevenueAnalytics();
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Revenue error" });
      await expect(getRevenueAnalytics()).rejects.toThrow("Revenue error");
    });
  });

  describe("getWalletData", () => {
    it("returns wallet data for shopee", async () => {
      const data = {
        total_balance: 500,
        pending_balance: 100,
        available_balance: 400,
        currency: "IDR",
        updated_at: "2024-01-01",
      };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getWalletData("shopee");
      expect(mockGet).toHaveBeenCalledWith("/shopee/wallet/balance");
      expect(result).toEqual(data);
    });

    it("throws for shopee on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Wallet error" });
      await expect(getWalletData("shopee")).rejects.toThrow("Wallet error");
    });

    it("returns default for non-shopee platforms", async () => {
      const result = await getWalletData("lazada");
      expect(result.total_balance).toBe(0);
      expect(result.currency).toBe("IDR");
      expect(mockGet).not.toHaveBeenCalled();
    });
  });

  describe("getShippingFeeData", () => {
    it("returns shipping fee data on success", async () => {
      const data = {
        total_orders: 10,
        orders_with_difference: 3,
        total_profit: 500,
        total_loss: 100,
        net_impact: 400,
      };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getShippingFeeData("shopee");
      expect(mockGet).toHaveBeenCalledWith("/analytics/shopee/shipping-fee");
      expect(result).toEqual(data);
    });

    it("returns default empty data on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Not found" });
      const result = await getShippingFeeData("lazada");
      expect(result).toEqual({
        total_orders: 0,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      });
    });
  });

  describe("getSyncStatus", () => {
    it("returns sync status on success", async () => {
      const data = {
        platform: "shopee",
        status: "synced",
        last_sync: "2024-01-01",
      };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getSyncStatus("shopee");
      expect(result).toEqual(data);
    });

    it("returns error status on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Sync error" });
      const result = await getSyncStatus("tiktok");
      expect(result.status).toBe("error");
      expect(result.platform).toBe("tiktok");
      expect(result.details).toBe("Sync error");
    });
  });
});

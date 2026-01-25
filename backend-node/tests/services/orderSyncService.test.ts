/**
 * OrderSyncService Unit Tests
 * Tests order synchronization across multiple platforms
 */

import {
  OrderSyncService,
  getOrderSyncService,
  clearOrderSyncInstances,
} from "../../src/services/orderSyncService";

// Mock dependencies
jest.mock("../../src/services/platformCoordinationService");
jest.mock("../../src/services/orders/orderRepository");
jest.mock("../../src/services/orders/shopeeOrderManager");
jest.mock("../../src/services/orders/lazadaOrderManager");
jest.mock("../../src/services/orders/tiktokOrderManager");

const TEST_TENANT_ID = "test_tenant";

describe("OrderSyncService", () => {
  let orderSyncService: OrderSyncService;
  let mockPlatformCoordinationService: any;
  let mockShopeeManager: any;
  let mockLazadaManager: any;
  let mockTiktokManager: any;

  beforeEach(() => {
    // Clear per-tenant instances
    clearOrderSyncInstances();

    orderSyncService = getOrderSyncService(TEST_TENANT_ID);

    // Setup mocks
    mockShopeeManager = {
      getOrderList: jest.fn(),
      getOrderDetails: jest.fn(),
    };

    mockLazadaManager = {
      getOrderList: jest.fn(),
      getOrderDetails: jest.fn(),
    };

    mockTiktokManager = {
      getOrderList: jest.fn(),
      getOrderDetails: jest.fn(),
    };

    mockPlatformCoordinationService = {
      initializePlatforms: jest.fn(),
      getClient: jest.fn(),
      reloadPlatformConfig: jest.fn(),
    };

    // Inject mocks
    (orderSyncService as any).orderManagers = new Map([
      ["shopee", mockShopeeManager],
      ["lazada", mockLazadaManager],
      ["tiktok", mockTiktokManager],
    ]);

    (orderSyncService as any).platformCoordinationService =
      mockPlatformCoordinationService;

    jest.clearAllMocks();
  });

  describe("Per-Tenant Pattern", () => {
    it("should return same instance for same tenant", () => {
      const instance1 = getOrderSyncService(TEST_TENANT_ID);
      const instance2 = getOrderSyncService(TEST_TENANT_ID);

      expect(instance1).toBe(instance2);
    });

    it("should return different instance for different tenant", () => {
      const instance1 = getOrderSyncService("tenant_a");
      const instance2 = getOrderSyncService("tenant_b");

      expect(instance1).not.toBe(instance2);
    });
  });

  describe("initialize", () => {
    it("should initialize all platform clients", async () => {
      mockPlatformCoordinationService.initializePlatforms.mockResolvedValue(
        undefined
      );
      mockPlatformCoordinationService.getClient.mockImplementation(
        (platform: string) => {
          return { platform };
        }
      );

      await orderSyncService.initialize();

      expect(
        mockPlatformCoordinationService.initializePlatforms
      ).toHaveBeenCalled();
      expect(mockPlatformCoordinationService.getClient).toHaveBeenCalledWith(
        "shopee"
      );
      expect(mockPlatformCoordinationService.getClient).toHaveBeenCalledWith(
        "lazada"
      );
      expect(mockPlatformCoordinationService.getClient).toHaveBeenCalledWith(
        "tiktok"
      );
    });

    it("should throw error if initialization fails", async () => {
      mockPlatformCoordinationService.initializePlatforms.mockRejectedValue(
        new Error("Platform init failed")
      );

      await expect(orderSyncService.initialize()).rejects.toThrow(
        "Platform init failed"
      );
    });
  });

  describe("syncByCategory", () => {
    const mockShopeeOrders = [
      {
        order_sn: "SHOPEE123",
        order_status: "UNPAID",
        items: [{ item_id: 1, quantity: 2 }],
      },
      {
        order_sn: "SHOPEE456",
        order_status: "UNPAID",
        items: [{ item_id: 2, quantity: 1 }],
      },
    ];

    const mockLazadaOrders = [
      {
        order_id: "LAZADA789",
        status: "unpaid",
        items: [{ sku: "LAZ001", quantity: 3 }],
      },
    ];

    beforeEach(() => {
      mockShopeeManager.getOrderList.mockResolvedValue(mockShopeeOrders);
      mockLazadaManager.getOrderList.mockResolvedValue(mockLazadaOrders);
      mockTiktokManager.getOrderList.mockResolvedValue([]);
      mockPlatformCoordinationService.reloadPlatformConfig.mockResolvedValue(
        undefined
      );
    });

    it("should sync unpaid orders for all platforms", async () => {
      const result = await orderSyncService.syncByCategory("unpaid", 7);

      expect(result).toHaveProperty("shopee");
      expect(result).toHaveProperty("lazada");
      expect(result).toHaveProperty("tiktok");
      expect(
        mockPlatformCoordinationService.reloadPlatformConfig
      ).toHaveBeenCalledTimes(3);
      expect(mockShopeeManager.getOrderList).toHaveBeenCalledWith("UNPAID", 7);
      expect(mockLazadaManager.getOrderList).toHaveBeenCalledWith("unpaid", 7);
      expect(mockTiktokManager.getOrderList).toHaveBeenCalledWith("UNPAID", 7);
    });

    it("should sync unprocess orders with correct status mapping", async () => {
      await orderSyncService.syncByCategory("unprocess", 3);

      expect(mockShopeeManager.getOrderList).toHaveBeenCalledWith(
        "READY_TO_SHIP",
        3
      );
      expect(mockLazadaManager.getOrderList).toHaveBeenCalledWith("topack", 3);
      expect(mockTiktokManager.getOrderList).toHaveBeenCalledWith(
        "AWAITING_SHIPMENT",
        3
      );
    });

    it("should sync processed orders with correct status mapping", async () => {
      await orderSyncService.syncByCategory("processed", 14);

      expect(mockShopeeManager.getOrderList).toHaveBeenCalledWith(
        "PROCESSED",
        14
      );
      expect(mockLazadaManager.getOrderList).toHaveBeenCalledWith("toship", 14);
      expect(mockTiktokManager.getOrderList).toHaveBeenCalledWith(
        "AWAITING_COLLECTION",
        14
      );
    });

    it("should only sync specified platforms", async () => {
      await orderSyncService.syncByCategory("unpaid", 7, ["shopee"]);

      expect(mockShopeeManager.getOrderList).toHaveBeenCalled();
      expect(mockLazadaManager.getOrderList).not.toHaveBeenCalled();
      expect(mockTiktokManager.getOrderList).not.toHaveBeenCalled();
    });

    it("should handle individual platform failures gracefully", async () => {
      mockShopeeManager.getOrderList.mockRejectedValue(
        new Error("Shopee API error")
      );

      const result = await orderSyncService.syncByCategory("unpaid", 7);

      expect(result.shopee).toHaveProperty("error");
      expect(result.lazada).not.toHaveProperty("error");
      expect(result.tiktok).not.toHaveProperty("error");
    });

    it("should reload configs before syncing", async () => {
      await orderSyncService.syncByCategory("unpaid", 7);

      expect(
        mockPlatformCoordinationService.reloadPlatformConfig
      ).toHaveBeenCalledWith("shopee");
      expect(
        mockPlatformCoordinationService.reloadPlatformConfig
      ).toHaveBeenCalledWith("lazada");
      expect(
        mockPlatformCoordinationService.reloadPlatformConfig
      ).toHaveBeenCalledWith("tiktok");
    });

    it("should return correct counts in results", async () => {
      const result = await orderSyncService.syncByCategory("unpaid", 7);

      expect(result.shopee).toHaveProperty("count");
      expect(result.shopee.count).toBe(2);
      expect(result.lazada.count).toBe(1);
      expect(result.tiktok.count).toBe(0);
    });
  });

  describe("getOrderManager", () => {
    it("should return correct manager for shopee", () => {
      const manager = (orderSyncService as any).getOrderManager("shopee");
      expect(manager).toBe(mockShopeeManager);
    });

    it("should return correct manager for lazada", () => {
      const manager = (orderSyncService as any).getOrderManager("lazada");
      expect(manager).toBe(mockLazadaManager);
    });

    it("should return correct manager for tiktok", () => {
      const manager = (orderSyncService as any).getOrderManager("tiktok");
      expect(manager).toBe(mockTiktokManager);
    });

    it("should throw error for unknown platform", () => {
      expect(() => {
        (orderSyncService as any).getOrderManager("unknown");
      }).toThrow("Order manager not found for platform: unknown");
    });
  });

  describe("ORDER_STATUS_MAPPINGS", () => {
    it("should have correct Shopee status mappings", async () => {
      const mappings = {
        unpaid: "UNPAID",
        unprocess: "READY_TO_SHIP",
        processed: "PROCESSED",
      };

      // Test through syncByCategory to verify mappings
      await orderSyncService.syncByCategory("unpaid", 7, ["shopee"]);
      expect(mockShopeeManager.getOrderList).toHaveBeenCalledWith(
        mappings.unpaid,
        7
      );
    });

    it("should have correct Lazada status mappings", async () => {
      const mappings = {
        unpaid: "unpaid",
        unprocess: "topack",
        processed: "toship",
      };

      await orderSyncService.syncByCategory("unprocess", 7, ["lazada"]);
      expect(mockLazadaManager.getOrderList).toHaveBeenCalledWith(
        mappings.unprocess,
        7
      );
    });

    it("should have correct TikTok status mappings", async () => {
      const mappings = {
        unpaid: "UNPAID",
        unprocess: "AWAITING_SHIPMENT",
        processed: "AWAITING_COLLECTION",
      };

      await orderSyncService.syncByCategory("processed", 7, ["tiktok"]);
      expect(mockTiktokManager.getOrderList).toHaveBeenCalledWith(
        mappings.processed,
        7
      );
    });
  });

  describe("Error Handling", () => {
    it("should handle empty order list", async () => {
      mockShopeeManager.getOrderList.mockResolvedValue([]);
      mockLazadaManager.getOrderList.mockResolvedValue([]);
      mockTiktokManager.getOrderList.mockResolvedValue([]);

      const result = await orderSyncService.syncByCategory("unpaid", 7);

      expect(result.shopee.count).toBe(0);
      expect(result.lazada.count).toBe(0);
      expect(result.tiktok.count).toBe(0);
    });

    it("should handle null response from platform", async () => {
      mockShopeeManager.getOrderList.mockResolvedValue(null);

      const result = await orderSyncService.syncByCategory("unpaid", 7, [
        "shopee",
      ]);

      expect(result.shopee).toHaveProperty("error");
    });

    it("should handle network timeout", async () => {
      mockShopeeManager.getOrderList.mockRejectedValue(new Error("ETIMEDOUT"));

      const result = await orderSyncService.syncByCategory("unpaid", 7, [
        "shopee",
      ]);

      expect(result.shopee.error).toContain("ETIMEDOUT");
    });
  });
});

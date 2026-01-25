/**
 * Order Sync Service
 * SRP: Facade/Coordinator for order sync operations
 * Delegates to specialized services
 *
 * MULTI-TENANT: Each tenant has its own instance
 */

import { ShopeeOrderManager } from "./orders/shopeeOrderManager";
import { LazadaOrderManager } from "./orders/lazadaOrderManager";
import { TiktokOrderManager } from "./orders/tiktokOrderManager";
import { getPlatformCoordinationService } from "./platformCoordinationService";
import { OrderRepository } from "./orders/orderRepository";
import { OrderSyncOperations } from "./orders/orderSyncOperations";
import { OrderQueryService } from "./orders/orderQueryService";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import {
  PlatformType,
  OrderStatusCategory,
} from "./orders/orderStatusMappings";

const logger = getLogger("OrderSync");

export class OrderSyncService {
  private orderManagers: Map<PlatformType, any> = new Map();
  private orderRepository: OrderRepository;
  private syncOperations!: OrderSyncOperations;
  private queryService!: OrderQueryService;
  private tenantId: string;

  /**
   * @param tenantId - Tenant ID for this service instance
   */
  constructor(tenantId: string) {
    this.tenantId = tenantId;
    this.orderRepository = new OrderRepository();
    this.queryService = new OrderQueryService(this.orderRepository);
  }

  async initialize(): Promise<void> {
    try {
      logger.info(
        `🚀 Initializing OrderSyncService for tenant: ${this.tenantId}...`
      );

      // Get per-tenant platform coordination service
      const platformCoordinationService = getPlatformCoordinationService(
        this.tenantId
      );
      await platformCoordinationService.initializePlatforms();

      const shopeeClient = platformCoordinationService.getClient("shopee");
      this.orderManagers.set("shopee", new ShopeeOrderManager(shopeeClient));

      const lazadaClient = platformCoordinationService.getClient("lazada");
      this.orderManagers.set("lazada", new LazadaOrderManager(lazadaClient));

      const tiktokClient = platformCoordinationService.getClient("tiktok");
      this.orderManagers.set("tiktok", new TiktokOrderManager(tiktokClient));

      // Initialize operations with managers
      this.syncOperations = new OrderSyncOperations(
        this.orderManagers,
        this.orderRepository,
        this.tenantId
      );

      logger.info(
        `✅ OrderSyncService initialized successfully for tenant: ${this.tenantId}`
      );
    } catch (error) {
      logger.error(
        `Failed to initialize OrderSyncService for ${this.tenantId}: ${error}`
      );
      throw error;
    }
  }

  /**
   * Get the tenant ID for this service instance
   */
  getTenantId(): string {
    return this.tenantId;
  }

  /**
   * Check if service is initialized
   */
  isInitialized(): boolean {
    return !!this.syncOperations;
  }

  // Delegate sync operations
  async syncByCategory(
    category: OrderStatusCategory,
    days: number = 7,
    platforms: PlatformType[] = ["shopee", "lazada", "tiktok"]
  ): Promise<Record<string, any>> {
    return this.syncOperations.syncByCategory(category, days, platforms);
  }

  async syncPlatformOrders(
    platform: PlatformType,
    category: OrderStatusCategory,
    days: number = 7
  ): Promise<any[]> {
    return this.syncOperations.syncPlatformOrders(platform, category, days);
  }

  // Delegate query operations
  async getOrdersByCategory(
    category: OrderStatusCategory,
    platform?: PlatformType
  ): Promise<any[]> {
    return this.queryService.getOrdersByCategory(category, platform);
  }

  async getPlatformOrders(
    platform: PlatformType,
    category?: OrderStatusCategory
  ): Promise<any[]> {
    return this.queryService.getPlatformOrders(platform, category);
  }

  async getOrderDetails(
    platform: PlatformType,
    orderIds: string[]
  ): Promise<any[]> {
    try {
      logger.info(
        `🔍 Fetching details for ${orderIds.length} ${platform} orders`
      );
      const manager = this.orderManagers.get(platform);
      if (!manager) throw new Error(`Order manager not found: ${platform}`);
      return await manager.getOrderDetails(orderIds);
    } catch (error) {
      logger.error(`Failed to get order details: ${error}`);
      throw error;
    }
  }

  formatOrdersForExport(
    platform: PlatformType,
    orders: any[]
  ): Record<string, any>[] {
    try {
      const manager = this.orderManagers.get(platform);
      if (!manager) return [];
      return manager.formatOrdersForExport(orders);
    } catch (error) {
      logger.error(`Failed to format orders: ${error}`);
      return [];
    }
  }
}

// Per-tenant instance registry
const tenantInstances: Map<string, OrderSyncService> = new Map();
const initializingPromises: Map<string, Promise<void>> = new Map();

/**
 * Get OrderSyncService for specific tenant
 * Automatically initializes the service if not already initialized
 *
 * @param tenantId - Tenant ID (optional, will use context if not provided)
 * @returns Promise<OrderSyncService> instance for the tenant
 * @throws Error if no tenant ID available
 */
export async function getOrderSyncService(
  tenantId?: string
): Promise<OrderSyncService> {
  const resolvedTenantId =
    tenantId ?? tenantContext.getTenantIdOrNull() ?? undefined;

  if (!resolvedTenantId) {
    throw new Error(
      "OrderSyncService: No tenant ID provided and no tenant context available"
    );
  }

  let instance = tenantInstances.get(resolvedTenantId);
  if (!instance) {
    instance = new OrderSyncService(resolvedTenantId);
    tenantInstances.set(resolvedTenantId, instance);
    logger.info(
      `Created new OrderSyncService instance for tenant: ${resolvedTenantId}`
    );
  }

  // Check if already initialized (syncOperations exists)
  if (!instance.isInitialized()) {
    // Check if initialization is already in progress
    let initPromise = initializingPromises.get(resolvedTenantId);
    if (!initPromise) {
      initPromise = instance.initialize();
      initializingPromises.set(resolvedTenantId, initPromise);
    }
    await initPromise;
    initializingPromises.delete(resolvedTenantId);
  }

  return instance;
}

/**
 * Clear all tenant instances (for testing purposes)
 */
export function clearOrderSyncInstances(): void {
  tenantInstances.clear();
}

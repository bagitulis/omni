/**
 * Order Today Service
 * Handles fetching and storing processed orders with tracking info
 * For "Order Today" tab - orders that need to be shipped today
 */

import { PrismaClient } from "@prisma/client";
import { getPlatformCoordinationService } from "../platformCoordinationService";
import { ShopeeOrderManager } from "./shopeeOrderManager";
import { LazadaOrderManager } from "./lazadaOrderManager";
import { TiktokOrderManager } from "./tiktokOrderManager";
import { getLogger } from "../../utils/logger";

const logger = getLogger("OrderTodayService");

export interface OrderTodayItemData {
  platform: string;
  orderSn: string;
  trackingNo: string;
  courier: string;
  sellerSku: string;
  productName: string;
  variationName: string;
  quantity: number;
}

export class OrderTodayService {
  private prisma: PrismaClient;
  private tenantId: string;

  constructor(prisma: PrismaClient, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Fetch processed orders with tracking from all platforms
   * Returns flattened items ready for display
   */
  async fetchAndSaveOrdersToday(
    days: number = 7
  ): Promise<OrderTodayItemData[]> {
    logger.info(`🚀 Fetching Order Today for tenant: ${this.tenantId}`);
    const allItems: OrderTodayItemData[] = [];

    try {
      const coordination = getPlatformCoordinationService(this.tenantId);
      await coordination.initializePlatforms();

      // Fetch from each platform in parallel
      const [shopeeItems, lazadaItems, tiktokItems] = await Promise.allSettled([
        this.fetchShopeeOrders(coordination, days),
        this.fetchLazadaOrders(coordination, days),
        this.fetchTiktokOrders(coordination, days),
      ]);

      if (shopeeItems.status === "fulfilled") {
        allItems.push(...shopeeItems.value);
      } else {
        logger.warn(`Shopee fetch failed: ${shopeeItems.reason}`);
      }

      if (lazadaItems.status === "fulfilled") {
        allItems.push(...lazadaItems.value);
      } else {
        logger.warn(`Lazada fetch failed: ${lazadaItems.reason}`);
      }

      if (tiktokItems.status === "fulfilled") {
        allItems.push(...tiktokItems.value);
      } else {
        logger.warn(`TikTok fetch failed: ${tiktokItems.reason}`);
      }

      // Save to database
      await this.saveOrderTodayItems(allItems);

      logger.info(`✅ Fetched ${allItems.length} total items for Order Today`);
      return allItems;
    } catch (error) {
      logger.error(`Failed to fetch Order Today: ${error}`);
      throw error;
    }
  }

  private async fetchShopeeOrders(
    coordination: ReturnType<typeof getPlatformCoordinationService>,
    days: number
  ): Promise<OrderTodayItemData[]> {
    try {
      const client = coordination.getClient("shopee");
      if (!client) return [];

      const manager = new ShopeeOrderManager(client);
      const orders = await manager.getProcessedOrdersWithTracking(days);
      return this.transformShopeeOrders(orders);
    } catch (error) {
      logger.error(`Shopee error: ${error}`);
      return [];
    }
  }

  private async fetchLazadaOrders(
    coordination: ReturnType<typeof getPlatformCoordinationService>,
    days: number
  ): Promise<OrderTodayItemData[]> {
    try {
      const client = coordination.getClient("lazada");
      if (!client) return [];

      const manager = new LazadaOrderManager(client);
      const items = await manager.getProcessedOrdersWithTracking(days);
      return this.transformLazadaOrders(items);
    } catch (error) {
      logger.error(`Lazada error: ${error}`);
      return [];
    }
  }

  private async fetchTiktokOrders(
    coordination: ReturnType<typeof getPlatformCoordinationService>,
    days: number
  ): Promise<OrderTodayItemData[]> {
    try {
      const client = coordination.getClient("tiktok");
      if (!client) return [];

      const manager = new TiktokOrderManager(client);
      const items = await manager.getProcessedOrdersWithTracking(days);
      return this.transformTiktokOrders(items);
    } catch (error) {
      logger.error(`TikTok error: ${error}`);
      return [];
    }
  }

  private transformShopeeOrders(orders: any[]): OrderTodayItemData[] {
    const items: OrderTodayItemData[] = [];
    for (const order of orders) {
      for (const item of order.item_list || []) {
        items.push({
          platform: "shopee",
          orderSn: order.order_sn || "",
          trackingNo: order.tracking_number || "",
          courier: order.shipping_carrier || "",
          sellerSku: item.model_sku || item.item_sku || "",
          productName: item.item_name || "",
          variationName: item.model_name || "",
          quantity: item.model_quantity_purchased || 1,
        });
      }
    }
    return items;
  }

  private transformLazadaOrders(items: any[]): OrderTodayItemData[] {
    return items.map((item) => ({
      platform: "lazada",
      orderSn: item.order_sn || "",
      trackingNo: item.tracking_number || "",
      courier: item.shipping_carrier || "",
      sellerSku: item.seller_sku || "",
      productName: item.product_name || "",
      variationName: item.variation_name || "",
      quantity: item.quantity || 1,
    }));
  }

  private transformTiktokOrders(items: any[]): OrderTodayItemData[] {
    return items.map((item) => ({
      platform: "tiktok",
      orderSn: item.order_sn || "",
      trackingNo: item.tracking_number || "",
      courier: item.shipping_carrier || "",
      sellerSku: item.seller_sku || "",
      productName: item.product_name || "",
      variationName: item.variation_name || "",
      quantity: item.quantity || 1,
    }));
  }

  private async saveOrderTodayItems(
    items: OrderTodayItemData[]
  ): Promise<void> {
    // Clear existing items for this tenant (fresh fetch each time)
    await this.prisma.orderTodayItem.deleteMany({
      where: { tenantId: this.tenantId },
    });

    if (items.length === 0) return;

    // Deduplicate items by unique constraint fields to prevent duplicate insertion
    const uniqueMap = new Map<string, OrderTodayItemData>();
    for (const item of items) {
      const key = `${item.platform}|${item.orderSn}|${item.sellerSku || ""}`;
      if (!uniqueMap.has(key)) {
        uniqueMap.set(key, item);
      }
    }
    const uniqueItems = Array.from(uniqueMap.values());

    // Bulk insert new items
    await this.prisma.orderTodayItem.createMany({
      data: uniqueItems.map((item) => ({
        tenantId: this.tenantId,
        platform: item.platform,
        orderSn: item.orderSn,
        trackingNo: item.trackingNo || null,
        courier: item.courier || null,
        sellerSku: item.sellerSku || null,
        productName: item.productName || null,
        variationName: item.variationName || null,
        quantity: item.quantity,
      })),
    });

    logger.info(
      `💾 Saved ${uniqueItems.length} OrderTodayItems (deduplicated from ${items.length})`
    );
  }

  /**
   * Get saved Order Today items from database
   */
  async getOrderTodayItems(): Promise<OrderTodayItemData[]> {
    const items = await this.prisma.orderTodayItem.findMany({
      where: { tenantId: this.tenantId },
      orderBy: { syncedAt: "desc" },
    });

    return items.map((item) => ({
      platform: item.platform,
      orderSn: item.orderSn,
      trackingNo: item.trackingNo || "",
      courier: item.courier || "",
      sellerSku: item.sellerSku || "",
      productName: item.productName || "",
      variationName: item.variationName || "",
      quantity: item.quantity,
    }));
  }
}

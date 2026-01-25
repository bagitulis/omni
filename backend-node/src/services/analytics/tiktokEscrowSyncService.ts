/**
 * TikTok Escrow Sync Service
 * SRP: Orchestrate sync process
 * Transformation delegated to TiktokTransactionTransformer
 */

import { getLogger } from "../../utils/logger";
import { TiktokAPIClient } from "../../api/clients/tiktokAPIClient";
import { TiktokDataRepository } from "./tiktokDataRepository";
import { tiktokTransactionTransformer } from "./tiktokTransactionTransformer";
import { SyncResult, TiktokOrder, TiktokTransaction, PrismaClientAny } from "./tiktokAnalyticsTypes";

const logger = getLogger("TiktokEscrowSyncService");

export class TiktokEscrowSyncService {
  private repository: TiktokDataRepository;
  private apiClient: TiktokAPIClient;
  private tenantId: string;
  private readonly PAGE_SIZE = 100;
  private readonly MAX_RETRIES = 3;
  private readonly RETRY_DELAY_MS = 1000;

  constructor(prisma: PrismaClientAny, apiClient: TiktokAPIClient, tenantId: string) {
    this.repository = new TiktokDataRepository(prisma, tenantId);
    this.apiClient = apiClient;
    this.tenantId = tenantId;
  }

  async isSynced(month: number, year: number): Promise<boolean> {
    return (await this.repository.findSync(month, year)) !== null;
  }

  async getSyncStatus(month: number, year: number) {
    return await this.repository.findSync(month, year);
  }

  async syncMonth(month: number, year: number, forceResync = false): Promise<SyncResult> {
    logger.info(`Starting TikTok sync for ${year}-${month}, tenant: ${this.tenantId}`);

    if (!forceResync && (await this.isSynced(month, year))) {
      return this.createResult(false, month, year, 0, 0, 0, "Already synced. Use force resync to update.");
    }

    if (forceResync) await this.deleteMonthData(month, year);

    try {
      const orders = await this.fetchOrdersByMonth(month, year);
      logger.info(`Fetched ${orders.length} completed orders`);

      if (orders.length === 0) {
        return this.createResult(true, month, year, 0, 0, 0, "No completed orders found for this month");
      }

      const { totalItems, failedOrders, processedOrders } = await this.processOrders(orders, month, year);
      await this.repository.upsertSyncRecord(month, year, processedOrders - failedOrders);

      return this.createResult(
        true, month, year, processedOrders - failedOrders, totalItems, failedOrders,
        `Synced ${processedOrders - failedOrders}/${orders.length} orders, ${totalItems} items`
      );
    } catch (error) {
      logger.error(`Sync failed: ${error}`);
      await this.deleteMonthData(month, year);
      throw error;
    }
  }

  private async processOrders(orders: TiktokOrder[], month: number, year: number) {
    let totalItems = 0, failedOrders = 0, processedOrders = 0;

    for (const order of orders) {
      try {
        const orderId = order.id;
        const transaction = await this.fetchTransactionWithRetry(orderId);
        if (transaction) {
          const savedOrder = await this.repository.upsertOrder(order, transaction, month, year);
          const items = transaction.order_line_item_list || [];
          for (const item of items) {
            await this.repository.upsertItem(savedOrder.id, orderId, item);
          }
          totalItems += items.length;
        } else {
          failedOrders++;
          logger.warn(`No settlement transaction for order ${orderId}`);
        }
        processedOrders++;
        if (processedOrders < orders.length) await this.sleep(500);
      } catch (error) {
        failedOrders++;
        logger.error(`Failed to process order ${order.id}: ${error}`);
      }
    }
    return { totalItems, failedOrders, processedOrders };
  }

  private async fetchOrdersByMonth(month: number, year: number): Promise<TiktokOrder[]> {
    const allOrders: TiktokOrder[] = [];
    const startDate = new Date(year, month - 1, 1, 0, 0, 0);
    const endDate = new Date(year, month, 1, 0, 0, 0);
    const createTimeGe = Math.floor(startDate.getTime() / 1000);
    const createTimeLt = Math.floor(endDate.getTime() / 1000);
    let nextPageToken: string | null = null;
    let pageCount = 0;

    do {
      const requestBody = { order_status: "COMPLETED", create_time_ge: createTimeGe, create_time_lt: createTimeLt };
      const params: Record<string, unknown> = { page_size: this.PAGE_SIZE };
      if (nextPageToken) params.page_token = nextPageToken;

      const response = await this.apiClient.searchOrders(requestBody, params);
      if (response?.data?.orders) allOrders.push(...response.data.orders);
      nextPageToken = response?.data?.next_page_token || null;
      pageCount++;
      logger.info(`Fetched page ${pageCount}: ${response?.data?.orders?.length || 0} orders`);
      if (nextPageToken) await this.sleep(500);
    } while (nextPageToken);

    return allOrders;
  }

  private async fetchTransactionWithRetry(orderId: string): Promise<TiktokTransaction | null> {
    for (let attempts = 1; attempts <= this.MAX_RETRIES; attempts++) {
      try {
        const response = await this.apiClient.getOrderTransactions(orderId);
        return tiktokTransactionTransformer.transformV202501Response(orderId, response?.data);
      } catch (error) {
        logger.warn(`Fetch failed for order ${orderId}, attempt ${attempts}/${this.MAX_RETRIES}`);
        if (attempts < this.MAX_RETRIES) {
          await this.sleep(this.RETRY_DELAY_MS * attempts);
        } else {
          logger.error(`Failed to fetch order ${orderId} after ${this.MAX_RETRIES} attempts - SKIPPING`);
          return null;
        }
      }
    }
    return null;
  }

  async deleteMonthData(month: number, year: number): Promise<void> {
    await this.repository.deleteMonthData(month, year);
  }

  private createResult(
    success: boolean, month: number, year: number, totalOrders: number,
    totalItems: number, failedOrders: number, message: string
  ): SyncResult {
    return { success, month, year, totalOrders, totalItems, failedOrders, message };
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

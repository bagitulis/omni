/**
 * Shopee Escrow Sync Service
 * Syncs escrow data from Shopee API to database
 * Single Responsibility: Sync escrow data
 */

import { getLogger } from "../../utils/logger";
import { ShopeeShippingFeeService } from "../orders/shopeeShippingFeeService";
import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";

const logger = getLogger("ShopeeEscrowSyncService");

// Use 'any' for Prisma client to avoid type issues with dynamically added models
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

interface SyncResult {
  success: boolean;
  month: number;
  year: number;
  totalOrders: number;
  totalItems: number;
  message: string;
}

interface WalletTransaction {
  order_sn: string;
  amount: string;
  create_time: number;
  buyer_name?: string;
}

export class ShopeeEscrowSyncService {
  private prisma: PrismaClientAny;
  private shippingFeeService: ShopeeShippingFeeService;
  private tenantId: string;
  private readonly BATCH_SIZE = 20;

  constructor(
    prisma: PrismaClientAny,
    apiClient: ShopeeAPIClient,
    tenantId: string
  ) {
    this.prisma = prisma;
    this.shippingFeeService = new ShopeeShippingFeeService(apiClient);
    this.tenantId = tenantId;
  }

  /**
   * Check if month is already synced
   */
  async isSynced(month: number, year: number): Promise<boolean> {
    const sync = await this.prisma.shopeeEscrowSync.findUnique({
      where: {
        tenantId_month_year: {
          tenantId: this.tenantId,
          month,
          year,
        },
      },
    });
    return sync !== null; // If record exists, it's synced
  }

  /**
   * Get sync status for a month
   */
  async getSyncStatus(month: number, year: number) {
    return await this.prisma.shopeeEscrowSync.findUnique({
      where: {
        tenantId_month_year: {
          tenantId: this.tenantId,
          month,
          year,
        },
      },
    });
  }

  /**
   * Sync escrow data for a specific month
   */
  async syncMonth(
    month: number,
    year: number,
    walletTransactions: WalletTransaction[],
    forceResync: boolean = false
  ): Promise<SyncResult> {
    logger.info(`Starting sync for ${year}-${month}, tenant: ${this.tenantId}`);

    // Check if already synced
    if (!forceResync && (await this.isSynced(month, year))) {
      return {
        success: false,
        month,
        year,
        totalOrders: 0,
        totalItems: 0,
        message: "Already synced. Use force resync to update.",
      };
    }

    // Create/update sync record - only after successful sync (atomic)
    // We'll delete existing data first if force resync
    if (forceResync) {
      await this.deleteMonthData(month, year);
    }

    try {
      // Extract order SNs from wallet transactions
      const orderSns = walletTransactions.map((tx) => tx.order_sn);
      logger.info(`Found ${orderSns.length} orders to sync`);

      let totalItems = 0;
      let processedOrders = 0;

      // Process in batches
      for (let i = 0; i < orderSns.length; i += this.BATCH_SIZE) {
        const batch = orderSns.slice(i, i + this.BATCH_SIZE);
        const itemsCount = await this.processBatch(batch, month, year);
        totalItems += itemsCount;
        processedOrders += batch.length;
        logger.info(
          `Processed batch ${Math.floor(i / this.BATCH_SIZE) + 1}: ${
            batch.length
          } orders`
        );
        if (i + this.BATCH_SIZE < orderSns.length) await this.sleep(500);
      }

      // Create sync record ONLY after successful completion (atomic)
      await this.prisma.shopeeEscrowSync.upsert({
        where: {
          tenantId_month_year: { tenantId: this.tenantId, month, year },
        },
        create: {
          tenantId: this.tenantId,
          month,
          year,
          totalOrders: processedOrders,
          syncedAt: new Date(),
        },
        update: { totalOrders: processedOrders, syncedAt: new Date() },
      });

      return {
        success: true,
        month,
        year,
        totalOrders: processedOrders,
        totalItems,
        message: `Synced ${processedOrders} orders, ${totalItems} items`,
      };
    } catch (error) {
      logger.error(`Sync failed: ${error}`);
      await this.deleteMonthData(month, year);
      throw error;
    }
  }

  /**
   * Delete all data for a month (for resync or cleanup on failure)
   * Note: ShopeeEscrowItem is auto-deleted via cascade when order is deleted
   */
  async deleteMonthData(month: number, year: number): Promise<void> {
    // Delete orders first - items will be cascade-deleted automatically
    // (ShopeeEscrowItem has onDelete: Cascade relation to ShopeeEscrowOrder)
    await this.prisma.shopeeEscrowOrder.deleteMany({
      where: {
        tenantId: this.tenantId,
        month,
        year,
      },
    });

    // Delete sync record
    await this.prisma.shopeeEscrowSync.deleteMany({
      where: {
        tenantId: this.tenantId,
        month,
        year,
      },
    });
  }

  /**
   * Process a batch of orders
   */
  private async processBatch(
    orderSns: string[],
    month: number,
    year: number
  ): Promise<number> {
    // Fetch escrow details from API
    const escrowDetails =
      await this.shippingFeeService.getEscrowDetailsBatch(orderSns);

    let totalItems = 0;

    for (let i = 0; i < escrowDetails.length; i++) {
      const escrow = escrowDetails[i];
      if (!escrow) continue;

      const orderSn = orderSns[i];
      const itemsCount = await this.saveEscrowOrder(
        escrow,
        orderSn,
        month,
        year
      );
      totalItems += itemsCount;
    }

    return totalItems;
  }

  /**
   * Save escrow order and items to database
   */
  private async saveEscrowOrder(
    escrow: any,
    orderSn: string,
    month: number,
    year: number
  ): Promise<number> {
    const orderIncome = escrow.order_income || {};
    const buyerPaymentInfo = escrow.buyer_payment_info || {};
    const items = orderIncome.items || [];

    // Upsert order
    const savedOrder = await this.prisma.shopeeEscrowOrder.upsert({
      where: {
        tenantId_orderSn: {
          tenantId: this.tenantId,
          orderSn,
        },
      },
      create: {
        tenantId: this.tenantId,
        orderSn,
        month,
        year,
        buyerUserName: escrow.buyer_user_name,
        escrowAmount: orderIncome.escrow_amount || 0,
        commissionFee: orderIncome.commission_fee || 0,
        serviceFee: orderIncome.service_fee || 0,
        sellerProcessingFee: orderIncome.seller_order_processing_fee || 0,
        buyerPaidShippingFee: orderIncome.buyer_paid_shipping_fee || 0,
        actualShippingFee: orderIncome.actual_shipping_fee || 0,
        shopeeShippingRebate: orderIncome.shopee_shipping_rebate || 0,
        estimatedShippingFee: orderIncome.estimated_shipping_fee || 0,
        buyerTotalAmount: orderIncome.buyer_total_amount || 0,
        buyerPaymentMethod: orderIncome.buyer_payment_method,
        rawOrderIncome: JSON.stringify(orderIncome),
        rawBuyerPaymentInfo: JSON.stringify(buyerPaymentInfo),
      },
      update: {
        escrowAmount: orderIncome.escrow_amount || 0,
        commissionFee: orderIncome.commission_fee || 0,
        serviceFee: orderIncome.service_fee || 0,
        sellerProcessingFee: orderIncome.seller_order_processing_fee || 0,
        buyerPaidShippingFee: orderIncome.buyer_paid_shipping_fee || 0,
        actualShippingFee: orderIncome.actual_shipping_fee || 0,
        shopeeShippingRebate: orderIncome.shopee_shipping_rebate || 0,
        estimatedShippingFee: orderIncome.estimated_shipping_fee || 0,
        buyerTotalAmount: orderIncome.buyer_total_amount || 0,
        buyerPaymentMethod: orderIncome.buyer_payment_method,
        rawOrderIncome: JSON.stringify(orderIncome),
        rawBuyerPaymentInfo: JSON.stringify(buyerPaymentInfo),
        updatedAt: new Date(),
      },
    });

    // Delete existing items and re-create
    await this.prisma.shopeeEscrowItem.deleteMany({
      where: { escrowOrderId: savedOrder.id },
    });

    // Save items
    for (const item of items) {
      await this.prisma.shopeeEscrowItem.create({
        data: {
          tenantId: this.tenantId,
          orderId: savedOrder.id, // Legacy column (same as escrowOrderId)
          orderSn: orderSn, // Legacy column
          month: month, // Legacy column
          year: year, // Legacy column
          escrowOrderId: savedOrder.id,
          itemId: item.item_id ? BigInt(item.item_id) : null,
          modelId: item.model_id ? BigInt(item.model_id) : null,
          sku: item.item_sku || null,
          modelSku: item.model_sku || null,
          itemName: item.item_name,
          modelName: item.model_name,
          quantity: item.quantity_purchased || 0,
          originalPrice: item.original_price || 0,
          sellingPrice: item.selling_price || 0,
          discountedPrice: item.discounted_price || 0,
          sellerDiscount: item.seller_discount || 0,
          shopeeDiscount: item.shopee_discount || 0,
          discountFromCoin: item.discount_from_coin || 0,
          discountFromVoucherSeller: item.discount_from_voucher_seller || 0,
          discountFromVoucherShopee: item.discount_from_voucher_shopee || 0,
          amsCommissionFee: item.ams_commission_fee || 0,
          sellerOrderProcessingFee: item.seller_order_processing_fee || 0,
          rawItemData: JSON.stringify(item),
        },
      });
    }

    return items.length;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * TikTok Data Repository
 * Database operations for TikTok escrow data
 * Single Responsibility: Prisma database CRUD
 */

import { getLogger } from "../../utils/logger";
import {
  TiktokOrder,
  TiktokTransaction,
  PrismaClientAny,
  parseAmount,
} from "./tiktokAnalyticsTypes";

const logger = getLogger("TiktokDataRepository");

export class TiktokDataRepository {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  async findSync(month: number, year: number) {
    return await this.prisma.tiktokEscrowSync.findFirst({
      where: { tenantId: this.tenantId, month, year },
    });
  }

  async upsertOrder(
    order: TiktokOrder,
    transaction: TiktokTransaction,
    month: number,
    year: number
  ) {
    const feeDetails = transaction.fee_details || {};
    const orderId = order.id; // TikTok API returns 'id', not 'order_id'

    return await this.prisma.tiktokEscrowOrder.upsert({
      where: { id: `${this.tenantId}_${orderId}` },
      create: {
        id: `${this.tenantId}_${orderId}`,
        tenantId: this.tenantId,
        orderId: orderId,
        month,
        year,
        orderStatus: order.order_status,
        orderDate: new Date(order.create_time * 1000),
        transactionId: transaction.transaction_id,
        transactionType: transaction.transaction_type,
        statementTime: new Date(transaction.statement_time * 1000),
        totalSettlementAmount: parseAmount(transaction.total_settlement_amount),
        platformCommission: parseAmount(feeDetails.platform_commission),
        transactionFee: parseAmount(feeDetails.transaction_fee),
        shippingFeeCustomerPaid: parseAmount(
          feeDetails.customer_paid_shipping_fee_after_discount
        ),
        shippingFeeActual: parseAmount(feeDetails.shipping_cost),
        shippingFeePlatformDiscount: parseAmount(
          feeDetails.shipping_fee_subsidy
        ),
        buyerTotalAmount: order.payment_info?.total_amount || 0,
        currency: order.payment_info?.currency || "IDR",
        rawTransactionData: JSON.stringify(transaction),
        rawOrderData: JSON.stringify(order),
        syncedAt: new Date(),
      },
      update: {
        orderStatus: order.order_status,
        transactionId: transaction.transaction_id,
        transactionType: transaction.transaction_type,
        statementTime: new Date(transaction.statement_time * 1000),
        totalSettlementAmount: parseAmount(transaction.total_settlement_amount),
        platformCommission: parseAmount(feeDetails.platform_commission),
        transactionFee: parseAmount(feeDetails.transaction_fee),
        shippingFeeCustomerPaid: parseAmount(
          feeDetails.customer_paid_shipping_fee_after_discount
        ),
        shippingFeeActual: parseAmount(feeDetails.shipping_cost),
        shippingFeePlatformDiscount: parseAmount(
          feeDetails.shipping_fee_subsidy
        ),
        buyerTotalAmount: order.payment_info?.total_amount || 0,
        rawTransactionData: JSON.stringify(transaction),
        rawOrderData: JSON.stringify(order),
        updatedAt: new Date(),
        syncedAt: new Date(),
      },
    });
  }

  async upsertItem(
    escrowOrderId: string,
    orderId: string,
    item: TransactionItemInput
  ) {
    return await this.prisma.tiktokEscrowItem.upsert({
      where: { id: `${escrowOrderId}_${item.sku_id}` },
      create: {
        id: `${escrowOrderId}_${item.sku_id}`,
        tenantId: this.tenantId,
        escrowOrderId,
        orderId,
        productId: item.product_id,
        productName: item.product_name,
        skuId: item.sku_id,
        sellerSku: item.seller_sku || null,
        quantity: item.quantity || 1,
        salePrice: parseAmount(item.sale_price),
        originalPrice: parseAmount(item.original_price),
        subtotalAfterSellerDiscount: parseAmount(
          item.subtotal_after_seller_discount
        ),
        platformDiscount: parseAmount(item.platform_discount),
        sellerDiscount: parseAmount(item.seller_discount),
        settlementAmount: parseAmount(item.sku_settlement_amount),
        rawItemData: JSON.stringify(item),
        syncedAt: new Date(),
      },
      update: {
        productName: item.product_name,
        sellerSku: item.seller_sku || null,
        quantity: item.quantity || 1,
        salePrice: parseAmount(item.sale_price),
        originalPrice: parseAmount(item.original_price),
        subtotalAfterSellerDiscount: parseAmount(
          item.subtotal_after_seller_discount
        ),
        platformDiscount: parseAmount(item.platform_discount),
        sellerDiscount: parseAmount(item.seller_discount),
        settlementAmount: parseAmount(item.sku_settlement_amount),
        rawItemData: JSON.stringify(item),
        updatedAt: new Date(),
        syncedAt: new Date(),
      },
    });
  }

  async upsertSyncRecord(month: number, year: number, totalOrders: number) {
    return await this.prisma.tiktokEscrowSync.upsert({
      where: { id: `${this.tenantId}_${month}_${year}` },
      create: {
        id: `${this.tenantId}_${month}_${year}`,
        tenantId: this.tenantId,
        month,
        year,
        totalOrders,
        syncedAt: new Date(),
      },
      update: {
        totalOrders,
        syncedAt: new Date(),
        updatedAt: new Date(),
      },
    });
  }

  async deleteMonthData(month: number, year: number): Promise<void> {
    const orders = await this.prisma.tiktokEscrowOrder.findMany({
      where: { tenantId: this.tenantId, month, year },
      select: { id: true },
    });
    const orderIds = orders.map((o: { id: string }) => o.id);

    if (orderIds.length > 0) {
      await this.prisma.tiktokEscrowItem.deleteMany({
        where: { escrowOrderId: { in: orderIds } },
      });
    }
    await this.prisma.tiktokEscrowOrder.deleteMany({
      where: { tenantId: this.tenantId, month, year },
    });
    await this.prisma.tiktokEscrowSync.deleteMany({
      where: { tenantId: this.tenantId, month, year },
    });
    logger.info(`Deleted data for ${month}/${year}`);
  }
}

interface TransactionItemInput {
  product_id: string;
  product_name: string;
  sku_id: string;
  seller_sku?: string;
  quantity?: number;
  sale_price?: string;
  original_price?: string;
  subtotal_after_seller_discount?: string;
  platform_discount?: string;
  seller_discount?: string;
  sku_settlement_amount?: string;
}

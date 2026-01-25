/**
 * Shipping Fee Analysis Service
 * Analyzes shipping fee differences from synced escrow data
 * Single Responsibility: Calculate and report shipping fee discrepancies
 */

import { getLogger } from "../../utils/logger";

const logger = getLogger("ShippingFeeAnalysisService");

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

export interface ShippingFeeAnalysisResult {
  orderSn: string;
  orderDate: string | null;
  buyerPaid: number;
  actualFee: number;
  shopeeRebate: number;
  difference: number;
  buyerName: string | null;
  paymentMethod: string | null;
}

export interface ShippingFeeSummary {
  totalOrders: number;
  ordersWithDifference: number;
  totalProfit: number;
  totalLoss: number;
  netImpact: number;
}

export interface ShippingFeeAnalysisResponse {
  summary: ShippingFeeSummary;
  orders: ShippingFeeAnalysisResult[];
}

export class ShippingFeeAnalysisService {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Analyze shipping fee differences for a month
   * Only returns orders with difference != 0
   * Sorted by payment method (carrier proxy)
   */
  async analyzeMonth(
    month: number,
    year: number
  ): Promise<ShippingFeeAnalysisResponse> {
    logger.info(
      `Analyzing shipping fees for ${year}-${month}, tenant: ${this.tenantId}`
    );

    // Fetch all orders for the month
    const orders = await this.prisma.shopeeEscrowOrder.findMany({
      where: {
        tenantId: this.tenantId,
        month,
        year,
      },
      select: {
        orderSn: true,
        orderDate: true,
        buyerUserName: true,
        buyerPaidShippingFee: true,
        actualShippingFee: true,
        shopeeShippingRebate: true,
        buyerPaymentMethod: true,
        syncedAt: true,
      },
      orderBy: [
        { buyerPaymentMethod: "asc" },
        { orderDate: "desc" },
      ],
    });

    // Calculate differences and filter
    const ordersWithDiff: ShippingFeeAnalysisResult[] = [];
    let totalProfit = 0;
    let totalLoss = 0;

    for (const order of orders) {
      const buyerPaid = order.buyerPaidShippingFee || 0;
      const actual = order.actualShippingFee || 0;
      const rebate = order.shopeeShippingRebate || 0;

      // Difference = Buyer Paid - Actual + Rebate
      // Positive = profit (buyer paid more than actual)
      // Negative = loss (actual more than buyer paid)
      const difference = buyerPaid - actual + rebate;

      // Only include orders with difference
      if (difference !== 0) {
        ordersWithDiff.push({
          orderSn: order.orderSn,
          orderDate: order.orderDate
            ? order.orderDate.toISOString().split("T")[0]
            : order.syncedAt?.toISOString().split("T")[0] || null,
          buyerPaid,
          actualFee: actual,
          shopeeRebate: rebate,
          difference,
          buyerName: order.buyerUserName,
          paymentMethod: order.buyerPaymentMethod,
        });

        if (difference > 0) {
          totalProfit += difference;
        } else {
          totalLoss += Math.abs(difference);
        }
      }
    }

    const summary: ShippingFeeSummary = {
      totalOrders: orders.length,
      ordersWithDifference: ordersWithDiff.length,
      totalProfit,
      totalLoss,
      netImpact: totalProfit - totalLoss,
    };

    logger.info(
      `Found ${ordersWithDiff.length} orders with shipping fee difference`
    );

    return {
      summary,
      orders: ordersWithDiff,
    };
  }
}

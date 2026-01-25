/**
 * TikTok Shipping Fee Analysis Service
 * Analyzes shipping fee differences from synced TikTok escrow data
 * Single Responsibility: Calculate and report shipping fee discrepancies
 */

import { getLogger } from "../../utils/logger";

const logger = getLogger("TiktokShippingFeeAnalysisService");

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

export interface ShippingFeeAnalysisResult {
  orderId: string;
  orderDate: string | null;
  originalFee: number;
  buyerPaid: number;
  platformDiscount: number;
  sellerPays: number;
  orderStatus: string | null;
  currency: string;
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

export class TiktokShippingFeeAnalysisService {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Analyze shipping fee differences for a month
   * Only returns orders with difference != 0
   * Sorted by order date descending
   */
  async analyzeMonth(
    month: number,
    year: number
  ): Promise<ShippingFeeAnalysisResponse> {
    logger.info(
      `Analyzing TikTok shipping fees for ${year}-${month}, tenant: ${this.tenantId}`
    );

    // Fetch all TikTok orders for the month (no filter - include all orders)
    const orders = await this.prisma.tiktokEscrowOrder.findMany({
      where: {
        tenantId: this.tenantId,
        month,
        year,
      },
      select: {
        orderId: true,
        orderDate: true,
        orderStatus: true,
        rawTransactionData: true, // Real transaction data (shipping_cost, subsidy)
        rawOrderData: true, // Order payment data (buyer shipping_fee)
        currency: true,
        syncedAt: true,
      },
      orderBy: [{ orderDate: "desc" }],
    });

    // Calculate differences and filter
    const ordersWithDiff: ShippingFeeAnalysisResult[] = [];
    let totalSellerPays = 0;

    for (const order of orders) {
      let originalFee = 0; // From rawTransactionData.fee_details.shipping_cost (absolute)
      let platformSubsidy = 0; // From rawTransactionData.fee_details.shipping_fee_subsidy
      let buyerPaid = 0; // From rawOrderData.payment.shipping_fee

      try {
        // Get real shipping cost from transaction data (after transaction complete)
        const transactionData = order.rawTransactionData
          ? JSON.parse(order.rawTransactionData)
          : null;
        if (transactionData?.fee_details) {
          const fees = transactionData.fee_details;
          // shipping_cost is negative (e.g., "-51000"), so take absolute
          originalFee = Math.abs(parseFloat(fees.shipping_cost || "0"));
          platformSubsidy = parseFloat(fees.shipping_fee_subsidy || "0");
        }

        // Get buyer paid shipping from order payment data
        const orderData = order.rawOrderData
          ? JSON.parse(order.rawOrderData)
          : null;
        if (orderData?.payment) {
          buyerPaid = parseFloat(orderData.payment.shipping_fee || "0");
        }
      } catch {
        // Skip orders with invalid data
        continue;
      }

      // Seller pays = Original Shipping Cost - Platform Subsidy - Buyer Paid
      // Example: 51000 - 25000 - 26000 = 0 (no difference)
      // Example: 36500 - 30000 - 0 = 6500 (seller pays 6500)
      const sellerPays = originalFee - platformSubsidy - buyerPaid;

      // Only include orders where seller pays extra
      if (sellerPays > 0) {
        ordersWithDiff.push({
          orderId: order.orderId,
          orderDate: order.orderDate
            ? order.orderDate.toISOString().split("T")[0]
            : order.syncedAt?.toISOString().split("T")[0] || null,
          originalFee,
          buyerPaid,
          platformDiscount: platformSubsidy,
          sellerPays,
          orderStatus: order.orderStatus,
          currency: order.currency || "IDR",
        });

        totalSellerPays += sellerPays;
      }
    }

    const summary: ShippingFeeSummary = {
      totalOrders: orders.length,
      ordersWithDifference: ordersWithDiff.length,
      totalProfit: 0, // Not applicable - seller only pays, never profits from shipping
      totalLoss: totalSellerPays,
      netImpact: -totalSellerPays,
    };

    logger.info(
      `Found ${ordersWithDiff.length} TikTok orders with shipping fee difference`
    );

    return {
      summary,
      orders: ordersWithDiff,
    };
  }
}

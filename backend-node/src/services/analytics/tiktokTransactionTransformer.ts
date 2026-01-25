/**
 * TikTok Transaction Transformer
 * SRP: Transform v202501 Finance API response to internal format
 */

import { getLogger } from "../../utils/logger";
import { TiktokTransaction } from "./tiktokAnalyticsTypes";

const logger = getLogger("TiktokTransactionTransformer");

export interface TransactionLineItem {
  product_id: string;
  product_name: string;
  sku_id: string;
  seller_sku?: string;
  quantity: number;
  sale_price: string;
  original_price?: string;
  subtotal_after_seller_discount?: string;
  platform_discount?: string;
  seller_discount?: string;
  sku_settlement_amount?: string;
}

export interface AggregatedFeeDetails {
  platform_commission?: string;
  affiliate_commission?: string;
  affiliate_partner_commission?: string;
  transaction_fee?: string;
  shipping_fee?: string;
  customer_paid_shipping_fee_after_discount?: string;
  fbt_fulfillment_fee?: string;
  customer_shipping_fee?: string;
  shipping_cost?: string;
  shipping_fee_subsidy?: string;
  return_shipping_fee?: string;
  adjustment_amount?: string;
}

export class TiktokTransactionTransformer {
  /**
   * Transform v202501 Finance API response to TiktokTransaction format
   */
  transformV202501Response(
    orderId: string,
    data: Record<string, unknown> | null
  ): TiktokTransaction | null {
    if (!data || !data.sku_transactions) {
      logger.warn(`No sku_transactions in response for order ${orderId}`);
      return null;
    }

    const skuTransactions = data.sku_transactions as Array<
      Record<string, unknown>
    >;
    if (skuTransactions.length === 0) {
      logger.warn(`Empty sku_transactions for order ${orderId}`);
      return null;
    }

    const firstSku = skuTransactions[0];
    const statementId = (firstSku.statement_id as string) || orderId;
    const aggregatedFees = this.aggregateFeeDetails(skuTransactions);
    const orderLineItems = skuTransactions.map((sku) =>
      this.transformSkuToLineItem(sku)
    );

    return {
      transaction_id: statementId,
      transaction_type: "ORDER_SETTLEMENT",
      order_id: (data.order_id as string) || orderId,
      statement_time: (data.order_create_time as number) || 0,
      total_settlement_amount: (data.settlement_amount as string) || "0",
      revenue_details: {
        product_refund_amount: "0",
        customer_order_refund: "0",
        customer_shipping_fee_refund: "0",
        tiktok_shipping_incentive_refund: "0",
        other_refund: "0",
      },
      fee_details: aggregatedFees,
      order_line_item_list: orderLineItems,
    };
  }

  /**
   * Aggregate fee details from all SKU transactions
   */
  aggregateFeeDetails(
    skuTransactions: Array<Record<string, unknown>>
  ): AggregatedFeeDetails {
    let platformCommission = 0;
    let affiliateCommission = 0;
    let transactionFee = 0;
    let shippingCost = 0;
    let shippingFeeSubsidy = 0;

    for (const sku of skuTransactions) {
      const breakdown = sku.fee_tax_breakdown as Record<string, unknown>;
      if (breakdown?.fee) {
        const fee = breakdown.fee as Record<string, string>;
        platformCommission += parseFloat(fee.platform_commission_amount || "0");
        affiliateCommission += parseFloat(
          fee.affiliate_commission_amount || "0"
        );
        transactionFee += parseFloat(fee.transaction_fee_amount || "0");
      }
      const shippingBreakdown = sku.shipping_cost_breakdown as Record<
        string,
        unknown
      >;
      if (shippingBreakdown) {
        shippingCost += parseFloat(
          (shippingBreakdown.actual_shipping_fee_amount as string) || "0"
        );
        shippingFeeSubsidy += parseFloat(
          (shippingBreakdown.shipping_fee_discount_amount as string) || "0"
        );
      }
    }

    return {
      platform_commission: String(platformCommission),
      affiliate_commission: String(affiliateCommission),
      transaction_fee: String(transactionFee),
      shipping_cost: String(shippingCost),
      shipping_fee_subsidy: String(shippingFeeSubsidy),
    };
  }

  /**
   * Transform v202501 sku_transaction to TransactionItem format
   */
  transformSkuToLineItem(sku: Record<string, unknown>): TransactionLineItem {
    const revenueBreakdown =
      (sku.revenue_breakdown as Record<string, string>) || {};
    return {
      product_id: (sku.sku_id as string) || "",
      product_name: (sku.product_name as string) || "",
      sku_id: (sku.sku_id as string) || "",
      seller_sku: (sku.sku_name as string) || undefined,
      quantity: parseInt((sku.quantity as string) || "1", 10),
      sale_price: (sku.revenue_amount as string) || "0",
      original_price: revenueBreakdown.subtotal_before_discount_amount || "0",
      subtotal_after_seller_discount: (sku.revenue_amount as string) || "0",
      platform_discount: "0",
      seller_discount: revenueBreakdown.seller_discount_amount || "0",
      sku_settlement_amount: (sku.settlement_amount as string) || "0",
    };
  }
}

export const tiktokTransactionTransformer = new TiktokTransactionTransformer();

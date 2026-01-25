/**
 * Shopee Shipping Fee Service
 * Handles shipping fee calculation and processing
 * Mirrors Python: ShopeeShippingFee
 */

import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { getLogger } from "../../utils/logger";

export interface ShippingFeeResult {
  "Create Time": string;
  "Order SN": string;
  "Buyer Paid": number;
  Actual: number;
  "Shopee Rebate": number;
  Difference: number;
  "Shipping Carrier": string;
}

export interface EscrowDetail {
  order_sn: string;
  order_income?: {
    buyer_paid_shipping_fee: number;
    actual_shipping_fee: number;
    shopee_shipping_rebate: number;
  };
}

export class ShopeeShippingFeeService {
  private api: ShopeeAPIClient;
  private logger: any;
  private orderManager: any; // Reference to order manager for order details

  private readonly BATCH_SIZE = 50;
  private readonly RATE_LIMIT_DELAY = 500; // ms

  constructor(apiClient: ShopeeAPIClient) {
    this.api = apiClient;
    this.logger = getLogger("ShopeeShippingFeeService");
  }

  /**
   * Set order manager reference (for getting order details)
   */
  setOrderManager(orderManager: any): void {
    this.orderManager = orderManager;
  }

  /**
   * Batch fetch escrow details from API
   */
  async getEscrowDetailsBatch(
    orderSnList: string[]
  ): Promise<(EscrowDetail | null)[]> {
    if (
      !orderSnList ||
      orderSnList.length === 0 ||
      orderSnList.length > this.BATCH_SIZE
    ) {
      throw new Error(`Batch size must be 1-${this.BATCH_SIZE} orders`);
    }

    try {
      // POST request: params={}, data={order_sn_list}
      const response = await this.api.request(
        "/api/v2/payment/get_escrow_detail_batch",
        "POST",
        {},
        { order_sn_list: orderSnList }
      );

      if (!response?.response) {
        this.logger.warn(`No escrow response for batch`);
        return new Array(orderSnList.length).fill(null);
      }

      const resultMap = new Map<string, EscrowDetail>();
      for (const detail of response.response) {
        if (detail.escrow_detail) {
          resultMap.set(detail.escrow_detail.order_sn, detail.escrow_detail);
        }
      }

      return orderSnList.map((sn) => resultMap.get(sn) || null);
    } catch (error) {
      this.logger.error(`Batch API failed: ${error}`);
      return new Array(orderSnList.length).fill(null);
    }
  }

  /**
   * Process shipping fee differences for orders
   */
  async processShippingFeeDifference(
    orderSnList: string[],
    _outputFile?: string,
    startFrom: number = 0
  ): Promise<ShippingFeeResult[]> {
    if (!orderSnList || orderSnList.length === 0) {
      this.logger.error("Order SN list is empty");
      return [];
    }

    if (!this.orderManager) {
      this.logger.error("Order manager not set");
      return [];
    }

    const results: ShippingFeeResult[] = [];
    const totalOrders = orderSnList.length;

    for (let i = startFrom; i < totalOrders; i += this.BATCH_SIZE) {
      const batch = orderSnList.slice(i, i + this.BATCH_SIZE);

      if (batch.length === 0) {
        continue;
      }

      try {
        const batchResults = await this._processShippingBatch(batch);
        if (batchResults.length > 0) {
          results.push(...batchResults);
          this.logger.info(
            `Processed batch ${Math.floor(i / this.BATCH_SIZE) + 1}: ${
              batchResults.length
            } orders`
          );
        }
      } catch (error) {
        this.logger.error(
          `Batch ${Math.floor(i / this.BATCH_SIZE) + 1} failed: ${error}`
        );
        continue;
      }

      // Rate limiting delay
      if (i + this.BATCH_SIZE < totalOrders) {
        await this.sleep(this.RATE_LIMIT_DELAY);
      }
    }

    this.logger.info(`Completed processing ${results.length} orders`);
    return results;
  }

  /**
   * Process a batch of shipping fee calculations
   */
  private async _processShippingBatch(
    batch: string[]
  ): Promise<ShippingFeeResult[]> {
    const results: ShippingFeeResult[] = [];

    try {
      // Get escrow details
      const escrowResults = await this.getEscrowDetailsBatch(batch);

      // Get order details
      const orderDetails = await this.orderManager.getOrderDetails(batch);

      let processedCount = 0;

      for (let idx = 0; idx < escrowResults.length; idx++) {
        const escrowData = escrowResults[idx];
        const orderSn = batch[idx];

        if (!escrowData) {
          continue;
        }

        let shippingCarrier = "";
        let createTime = "";

        // Find matching order details
        const matchingOrder = orderDetails.find(
          (o: any) => o.order_sn === orderSn
        );

        if (matchingOrder) {
          const timestamp = matchingOrder.create_time || 0;
          if (timestamp) {
            createTime = new Date(timestamp * 1000).toISOString().split("T")[0];
          }

          const packageList = matchingOrder.package_list || [];
          if (Array.isArray(packageList) && packageList.length > 0) {
            shippingCarrier = packageList[0].shipping_carrier || "";
          }
        }

        const orderIncome = (escrowData.order_income || {}) as any;
        const buyerPaid = parseFloat(
          (orderIncome.buyer_paid_shipping_fee || 0).toString()
        );
        const actual = parseFloat(
          (orderIncome.actual_shipping_fee || 0).toString()
        );
        const rebate = parseFloat(
          (orderIncome.shopee_shipping_rebate || 0).toString()
        );

        results.push({
          "Create Time": createTime,
          "Order SN": orderSn,
          "Buyer Paid": buyerPaid,
          Actual: actual,
          "Shopee Rebate": rebate,
          Difference: buyerPaid - actual + rebate,
          "Shipping Carrier": shippingCarrier,
        });

        processedCount++;
      }

      if (processedCount > 0) {
        this.logger.info(`✅ Shipping fees processed`);
      }
    } catch (error) {
      this.logger.error(`❌ Error processing shipping batch: ${error}`);
    }

    return results;
  }

  /**
   * Export shipping fee results to Google Sheets
   */
  async exportShippingFeeToGoogleSheets(
    results: ShippingFeeResult[],
    spreadsheetId: string,
    sheetName: string,
    googleSheetsManager: any
  ): Promise<boolean> {
    try {
      if (!results || results.length === 0) {
        this.logger.warn("No results to export");
        return false;
      }

      // Prepare headers
      const headers = [
        "Create Time",
        "Order SN",
        "Buyer Paid",
        "Actual",
        "Shopee Rebate",
        "Difference",
        "Shipping Carrier",
      ];

      // Prepare data rows
      const data = [
        headers,
        ...results.map((result) => [
          result["Create Time"],
          result["Order SN"],
          result["Buyer Paid"],
          result["Actual"],
          result["Shopee Rebate"],
          result["Difference"],
          result["Shipping Carrier"],
        ]),
      ];

      // Upload to Google Sheets
      const success = await googleSheetsManager.uploadToSheet({
        spreadsheetId,
        sheetName,
        data,
      });

      if (success) {
        this.logger.info(
          `✅ Successfully exported ${results.length} shipping fee records to Google Sheets`
        );
        return true;
      } else {
        this.logger.error("❌ Failed to export to Google Sheets");
        return false;
      }
    } catch (error) {
      this.logger.error(
        `❌ Error in exportShippingFeeToGoogleSheets: ${error}`
      );
      return false;
    }
  }

  /**
   * Sleep helper
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

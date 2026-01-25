/**
 * Shopee Combined Operations Routes
 * Handles combined wallet and order operations
 * Single Responsibility: Combined operations only
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";
import { getTokenManager } from "../services/tokenManager";

const logger = getLogger("ShopeeOperations");
export const shopeeOperationsRouter = Router();

/**
 * POST /api/shopee/operations/wallet-with-orders
 * Get wallet transactions with order details
 */
shopeeOperationsRouter.post(
  "/wallet-with-orders",
  async (req: Request, res: Response) => {
    try {
      logger.info("[SHOPEE] Fetching wallet transactions with order details");

      const { month, year, transactionType } = req.body;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!month || !year) {
        return res.status(400).json({
          success: false,
          error: "Month and year are required",
        });
      }

      if (!tenantId) {
        return res.status(400).json({
          success: false,
          error: "x-tenant-id header is required",
        });
      }

      const tokenManager = await getTokenManager(tenantId);
      const shopeeApiClient = tokenManager.getShopeeClient();

      if (!shopeeApiClient) {
        return res.status(500).json({
          success: false,
          error: "Shopee API client not initialized",
        });
      }

      const orderManager = new ShopeeOrderManager(shopeeApiClient);
      const rawTransactions = await orderManager.wallet.getTransactions(
        month,
        year,
        transactionType
      );

      if (!rawTransactions || rawTransactions.length === 0) {
        return res.json({
          success: true,
          message: "No transactions found",
          data: { transactions: [], orderNumbers: [] },
        });
      }

      const processedData =
        orderManager.wallet.processTransactions(rawTransactions);
      const orderNumbers = Array.from(
        orderManager.wallet.extractOrderNumbers(rawTransactions)
      );

      let orderDetails: any[] = [];
      if (orderNumbers.length > 0) {
        orderDetails = await orderManager.getOrderDetails(orderNumbers);
      }

      logger.info(
        `[SHOPEE] Fetched ${processedData.count} transactions and ${orderDetails.length} order details`
      );

      return res.json({
        success: true,
        message: "Wallet transactions with order details fetched",
        data: {
          transactions: processedData.transactions,
          orders: orderDetails,
          orderNumbers,
          totalAmount: processedData.totalAmount,
          transactionCount: processedData.count,
        },
      });
    } catch (error) {
      logger.error(`[SHOPEE] Wallet with orders failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default shopeeOperationsRouter;

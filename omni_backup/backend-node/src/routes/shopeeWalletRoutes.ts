/**
 * Shopee Wallet Operations Routes
 * Handles wallet transactions and reporting
 * Single Responsibility: Wallet operations only
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getTokenManager } from "../services/tokenManager";
import { ShopeeShippingFeeService } from "../services/orders/shopeeShippingFeeService";

const logger = getLogger("ShopeeWallet");
export const shopeeWalletRouter = Router();

/**
 * POST /api/shopee/operations/wallet-report
 * Get wallet transactions from Shopee
 */
shopeeWalletRouter.post("/report", async (req: Request, res: Response) => {
  try {
    logger.info("[SHOPEE] Executing wallet report operation");

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
      logger.info("No transactions found");
      return res.json({
        success: true,
        message: "No transactions found for the selected period",
        data: [],
      });
    }

    const processedData =
      orderManager.wallet.processTransactions(rawTransactions);
    const orderNumbers =
      orderManager.wallet.extractOrderNumbers(rawTransactions);

    logger.info(
      `[SHOPEE] Processed ${processedData.count} wallet transactions`
    );

    return res.json({
      success: true,
      message: "Wallet transactions fetched successfully",
      data: {
        transactions: processedData.transactions,
        totalAmount: processedData.totalAmount,
        count: processedData.count,
        orderNumbers: Array.from(orderNumbers),
      },
    });
  } catch (error) {
    logger.error(`[SHOPEE] Wallet report failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * POST /api/shopee/operations/export-wallet
 * Export wallet transactions to Google Sheets
 */
shopeeWalletRouter.post("/export", async (req: Request, res: Response) => {
  try {
    logger.info("[SHOPEE] Exporting wallet to Google Sheets");

    const { month, year, transactionType, spreadsheetId, sheetName } = req.body;
    const tenantId = req.headers["x-tenant-id"] as string;

    if (!month || !year || !spreadsheetId || !sheetName) {
      return res.status(400).json({
        success: false,
        error: "Month, year, spreadsheetId, and sheetName are required",
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
      return res.status(400).json({
        success: false,
        error: "No transactions found",
      });
    }

    const processedData =
      orderManager.wallet.processTransactions(rawTransactions);

    const googleSheetsService = getGoogleSheetsService();
    const headers = [
      "Date",
      "Order SN",
      "Description",
      "Amount",
      "Status",
      "Transaction Type",
      "Tab Type",
      "Buyer Name",
    ];

    const rows = processedData.transactions.map((tx) => [
      tx.Date,
      tx["Order SN"],
      tx.Description,
      tx.Amount,
      tx.Status,
      tx["Transaction Type"],
      tx["Tab Type"],
      tx["Buyer Name"],
    ]);

    const data = [headers, ...rows];

    const success = await googleSheetsService.uploadToSheet({
      spreadsheetId,
      sheetName,
      data,
    });

    if (!success) {
      return res.status(400).json({
        success: false,
        error: "Failed to export to Google Sheets",
      });
    }

    logger.info(
      `[SHOPEE] Exported ${processedData.count} transactions to Google Sheets`
    );

    return res.json({
      success: true,
      message: "Wallet transactions exported to Google Sheets",
      data: {
        count: processedData.count,
        totalAmount: processedData.totalAmount,
      },
    });
  } catch (error) {
    logger.error(`[SHOPEE] Export wallet failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * POST /api/shopee/wallet/escrow-detail
 * Get single escrow detail for an order
 */
shopeeWalletRouter.post(
  "/escrow-detail",
  async (req: Request, res: Response) => {
    try {
      logger.info("[SHOPEE] Fetching single escrow detail");

      const { order_sn } = req.body;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!order_sn) {
        return res.status(400).json({
          success: false,
          error: "order_sn is required",
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

      const shippingFeeService = new ShopeeShippingFeeService(shopeeApiClient);
      const escrowDetail = await shippingFeeService.getEscrowDetailsBatch([
        order_sn,
      ]);

      if (!escrowDetail || !escrowDetail[0]) {
        logger.warn(`No escrow detail found for order: ${order_sn}`);
        return res.json({
          success: true,
          message: "No escrow detail found",
          data: {
            response: null,
          },
        });
      }

      logger.info(`[SHOPEE] Escrow detail fetched for order: ${order_sn}`);

      return res.json({
        success: true,
        message: "Escrow detail fetched successfully",
        data: {
          response: escrowDetail[0],
        },
      });
    } catch (error) {
      logger.error(`[SHOPEE] Escrow detail fetch failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

/**
 * POST /api/shopee/wallet/escrow-detail-batch
 * Get escrow details for multiple orders
 */
shopeeWalletRouter.post(
  "/escrow-detail-batch",
  async (req: Request, res: Response) => {
    try {
      logger.info("[SHOPEE] Fetching batch escrow details");

      const { order_sn_list } = req.body;
      const tenantId = req.headers["x-tenant-id"] as string;

      if (!order_sn_list || !Array.isArray(order_sn_list)) {
        return res.status(400).json({
          success: false,
          error: "order_sn_list is required and must be an array",
        });
      }

      if (order_sn_list.length === 0) {
        return res.status(400).json({
          success: false,
          error: "order_sn_list cannot be empty",
        });
      }

      if (!tenantId) {
        return res.status(400).json({
          success: false,
          error: "x-tenant-id header is required",
        });
      }

      // Validate batch size (Shopee limit is typically 50)
      const MAX_BATCH_SIZE = 50;
      if (order_sn_list.length > MAX_BATCH_SIZE) {
        return res.status(400).json({
          success: false,
          error: `Batch size cannot exceed ${MAX_BATCH_SIZE} orders`,
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

      const shippingFeeService = new ShopeeShippingFeeService(shopeeApiClient);
      const escrowDetails =
        await shippingFeeService.getEscrowDetailsBatch(order_sn_list);

      const responseData = escrowDetails.map((detail, index) => ({
        order_sn: order_sn_list[index],
        escrow_detail: detail,
      }));

      logger.info(
        `[SHOPEE] Batch escrow details fetched for ${order_sn_list.length} orders`
      );

      return res.json({
        success: true,
        message: "Batch escrow details fetched successfully",
        data: {
          response: responseData,
          count: order_sn_list.length,
          withData: escrowDetails.filter((d) => d !== null).length,
        },
      });
    } catch (error) {
      logger.error(`[SHOPEE] Batch escrow detail fetch failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default shopeeWalletRouter;

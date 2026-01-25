/**
 * Wallet Google Sheets Routes
 * Handles wallet transaction export to Google Sheets
 * Single Responsibility: Wallet sheets operations only
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getTokenManager } from "../services/tokenManager";

const logger = getLogger("WalletSheetsRoutes");
export const walletSheetsRouter = Router();

/**
 * Generate safe sheet name for Google Sheets
 * Uses simpler naming to avoid length issues
 */
function generateWalletSheetName(
  month?: number,
  year?: number,
  transactionType?: string
): string {
  const now = new Date();
  const m = month || now.getMonth() + 1;
  const y = year || now.getFullYear();

  // Use short abbreviations for transaction types
  const typeAbbrev: Record<string, string> = {
    wallet_order_income: "income",
    wallet_adjustment_filter: "adjust",
    wallet_wallet_payment: "payment",
    wallet_refund_from_order: "refund",
    wallet_withdrawals: "withdraw",
    fast_escrow_repayment: "escrow",
    fast_pay: "fastpay",
    seller_loan: "loan",
    corporate_loan: "corp_loan",
  };

  const typeStr = transactionType
    ? typeAbbrev[transactionType] || transactionType.substring(0, 5)
    : "all";

  const baseName = `Wallet_${m.toString().padStart(2, "0")}_${y}_${typeStr}`;
  // Clean for Google Sheets requirements (max 31 chars)
  const safeName = baseName
    .replace(/[^\w-]/g, "_")
    .replace(/_+/g, "_")
    .substring(0, 31);

  return safeName;
}

/**
 * POST /api/sheets/operations/wallet-to-sheets
 * Export Shopee wallet transactions to Google Sheets
 */
walletSheetsRouter.post(
  "/wallet-to-sheets",
  async (req: Request, res: Response) => {
    try {
      logger.info("[SHEETS] Executing wallet-to-sheets operation");

      const { month, year, transaction_type } = req.body;
      const tenantId = req.headers["x-tenant-id"] as string;

      logger.info(`Month: ${month}, Year: ${year}, Type: ${transaction_type}`);

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

      // Get Shopee API client
      const tokenManager = await getTokenManager(tenantId);
      const shopeeApiClient = tokenManager.getShopeeClient();

      if (!shopeeApiClient) {
        return res.status(500).json({
          success: false,
          error: "Shopee API client not initialized",
        });
      }

      // Initialize order manager
      const orderManager = new ShopeeOrderManager(shopeeApiClient);

      // Get settings from Google Sheets
      const googleSheetsService = getGoogleSheetsService();

      // Try to get spreadsheetId from request, then from settings
      let walletSpreadsheetId = (req.body as any).spreadsheetId || null;
      let walletSheetName = (req.body as any).sheetName;

      // If not provided in request, try to get from Google Sheets settings
      if (!walletSpreadsheetId) {
        const settings = await googleSheetsService.getDetailedSettings();
        walletSpreadsheetId = settings?.wallet_spreadsheet_id || null;
      }

      if (!walletSpreadsheetId) {
        return res.status(400).json({
          success: false,
          error:
            "Google Sheets spreadsheet ID not found in request or settings (spreadsheetId)",
        });
      }

      // Generate safe sheet name if not provided (mirroring Python backend)
      if (!walletSheetName) {
        walletSheetName = generateWalletSheetName(
          month,
          year,
          transaction_type
        );
      }

      // Get and process transactions
      const rawTransactions = await orderManager.wallet.getTransactions(
        month,
        year,
        transaction_type
      );

      if (!rawTransactions || rawTransactions.length === 0) {
        return res.status(400).json({
          success: false,
          error: "No transactions found",
        });
      }

      // Process transactions
      const processedData =
        orderManager.wallet.processTransactions(rawTransactions);

      // Prepare headers and data
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

      // Upload to Google Sheets
      const success = await googleSheetsService.uploadToSheet({
        spreadsheetId: walletSpreadsheetId,
        sheetName: walletSheetName,
        data,
      });

      if (!success) {
        return res.status(400).json({
          success: false,
          error: "Failed to export to Google Sheets",
        });
      }

      logger.info(
        `[SHEETS] Wallet to sheets completed successfully (${processedData.count} records)`
      );

      return res.json({
        success: true,
        message: "Wallet transactions exported to Google Sheets",
        data: {
          count: processedData.count,
          totalAmount: processedData.totalAmount,
          spreadsheetId: walletSpreadsheetId,
          sheetName: walletSheetName,
        },
      });
    } catch (error) {
      logger.error(`[SHEETS] Wallet to sheets failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default walletSheetsRouter;

/**
 * Shipping Fee Google Sheets Routes
 * Handles shipping fee export to Google Sheets
 * Single Responsibility: Shipping fee sheets operations only
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getTokenManager } from "../services/tokenManager";

const logger = getLogger("ShippingFeeSheetsRoutes");
export const shippingFeeSheetsRouter = Router();

/**
 * Generate safe sheet name for shipping fee export
 */
function generateShippingFeeSheetName(month?: number, year?: number): string {
  const now = new Date();
  const m = month || now.getMonth() + 1;
  const y = year || now.getFullYear();

  const safeName = `ShippingFee_${m.toString().padStart(2, "0")}_${y}`;
  return safeName.substring(0, 31);
}

/**
 * POST /api/sheets/operations/shipping-fee-to-sheets
 * Export Shopee shipping fee differences to Google Sheets
 */
shippingFeeSheetsRouter.post(
  "/shipping-fee-to-sheets",
  async (req: Request, res: Response) => {
    try {
      logger.info("[SHEETS] Executing shipping-fee-to-sheets operation");

      const { month, year, spreadsheetId, sheetName } = req.body;
      const tenantId = req.headers["x-tenant-id"] as string;

      logger.info(`Month: ${month}, Year: ${year}`);

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

      // Get Google Sheets service for settings
      const googleSheetsService = getGoogleSheetsService();

      // Try to get spreadsheetId from request, then from settings
      let shippingSpreadsheetId = spreadsheetId || null;
      let shippingSheetName = sheetName;

      // If not provided in request, try to get from Google Sheets settings
      if (!shippingSpreadsheetId) {
        const settings = await googleSheetsService.getDetailedSettings();
        shippingSpreadsheetId = settings?.shipping_spreadsheet_id || null;
      }

      if (!shippingSpreadsheetId) {
        return res.status(400).json({
          success: false,
          error:
            "Google Sheets spreadsheet ID not found in request or settings (spreadsheetId)",
        });
      }

      // Generate safe sheet name if not provided (mirroring Python backend)
      if (!shippingSheetName) {
        shippingSheetName = generateShippingFeeSheetName(month, year);
      }

      // Get wallet transactions to extract order numbers
      const rawTransactions = await orderManager.wallet.getTransactions(
        month,
        year
      );

      if (!rawTransactions || rawTransactions.length === 0) {
        return res.status(400).json({
          success: false,
          error: "No transactions found",
        });
      }

      // Process raw transactions to get formatted data with "Order SN" key
      const processedData =
        orderManager.wallet.processTransactions(rawTransactions);

      // Extract order numbers from processed transactions
      const orderNumbers = Array.from(
        orderManager.wallet.extractOrderNumbers(processedData.transactions)
      );

      if (orderNumbers.length === 0) {
        return res.status(400).json({
          success: false,
          error: "No order numbers found in transactions",
        });
      }

      // Process shipping fees
      const results =
        await orderManager.shippingFee.processShippingFeeDifference(
          orderNumbers
        );

      if (results.length === 0) {
        return res.status(400).json({
          success: false,
          error: "No shipping fee data processed",
        });
      }

      // Export to Google Sheets
      const success =
        await orderManager.shippingFee.exportShippingFeeToGoogleSheets(
          results,
          shippingSpreadsheetId,
          shippingSheetName,
          googleSheetsService
        );

      if (!success) {
        return res.status(400).json({
          success: false,
          error: "Failed to export to Google Sheets",
        });
      }

      logger.info(
        `[SHEETS] Shipping fee to sheets completed successfully (${results.length} records)`
      );

      return res.json({
        success: true,
        message: "Shipping fees exported to Google Sheets",
        data: {
          count: results.length,
          spreadsheetId: shippingSpreadsheetId,
          sheetName: shippingSheetName,
        },
      });
    } catch (error) {
      logger.error(`[SHEETS] Shipping fee to sheets failed: ${error}`);
      return res.status(500).json({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);

export default shippingFeeSheetsRouter;

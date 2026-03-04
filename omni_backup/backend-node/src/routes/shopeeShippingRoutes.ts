/**
 * Shopee Shipping Fee Routes
 * Handles shipping fee processing and reporting
 * Single Responsibility: Shipping fee operations only
 */

import { Router, Request, Response } from "express";
import { getLogger } from "../utils/logger";
import { ShopeeOrderManager } from "../services/orders/shopeeOrderManager";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getTokenManager } from "../services/tokenManager";

const logger = getLogger("ShopeeShipping");
export const shopeeShippingRouter = Router();

/**
 * POST /api/shopee/operations/shipping-fee
 * Process shipping fee differences
 */
shopeeShippingRouter.post("/fee", async (req: Request, res: Response) => {
  try {
    logger.info("[SHOPEE] Processing shipping fee differences");

    const { orderSnList } = req.body;
    const tenantId = req.headers["x-tenant-id"] as string;

    if (
      !orderSnList ||
      !Array.isArray(orderSnList) ||
      orderSnList.length === 0
    ) {
      return res.status(400).json({
        success: false,
        error: "orderSnList array is required",
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
    const results =
      await orderManager.shippingFee.processShippingFeeDifference(orderSnList);

    logger.info(
      `[SHOPEE] Processed shipping fees for ${results.length} orders`
    );

    return res.json({
      success: true,
      message: "Shipping fees processed",
      data: {
        results,
        count: results.length,
      },
    });
  } catch (error) {
    logger.error(`[SHOPEE] Shipping fee processing failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

/**
 * POST /api/shopee/operations/export-shipping-fee
 * Export shipping fee to Google Sheets
 */
shopeeShippingRouter.post("/export", async (req: Request, res: Response) => {
  try {
    logger.info("[SHOPEE] Exporting shipping fees to Google Sheets");

    const { orderSnList, spreadsheetId, sheetName } = req.body;
    const tenantId = req.headers["x-tenant-id"] as string;

    if (
      !orderSnList ||
      !Array.isArray(orderSnList) ||
      !spreadsheetId ||
      !sheetName
    ) {
      return res.status(400).json({
        success: false,
        error: "orderSnList, spreadsheetId, and sheetName are required",
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
    const results =
      await orderManager.shippingFee.processShippingFeeDifference(orderSnList);

    if (results.length === 0) {
      return res.status(400).json({
        success: false,
        error: "No shipping fee data processed",
      });
    }

    const googleSheetsService = getGoogleSheetsService();
    const success =
      await orderManager.shippingFee.exportShippingFeeToGoogleSheets(
        results,
        spreadsheetId,
        sheetName,
        googleSheetsService
      );

    if (!success) {
      return res.status(400).json({
        success: false,
        error: "Failed to export to Google Sheets",
      });
    }

    logger.info(
      `[SHOPEE] Exported ${results.length} shipping fee records to Google Sheets`
    );

    return res.json({
      success: true,
      message: "Shipping fees exported to Google Sheets",
      data: {
        count: results.length,
      },
    });
  } catch (error) {
    logger.error(`[SHOPEE] Export shipping fee failed: ${error}`);
    return res.status(500).json({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

export default shopeeShippingRouter;

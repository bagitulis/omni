/**
 * Google Sheets Settings Controller
 * Handles Google Sheets settings configuration
 * Single Responsibility: Settings management only
 */

import { Request, Response } from "express";
import { getGoogleSheetsService } from "../services/googleSheetsService";
import { getGoogleSheetsAutoDetector } from "../services/googleSheetsAutoDetector";
import { getServiceAccountRotationManager } from "../services/googleSheetsServiceAccountRotationManager";
import { getLogger } from "../utils/logger";

const logger = getLogger("GoogleSheetsSettings");

/**
 * GET /api/google/settings/detailed
 * Get current detailed Google Sheets settings
 */
export async function getDetailedSettingsController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const tenantId = req.tenantId; // Get tenant from authMiddleware
    logger.info("🔄 Loading detailed Google Sheets settings");
    logger.info(`   Tenant: ${tenantId}`);

    const service = getGoogleSheetsService();
    const settings = await service.getDetailedSettings(tenantId);

    logger.info("✅ Detailed settings loaded successfully");
    res.json({
      success: true,
      data: settings,
    });
  } catch (error: any) {
    logger.error(`❌ Error loading detailed settings: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/google/settings/update-detailed
 * Update detailed Google Sheets settings
 */
export async function updateDetailedSettingsController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const data = req.body;
    const tenantId = req.tenantId; // Get tenant from authMiddleware

    if (!data) {
      logger.warn("❌ No data provided");
      res.status(400).json({
        success: false,
        error: "No data provided",
      });
      return;
    }

    logger.info("🔄 Updating detailed Google Sheets settings");
    logger.info(`   Tenant: ${tenantId}`);
    logger.info(`   Wallet ID: ${data.wallet_spreadsheet_id || "not set"}`);
    logger.info(`   Shipping ID: ${data.shipping_spreadsheet_id || "not set"}`);
    logger.info(
      `   Inventory ID: ${data.inventory_spreadsheet_id || "not set"}`
    );
    logger.info(
      `   Inventory Sheet: ${data.inventory_sheet_name || "not set"}`
    );
    logger.info(`   Order ID: ${data.order_spreadsheet_id || "not set"}`);

    const service = getGoogleSheetsService();
    const success = await service.updateDetailedSettings(
      data.wallet_spreadsheet_id,
      data.shipping_spreadsheet_id,
      data.inventory_spreadsheet_id,
      data.order_spreadsheet_id,
      data.inventory_sheet_name,
      data.inventory_selected_columns || [],
      data.inventory_available_worksheets || [],
      data.wallet_available_worksheets || [],
      data.shipping_available_worksheets || [],
      data.order_available_worksheets || [],
      tenantId
    );

    if (success) {
      logger.info("✅ Detailed settings updated successfully");
      res.json({
        success: true,
        message: "Settings updated successfully",
      });
    } else {
      logger.error("❌ Failed to update detailed settings");
      res.status(400).json({
        success: false,
        error: "Failed to update settings",
      });
    }
  } catch (error: any) {
    logger.error(`❌ Error updating detailed settings: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * GET /api/google/settings/test
 * Test current Google Sheets connection
 */
export async function testConnectionController(
  _req: Request,
  res: Response
): Promise<void> {
  try {
    logger.info("🔄 Testing Google Sheets connection");

    const service = getGoogleSheetsService();
    const connected = await service.testConnection();

    if (connected) {
      logger.info("✅ Google Sheets connection test successful");
    } else {
      logger.warn("❌ Google Sheets connection test failed");
    }

    res.json({
      success: true,
      connected,
      message: connected
        ? "Connection test successful"
        : "Connection test failed",
    });
  } catch (error: any) {
    logger.error(`❌ Error testing connection: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/google/settings/save-links
 * Save spreadsheet links (URLs) for different purposes
 */
export async function saveSpreadsheetLinksController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const { inventory, wallet, shipping, order } = req.body;
    const tenantId = req.tenantId; // Get tenant from authMiddleware

    if (!inventory && !wallet && !shipping && !order) {
      res.status(400).json({
        success: false,
        error: "At least one spreadsheet link is required",
      });
      return;
    }

    logger.info("🔄 Saving spreadsheet links");
    logger.info(`   Tenant: ${tenantId}`);
    logger.info(`   Inventory Link: ${inventory || "not set"}`);
    logger.info(`   Wallet Link: ${wallet || "not set"}`);
    logger.info(`   Shipping Link: ${shipping || "not set"}`);
    logger.info(`   Order Link: ${order || "not set"}`);

    const service = getGoogleSheetsService();
    const success = await service.saveSpreadsheetLinks({
      inventory,
      wallet,
      shipping,
      order,
    }, tenantId);

    if (success) {
      logger.info("✅ Spreadsheet links saved successfully");
      res.json({
        success: true,
        message: "Spreadsheet links saved successfully",
      });
    } else {
      logger.error("❌ Failed to save spreadsheet links");
      res.status(400).json({
        success: false,
        error: "Failed to save spreadsheet links",
      });
    }
  } catch (error: any) {
    logger.error(`❌ Error saving spreadsheet links: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * GET /api/google/settings/saved-links
 * Get saved spreadsheet links
 */
export async function getSavedLinksController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const tenantId = req.tenantId; // Get tenant from authMiddleware
    logger.info("🔄 Loading saved spreadsheet links");
    logger.info(`   Tenant: ${tenantId}`);

    const service = getGoogleSheetsService();
    const links = await service.getSavedSpreadsheetLinks(tenantId);

    logger.info("✅ Spreadsheet links loaded successfully");
    res.json({
      success: true,
      data: links,
    });
  } catch (error: any) {
    logger.error(`❌ Error loading spreadsheet links: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

/**
 * POST /api/google/settings/validate-link
 * Validate spreadsheet link and return metadata (sheets list)
 */
export async function validateSpreadsheetLinkController(
  req: Request,
  res: Response
): Promise<void> {
  try {
    const { spreadsheetUrl, type } = req.body;

    if (!spreadsheetUrl) {
      res.status(400).json({
        success: false,
        error: "Spreadsheet URL is required",
      });
      return;
    }

    if (!type || !["inventory", "wallet", "shipping", "order"].includes(type)) {
      res.status(400).json({
        success: false,
        error:
          "Valid spreadsheet type is required (inventory, wallet, shipping, order)",
      });
      return;
    }

    logger.info(`🔍 Validating ${type} spreadsheet link: ${spreadsheetUrl}`);

    // Get service account
    const manager = getServiceAccountRotationManager();
    const activeAccount = manager.getActiveAccount();

    if (!activeAccount || !activeAccount.authClient) {
      res.status(400).json({
        success: false,
        error: "No service account available",
      });
      return;
    }

    // Detect metadata from URL
    const detector = getGoogleSheetsAutoDetector();
    const metadata = await detector.detectMetadata(
      spreadsheetUrl,
      activeAccount.authClient
    );

    if (!metadata) {
      logger.warn(`❌ Failed to validate ${type} link`);
      res.status(400).json({
        success: false,
        error: `Invalid spreadsheet URL or no access. Make sure the service account has access to this spreadsheet.`,
      });
      return;
    }

    logger.info(
      `✅ ${type} spreadsheet validated: ${metadata.name} (${metadata.sheets.length} sheets)`
    );

    res.json({
      success: true,
      data: {
        spreadsheetId: metadata.spreadsheetId,
        name: metadata.name,
        sheets: metadata.sheets,
        type,
      },
    });
  } catch (error: any) {
    logger.error(`❌ Error validating spreadsheet link: ${error.message}`);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
}

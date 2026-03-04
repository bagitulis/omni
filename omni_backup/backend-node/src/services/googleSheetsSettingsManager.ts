import { googleSheetsSettingsService } from "./googleSheetsSettingsService";
import { getInventoryService } from "./inventoryService";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { GoogleSheetsAPIOperations } from "./googleSheetsAPIOperations";

export class GoogleSheetsSettingsManager {
  private apiOperations: GoogleSheetsAPIOperations;
  private logger: Logger;

  constructor(apiOperations: GoogleSheetsAPIOperations) {
    this.apiOperations = apiOperations;
    this.logger = getLogger("GoogleSheetsSettingsManager");
  }

  async updateDetailedSettings(
    walletSpreadsheetId: string,
    shippingSpreadsheetId: string,
    inventorySpreadsheetId: string,
    orderSpreadsheetId: string,
    inventorySheetName: string,
    inventorySelectedColumns: string[],
    inventoryAvailableWorksheets?: any[],
    walletAvailableWorksheets?: any[],
    shippingAvailableWorksheets?: any[],
    orderAvailableWorksheets?: any[],
    tenantId?: string
  ): Promise<boolean> {
    try {
      let columnsToSave = inventorySelectedColumns;
      let keyColumnToSave = "SKU"; // Default

      if (
        inventorySpreadsheetId &&
        inventorySheetName &&
        (!inventorySelectedColumns || inventorySelectedColumns.length === 0)
      ) {
        this.logger.info(
          `Auto-fetching columns from inventory sheet: ${inventorySheetName}`
        );
        const headers = await this.apiOperations.getColumnHeaders(
          inventorySpreadsheetId,
          inventorySheetName
        );
        if (headers.length > 0) {
          columnsToSave = headers;
          this.logger.info(`Auto-fetched columns: ${columnsToSave.join(", ")}`);

          // Auto-detect key column from new sheet
          keyColumnToSave = this.detectKeyColumn(headers);
          this.logger.info(`🔑 Auto-detected key column: ${keyColumnToSave}`);
        }
      }

      try {
        this.logger.info("Saving Google Sheets settings...");
        await googleSheetsSettingsService.saveSettings(
          {
            walletSpreadsheetId,
            shippingSpreadsheetId,
            inventorySpreadsheetId,
            inventorySheetName,
            inventorySelectedColumns: columnsToSave || [],
            orderSpreadsheetId,
            inventoryAvailableWorksheets: inventoryAvailableWorksheets || [],
            walletAvailableWorksheets: walletAvailableWorksheets || [],
            shippingAvailableWorksheets: shippingAvailableWorksheets || [],
            orderAvailableWorksheets: orderAvailableWorksheets || [],
          },
          tenantId
        );
        this.logger.info("✅ Google Sheets settings saved");
      } catch (dbError) {
        this.logger.error("Error saving Google Sheets settings");
      }

      try {
        this.logger.info("Saving inventory settings...");
        await getInventoryService().updateSettings(
          inventorySpreadsheetId,
          inventorySheetName,
          columnsToSave || [],
          keyColumnToSave,
          1,
          2,
          false,
          300
        );
        this.logger.info("✅ Inventory settings saved");
      } catch (dbError) {
        this.logger.error("Error saving inventory settings");
      }

      return true;
    } catch (error) {
      this.logger.error("Error updating settings");
      return false;
    }
  }

  /**
   * Auto-detect key column from available columns
   * Prioritize columns that contain "SKU" or "ID", otherwise use first column
   */
  private detectKeyColumn(headers: string[]): string {
    if (!headers || headers.length === 0) return "SKU";

    // Look for columns with "SKU" in name (case insensitive)
    const skuColumn = headers.find((h) => h.toLowerCase().includes("sku"));
    if (skuColumn) return skuColumn;

    // Look for columns with "ID" in name
    const idColumn = headers.find((h) => h.toLowerCase().includes("id"));
    if (idColumn) return idColumn;

    // Fall back to first column
    return headers[0];
  }

  async getDetailedSettings(tenantId?: string): Promise<Record<string, any>> {
    try {
      const settings = await googleSheetsSettingsService.getSettings(tenantId);
      return {
        wallet_spreadsheet_id: settings.walletSpreadsheetId || "",
        shipping_spreadsheet_id: settings.shippingSpreadsheetId || "",
        inventory_spreadsheet_id: settings.inventorySpreadsheetId || "",
        inventory_sheet_name: settings.inventorySheetName || "",
        inventory_selected_columns: settings.inventorySelectedColumns || [],
        inventory_available_worksheets:
          settings.inventoryAvailableWorksheets || [],
        wallet_available_worksheets: settings.walletAvailableWorksheets || [],
        shipping_available_worksheets:
          settings.shippingAvailableWorksheets || [],
        order_available_worksheets: settings.orderAvailableWorksheets || [],
        order_spreadsheet_id: settings.orderSpreadsheetId || "",
        spreadsheet_id: settings.spreadsheetId || "",
        selected_sheet: settings.selectedSheet || "",
        manual_mode: settings.manualMode || false,
        available_spreadsheets: settings.availableSpreadsheets || [],
      };
    } catch (error) {
      this.logger.error("Error getting detailed settings");
      throw error;
    }
  }

  /**
   * Save spreadsheet links (URLs) - extract ID from URL and save
   * @param links - Object with inventory, wallet, shipping, order URLs
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async saveSpreadsheetLinks(
    links: {
      inventory?: string;
      wallet?: string;
      shipping?: string;
      order?: string;
    },
    tenantId?: string
  ): Promise<boolean> {
    try {
      this.logger.info("🔄 Processing spreadsheet links");

      const extractIdFromUrl = (url: string): string => {
        if (!url) return "";
        // Extract ID from Google Sheets URL: https://docs.google.com/spreadsheets/d/{ID}/...
        const match = url.match(/\/spreadsheets\/d\/([a-zA-Z0-9-_]+)/);
        return match ? match[1] : url;
      };

      const inventoryId = extractIdFromUrl(links.inventory || "");
      const walletId = extractIdFromUrl(links.wallet || "");
      const shippingId = extractIdFromUrl(links.shipping || "");
      const orderId = extractIdFromUrl(links.order || "");

      this.logger.info(`   Inventory ID: ${inventoryId || "not set"}`);
      this.logger.info(`   Wallet ID: ${walletId || "not set"}`);
      this.logger.info(`   Shipping ID: ${shippingId || "not set"}`);
      this.logger.info(`   Order ID: ${orderId || "not set"}`);

      await googleSheetsSettingsService.saveSettings(
        {
          inventorySpreadsheetId: inventoryId,
          walletSpreadsheetId: walletId,
          shippingSpreadsheetId: shippingId,
          orderSpreadsheetId: orderId,
        },
        tenantId
      );

      this.logger.info("✅ Spreadsheet links saved successfully");
      return true;
    } catch (error) {
      this.logger.error("Error saving spreadsheet links");
      return false;
    }
  }

  /**
   * Get saved spreadsheet links
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async getSavedSpreadsheetLinks(tenantId?: string): Promise<any> {
    try {
      this.logger.info("🔄 Loading saved spreadsheet links");
      const settings = await googleSheetsSettingsService.getSettings(tenantId);

      const convertIdToUrl = (id: string): string => {
        if (!id) return "";
        // Convert ID to Google Sheets URL
        return `https://docs.google.com/spreadsheets/d/${id}/edit`;
      };

      return {
        inventory: convertIdToUrl(settings.inventorySpreadsheetId || ""),
        wallet: convertIdToUrl(settings.walletSpreadsheetId || ""),
        shipping: convertIdToUrl(settings.shippingSpreadsheetId || ""),
        order: convertIdToUrl(settings.orderSpreadsheetId || ""),
      };
    } catch (error) {
      this.logger.error("Error getting saved spreadsheet links");
      return {};
    }
  }
}

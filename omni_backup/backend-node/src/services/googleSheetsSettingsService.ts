import { getPrisma } from "./prismaClient";
import { getLogger } from "../utils/logger";

const logger = getLogger("GoogleSheetsSettings");

export interface GoogleSheetsSettingsData {
  spreadsheetId?: string;
  selectedSheet?: string;
  manualMode?: boolean;
  availableSpreadsheets?: any[];
  availableWorksheets?: any[];
  inventoryAvailableWorksheets?: any[];
  walletAvailableWorksheets?: any[];
  shippingAvailableWorksheets?: any[];
  orderAvailableWorksheets?: any[];
  walletSpreadsheetId?: string;
  shippingSpreadsheetId?: string;
  inventorySpreadsheetId?: string;
  inventorySheetName?: string;
  inventorySelectedColumns?: string[];
  orderSpreadsheetId?: string;
}

class GoogleSheetsSettingsService {
  /**
   * Get current Google Sheets settings
   * Safely handles missing columns in older database schemas
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async getSettings(tenantId?: string): Promise<GoogleSheetsSettingsData> {
    try {
      const prisma = getPrisma(tenantId);

      // Try to get settings, but handle column errors gracefully
      let settings: any;
      try {
        settings = await prisma.googleSheetsSettings.findFirst();
      } catch (error: any) {
        // If column doesn't exist, return empty settings
        if (
          error.code === "P2022" ||
          error.message.includes("does not exist")
        ) {
          logger.warn(
            "GoogleSheetsSettings table or columns not yet created, returning empty settings"
          );
          return this.getEmptySettings();
        }
        throw error;
      }

      if (!settings) {
        return this.getEmptySettings();
      }

      return {
        spreadsheetId: settings.spreadsheetId || "",
        selectedSheet: settings.selectedSheet || "",
        manualMode: settings.manualMode || false,
        availableSpreadsheets: settings.availableSpreadsheets
          ? JSON.parse(settings.availableSpreadsheets)
          : [],
        availableWorksheets: settings.availableWorksheets
          ? JSON.parse(settings.availableWorksheets)
          : [],
        inventoryAvailableWorksheets: settings.inventoryAvailableWorksheets
          ? JSON.parse(settings.inventoryAvailableWorksheets)
          : [],
        walletAvailableWorksheets: settings.walletAvailableWorksheets
          ? JSON.parse(settings.walletAvailableWorksheets)
          : [],
        shippingAvailableWorksheets: settings.shippingAvailableWorksheets
          ? JSON.parse(settings.shippingAvailableWorksheets)
          : [],
        orderAvailableWorksheets: settings.orderAvailableWorksheets
          ? JSON.parse(settings.orderAvailableWorksheets)
          : [],
        walletSpreadsheetId: settings.walletSpreadsheetId || "",
        shippingSpreadsheetId: settings.shippingSpreadsheetId || "",
        inventorySpreadsheetId: settings.inventorySpreadsheetId || "",
        inventorySheetName: settings.inventorySheetName || "",
        inventorySelectedColumns: settings.inventorySelectedColumns
          ? JSON.parse(settings.inventorySelectedColumns)
          : [],
        orderSpreadsheetId: settings.orderSpreadsheetId || "",
      };
    } catch (error) {
      logger.error(`Error getting Google Sheets settings: ${error}`);
      throw error;
    }
  }

  private getEmptySettings(): GoogleSheetsSettingsData {
    return {
      spreadsheetId: "",
      selectedSheet: "",
      manualMode: false,
      availableSpreadsheets: [],
      availableWorksheets: [],
      inventoryAvailableWorksheets: [],
      walletAvailableWorksheets: [],
      shippingAvailableWorksheets: [],
      orderAvailableWorksheets: [],
      walletSpreadsheetId: "",
      shippingSpreadsheetId: "",
      inventorySpreadsheetId: "",
      inventorySheetName: "",
      inventorySelectedColumns: [],
      orderSpreadsheetId: "",
    };
  }

  /**
   * Save Google Sheets settings
   * Safely handles missing columns in older database schemas
   * @param data - Settings data to save
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async saveSettings(
    data: GoogleSheetsSettingsData,
    tenantId?: string
  ): Promise<void> {
    try {
      const prisma = getPrisma(tenantId);

      // Try to get or create settings record, but handle column errors
      let settings: any = null;
      try {
        settings = await prisma.googleSheetsSettings.findFirst();
      } catch (error: any) {
        // If table or column doesn't exist, try to create it
        if (
          error.code === "P2022" ||
          error.message.includes("does not exist")
        ) {
          logger.warn(
            "GoogleSheetsSettings table columns not yet synced, attempting create..."
          );
          // Continue with create operation
        } else {
          throw error;
        }
      }

      const updateData = {
        spreadsheetId: data.spreadsheetId || null,
        selectedSheet: data.selectedSheet || null,
        manualMode: data.manualMode ?? false,
        availableSpreadsheets: data.availableSpreadsheets
          ? JSON.stringify(data.availableSpreadsheets)
          : null,
        availableWorksheets: data.availableWorksheets
          ? JSON.stringify(data.availableWorksheets)
          : null,
        inventoryAvailableWorksheets: data.inventoryAvailableWorksheets
          ? JSON.stringify(data.inventoryAvailableWorksheets)
          : null,
        walletAvailableWorksheets: data.walletAvailableWorksheets
          ? JSON.stringify(data.walletAvailableWorksheets)
          : null,
        shippingAvailableWorksheets: data.shippingAvailableWorksheets
          ? JSON.stringify(data.shippingAvailableWorksheets)
          : null,
        orderAvailableWorksheets: data.orderAvailableWorksheets
          ? JSON.stringify(data.orderAvailableWorksheets)
          : null,
        // Only update these if explicitly provided (not undefined)
        // If undefined, keep existing value from database
        walletSpreadsheetId:
          data.walletSpreadsheetId !== undefined
            ? data.walletSpreadsheetId
            : settings?.walletSpreadsheetId,
        shippingSpreadsheetId:
          data.shippingSpreadsheetId !== undefined
            ? data.shippingSpreadsheetId
            : settings?.shippingSpreadsheetId,
        inventorySpreadsheetId:
          data.inventorySpreadsheetId !== undefined
            ? data.inventorySpreadsheetId
            : settings?.inventorySpreadsheetId,
        inventorySheetName:
          data.inventorySheetName !== undefined
            ? data.inventorySheetName
            : settings?.inventorySheetName,
        inventorySelectedColumns: data.inventorySelectedColumns
          ? JSON.stringify(data.inventorySelectedColumns)
          : null,
        orderSpreadsheetId:
          data.orderSpreadsheetId !== undefined
            ? data.orderSpreadsheetId
            : settings?.orderSpreadsheetId,
      };

      try {
        if (settings) {
          await prisma.googleSheetsSettings.update({
            where: { id: settings.id },
            data: updateData,
          });
        } else {
          await prisma.googleSheetsSettings.create({
            data: updateData,
          });
        }
      } catch (dbError: any) {
        // If schema mismatch, silently continue - settings are saved in frontend cache
        if (
          dbError.code === "P2022" ||
          dbError.message.includes("does not exist")
        ) {
          logger.warn(
            "Schema mismatch - settings will be stored in frontend cache only"
          );
        } else {
          throw dbError;
        }
      }
    } catch (error) {
      logger.error(`Error saving Google Sheets settings: ${error}`);
      throw error;
    }
  }

  /**
   * Update available spreadsheets
   * @param spreadsheets - List of spreadsheets
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async updateAvailableSpreadsheets(
    spreadsheets: any[],
    tenantId?: string
  ): Promise<void> {
    try {
      const prisma = getPrisma(tenantId);
      const settings = await prisma.googleSheetsSettings.findFirst();

      if (!settings) {
        await prisma.googleSheetsSettings.create({
          data: {
            availableSpreadsheets: JSON.stringify(spreadsheets),
          },
        });
      } else {
        await prisma.googleSheetsSettings.update({
          where: { id: settings.id },
          data: {
            availableSpreadsheets: JSON.stringify(spreadsheets),
          },
        });
      }
    } catch (error) {
      logger.error(`Error updating available spreadsheets: ${error}`);
      throw error;
    }
  }

  /**
   * Clear settings
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async clearSettings(tenantId?: string): Promise<void> {
    try {
      const prisma = getPrisma(tenantId);
      await prisma.googleSheetsSettings.deleteMany();
    } catch (error) {
      logger.error(`Error clearing settings: ${error}`);
      throw error;
    }
  }
}

export const googleSheetsSettingsService = new GoogleSheetsSettingsService();

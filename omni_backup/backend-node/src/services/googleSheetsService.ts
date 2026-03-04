import { google } from "googleapis";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { getGoogleSheetsAuthService } from "./googleSheetsAuthService";
import { GoogleSheetsAPIOperations } from "./googleSheetsAPIOperations";
import { GoogleSheetsSettingsManager } from "./googleSheetsSettingsManager";

class GoogleSheetsService {
  private apiOperations: GoogleSheetsAPIOperations | null = null;
  private settingsManager: GoogleSheetsSettingsManager | null = null;
  private authService = getGoogleSheetsAuthService();
  private logger: Logger;

  constructor() {
    this.logger = getLogger("GoogleSheetsService");
    this.initializeApis();
  }

  private initializeApis(): void {
    try {
      const authClient = this.authService.getActiveAuthClient();

      if (!authClient) {
        this.logger.warn("Service account not initialized");
        return;
      }

      const sheetsAPI = google.sheets({ version: "v4", auth: authClient });
      const driveAPI = google.drive({ version: "v3", auth: authClient });

      this.apiOperations = new GoogleSheetsAPIOperations(sheetsAPI, driveAPI);
      this.apiOperations.setAuthService(this.authService);

      this.settingsManager = new GoogleSheetsSettingsManager(
        this.apiOperations
      );

      this.logger.info("✅ Google Sheets APIs initialized");
    } catch (error) {
      this.logger.error("Error initializing APIs");
    }
  }

  /**
   * Ensure service account is initialized
   */
  private async ensureReady(): Promise<boolean> {
    try {
      const isReady = await this.authService.ensureInitialized();
      if (!isReady) {
        this.logger.warn("⚠️ Service account not initialized");
        return false;
      }

      // Initialize APIs if not already done
      if (!this.apiOperations) {
        this.initializeApis();
      }

      return this.apiOperations !== null;
    } catch (error) {
      this.logger.error("Error ensuring readiness", error);
      return false;
    }
  }

  async getAuthStatus() {
    return this.authService.getAuthStatus();
  }

  isAuthenticated(): boolean {
    return this.authService.isAuthenticated();
  }

  async getSpreadsheets() {
    await this.ensureReady();
    return this.apiOperations?.getSpreadsheets() || [];
  }

  async getWorksheets(spreadsheetId: string) {
    await this.ensureReady();
    return this.apiOperations?.getWorksheets(spreadsheetId) || [];
  }

  async getColumnHeaders(spreadsheetId: string, sheetName: string) {
    await this.ensureReady();
    return this.apiOperations?.getColumnHeaders(spreadsheetId, sheetName) || [];
  }

  async getAllData(
    spreadsheetId: string,
    sheetName: string,
    startRow: number = 2
  ) {
    await this.ensureReady();
    return (
      this.apiOperations?.getAllData(spreadsheetId, sheetName, startRow) || []
    );
  }

  async clearAndWriteData(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    await this.ensureReady();

    if (!this.apiOperations) return false;

    const cleanedSheetName = this.apiOperations.cleanSheetName(sheetName);
    const ensured = await this.apiOperations.ensureSheetExists(
      spreadsheetId,
      cleanedSheetName
    );

    if (!ensured) {
      this.logger.error("Failed to ensure sheet exists");
      return false;
    }

    return this.apiOperations.clearAndWriteData(
      spreadsheetId,
      cleanedSheetName,
      values
    );
  }

  async createSpreadsheet(title: string) {
    await this.ensureReady();
    return this.apiOperations?.createSpreadsheet(title) || null;
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
    await this.ensureReady();
    return (
      this.settingsManager?.updateDetailedSettings(
        walletSpreadsheetId,
        shippingSpreadsheetId,
        inventorySpreadsheetId,
        orderSpreadsheetId,
        inventorySheetName,
        inventorySelectedColumns,
        inventoryAvailableWorksheets,
        walletAvailableWorksheets,
        shippingAvailableWorksheets,
        orderAvailableWorksheets,
        tenantId
      ) || false
    );
  }

  async getDetailedSettings(tenantId?: string): Promise<Record<string, any>> {
    await this.ensureReady();
    return this.settingsManager?.getDetailedSettings(tenantId) || {};
  }

  async testConnection(spreadsheetId?: string): Promise<boolean> {
    await this.ensureReady();
    return this.apiOperations?.testConnection(spreadsheetId) || false;
  }

  async refreshSpreadsheets() {
    return this.getSpreadsheets();
  }

  async getSheetData(spreadsheetId: string, sheetName: string) {
    await this.ensureReady();
    return this.apiOperations?.getSheetData(spreadsheetId, sheetName) || [];
  }

  async appendRows(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    await this.ensureReady();
    return (
      this.apiOperations?.appendRows(spreadsheetId, sheetName, values) || false
    );
  }

  async updateRange(
    spreadsheetId: string,
    range: string,
    values: any[][]
  ): Promise<boolean> {
    await this.ensureReady();
    return (
      this.apiOperations?.updateRange(spreadsheetId, range, values) || false
    );
  }

  async uploadToSheet(options: {
    spreadsheetId: string;
    sheetName: string;
    data: any[][];
  }): Promise<boolean> {
    return this.clearAndWriteData(
      options.spreadsheetId,
      options.sheetName,
      options.data
    );
  }

  /**
   * Save spreadsheet links (URLs) for different purposes
   * @param links - Object with URLs
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
    await this.ensureReady();
    return this.settingsManager?.saveSpreadsheetLinks(links, tenantId) || false;
  }

  /**
   * Get saved spreadsheet links
   * @param tenantId - Optional tenant ID for multi-tenant support
   */
  async getSavedSpreadsheetLinks(tenantId?: string): Promise<any> {
    await this.ensureReady();
    return this.settingsManager?.getSavedSpreadsheetLinks(tenantId) || {};
  }
}

let googleSheetsServiceInstance: GoogleSheetsService | null = null;

export function getGoogleSheetsService(): GoogleSheetsService {
  if (!googleSheetsServiceInstance) {
    googleSheetsServiceInstance = new GoogleSheetsService();
  }
  return googleSheetsServiceInstance;
}

export { GoogleSheetsService };

/**
 * Google Sheets API Operations Coordinator
 * Coordinates API operations with retry and error handling
 *
 * Single Responsibility: Coordinating API calls with resilience
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { GoogleSheetsErrorHandler } from "../utils/googleSheetsErrorHandler";
import { GoogleSheetsAPIRetryHandler } from "../utils/googleSheetsAPIRetryHandler";
import { GoogleSheetsDataOperations } from "./googleSheetsDataOperations";
import {
  GoogleSheetsSheetOperations,
  Sheet,
} from "./googleSheetsSheetOperations";

export interface Spreadsheet {
  id: string;
  name: string;
}

export { Sheet } from "./googleSheetsSheetOperations";

export class GoogleSheetsAPIOperations {
  private sheetsAPI: any;
  private driveAPI: any;
  private logger: Logger;
  private errorHandler: GoogleSheetsErrorHandler;
  private retryHandler: GoogleSheetsAPIRetryHandler;
  private dataOps: GoogleSheetsDataOperations;
  private sheetOps: GoogleSheetsSheetOperations;

  constructor(sheetsAPI: any, driveAPI: any) {
    this.sheetsAPI = sheetsAPI;
    this.driveAPI = driveAPI;
    this.logger = getLogger("GoogleSheetsAPIOperations");
    this.errorHandler = new GoogleSheetsErrorHandler(this.logger);
    this.retryHandler = new GoogleSheetsAPIRetryHandler(this.logger);
    this.dataOps = new GoogleSheetsDataOperations(sheetsAPI);
    this.sheetOps = new GoogleSheetsSheetOperations(sheetsAPI);
  }

  /**
   * Set auth service for token refresh
   */
  setAuthService(authService: any): void {
    this.retryHandler.setAuthService(authService);
  }

  async getSpreadsheets(): Promise<Spreadsheet[]> {
    try {
      if (!this.driveAPI) {
        this.logger.warn("Drive API not initialized");
        return [];
      }

      return (
        (await this.retryHandler.execute(async () => {
          const baseParams: any = {
            q: "mimeType='application/vnd.google-apps.spreadsheet' and trashed=false",
            fields: "files(id, name, driveId)",
            pageSize: 100,
            orderBy: "name",
            supportsAllDrives: true,
            includeItemsFromAllDrives: true,
          };

          const userResp = await this.driveAPI.files.list({
            ...baseParams,
            corpora: "user",
            spaces: "drive",
          });

          let files: any[] = userResp.data.files || [];

          if (!files.length) {
            const allDrivesResp = await this.driveAPI.files.list({
              ...baseParams,
              corpora: "allDrives",
            });
            files = allDrivesResp.data.files || [];
          }

          return files.map((file: any) => ({ id: file.id, name: file.name }));
        }, "GoogleSheetsAPIOperations.getSpreadsheets")) || []
      );
    } catch (error) {
      this.errorHandler.logError(
        "getSpreadsheets",
        error,
        "error"
      );
      if (this.errorHandler.isTokenError(error)) {
        this.logger.warn(
          `Invalid_grant detected. User must re-authenticate at /api/google/auth/initiate`
        );
      }
      return [];
    }
  }

  async getWorksheets(spreadsheetId: string): Promise<Sheet[]> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return [];
      }

      return (
        (await this.retryHandler.execute(async () => {
          const response = await this.sheetsAPI.spreadsheets.get({
            spreadsheetId,
          });
          return (response.data.sheets || []).map((sheet: any) => ({
            title: sheet.properties.title,
            id: sheet.properties.sheetId,
            index: sheet.properties.index,
            sheetType: sheet.properties.sheetType,
          }));
        }, `GoogleSheetsAPIOperations.getWorksheets(${spreadsheetId})`)) || []
      );
    } catch (error) {
      this.errorHandler.logError(
        "getWorksheets",
        error,
        "error"
      );
      if (this.errorHandler.isTokenError(error)) {
        this.logger.warn(
          `Invalid_grant detected. User must re-authenticate at /api/google/auth/initiate`
        );
      }
      return [];
    }
  }

  async getColumnHeaders(
    spreadsheetId: string,
    sheetName: string
  ): Promise<string[]> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return [];
      }

      return (
        (await this.retryHandler.execute(async () => {
          const range = `${sheetName}!1:1`;
          const response = await this.sheetsAPI.spreadsheets.values.get({
            spreadsheetId,
            range,
          });

          return (response.data.values?.[0] || []).filter(
            (h: string) => h && h.trim()
          );
        }, `GoogleSheetsAPIOperations.getColumnHeaders(${sheetName})`)) || []
      );
    } catch (error) {
      this.errorHandler.logError(
        "getColumnHeaders",
        error,
        "error"
      );
      if (this.errorHandler.isTokenError(error)) {
        this.logger.warn(
          `Invalid_grant detected. User must re-authenticate at /api/google/auth/initiate`
        );
      }
      return [];
    }
  }

  // Delegate data operations
  async getAllData(
    spreadsheetId: string,
    sheetName: string,
    startRow?: number
  ): Promise<any[]> {
    return this.dataOps.getAllData(spreadsheetId, sheetName, startRow);
  }

  async getSheetData(
    spreadsheetId: string,
    sheetName: string
  ): Promise<any[][]> {
    return this.dataOps.getSheetData(spreadsheetId, sheetName);
  }

  async clearAndWriteData(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    return this.dataOps.clearAndWriteData(spreadsheetId, sheetName, values);
  }

  async appendRows(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    return this.dataOps.appendRows(spreadsheetId, sheetName, values);
  }

  async updateRange(
    spreadsheetId: string,
    range: string,
    values: any[][]
  ): Promise<boolean> {
    return this.dataOps.updateRange(spreadsheetId, range, values);
  }

  // Delegate sheet operations
  async ensureSheetExists(
    spreadsheetId: string,
    sheetName: string
  ): Promise<boolean> {
    return this.sheetOps.ensureSheetExists(spreadsheetId, sheetName);
  }

  async createSpreadsheet(title: string): Promise<Spreadsheet | null> {
    return this.sheetOps.createSpreadsheet(title);
  }

  cleanSheetName(name: string): string {
    return this.sheetOps.cleanSheetName(name);
  }

  async testConnection(spreadsheetId?: string): Promise<boolean> {
    try {
      if (spreadsheetId) {
        await this.sheetsAPI.spreadsheets.get({
          spreadsheetId,
          fields: "spreadsheetId",
        });
      } else {
        await this.driveAPI.files.list({ pageSize: 1, fields: "files(id)" });
      }
      return true;
    } catch (error) {
      this.logger.error("Connection test failed");
      return false;
    }
  }
}

/**
 * Google Sheets Data Service
 * Handles data import/export with service account rotation
 *
 * Single Responsibility: Data operations with rotation support
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { getGoogleSheetsAuthService } from "./googleSheetsAuthService";
import { google } from "googleapis";
import { backendLogger } from "../utils/backendLogger";

export class GoogleSheetsDataService {
  private logger: Logger;
  private authService = getGoogleSheetsAuthService();

  constructor() {
    this.logger = getLogger("GoogleSheetsDataService");
  }

  /**
   * Read data from spreadsheet with auto-rotation on error
   */
  async readSheetData(
    spreadsheetId: string,
    sheetName: string,
    range: string = "A1:Z1000"
  ): Promise<any[][] | null> {
    try {
      await this.authService.ensureInitialized();

      let authClient = this.authService.getActiveAuthClient();
      if (!authClient) {
        this.logger.error("No service account available");
        return null;
      }

      const sheets = google.sheets({ version: "v4", auth: authClient });

      try {
        const response = await sheets.spreadsheets.values.get({
          spreadsheetId,
          range: `${sheetName}!${range}`,
        });

        this.authService.recordUsage();
        return response.data.values || [];
      } catch (error: any) {
        // Check if it's a quota error
        if (error.message?.includes("quota") || error.status === 429) {
          this.logger.warn("Quota limit reached, rotating service account");
          const rotated = await this.authService.rotateServiceAccount("limit");

          if (rotated) {
            authClient = this.authService.getActiveAuthClient();
            if (!authClient) return null;

            const sheetsRetry = google.sheets({
              version: "v4",
              auth: authClient,
            });
            const response = await sheetsRetry.spreadsheets.values.get({
              spreadsheetId,
              range: `${sheetName}!${range}`,
            });

            this.authService.recordUsage();
            return response.data.values || [];
          }
        }

        throw error;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error reading sheet data: ${errorMsg}`);
      return null;
    }
  }

  /**
   * Write data to spreadsheet with auto-rotation on error
   */
  async writeSheetData(
    spreadsheetId: string,
    sheetName: string,
    range: string,
    values: any[][]
  ): Promise<boolean> {
    try {
      await this.authService.ensureInitialized();

      let authClient = this.authService.getActiveAuthClient();
      if (!authClient) {
        this.logger.error("No service account available");
        return false;
      }

      const sheets = google.sheets({ version: "v4", auth: authClient });

      try {
        await sheets.spreadsheets.values.update({
          spreadsheetId,
          range: `${sheetName}!${range}`,
          valueInputOption: "RAW",
          requestBody: { values },
        });

        this.authService.recordUsage();
        backendLogger.success(
          "GoogleSheetsDataService",
          `✅ Data written to ${sheetName}`
        );
        return true;
      } catch (error: any) {
        // Check if it's a quota error
        if (error.message?.includes("quota") || error.status === 429) {
          this.logger.warn("Quota limit reached, rotating service account");
          const rotated = await this.authService.rotateServiceAccount("limit");

          if (rotated) {
            authClient = this.authService.getActiveAuthClient();
            if (!authClient) return false;

            const sheetsRetry = google.sheets({
              version: "v4",
              auth: authClient,
            });
            await sheetsRetry.spreadsheets.values.update({
              spreadsheetId,
              range: `${sheetName}!${range}`,
              valueInputOption: "RAW",
              requestBody: { values },
            });

            this.authService.recordUsage();
            return true;
          }
        }

        throw error;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error writing sheet data: ${errorMsg}`);
      return false;
    }
  }

  /**
   * Append rows to spreadsheet
   */
  async appendRows(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    try {
      await this.authService.ensureInitialized();

      const authClient = this.authService.getActiveAuthClient();
      if (!authClient) {
        this.logger.error("No service account available");
        return false;
      }

      const sheets = google.sheets({ version: "v4", auth: authClient });

      try {
        await sheets.spreadsheets.values.append({
          spreadsheetId,
          range: `${sheetName}!A:Z`,
          valueInputOption: "RAW",
          requestBody: { values },
        });

        this.authService.recordUsage();
        return true;
      } catch (error: any) {
        if (error.message?.includes("quota") || error.status === 429) {
          const rotated = await this.authService.rotateServiceAccount("limit");
          if (!rotated) return false;

          const authClientRetry = this.authService.getActiveAuthClient();
          if (!authClientRetry) return false;

          const sheetsRetry = google.sheets({
            version: "v4",
            auth: authClientRetry,
          });
          await sheetsRetry.spreadsheets.values.append({
            spreadsheetId,
            range: `${sheetName}!A:Z`,
            valueInputOption: "RAW",
            requestBody: { values },
          });

          this.authService.recordUsage();
          return true;
        }
        throw error;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error appending rows: ${errorMsg}`);
      return false;
    }
  }

  /**
   * Clear range
   */
  async clearRange(spreadsheetId: string, range: string): Promise<boolean> {
    try {
      await this.authService.ensureInitialized();

      const authClient = this.authService.getActiveAuthClient();
      if (!authClient) return false;

      const sheets = google.sheets({ version: "v4", auth: authClient });

      await sheets.spreadsheets.values.clear({
        spreadsheetId,
        range,
      });

      this.authService.recordUsage();
      return true;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error clearing range: ${errorMsg}`);
      return false;
    }
  }

  /**
   * Get account statistics
   */
  getAccountStats(): Record<string, any> {
    return this.authService.getAccountStats();
  }
}

let dataServiceInstance: GoogleSheetsDataService | null = null;

export function getGoogleSheetsDataService(): GoogleSheetsDataService {
  if (!dataServiceInstance) {
    dataServiceInstance = new GoogleSheetsDataService();
  }
  return dataServiceInstance;
}

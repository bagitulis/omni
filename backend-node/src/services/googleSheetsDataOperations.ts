/**
 * Google Sheets Data Operations
 * Handles data read/write operations with auto-rotation on quota limits
 *
 * Single Responsibility: Data operations with retry logic
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { getServiceAccountRotationManager } from "./googleSheetsServiceAccountRotationManager";

const RETRY_DELAY_MS = 2000; // Wait 2 seconds before retry after 429

export class GoogleSheetsDataOperations {
  private sheetsAPI: any;
  private logger: Logger;

  constructor(sheetsAPI: any) {
    this.sheetsAPI = sheetsAPI;
    this.logger = getLogger("GoogleSheetsDataOperations");
  }

  async getAllData(
    spreadsheetId: string,
    sheetName: string,
    startRow: number = 2
  ): Promise<any[]> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return [];
      }

      const range = `${sheetName}!${startRow}:1000`;
      const response = await this.sheetsAPI.spreadsheets.values.get({
        spreadsheetId,
        range,
      });

      return response.data.values || [];
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(
        `Error getting all data from '${sheetName}': ${errorMsg}`
      );
      return [];
    }
  }

  async getSheetData(
    spreadsheetId: string,
    sheetName: string
  ): Promise<any[][]> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return [];
      }

      const range = `${sheetName}!A:Z`;
      const response = await this.sheetsAPI.spreadsheets.values.get({
        spreadsheetId,
        range,
      });

      return response.data.values || [];
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(
        `Error getting sheet data from '${sheetName}': ${errorMsg}`
      );
      return [];
    }
  }

  async clearAndWriteData(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return false;
      }

      const clearRange = `${sheetName}!A:Z`;
      await this.sheetsAPI.spreadsheets.values.clear({
        spreadsheetId,
        range: clearRange,
      });

      const range = `${sheetName}!A1`;
      await this.sheetsAPI.spreadsheets.values.update({
        spreadsheetId,
        range,
        valueInputOption: "RAW",
        requestBody: { values },
      });

      return true;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      const errorCode = (error as any)?.code || "UNKNOWN";
      this.logger.error(
        `Error clearing/writing data to '${sheetName}': ${errorMsg} (Code: ${errorCode})`
      );
      return false;
    }
  }

  async appendRows(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return false;
      }

      const range = `${sheetName}!A:Z`;
      await this.sheetsAPI.spreadsheets.values.append({
        spreadsheetId,
        range,
        valueInputOption: "RAW",
        requestBody: { values },
      });

      return true;
    } catch (error: any) {
      const status = error?.status || error?.code || "UNKNOWN";
      const errorMsg = error instanceof Error ? error.message : String(error);

      // Handle quota exceeded (429)
      if (status === 429) {
        return this.handleQuotaExceededAppend(spreadsheetId, sheetName, values);
      }

      // Other errors
      this.logger.error(
        `Error appending rows to '${sheetName}': ${errorMsg} (Status: ${status})`
      );
      return false;
    }
  }

  /**
   * Handle 429 quota exceeded errors for append operations
   */
  private async handleQuotaExceededAppend(
    spreadsheetId: string,
    sheetName: string,
    values: any[][]
  ): Promise<boolean> {
    const rotationManager = getServiceAccountRotationManager();
    const activeAccount = rotationManager.getActiveAccount();

    this.logger.warn(
      `⚠️ Quota exceeded on append for ${activeAccount?.clientEmail} - attempting rotation...`
    );

    // Try to rotate
    const rotated = await rotationManager.rotateServiceAccount("limit");

    if (!rotated) {
      this.logger.error(
        `❌ Could not rotate service account for append (only ${
          rotationManager.getAllAccounts().length
        } available)`
      );
      return false;
    }

    // Wait before retrying
    await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY_MS));

    try {
      const newActiveAccount = rotationManager.getActiveAccount();
      const range = `${sheetName}!A:Z`;

      this.logger.info(
        `🔄 Retrying append with rotated account: ${newActiveAccount?.clientEmail}`
      );

      await this.sheetsAPI.spreadsheets.values.append({
        spreadsheetId,
        range,
        valueInputOption: "RAW",
        requestBody: { values },
      });

      this.logger.info(`✅ Append succeeded with rotated account`);
      return true;
    } catch (retryError: any) {
      const retryStatus = retryError?.status || retryError?.code || "UNKNOWN";
      const retryMsg =
        retryError instanceof Error ? retryError.message : String(retryError);

      this.logger.error(
        `❌ Retry append failed: ${retryMsg} (Status: ${retryStatus})`
      );
      return false;
    }
  }

  async updateRange(
    spreadsheetId: string,
    range: string,
    values: any[][]
  ): Promise<boolean> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return false;
      }

      await this.sheetsAPI.spreadsheets.values.update({
        spreadsheetId,
        range,
        valueInputOption: "RAW",
        requestBody: { values },
      });

      return true;
    } catch (error: any) {
      const status = error?.status || error?.code || "UNKNOWN";
      const errorMsg = error instanceof Error ? error.message : String(error);

      // Handle quota exceeded (429)
      if (status === 429) {
        return this.handleQuotaExceeded(
          "updateRange",
          spreadsheetId,
          range,
          values
        );
      }

      // Other errors
      this.logger.error(
        `Error updating range '${range}': ${errorMsg} (Code: ${status})`
      );
      return false;
    }
  }

  /**
   * Handle 429 quota exceeded errors with service account rotation
   */
  private async handleQuotaExceeded(
    operation: string,
    spreadsheetId: string,
    range: string,
    values: any[][]
  ): Promise<boolean> {
    const rotationManager = getServiceAccountRotationManager();
    const activeAccount = rotationManager.getActiveAccount();

    this.logger.warn(
      `⚠️ Quota exceeded for ${activeAccount?.clientEmail} - attempting rotation...`
    );

    // Try to rotate to next service account
    const rotated = await rotationManager.rotateServiceAccount("limit");

    if (!rotated) {
      this.logger.error(
        `❌ Could not rotate service account (only ${
          rotationManager.getAllAccounts().length
        } available)`
      );
      return false;
    }

    // Wait before retrying
    await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY_MS));

    // Retry the operation (without recursion to prevent infinite loops)
    try {
      const newActiveAccount = rotationManager.getActiveAccount();
      this.logger.info(
        `🔄 Retrying ${operation} with rotated account: ${newActiveAccount?.clientEmail}`
      );

      await this.sheetsAPI.spreadsheets.values.update({
        spreadsheetId,
        range,
        valueInputOption: "RAW",
        requestBody: { values },
      });

      this.logger.info(`✅ ${operation} succeeded with rotated account`);
      return true;
    } catch (retryError: any) {
      const retryStatus = retryError?.status || retryError?.code || "UNKNOWN";
      const retryMsg =
        retryError instanceof Error ? retryError.message : String(retryError);

      this.logger.error(
        `❌ Retry failed for ${operation}: ${retryMsg} (Status: ${retryStatus})`
      );
      return false;
    }
  }
}

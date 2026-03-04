/**
 * Google Sheets Sheet Management Operations
 * Handles sheet creation and management
 *
 * Single Responsibility: Sheet management only
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";

export interface Sheet {
  title: string;
  id: string;
  index: number;
  sheetType: string;
}

export class GoogleSheetsSheetOperations {
  private sheetsAPI: any;
  private logger: Logger;

  constructor(sheetsAPI: any) {
    this.sheetsAPI = sheetsAPI;
    this.logger = getLogger("GoogleSheetsSheetOperations");
  }

  async ensureSheetExists(
    spreadsheetId: string,
    sheetName: string
  ): Promise<boolean> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return false;
      }

      const spreadsheet = await this.sheetsAPI.spreadsheets.get({
        spreadsheetId,
      });
      const sheets = spreadsheet.data.sheets || [];
      const sheetExists = sheets.some(
        (sheet: any) => sheet.properties?.title === sheetName
      );

      if (!sheetExists) {
        this.logger.info(`Creating new worksheet: '${sheetName}'`);
        await this.sheetsAPI.spreadsheets.batchUpdate({
          spreadsheetId,
          requestBody: {
            requests: [{ addSheet: { properties: { title: sheetName } } }],
          },
        });
      }

      return true;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(
        `Error ensuring sheet exists '${sheetName}': ${errorMsg}`
      );
      return false;
    }
  }

  async createSpreadsheet(
    title: string
  ): Promise<{ id: string; name: string } | null> {
    try {
      if (!this.sheetsAPI) {
        this.logger.warn("Sheets API not initialized");
        return null;
      }

      const response = await this.sheetsAPI.spreadsheets.create({
        resource: { properties: { title } },
      });

      return {
        id: response.data.spreadsheetId,
        name: response.data.properties.title,
      };
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error creating spreadsheet '${title}': ${errorMsg}`);
      return null;
    }
  }

  cleanSheetName(name: string): string {
    const invalidChars = [":", "/", "?", "*", "[", "]"];
    let cleaned = name;
    for (const char of invalidChars) {
      cleaned = cleaned.replace(new RegExp(`\\${char}`, "g"), "");
    }
    return cleaned.trim().substring(0, 31);
  }
}

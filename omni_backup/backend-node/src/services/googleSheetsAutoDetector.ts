/**
 * Google Sheets Auto Detector
 * Detects spreadsheet metadata from URL and sheets structure
 *
 * Single Responsibility: Auto-detect spreadsheet metadata
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import { google } from "googleapis";
import { SheetMetadata } from "./spreadsheetRegistryService";

export class GoogleSheetsAutoDetector {
  private logger: Logger;

  constructor() {
    this.logger = getLogger("GoogleSheetsAutoDetector");
  }

  /**
   * Extract spreadsheet ID from Google Sheets URL
   */
  extractSpreadsheetId(url: string): string | null {
    try {
      // Match patterns: /d/{id}/ or /d/{id}
      const match = url.match(/\/d\/([a-zA-Z0-9-_]+)/);
      return match ? match[1] : null;
    } catch (error) {
      this.logger.error("Error extracting spreadsheet ID", error);
      return null;
    }
  }

  /**
   * Detect spreadsheet metadata from URL using service account
   */
  async detectMetadata(
    spreadsheetUrl: string,
    authClient: any
  ): Promise<{
    spreadsheetId: string;
    name: string;
    sheets: SheetMetadata[];
  } | null> {
    try {
      const spreadsheetId = this.extractSpreadsheetId(spreadsheetUrl);

      if (!spreadsheetId) {
        this.logger.warn("Invalid spreadsheet URL");
        return null;
      }

      const sheets = google.sheets({
        version: "v4",
        auth: authClient,
      });

      const response = await sheets.spreadsheets.get({
        spreadsheetId,
        fields: "properties,sheets",
      });

      const spreadsheetName = response.data.properties?.title || "Untitled";
      const sheetsMetadata: SheetMetadata[] = (response.data.sheets || []).map(
        (sheet: any, index: number) => ({
          name: sheet.properties?.title || `Sheet${index + 1}`,
          sheetId: sheet.properties?.sheetId || index,
          index: sheet.properties?.index || index,
          columnCount: sheet.properties?.gridProperties?.columnCount,
          rowCount: sheet.properties?.gridProperties?.rowCount,
        })
      );

      this.logger.info(
        `✅ Detected spreadsheet: ${spreadsheetName} with ${sheetsMetadata.length} sheets`
      );

      return {
        spreadsheetId,
        name: spreadsheetName,
        sheets: sheetsMetadata,
      };
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error detecting metadata: ${errorMsg}`);
      return null;
    }
  }

  /**
   * Validate spreadsheet is accessible
   */
  async validateAccess(
    spreadsheetUrl: string,
    authClient: any
  ): Promise<boolean> {
    try {
      const spreadsheetId = this.extractSpreadsheetId(spreadsheetUrl);
      if (!spreadsheetId) return false;

      const sheets = google.sheets({
        version: "v4",
        auth: authClient,
      });

      await sheets.spreadsheets.get({
        spreadsheetId,
        fields: "properties",
      });

      return true;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.warn(`Spreadsheet not accessible: ${errorMsg}`);
      return false;
    }
  }

  /**
   * Get specific sheet data
   */
  async getSheetData(
    spreadsheetId: string,
    sheetName: string,
    authClient: any,
    range: string = "A1:Z1000"
  ): Promise<any[][] | null> {
    try {
      const sheets = google.sheets({
        version: "v4",
        auth: authClient,
      });

      const response = await sheets.spreadsheets.values.get({
        spreadsheetId,
        range: `${sheetName}!${range}`,
      });

      return response.data.values || [];
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error getting sheet data: ${errorMsg}`);
      return null;
    }
  }

  /**
   * Detect column headers from sheet
   */
  async detectColumnHeaders(
    spreadsheetId: string,
    sheetName: string,
    authClient: any
  ): Promise<string[] | null> {
    try {
      const data = await this.getSheetData(
        spreadsheetId,
        sheetName,
        authClient,
        "A1:Z1"
      );

      if (!data || data.length === 0) {
        this.logger.warn("No data found in sheet");
        return [];
      }

      return data[0].filter((col: any) => col !== undefined && col !== "");
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      this.logger.error(`Error detecting column headers: ${errorMsg}`);
      return null;
    }
  }
}

// Singleton instance
let autoDetectorInstance: GoogleSheetsAutoDetector | null = null;

export function getGoogleSheetsAutoDetector(): GoogleSheetsAutoDetector {
  if (!autoDetectorInstance) {
    autoDetectorInstance = new GoogleSheetsAutoDetector();
  }
  return autoDetectorInstance;
}

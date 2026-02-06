/**
 * Google Sheets Storage Service
 * Handles persistent storage of Google Sheets data
 * Uses centralized cache service for consistency
 * Single Responsibility: Google Sheets data persistence
 */

import cacheService from "./cacheService";

interface SheetMetadata {
  name: string;
  sheetId: number;
  index: number;
  columnCount: number;
  rowCount: number;
}

interface SheetsCache {
  timestamp: number;
  data: SheetMetadata[];
  spreadsheetId: string;
  type: "inventory" | "wallet" | "shipping" | "order";
}

const STORAGE_KEYS = {
  SHEETS_CACHE: "google_sheets_cache",
  LINKS: "google_sheets_links",
  SELECTED_SHEETS: "google_sheets_selected_sheets",
  SETTINGS: "google_sheets_settings",
};

const CACHE_TTL = 30 * 60 * 1000; // 30 minutes

class GoogleSheetsStorageService {
  /**
   * Save sheets metadata to cache
   */
  saveSheetsCacheLocal(
    type: "inventory" | "wallet" | "shipping" | "order",
    sheets: SheetMetadata[],
    spreadsheetId: string
  ): void {
    try {
      const cache = this.getSheetsCacheLocal();
      cache[type] = {
        timestamp: Date.now(),
        data: sheets,
        spreadsheetId,
        type,
      };
      cacheService.set(STORAGE_KEYS.SHEETS_CACHE, cache, { ttl: CACHE_TTL });
    } catch (error) {
      console.warn("Failed to save sheets cache", error);
    }
  }

  /**
   * Get cached sheets metadata
   */
  getSheetsCacheLocal(
    type?: "inventory" | "wallet" | "shipping" | "order"
  ): Record<string, SheetsCache> {
    try {
      const cache =
        cacheService.get<Record<string, SheetsCache>>(
          STORAGE_KEYS.SHEETS_CACHE,
          { ttl: CACHE_TTL }
        ) || {};

      return type ? (cache[type] ? { [type]: cache[type] } : {}) : cache;
    } catch (error) {
      console.warn("Failed to read sheets cache", error);
      return {};
    }
  }

  /**
   * Clear expired cache entries
   */
  clearExpiredCache(): void {
    cacheService.removeByPrefix(STORAGE_KEYS.SHEETS_CACHE);
  }

  /**
   * Save spreadsheet links
   */
  saveLinksLocal(links: {
    inventory?: string;
    wallet?: string;
    shipping?: string;
    order?: string;
  }): void {
    try {
      cacheService.set(STORAGE_KEYS.LINKS, links);
    } catch (error) {
      console.warn("Failed to save links", error);
    }
  }

  /**
   * Get saved spreadsheet links
   */
  getLinksLocal(): {
    inventory?: string;
    wallet?: string;
    shipping?: string;
    order?: string;
  } {
    try {
      return cacheService.get<any>(STORAGE_KEYS.LINKS) || {};
    } catch (error) {
      console.warn("Failed to get links", error);
      return {};
    }
  }

  /**
   * Save selected sheet name per type
   */
  saveSelectedSheetsLocal(selections: Record<string, string>): void {
    try {
      cacheService.set(STORAGE_KEYS.SELECTED_SHEETS, selections);
    } catch (error) {
      console.warn("Failed to save selected sheets", error);
    }
  }

  /**
   * Get selected sheet names
   */
  getSelectedSheetsLocal(): Record<string, string> {
    try {
      return (
        cacheService.get<Record<string, string>>(
          STORAGE_KEYS.SELECTED_SHEETS
        ) || {}
      );
    } catch (error) {
      console.warn("Failed to get selected sheets", error);
      return {};
    }
  }

  /**
   * Clear all Google Sheets storage
   */
  clearAll(): void {
    cacheService.clearType("googleSheets");
  }

  /**
   * Verify cache validity
   */
  isCacheValid(type: "inventory" | "wallet" | "shipping" | "order"): boolean {
    const cache = this.getSheetsCacheLocal(type);
    return cache[type] !== undefined;
  }
}

export default new GoogleSheetsStorageService();

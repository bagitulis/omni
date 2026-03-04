/**
 * TikTok Ads Excel Parser
 * Parse Excel files from TikTok Ads export
 * Single Responsibility: Excel parsing only
 */

import * as XLSX from "xlsx";
import { getLogger } from "../../utils/logger";
import { CreativeDataWithPeriod } from "./tiktokAdsTypes";

const logger = getLogger("TiktokAdsExcelParser");

export class TiktokAdsExcelParser {
  /**
   * Parse Excel file buffer and extract creative data
   */
  parseExcelBuffer(
    buffer: Buffer,
    fileName: string
  ): { data: CreativeDataWithPeriod[]; period: { start: Date; end: Date } } {
    const workbook = XLSX.read(buffer, { type: "buffer" });
    const sheetName = workbook.SheetNames[0];
    const worksheet = workbook.Sheets[sheetName];

    // Convert to JSON with header mapping
    const rawData = XLSX.utils.sheet_to_json(worksheet) as Record<
      string,
      unknown
    >[];

    // Extract period from filename
    const period = this.extractPeriodFromFileName(fileName);

    // Map Excel columns to our schema
    const mappedData = rawData
      .map((row) => this.mapRowToCreativeData(row, period))
      .filter((item): item is CreativeDataWithPeriod => item !== null);

    logger.info(
      `Parsed ${mappedData.length} rows from ${fileName} (${rawData.length} raw)`
    );

    return { data: mappedData, period };
  }

  /**
   * Extract period start/end from filename
   * Format: "creative data for product campaigns YYYY-MM-DD HH ~ YYYY-MM-DD HH.xlsx"
   */
  extractPeriodFromFileName(fileName: string): { start: Date; end: Date } {
    // Default to current month if parsing fails
    const now = new Date();
    const defaultStart = new Date(now.getFullYear(), now.getMonth(), 1);
    const defaultEnd = new Date(now.getFullYear(), now.getMonth() + 1, 0);

    try {
      // Match pattern: YYYY-MM-DD HH ~ YYYY-MM-DD HH
      const match = fileName.match(
        /(\d{4}-\d{2}-\d{2})\s+\d+\s*~\s*(\d{4}-\d{2}-\d{2})\s+\d+/
      );

      if (match) {
        const startDate = new Date(match[1]);
        const endDate = new Date(match[2]);
        // Set end date to end of day
        endDate.setHours(23, 59, 59, 999);
        return { start: startDate, end: endDate };
      }
    } catch (error) {
      logger.warn(`Failed to parse period from filename: ${fileName}`, error);
    }

    return { start: defaultStart, end: defaultEnd };
  }

  /**
   * Map a single Excel row to CreativeData
   */
  private mapRowToCreativeData(
    row: Record<string, unknown>,
    period: { start: Date; end: Date }
  ): CreativeDataWithPeriod | null {
    try {
      // Required fields validation
      const campaignId = this.toString(row["ID Campaign"]);
      const productId = this.toString(row["ID produk"]);

      if (!campaignId || !productId) {
        return null; // Skip rows without required fields
      }

      // Skip invalid product IDs (like -1)
      if (productId === "-1" || productId === "") {
        return null;
      }

      const data: CreativeDataWithPeriod = {
        periodStart: period.start,
        periodEnd: period.end,
        campaignName: this.toString(row["Nama kampanye"]) || "",
        campaignId,
        productId,
        creativeType: this.toString(row["Jenis materi iklan"]) || "Unknown",
        videoTitle: this.toNullableString(row["Judul video"]),
        videoId: this.toNullableString(row["ID video"]),
        tiktokAccount: this.toNullableString(row["Akun TikTok"]),
        postingTime: this.toNullableDate(row["Waktu posting"]),
        status: this.toNullableString(row["Status"]),
        authorizationType: this.toNullableString(row["Jenis otorisasi"]),
        cost: this.toNumber(row["Biaya"]),
        ordersSku: this.toNumber(row["Pesanan SKU"]),
        costPerOrder: this.toNumber(row["Biaya per pesanan"]),
        grossRevenue: this.toNumber(row["Pendapatan kotor"]),
        roi: this.toNumber(row["ROI"]),
        impressions: this.toNumber(row["Impresi iklan produk"]),
        clicks: this.toNumber(row["Jumlah klik iklan produk"]),
        ctr: this.toNumber(row["Tingkat klik iklan produk"]),
        conversionRate: this.toNumber(row["Rasio konversi iklan"]),
        watchRate2s: this.toNullableNumber(
          row["Rasio tayang video iklan 2 detik"]
        ),
        watchRate6s: this.toNullableNumber(
          row["Rasio tayang video iklan 6 detik"]
        ),
        watchRate25pct: this.toNullableNumber(
          row["Rasio tayang video iklan 25%"]
        ),
        watchRate50pct: this.toNullableNumber(
          row["Rasio tayang video iklan 50%"]
        ),
        watchRate75pct: this.toNullableNumber(
          row["Rasio tayang video iklan 75%"]
        ),
        watchRate100pct: this.toNullableNumber(
          row["Rasio tayang video iklan 100%"]
        ),
        currency: this.toString(row["Mata uang"]) || "IDR",
      };

      return data;
    } catch (error) {
      logger.warn("Failed to parse row", error);
      return null;
    }
  }

  private toString(value: unknown): string {
    if (value === null || value === undefined || value === "-") {
      return "";
    }
    return String(value).trim();
  }

  private toNullableString(value: unknown): string | null {
    if (value === null || value === undefined || value === "-") {
      return null;
    }
    const str = String(value).trim();
    return str === "" ? null : str;
  }

  private toNumber(value: unknown): number {
    if (value === null || value === undefined || value === "-") {
      return 0;
    }
    const num = Number(value);
    return isNaN(num) ? 0 : num;
  }

  private toNullableNumber(value: unknown): number | null {
    if (value === null || value === undefined || value === "-") {
      return null;
    }
    const num = Number(value);
    return isNaN(num) ? null : num;
  }

  private toNullableDate(value: unknown): Date | null {
    if (value === null || value === undefined || value === "-") {
      return null;
    }
    try {
      const date = new Date(String(value));
      return isNaN(date.getTime()) ? null : date;
    } catch {
      return null;
    }
  }
}

// Singleton instance
let parserInstance: TiktokAdsExcelParser | null = null;

export function getTiktokAdsExcelParser(): TiktokAdsExcelParser {
  if (!parserInstance) {
    parserInstance = new TiktokAdsExcelParser();
  }
  return parserInstance;
}

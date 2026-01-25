/**
 * Shopee Ads CSV Parser
 * Parse Shopee Ads CSV files
 * Single Responsibility: CSV parsing only
 */

import { ShopeeAdsProductRow } from "./shopeeAdsTypes";

interface ParseResult {
  data: ShopeeAdsProductRow[];
  period: { start: string; end: string; label: string };
}

// Column mapping: CSV header → internal field
const COLUMN_MAP: Record<string, keyof ShopeeAdsProductRow> = {
  "Kode Produk": "productId",
  "Nama Iklan": "productName",
  Status: "status",
  "Mode Bidding": "biddingMode",
  "Penempatan Iklan": "placement",
  "Tanggal Mulai": "startDate",
  "Tanggal Selesai": "endDate",
  Dilihat: "impressions",
  "Jumlah Klik": "clicks",
  "Persentase Klik": "ctr",
  Konversi: "conversions",
  "Konversi Langsung": "directConversions",
  "Tingkat konversi": "conversionRate",
  "Tingkat Konversi Langsung": "directConversionRate",
  "Biaya per Konversi": "costPerConversion",
  "Biaya per Konversi Langsung": "costPerDirectConversion",
  "Produk Terjual": "unitsSold",
  "Terjual Langsung": "directUnitsSold",
  "Omzet Penjualan": "revenue",
  "Penjualan Langsung": "directRevenue",
  Biaya: "cost",
  "Efektifitas Iklan": "roas",
  "Efektivitas Langsung": "directRoas",
  ACOS: "acos",
  "ACOS Langsung": "directAcos",
};

class ShopeeAdsCsvParser {
  /**
   * Extract period from filename
   * Format: Data-+Semua-Iklan-Produk-DD_MM_YYYY-DD_MM_YYYY.csv
   */
  extractPeriodFromFilename(filename: string): {
    start: string;
    end: string;
    label: string;
  } | null {
    const pattern = /(\d{2})_(\d{2})_(\d{4})-(\d{2})_(\d{2})_(\d{4})/;
    const match = filename.match(pattern);

    if (!match) return null;

    const [, day1, month1, year1, day2, month2, year2] = match;
    const startDate = new Date(
      parseInt(year1),
      parseInt(month1) - 1,
      parseInt(day1),
    );
    const endDate = new Date(
      parseInt(year2),
      parseInt(month2) - 1,
      parseInt(day2),
    );

    // ISO week label
    const weekNum = this.getISOWeek(startDate);
    const label = `${year1}-W${weekNum.toString().padStart(2, "0")}`;

    return {
      start: startDate.toISOString().split("T")[0],
      end: endDate.toISOString().split("T")[0],
      label,
    };
  }

  /**
   * Parse CSV buffer to product data
   */
  parseCsvBuffer(buffer: Buffer, filename: string): ParseResult {
    const period = this.extractPeriodFromFilename(filename);
    if (!period) {
      throw new Error(`Cannot extract period from filename: ${filename}`);
    }

    // Decode buffer (handle BOM)
    let content = buffer.toString("utf-8");
    if (content.charCodeAt(0) === 0xfeff) {
      content = content.slice(1);
    }

    const lines = content.split("\n").filter((line) => line.trim());

    // Skip first 5 rows (Shopee header rows)
    const dataLines = lines.slice(5);
    if (dataLines.length === 0) {
      return { data: [], period };
    }

    // Parse header
    const headerLine = dataLines[0];
    const headers = this.parseCSVLine(headerLine);

    // Map headers to column indices
    const columnIndices: Record<string, number> = {};
    headers.forEach((header, index) => {
      const cleanHeader = header.trim();
      if (COLUMN_MAP[cleanHeader]) {
        columnIndices[COLUMN_MAP[cleanHeader]] = index;
      }
    });

    // Parse data rows
    const data: ShopeeAdsProductRow[] = [];
    for (let i = 1; i < dataLines.length; i++) {
      const line = dataLines[i].trim();
      if (!line) continue;

      const values = this.parseCSVLine(line);
      const row = this.parseRow(values, columnIndices, period);
      if (row && row.productId) {
        data.push(row);
      }
    }

    return { data, period };
  }

  private parseRow(
    values: string[],
    indices: Record<string, number>,
    period: { start: string; end: string; label: string },
  ): ShopeeAdsProductRow | null {
    const getValue = (field: string): string => {
      const idx = indices[field];
      return idx !== undefined ? values[idx]?.trim() || "" : "";
    };

    const productId = getValue("productId");
    if (!productId) return null;

    return {
      productId,
      productName: getValue("productName") || productId,
      status: getValue("status") || null,
      biddingMode: getValue("biddingMode") || null,
      placement: getValue("placement") || null,
      startDate: this.parseDate(getValue("startDate")),
      endDate: getValue("endDate") || null,
      impressions: this.parseNumber(getValue("impressions")),
      clicks: this.parseNumber(getValue("clicks")),
      ctr: this.parsePercentage(getValue("ctr")),
      conversions: this.parseNumber(getValue("conversions")),
      directConversions: this.parseNumber(getValue("directConversions")),
      conversionRate: this.parsePercentage(getValue("conversionRate")),
      directConversionRate: this.parsePercentage(
        getValue("directConversionRate"),
      ),
      costPerConversion: this.parseCurrency(getValue("costPerConversion")),
      costPerDirectConversion: this.parseCurrency(
        getValue("costPerDirectConversion"),
      ),
      unitsSold: this.parseNumber(getValue("unitsSold")),
      directUnitsSold: this.parseNumber(getValue("directUnitsSold")),
      revenue: this.parseCurrency(getValue("revenue")),
      directRevenue: this.parseCurrency(getValue("directRevenue")),
      cost: this.parseCurrency(getValue("cost")),
      roas: this.parseFloat(getValue("roas")),
      directRoas: this.parseFloat(getValue("directRoas")),
      acos: this.parsePercentage(getValue("acos")),
      directAcos: this.parsePercentage(getValue("directAcos")),
      periodStart: new Date(period.start),
      periodEnd: new Date(period.end),
      periodLabel: period.label,
    };
  }

  private parseCSVLine(line: string): string[] {
    const result: string[] = [];
    let current = "";
    let inQuotes = false;

    for (const char of line) {
      if (char === '"') {
        inQuotes = !inQuotes;
      } else if (char === "," && !inQuotes) {
        result.push(current);
        current = "";
      } else {
        current += char;
      }
    }
    result.push(current);
    return result;
  }

  private parseNumber(value: string): number {
    if (!value) return 0;
    const cleaned = value.replace(/[^\d.-]/g, "");
    return parseInt(cleaned) || 0;
  }

  private parseFloat(value: string): number {
    if (!value) return 0;
    const cleaned = value.replace(/[^\d.-]/g, "");
    return parseFloat(cleaned) || 0;
  }

  private parsePercentage(value: string): number {
    if (!value) return 0;
    const cleaned = value.replace(/[%,]/g, "").trim();
    return parseFloat(cleaned) / 100 || 0;
  }

  private parseCurrency(value: string): number {
    if (!value) return 0;
    const cleaned = value.replace(/[Rp.\s,]/g, "").trim();
    return parseFloat(cleaned) || 0;
  }

  private parseDate(value: string): Date | null {
    if (!value || value === "-") return null;
    // Format: DD/MM/YYYY
    const match = value.match(/(\d{2})\/(\d{2})\/(\d{4})/);
    if (!match) return null;
    const [, day, month, year] = match;
    return new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
  }

  private getISOWeek(date: Date): number {
    const d = new Date(date);
    d.setHours(0, 0, 0, 0);
    d.setDate(d.getDate() + 3 - ((d.getDay() + 6) % 7));
    const week1 = new Date(d.getFullYear(), 0, 4);
    return (
      1 +
      Math.round(
        ((d.getTime() - week1.getTime()) / 86400000 -
          3 +
          ((week1.getDay() + 6) % 7)) /
          7,
      )
    );
  }
}

// Singleton instance
let parserInstance: ShopeeAdsCsvParser | null = null;

export function getShopeeAdsCsvParser(): ShopeeAdsCsvParser {
  if (!parserInstance) {
    parserInstance = new ShopeeAdsCsvParser();
  }
  return parserInstance;
}

export { ShopeeAdsCsvParser };

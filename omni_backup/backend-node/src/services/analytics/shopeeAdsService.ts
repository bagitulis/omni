/**
 * Shopee Ads Service
 * Business logic for Shopee Ads analytics
 * Single Responsibility: Orchestrate upload, aggregation, and queries
 */

import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";
import { ShopeeAdsRepository } from "./shopeeAdsRepository";
import { getShopeeAdsCsvParser } from "./shopeeAdsCsvParser";
import {
  UploadBatchResult,
  UploadOptions,
  ProductDataFilter,
  DashboardSummary,
} from "./shopeeAdsTypes";

const logger = getLogger("ShopeeAdsService");

type PrismaClientAny = PrismaClient | ReturnType<PrismaClient["$extends"]>;

export class ShopeeAdsService {
  private repository: ShopeeAdsRepository;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    if (!tenantId) {
      throw new Error("Missing tenantId - authentication required");
    }
    this.repository = new ShopeeAdsRepository(prisma, tenantId);
  }

  // ============================================
  // Upload & Import
  // ============================================

  async uploadCsvFile(
    buffer: Buffer,
    fileName: string,
    options: UploadOptions,
  ): Promise<UploadBatchResult> {
    const parser = getShopeeAdsCsvParser();
    const errors: string[] = [];

    try {
      // Parse CSV file
      const { data, period } = parser.parseCsvBuffer(buffer, fileName);

      // Use provided periodLabel or fall back to parsed one
      const finalPeriodLabel = options.periodLabel || period.label;

      if (data.length === 0) {
        return {
          success: false,
          batchId: "",
          fileName,
          periodStart: period.start,
          periodEnd: period.end,
          periodLabel: finalPeriodLabel,
          totalRows: 0,
          insertedRows: 0,
          skippedRows: 0,
          updatedRows: 0,
          errors: ["No valid data found in CSV file"],
        };
      }

      // Create upload batch record
      const batch = await this.repository.createUploadBatch({
        fileName,
        periodStart: period.start,
        periodEnd: period.end,
        periodLabel: finalPeriodLabel,
        totalRows: data.length,
        uploadedBy: options.uploadedBy,
      });

      // Process each row
      let insertedRows = 0;
      let skippedRows = 0;
      let updatedRows = 0;

      for (const row of data) {
        try {
          const result = await this.repository.upsertProductData(
            row,
            batch.id,
            options.mode,
          );

          if (result.action === "inserted") insertedRows++;
          else if (result.action === "skipped") skippedRows++;
          else if (result.action === "updated") updatedRows++;
        } catch (error) {
          const msg = error instanceof Error ? error.message : String(error);
          errors.push(`Row error: ${msg}`);
          logger.warn(`Failed to process row`, { error: msg });
        }
      }

      // Update batch status
      await this.repository.updateUploadBatch(batch.id, {
        insertedRows,
        skippedRows,
        updatedRows,
        status: errors.length > 0 ? "completed_with_errors" : "completed",
        errorMessage:
          errors.length > 0 ? errors.slice(0, 5).join("; ") : undefined,
      });

      logger.info(`Upload completed`, {
        batchId: batch.id,
        insertedRows,
        skippedRows,
        updatedRows,
      });

      return {
        success: true,
        batchId: batch.id,
        fileName,
        periodStart: period.start,
        periodEnd: period.end,
        periodLabel: period.label,
        totalRows: data.length,
        insertedRows,
        skippedRows,
        updatedRows,
        errors,
      };
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error(`Upload failed`, { error: msg });
      return {
        success: false,
        batchId: "",
        fileName,
        periodStart: "",
        periodEnd: "",
        periodLabel: "",
        totalRows: 0,
        insertedRows: 0,
        skippedRows: 0,
        updatedRows: 0,
        errors: [msg],
      };
    }
  }

  // ============================================
  // Query Operations
  // ============================================

  async getDashboard(
    periodStart?: string,
    periodEnd?: string,
  ): Promise<DashboardSummary> {
    return this.repository.getDashboardSummary(periodStart, periodEnd);
  }

  async getProductData(filter: ProductDataFilter) {
    return this.repository.getProductData(filter);
  }

  async getUploadHistory(limit = 20) {
    return this.repository.getUploadHistory(limit);
  }
}

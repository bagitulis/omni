/**
 * TikTok Ads Service
 * Single Responsibility: Orchestrate upload and delegate to specialized services
 */

import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";
import { TiktokAdsRepository } from "./tiktokAdsRepository";
import { TiktokAdsDashboardService } from "./tiktokAdsDashboardService";
import { TiktokAdsSummaryService } from "./tiktokAdsSummaryService";
import { getTiktokAdsExcelParser } from "./tiktokAdsExcelParser";
import {
  UploadBatchResult,
  UploadOptions,
  CreativeDataFilter,
  DashboardSummary,
  CreativeDataWithPeriod,
} from "./tiktokAdsTypes";

const logger = getLogger("TiktokAdsService");

type PrismaClientAny = PrismaClient | ReturnType<PrismaClient["$extends"]>;

export class TiktokAdsService {
  private repository: TiktokAdsRepository;
  private dashboardService: TiktokAdsDashboardService;
  private summaryService: TiktokAdsSummaryService;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    if (!tenantId) {
      throw new Error("Missing tenantId - authentication required");
    }
    this.tenantId = tenantId;
    this.repository = new TiktokAdsRepository(prisma, tenantId);
    this.dashboardService = new TiktokAdsDashboardService(prisma, tenantId);
    this.summaryService = new TiktokAdsSummaryService(prisma, tenantId);
  }

  // ============================================
  // Upload & Import
  // ============================================

  async uploadExcelFile(
    buffer: Buffer,
    fileName: string,
    options: UploadOptions,
  ): Promise<UploadBatchResult> {
    const parser = getTiktokAdsExcelParser();
    const errors: string[] = [];

    try {
      const { data, period } = parser.parseExcelBuffer(buffer, fileName);

      if (data.length === 0) {
        return this.createEmptyResult(fileName, period, [
          "No valid data found in Excel file",
        ]);
      }

      const batch = await this.repository.createUploadBatch({
        fileName,
        periodStart: period.start,
        periodEnd: period.end,
        totalRows: data.length,
        uploadedBy: options.uploadedBy,
      });

      const stats = await this.processRows(
        data,
        batch.id,
        options.mode,
        errors,
      );

      await this.repository.updateUploadBatch(batch.id, {
        ...stats,
        status: errors.length > 0 ? "completed_with_errors" : "completed",
        errorMessage: errors.length > 0 ? errors.join("; ") : undefined,
      });

      // Refresh summary in background (non-blocking)
      this.refreshProductSummary().catch((err) => {
        logger.error("Failed to refresh product summary after upload", {
          error: err instanceof Error ? err.message : String(err),
        });
      });

      logger.info(`Upload completed`, { batchId: batch.id, ...stats });

      return {
        success: true,
        batchId: batch.id,
        fileName,
        periodStart: period.start,
        periodEnd: period.end,
        totalRows: data.length,
        ...stats,
        errors,
      };
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error(`Upload failed`, { error: msg });
      return this.createEmptyResult(
        fileName,
        { start: new Date(), end: new Date() },
        [msg],
      );
    }
  }

  /**
   * Process rows and return insert/update/skip counts
   */
  private async processRows(
    data: CreativeDataWithPeriod[],
    batchId: string,
    mode: "skip" | "update",
    errors: string[],
  ): Promise<{
    insertedRows: number;
    skippedRows: number;
    updatedRows: number;
  }> {
    let insertedRows = 0;
    let skippedRows = 0;
    let updatedRows = 0;

    for (const row of data) {
      try {
        const result = await this.repository.upsertCreativeData(
          row,
          batchId,
          mode,
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

    return { insertedRows, skippedRows, updatedRows };
  }

  /**
   * Create empty upload result for error cases
   */
  private createEmptyResult(
    fileName: string,
    period: { start: Date; end: Date },
    errors: string[],
  ): UploadBatchResult {
    return {
      success: false,
      batchId: "",
      fileName,
      periodStart: period.start,
      periodEnd: period.end,
      totalRows: 0,
      insertedRows: 0,
      skippedRows: 0,
      updatedRows: 0,
      errors,
    };
  }

  // ============================================
  // Dashboard (delegated to DashboardService)
  // ============================================

  async getDashboardSummary(
    periodStart?: Date,
    periodEnd?: Date,
  ): Promise<DashboardSummary> {
    return this.dashboardService.getDashboardSummary(periodStart, periodEnd);
  }

  // ============================================
  // Query Methods (delegated to Repository)
  // ============================================

  async getCreativeData(filter: CreativeDataFilter) {
    return await this.repository.getCreativeData(filter);
  }

  async getUploadHistory(limit = 20) {
    return await this.repository.getUploadBatches(limit);
  }

  async getDataCount(periodStart?: Date, periodEnd?: Date): Promise<number> {
    return await this.repository.countCreativeData({
      tenantId: this.tenantId,
      periodStart,
      periodEnd,
    });
  }

  // ============================================
  // Summary Maintenance (delegated to SummaryService)
  // ============================================

  async refreshProductSummary(): Promise<void> {
    return this.summaryService.refreshProductSummary();
  }
}

// ============================================
// Factory Function
// ============================================

export function createTiktokAdsService(
  prisma: PrismaClientAny,
  tenantId: string,
): TiktokAdsService {
  if (!tenantId) {
    throw new Error("Missing tenantId - authentication required");
  }
  return new TiktokAdsService(prisma, tenantId);
}

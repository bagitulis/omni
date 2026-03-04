/**
 * TikTok Ads Repository
 * Database operations for TikTok Ads data
 * Single Responsibility: Prisma CRUD operations only
 */

import { PrismaClient } from "@prisma/client";
import {
  CreativeDataWithPeriod,
  CreativeDataFilter,
  ProductSummary,
} from "./tiktokAdsTypes";

type PrismaClientAny = PrismaClient | ReturnType<PrismaClient["$extends"]>;

export class TiktokAdsRepository {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  // ============================================
  // Upload Batch Operations
  // ============================================

  async createUploadBatch(data: {
    fileName: string;
    periodStart: Date;
    periodEnd: Date;
    totalRows: number;
    uploadedBy?: string;
  }) {
    return await (this.prisma as PrismaClient).tiktokAdsUploadBatch.create({
      data: {
        tenantId: this.tenantId,
        fileName: data.fileName,
        periodStart: data.periodStart,
        periodEnd: data.periodEnd,
        totalRows: data.totalRows,
        uploadedBy: data.uploadedBy,
        status: "processing",
      },
    });
  }

  async updateUploadBatch(
    batchId: string,
    data: {
      insertedRows?: number;
      skippedRows?: number;
      updatedRows?: number;
      status?: string;
      errorMessage?: string;
    }
  ) {
    return await (this.prisma as PrismaClient).tiktokAdsUploadBatch.update({
      where: { id: batchId },
      data,
    });
  }

  async getUploadBatches(limit = 20) {
    return await (this.prisma as PrismaClient).tiktokAdsUploadBatch.findMany({
      where: { tenantId: this.tenantId },
      orderBy: { createdAt: "desc" },
      take: limit,
    });
  }

  // ============================================
  // Creative Data Operations
  // ============================================

  async upsertCreativeData(
    data: CreativeDataWithPeriod,
    uploadBatchId: string,
    mode: "skip" | "update"
  ): Promise<{ action: "inserted" | "updated" | "skipped" }> {
    const prismaClient = this.prisma as PrismaClient;

    // Check for existing record
    const existing = await prismaClient.tiktokAdsCreativeData.findFirst({
      where: {
        tenantId: this.tenantId,
        campaignId: data.campaignId,
        productId: data.productId,
        videoId: data.videoId || "",
        periodStart: data.periodStart,
        periodEnd: data.periodEnd,
      },
    });

    if (existing) {
      if (mode === "skip") {
        return { action: "skipped" };
      }

      // Update existing record
      await prismaClient.tiktokAdsCreativeData.update({
        where: { id: existing.id },
        data: this.mapToDbData(data, uploadBatchId),
      });
      return { action: "updated" };
    }

    // Insert new record
    await prismaClient.tiktokAdsCreativeData.create({
      data: {
        tenantId: this.tenantId,
        ...this.mapToDbData(data, uploadBatchId),
      },
    });
    return { action: "inserted" };
  }

  private mapToDbData(data: CreativeDataWithPeriod, uploadBatchId: string) {
    return {
      uploadBatchId,
      periodStart: data.periodStart,
      periodEnd: data.periodEnd,
      campaignId: data.campaignId,
      campaignName: data.campaignName,
      productId: data.productId,
      creativeType: data.creativeType,
      videoTitle: data.videoTitle,
      videoId: data.videoId || "",
      tiktokAccount: data.tiktokAccount,
      postingTime: data.postingTime,
      status: data.status,
      authorizationType: data.authorizationType,
      cost: data.cost,
      ordersSku: data.ordersSku,
      costPerOrder: data.costPerOrder,
      grossRevenue: data.grossRevenue,
      roi: data.roi,
      impressions: data.impressions,
      clicks: data.clicks,
      ctr: data.ctr,
      conversionRate: data.conversionRate,
      watchRate2s: data.watchRate2s,
      watchRate6s: data.watchRate6s,
      watchRate25pct: data.watchRate25pct,
      watchRate50pct: data.watchRate50pct,
      watchRate75pct: data.watchRate75pct,
      watchRate100pct: data.watchRate100pct,
      currency: data.currency,
    };
  }

  async getCreativeData(filter: CreativeDataFilter) {
    const prismaClient = this.prisma as PrismaClient;

    const where: Record<string, unknown> = {
      tenantId: filter.tenantId,
    };

    if (filter.periodStart) {
      where.periodStart = { gte: filter.periodStart };
    }
    if (filter.periodEnd) {
      where.periodEnd = { lte: filter.periodEnd };
    }
    if (filter.campaignId) {
      where.campaignId = filter.campaignId;
    }
    if (filter.productId) {
      where.productId = filter.productId;
    }
    if (filter.creativeType) {
      where.creativeType = filter.creativeType;
    }
    if (filter.minRoi !== undefined) {
      where.roi = { ...(where.roi as object), gte: filter.minRoi };
    }
    if (filter.maxRoi !== undefined) {
      where.roi = { ...(where.roi as object), lte: filter.maxRoi };
    }

    const orderBy: Record<string, string> = {};
    if (filter.orderBy) {
      const fieldMap: Record<string, string> = {
        cost: "cost",
        revenue: "grossRevenue",
        roi: "roi",
        orders: "ordersSku",
      };
      orderBy[fieldMap[filter.orderBy]] = filter.orderDir || "desc";
    }

    return await prismaClient.tiktokAdsCreativeData.findMany({
      where,
      orderBy:
        Object.keys(orderBy).length > 0 ? orderBy : { createdAt: "desc" },
      take: filter.limit || 100,
      skip: filter.offset || 0,
    });
  }

  async countCreativeData(filter: CreativeDataFilter): Promise<number> {
    const prismaClient = this.prisma as PrismaClient;

    const where: Record<string, unknown> = {
      tenantId: filter.tenantId,
    };

    if (filter.periodStart) {
      where.periodStart = { gte: filter.periodStart };
    }
    if (filter.periodEnd) {
      where.periodEnd = { lte: filter.periodEnd };
    }

    return await prismaClient.tiktokAdsCreativeData.count({ where });
  }

  // ============================================
  // Product Summary Operations
  // ============================================

  async upsertProductSummary(summary: ProductSummary) {
    const prismaClient = this.prisma as PrismaClient;

    return await prismaClient.tiktokAdsProductSummary.upsert({
      where: {
        tenantId_productId_periodType_periodDate: {
          tenantId: this.tenantId,
          productId: summary.productId,
          periodType: summary.periodType,
          periodDate: summary.periodDate,
        },
      },
      create: {
        tenantId: this.tenantId,
        ...summary,
      },
      update: summary,
    });
  }

  async getProductSummaries(periodType: string, periodDate?: Date) {
    const prismaClient = this.prisma as PrismaClient;

    const where: Record<string, unknown> = {
      tenantId: this.tenantId,
      periodType,
    };

    if (periodDate) {
      where.periodDate = periodDate;
    }

    return await prismaClient.tiktokAdsProductSummary.findMany({
      where,
      orderBy: { totalRevenue: "desc" },
    });
  }
}

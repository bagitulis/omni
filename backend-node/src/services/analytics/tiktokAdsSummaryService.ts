/**
 * TikTok Ads Summary Service
 * Single Responsibility: Maintain pre-aggregated summary table
 */

import { PrismaClient } from "@prisma/client";
import { getLogger } from "../../utils/logger";

const logger = getLogger("TiktokAdsSummaryService");

type PrismaClientAny = PrismaClient | ReturnType<PrismaClient["$extends"]>;

interface ProductAggregation {
  productId: string;
  totalCost: number;
  totalRevenue: number;
  totalOrders: number;
  totalImpressions: number;
  totalClicks: number;
}

interface CreativeTypeRow {
  productId: string;
  creativeType: string;
  revenue: number;
}

export class TiktokAdsSummaryService {
  private prisma: PrismaClientAny;
  private tenantId: string;

  constructor(prisma: PrismaClientAny, tenantId: string) {
    if (!tenantId) {
      throw new Error("Missing tenantId - authentication required");
    }
    this.prisma = prisma;
    this.tenantId = tenantId;
  }

  /**
   * Refresh the TiktokAdsProductSummary table
   * Pre-aggregates data for fast dashboard queries
   */
  async refreshProductSummary(): Promise<void> {
    const prismaClient = this.prisma as PrismaClient;
    logger.info(`Refreshing product summary for tenant: ${this.tenantId}`);
    const startTime = Date.now();

    try {
      const periodDate = await this.getMinPeriodDate(prismaClient);
      if (!periodDate) {
        logger.info(`No creative data found for tenant: ${this.tenantId}`);
        return;
      }

      const aggregations = await this.aggregateByProduct(prismaClient);
      const bestCreativeMap = await this.getBestCreativeTypes(prismaClient);

      await this.clearAndInsertSummaries(
        prismaClient,
        aggregations,
        bestCreativeMap,
        periodDate,
      );

      const duration = (Date.now() - startTime) / 1000;
      logger.info(`Product summary refreshed in ${duration.toFixed(2)}s`, {
        tenantId: this.tenantId,
        products: aggregations.length,
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      logger.error(`Failed to refresh product summary`, { error: msg });
      throw error;
    }
  }

  /**
   * Get minimum period date from creative data
   */
  private async getMinPeriodDate(prisma: PrismaClient): Promise<Date | null> {
    const dateRange = await prisma.tiktokAdsCreativeData.aggregate({
      where: { tenantId: this.tenantId },
      _min: { periodStart: true },
    });

    const minDate = dateRange._min.periodStart;
    if (!minDate) return null;

    const periodDate = new Date(minDate);
    periodDate.setHours(0, 0, 0, 0);
    return periodDate;
  }

  /**
   * Aggregate creative data by productId using raw SQL for performance
   */
  private async aggregateByProduct(
    prisma: PrismaClient,
  ): Promise<ProductAggregation[]> {
    return await prisma.$queryRaw<ProductAggregation[]>`
      SELECT 
        productId,
        SUM(cost) as totalCost,
        SUM(grossRevenue) as totalRevenue,
        SUM(ordersSku) as totalOrders,
        SUM(impressions) as totalImpressions,
        SUM(clicks) as totalClicks
      FROM TiktokAdsCreativeData
      WHERE tenantId = ${this.tenantId}
      GROUP BY productId
    `;
  }

  /**
   * Get best creative type for each product (by revenue)
   */
  private async getBestCreativeTypes(
    prisma: PrismaClient,
  ): Promise<Map<string, string>> {
    const creativeTypes = await prisma.$queryRaw<CreativeTypeRow[]>`
      SELECT productId, creativeType, SUM(grossRevenue) as revenue
      FROM TiktokAdsCreativeData
      WHERE tenantId = ${this.tenantId}
      GROUP BY productId, creativeType
      ORDER BY productId, revenue DESC
    `;

    const map = new Map<string, string>();
    for (const row of creativeTypes) {
      if (!map.has(row.productId)) {
        map.set(row.productId, row.creativeType);
      }
    }
    return map;
  }

  /**
   * Clear existing summaries and insert new ones in batches
   */
  private async clearAndInsertSummaries(
    prisma: PrismaClient,
    aggregations: ProductAggregation[],
    bestCreativeMap: Map<string, string>,
    periodDate: Date,
  ): Promise<void> {
    await prisma.tiktokAdsProductSummary.deleteMany({
      where: { tenantId: this.tenantId },
    });

    const batchSize = 100;
    for (let i = 0; i < aggregations.length; i += batchSize) {
      const batch = aggregations.slice(i, i + batchSize);
      await prisma.tiktokAdsProductSummary.createMany({
        data: batch.map((agg) =>
          this.mapToSummary(agg, bestCreativeMap, periodDate),
        ),
      });
    }
  }

  /**
   * Map aggregation to summary record
   */
  private mapToSummary(
    agg: ProductAggregation,
    bestCreativeMap: Map<string, string>,
    periodDate: Date,
  ) {
    const cost = Number(agg.totalCost) || 0;
    const revenue = Number(agg.totalRevenue) || 0;
    const orders = Number(agg.totalOrders) || 0;
    const impressions = Number(agg.totalImpressions) || 0;
    const clicks = Number(agg.totalClicks) || 0;

    return {
      tenantId: this.tenantId,
      productId: agg.productId,
      periodType: "all",
      periodDate,
      totalCost: cost,
      totalRevenue: revenue,
      totalOrders: orders,
      totalImpressions: impressions,
      totalClicks: clicks,
      avgRoi: cost > 0 ? revenue / cost : 0,
      avgCtr: impressions > 0 ? clicks / impressions : 0,
      avgConversionRate: clicks > 0 ? orders / clicks : 0,
      bestCreativeType: bestCreativeMap.get(agg.productId) || null,
    };
  }
}

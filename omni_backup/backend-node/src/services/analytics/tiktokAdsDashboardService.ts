/**
 * TikTok Ads Dashboard Service
 * Single Responsibility: Dashboard queries and aggregations
 */

import { PrismaClient } from "@prisma/client";
import { DashboardSummary, CreativeTypeStats } from "./tiktokAdsTypes";

type PrismaClientAny = PrismaClient | ReturnType<PrismaClient["$extends"]>;

export class TiktokAdsDashboardService {
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
   * Get dashboard summary with auto-detection of data source
   * Uses pre-aggregated summary table if available, falls back to raw data
   */
  async getDashboardSummary(
    periodStart?: Date,
    periodEnd?: Date,
  ): Promise<DashboardSummary> {
    const prismaClient = this.prisma as PrismaClient;

    const summaryWhere = {
      tenantId: this.tenantId,
      periodType: "all",
    };

    const summaryCount = await prismaClient.tiktokAdsProductSummary.count({
      where: summaryWhere,
    });

    if (summaryCount > 0) {
      return await this.getFromSummaryTable(prismaClient, summaryWhere);
    }

    return await this.getFromRawData(prismaClient, periodStart, periodEnd);
  }

  /**
   * FAST PATH: Query from pre-aggregated TiktokAdsProductSummary table
   */
  private async getFromSummaryTable(
    prisma: PrismaClient,
    where: Record<string, unknown>,
  ): Promise<DashboardSummary> {
    const totals = await prisma.tiktokAdsProductSummary.aggregate({
      where,
      _sum: {
        totalCost: true,
        totalRevenue: true,
        totalOrders: true,
        totalImpressions: true,
        totalClicks: true,
      },
    });

    const metrics = this.extractMetrics(totals._sum);

    const topProductSummaries = await prisma.tiktokAdsProductSummary.findMany({
      where,
      orderBy: { totalRevenue: "desc" },
      take: 10,
    });

    const topProducts = topProductSummaries.map((p) => ({
      productId: p.productId,
      cost: p.totalCost,
      revenue: p.totalRevenue,
      orders: p.totalOrders,
      roi: p.avgRoi,
    }));

    const creativeTypeComparison = await this.getCreativeTypeStats(prisma);

    return this.buildDashboardResponse(
      metrics,
      topProducts,
      creativeTypeComparison,
    );
  }

  /**
   * SLOW PATH: Query from raw TiktokAdsCreativeData table
   */
  private async getFromRawData(
    prisma: PrismaClient,
    periodStart?: Date,
    periodEnd?: Date,
  ): Promise<DashboardSummary> {
    const where: Record<string, unknown> = { tenantId: this.tenantId };
    if (periodStart) where.periodStart = { gte: periodStart };
    if (periodEnd) where.periodEnd = { lte: periodEnd };

    const totals = await prisma.tiktokAdsCreativeData.aggregate({
      where,
      _sum: {
        cost: true,
        grossRevenue: true,
        ordersSku: true,
        impressions: true,
        clicks: true,
      },
    });

    const metrics = {
      totalCost: totals._sum.cost || 0,
      totalRevenue: totals._sum.grossRevenue || 0,
      totalOrders: totals._sum.ordersSku || 0,
      totalImpressions: totals._sum.impressions || 0,
      totalClicks: totals._sum.clicks || 0,
    };

    const productGroups = await prisma.tiktokAdsCreativeData.groupBy({
      by: ["productId"],
      where,
      _sum: { cost: true, grossRevenue: true, ordersSku: true },
      orderBy: { _sum: { grossRevenue: "desc" } },
      take: 10,
    });

    const topProducts = productGroups.map((p) => ({
      productId: p.productId,
      cost: p._sum.cost || 0,
      revenue: p._sum.grossRevenue || 0,
      orders: p._sum.ordersSku || 0,
      roi:
        (p._sum.cost || 0) > 0
          ? (p._sum.grossRevenue || 0) / (p._sum.cost || 0)
          : 0,
    }));

    const creativeTypeComparison = await this.getCreativeTypeStats(prisma);

    return this.buildDashboardResponse(
      metrics,
      topProducts,
      creativeTypeComparison,
    );
  }

  /**
   * Get creative type statistics (shared by both paths)
   */
  private async getCreativeTypeStats(
    prisma: PrismaClient,
  ): Promise<CreativeTypeStats[]> {
    const groups = await prisma.tiktokAdsCreativeData.groupBy({
      by: ["creativeType"],
      where: { tenantId: this.tenantId },
      _sum: { cost: true, grossRevenue: true, ordersSku: true },
    });

    return groups
      .map((c) => {
        const cost = c._sum.cost || 0;
        const revenue = c._sum.grossRevenue || 0;
        const orders = c._sum.ordersSku || 0;
        return {
          creativeType: c.creativeType,
          cost,
          revenue,
          orders,
          roi: cost > 0 ? revenue / cost : 0,
          costPerOrder: orders > 0 ? cost / orders : 0,
        };
      })
      .sort((a, b) => b.revenue - a.revenue);
  }

  /**
   * Extract metrics from aggregation result
   */
  private extractMetrics(sum: Record<string, number | null>) {
    return {
      totalCost: sum.totalCost || 0,
      totalRevenue: sum.totalRevenue || 0,
      totalOrders: sum.totalOrders || 0,
      totalImpressions: sum.totalImpressions || 0,
      totalClicks: sum.totalClicks || 0,
    };
  }

  /**
   * Build final dashboard response with computed ratios
   */
  private buildDashboardResponse(
    metrics: ReturnType<typeof this.extractMetrics>,
    topProducts: Array<{
      productId: string;
      cost: number;
      revenue: number;
      orders: number;
      roi: number;
    }>,
    creativeTypeComparison: CreativeTypeStats[],
  ): DashboardSummary {
    return {
      totalCost: metrics.totalCost,
      totalRevenue: metrics.totalRevenue,
      totalOrders: metrics.totalOrders,
      avgRoi:
        metrics.totalCost > 0 ? metrics.totalRevenue / metrics.totalCost : 0,
      totalImpressions: metrics.totalImpressions,
      totalClicks: metrics.totalClicks,
      avgCtr:
        metrics.totalImpressions > 0
          ? metrics.totalClicks / metrics.totalImpressions
          : 0,
      avgConversionRate:
        metrics.totalClicks > 0 ? metrics.totalOrders / metrics.totalClicks : 0,
      topProducts,
      creativeTypeComparison,
      periodTrend: [],
    };
  }
}

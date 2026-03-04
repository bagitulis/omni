/**
 * Shopee Ads Repository
 * Database access for Shopee Ads data
 * Single Responsibility: Database operations only
 */

import {
  ShopeeAdsProductRow,
  ProductDataFilter,
  DashboardSummary,
  ProductPerformance,
  BiddingModeStats,
} from "./shopeeAdsTypes";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PrismaClientAny = any;

export class ShopeeAdsRepository {
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
    periodStart: string;
    periodEnd: string;
    periodLabel: string;
    totalRows: number;
    uploadedBy?: string;
  }): Promise<{ id: string }> {
    const result = await this.prisma.shopeeAdsUploadBatch.create({
      data: {
        tenantId: this.tenantId,
        fileName: data.fileName,
        periodStart: new Date(data.periodStart),
        periodEnd: new Date(data.periodEnd),
        totalRows: data.totalRows,
        uploadedBy: data.uploadedBy,
        status: "processing",
      },
    });
    return { id: result.id as string };
  }

  async updateUploadBatch(
    batchId: string,
    data: {
      insertedRows?: number;
      skippedRows?: number;
      updatedRows?: number;
      status?: string;
      errorMessage?: string | null;
    },
  ): Promise<void> {
    await this.prisma.shopeeAdsUploadBatch.update({
      where: { id: batchId },
      data,
    });
  }

  async getUploadHistory(limit = 20): Promise<unknown[]> {
    const result = await this.prisma.shopeeAdsUploadBatch.findMany({
      where: { tenantId: this.tenantId },
      orderBy: { createdAt: "desc" },
      take: limit,
    });
    return result as unknown[];
  }

  // ============================================
  // Product Data Operations
  // ============================================

  async upsertProductData(
    row: ShopeeAdsProductRow,
    batchId: string,
    mode: "skip" | "update",
  ): Promise<{ action: "inserted" | "skipped" | "updated" }> {
    // Check existing
    const existing = await this.prisma.shopeeAdsProductData.findFirst({
      where: {
        tenantId: this.tenantId,
        productId: row.productId,
        periodStart: row.periodStart,
        periodEnd: row.periodEnd,
      },
    });

    if (existing) {
      if (mode === "skip") {
        return { action: "skipped" };
      }
      // Update
      const existingId = (existing as { id: string }).id;
      await this.prisma.shopeeAdsProductData.update({
        where: { id: existingId },
        data: this.mapRowToData(row, batchId),
      });
      return { action: "updated" };
    }

    // Insert
    await this.prisma.shopeeAdsProductData.create({
      data: {
        tenantId: this.tenantId,
        ...this.mapRowToData(row, batchId),
      },
    });
    return { action: "inserted" };
  }

  private mapRowToData(row: ShopeeAdsProductRow, batchId: string) {
    return {
      uploadBatchId: batchId,
      periodStart: row.periodStart,
      periodEnd: row.periodEnd,
      periodLabel: row.periodLabel,
      productId: row.productId,
      productName: row.productName,
      status: row.status,
      biddingMode: row.biddingMode,
      placement: row.placement,
      startDate: row.startDate,
      endDate: row.endDate,
      impressions: row.impressions,
      clicks: row.clicks,
      ctr: row.ctr,
      conversions: row.conversions,
      directConversions: row.directConversions,
      conversionRate: row.conversionRate,
      directConversionRate: row.directConversionRate,
      costPerConversion: row.costPerConversion,
      costPerDirectConversion: row.costPerDirectConversion,
      unitsSold: row.unitsSold,
      directUnitsSold: row.directUnitsSold,
      revenue: row.revenue,
      directRevenue: row.directRevenue,
      cost: row.cost,
      roas: row.roas,
      directRoas: row.directRoas,
      acos: row.acos,
      directAcos: row.directAcos,
    };
  }

  // ============================================
  // Query Operations
  // ============================================

  async getProductData(
    filter: ProductDataFilter,
  ): Promise<{ data: unknown[]; total: number }> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const where: Record<string, any> = { tenantId: this.tenantId };

    if (filter.periodStart) {
      where.periodStart = { gte: new Date(filter.periodStart) };
    }
    if (filter.periodEnd) {
      where.periodEnd = { lte: new Date(filter.periodEnd) };
    }
    if (filter.productId) {
      where.productId = filter.productId;
    }
    if (filter.biddingMode) {
      where.biddingMode = filter.biddingMode;
    }
    if (filter.minRoas !== undefined) {
      where.roas = { ...where.roas, gte: filter.minRoas };
    }
    if (filter.maxRoas !== undefined) {
      where.roas = { ...where.roas, lte: filter.maxRoas };
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const orderBy: Record<string, any> = {};
    if (filter.orderBy) {
      orderBy[filter.orderBy] = filter.orderDir || "desc";
    } else {
      orderBy.revenue = "desc";
    }

    const [data, total] = await Promise.all([
      this.prisma.shopeeAdsProductData.findMany({
        where,
        orderBy,
        take: filter.limit || 100,
        skip: filter.offset || 0,
      }),
      this.prisma.shopeeAdsProductData.count({ where }),
    ]);

    return { data: data as unknown[], total: total as number };
  }

  async getDashboardSummary(
    periodStart?: string,
    periodEnd?: string,
  ): Promise<DashboardSummary> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const where: Record<string, any> = { tenantId: this.tenantId };
    if (periodStart) where.periodStart = { gte: new Date(periodStart) };
    if (periodEnd) where.periodEnd = { lte: new Date(periodEnd) };

    const allData = (await this.prisma.shopeeAdsProductData.findMany({
      where,
    })) as Array<{
      cost?: number;
      revenue?: number;
      directRevenue?: number;
      conversions?: number;
      impressions?: number;
      clicks?: number;
      productId?: string;
      productName?: string;
      biddingMode?: string;
    }>;

    // Aggregate metrics
    interface Totals {
      cost: number;
      revenue: number;
      directRevenue: number;
      orders: number;
      impressions: number;
      clicks: number;
    }
    const totals = allData.reduce<Totals>(
      (acc: Totals, row) => ({
        cost: acc.cost + (row.cost || 0),
        revenue: acc.revenue + (row.revenue || 0),
        directRevenue: acc.directRevenue + (row.directRevenue || 0),
        orders: acc.orders + (row.conversions || 0),
        impressions: acc.impressions + (row.impressions || 0),
        clicks: acc.clicks + (row.clicks || 0),
      }),
      {
        cost: 0,
        revenue: 0,
        directRevenue: 0,
        orders: 0,
        impressions: 0,
        clicks: 0,
      },
    );

    // Top products by revenue
    const topProducts = this.aggregateTopProducts(allData);

    // Bidding mode comparison
    const biddingModeComparison = this.aggregateBiddingModes(allData);

    return {
      totalCost: totals.cost,
      totalRevenue: totals.revenue,
      totalDirectRevenue: totals.directRevenue,
      totalOrders: totals.orders,
      avgRoas: totals.cost > 0 ? totals.revenue / totals.cost : 0,
      avgDirectRoas: totals.cost > 0 ? totals.directRevenue / totals.cost : 0,
      totalImpressions: totals.impressions,
      totalClicks: totals.clicks,
      avgCtr: totals.impressions > 0 ? totals.clicks / totals.impressions : 0,
      avgConversionRate: totals.clicks > 0 ? totals.orders / totals.clicks : 0,
      topProducts,
      biddingModeComparison,
    };
  }

  private aggregateTopProducts(
    data: Array<{
      productId?: string;
      productName?: string;
      cost?: number;
      revenue?: number;
      conversions?: number;
    }>,
  ): ProductPerformance[] {
    interface ProductAgg {
      productId: string;
      productName: string;
      cost: number;
      revenue: number;
      orders: number;
    }
    const byProduct = new Map<string, ProductAgg>();

    for (const row of data) {
      const key = row.productId || "unknown";
      if (!byProduct.has(key)) {
        byProduct.set(key, {
          productId: row.productId || "unknown",
          productName: row.productName || "Unknown",
          cost: 0,
          revenue: 0,
          orders: 0,
        });
      }
      const p = byProduct.get(key)!;
      p.cost += row.cost || 0;
      p.revenue += row.revenue || 0;
      p.orders += row.conversions || 0;
    }

    return Array.from(byProduct.values())
      .map((p) => ({ ...p, roas: p.cost > 0 ? p.revenue / p.cost : 0 }))
      .sort((a, b) => b.revenue - a.revenue)
      .slice(0, 10);
  }

  private aggregateBiddingModes(
    data: Array<{
      biddingMode?: string;
      cost?: number;
      revenue?: number;
      conversions?: number;
    }>,
  ): BiddingModeStats[] {
    interface ModeAgg {
      biddingMode: string;
      cost: number;
      revenue: number;
      orders: number;
      productCount: number;
    }
    const byMode = new Map<string, ModeAgg>();

    for (const row of data) {
      const key = row.biddingMode || "Unknown";
      if (!byMode.has(key)) {
        byMode.set(key, {
          biddingMode: key,
          cost: 0,
          revenue: 0,
          orders: 0,
          productCount: 0,
        });
      }
      const m = byMode.get(key)!;
      m.cost += row.cost || 0;
      m.revenue += row.revenue || 0;
      m.orders += row.conversions || 0;
      m.productCount++;
    }

    return Array.from(byMode.values())
      .map((m) => ({
        ...m,
        roas: m.cost > 0 ? m.revenue / m.cost : 0,
        costPerOrder: m.orders > 0 ? m.cost / m.orders : 0,
      }))
      .sort((a, b) => b.revenue - a.revenue);
  }
}

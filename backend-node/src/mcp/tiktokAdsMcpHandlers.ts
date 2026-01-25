/**
 * TikTok Ads MCP Tool Handlers
 *
 * Handler implementations for MCP tools
 */

import { getDbManager } from "../services/dbConnectionManager";
import type { PrismaClient } from "@prisma/client";

// Types for MCP data
interface CreativeData {
  cost: number;
  grossRevenue: number;
  ordersSku: number;
  impressions: number;
  clicks: number;
  productId: string;
  creativeType: string;
  campaignName: string;
}

// Get Prisma for tenant
function getPrisma(tenantId: string): PrismaClient {
  const dbManager = getDbManager();
  return dbManager.getConnection(tenantId);
}

export async function handleDashboard(args: { tenantId: string }) {
  const prisma = getPrisma(args.tenantId);
  const data: CreativeData[] = await prisma.tiktokAdsCreativeData.findMany({
    where: { tenantId: args.tenantId },
    select: {
      cost: true,
      grossRevenue: true,
      ordersSku: true,
      impressions: true,
      clicks: true,
      productId: true,
      creativeType: true,
      campaignName: true,
    },
  });

  const totalCost = data.reduce(
    (sum: number, d: CreativeData) => sum + d.cost,
    0
  );
  const totalRevenue = data.reduce(
    (sum: number, d: CreativeData) => sum + d.grossRevenue,
    0
  );
  const totalOrders = data.reduce(
    (sum: number, d: CreativeData) => sum + d.ordersSku,
    0
  );
  const totalImpressions = data.reduce(
    (sum: number, d: CreativeData) => sum + d.impressions,
    0
  );
  const totalClicks = data.reduce(
    (sum: number, d: CreativeData) => sum + d.clicks,
    0
  );

  return {
    summary: {
      totalCost,
      totalRevenue,
      totalOrders,
      roi: totalCost > 0 ? totalRevenue / totalCost : 0,
      totalImpressions,
      totalClicks,
      ctr: totalImpressions > 0 ? totalClicks / totalImpressions : 0,
      recordCount: data.length,
    },
  };
}

export async function handleTopProducts(args: {
  tenantId: string;
  sortBy?: string;
  limit?: number;
}) {
  const prisma = getPrisma(args.tenantId);
  const sortBy = args.sortBy || "revenue";
  const limit = args.limit || 10;

  const data: CreativeData[] = await prisma.tiktokAdsCreativeData.findMany({
    where: { tenantId: args.tenantId },
    select: {
      cost: true,
      grossRevenue: true,
      ordersSku: true,
      impressions: true,
      clicks: true,
      productId: true,
      creativeType: true,
      campaignName: true,
    },
  });

  // Aggregate by productId
  const productMap = new Map<
    string,
    { revenue: number; cost: number; orders: number }
  >();
  for (const d of data) {
    const existing = productMap.get(d.productId) || {
      revenue: 0,
      cost: 0,
      orders: 0,
    };
    productMap.set(d.productId, {
      revenue: existing.revenue + d.grossRevenue,
      cost: existing.cost + d.cost,
      orders: existing.orders + d.ordersSku,
    });
  }

  const products = Array.from(productMap.entries()).map(
    ([productId, stats]) => ({
      productId,
      ...stats,
      roi: stats.cost > 0 ? stats.revenue / stats.cost : 0,
    })
  );

  products.sort((a, b) => {
    if (sortBy === "roi") return b.roi - a.roi;
    if (sortBy === "orders") return b.orders - a.orders;
    return b.revenue - a.revenue;
  });

  return { products: products.slice(0, limit) };
}

export async function handleCreativeComparison(args: { tenantId: string }) {
  const prisma = getPrisma(args.tenantId);
  const data: CreativeData[] = await prisma.tiktokAdsCreativeData.findMany({
    where: { tenantId: args.tenantId },
    select: {
      cost: true,
      grossRevenue: true,
      ordersSku: true,
      impressions: true,
      clicks: true,
      productId: true,
      creativeType: true,
      campaignName: true,
    },
  });

  const comparison: Record<
    string,
    { cost: number; revenue: number; orders: number }
  > = {
    Video: { cost: 0, revenue: 0, orders: 0 },
    "Kartu produk": { cost: 0, revenue: 0, orders: 0 },
  };

  for (const d of data) {
    const type = d.creativeType || "Unknown";
    if (comparison[type]) {
      comparison[type].cost += d.cost;
      comparison[type].revenue += d.grossRevenue;
      comparison[type].orders += d.ordersSku;
    }
  }

  return {
    comparison: Object.entries(comparison).map(([type, stats]) => ({
      creativeType: type,
      ...stats,
      roi: stats.cost > 0 ? stats.revenue / stats.cost : 0,
    })),
  };
}

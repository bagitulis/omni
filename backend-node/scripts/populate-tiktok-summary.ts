/**
 * Script to populate TiktokAdsProductSummary table from existing TiktokAdsCreativeData
 * This creates pre-aggregated data for fast dashboard queries
 *
 * Run: npx ts-node scripts/populate-tiktok-summary.ts
 */

import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

interface ProductAggregation {
  productId: string;
  totalCost: number;
  totalRevenue: number;
  totalOrders: number;
  totalImpressions: number;
  totalClicks: number;
  creativeTypeCounts: Record<string, number>;
}

async function populateSummary(tenantId: string): Promise<void> {
  console.log(`\n📊 Processing tenant: ${tenantId}`);

  // Step 1: Get count of creative data
  const totalRows = await prisma.tiktokAdsCreativeData.count({
    where: { tenantId },
  });

  console.log(`   Found ${totalRows.toLocaleString()} creative data rows`);

  if (totalRows === 0) {
    console.log(`   ⏭️ Skipping - no data`);
    return;
  }

  // Step 2: Get date range
  const dateRange = await prisma.tiktokAdsCreativeData.aggregate({
    where: { tenantId },
    _min: { periodStart: true },
    _max: { periodEnd: true },
  });

  const minDate = dateRange._min.periodStart;
  const maxDate = dateRange._max.periodEnd;

  if (!minDate || !maxDate) {
    console.log(`   ⚠️ No valid date range found`);
    return;
  }

  console.log(
    `   Date range: ${minDate.toISOString().split("T")[0]} to ${
      maxDate.toISOString().split("T")[0]
    }`,
  );

  // Step 3: Aggregate by productId using raw SQL for performance
  console.log(`   🔄 Aggregating by productId...`);
  const startTime = Date.now();

  // Use raw SQL for better performance on large datasets
  const aggregations = await prisma.$queryRaw<
    Array<{
      productId: string;
      totalCost: number;
      totalRevenue: number;
      totalOrders: number;
      totalImpressions: number;
      totalClicks: number;
      bestCreativeType: string | null;
    }>
  >`
    SELECT 
      productId,
      SUM(cost) as totalCost,
      SUM(grossRevenue) as totalRevenue,
      SUM(ordersSku) as totalOrders,
      SUM(impressions) as totalImpressions,
      SUM(clicks) as totalClicks,
      (
        SELECT creativeType 
        FROM TiktokAdsCreativeData t2 
        WHERE t2.tenantId = ${tenantId} 
          AND t2.productId = TiktokAdsCreativeData.productId
        GROUP BY creativeType
        ORDER BY SUM(grossRevenue) DESC
        LIMIT 1
      ) as bestCreativeType
    FROM TiktokAdsCreativeData
    WHERE tenantId = ${tenantId}
    GROUP BY productId
  `;

  const queryTime = Date.now() - startTime;
  console.log(
    `   ✅ Query completed in ${(queryTime / 1000).toFixed(2)}s - ${
      aggregations.length
    } products`,
  );

  // Step 4: Clear existing summaries for this tenant
  console.log(`   🗑️ Clearing existing summaries...`);
  await prisma.tiktokAdsProductSummary.deleteMany({
    where: { tenantId },
  });

  // Step 5: Insert new summaries
  console.log(`   📥 Inserting ${aggregations.length} product summaries...`);

  const periodDate = new Date(minDate);
  periodDate.setHours(0, 0, 0, 0);

  // Use batch insert for performance
  const batchSize = 100;
  let inserted = 0;

  for (let i = 0; i < aggregations.length; i += batchSize) {
    const batch = aggregations.slice(i, i + batchSize);

    await prisma.tiktokAdsProductSummary.createMany({
      data: batch.map((agg) => {
        const cost = Number(agg.totalCost) || 0;
        const revenue = Number(agg.totalRevenue) || 0;
        const orders = Number(agg.totalOrders) || 0;
        const impressions = Number(agg.totalImpressions) || 0;
        const clicks = Number(agg.totalClicks) || 0;

        return {
          tenantId,
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
          bestCreativeType: agg.bestCreativeType,
        };
      }),
      skipDuplicates: true,
    });

    inserted += batch.length;
    process.stdout.write(
      `\r   📥 Progress: ${inserted}/${aggregations.length}`,
    );
  }

  console.log(`\n   ✅ Inserted ${inserted} product summaries`);
}

async function main(): Promise<void> {
  console.log("═══════════════════════════════════════════════════════");
  console.log("  TikTok Ads Product Summary Population Script");
  console.log("═══════════════════════════════════════════════════════");

  try {
    // Get all tenants with TikTok ads data
    const tenants = await prisma.tiktokAdsCreativeData.groupBy({
      by: ["tenantId"],
      _count: true,
    });

    console.log(`\n📋 Found ${tenants.length} tenants with TikTok Ads data:`);
    tenants.forEach((t) => {
      console.log(`   - ${t.tenantId}: ${t._count.toLocaleString()} rows`);
    });

    for (const tenant of tenants) {
      await populateSummary(tenant.tenantId);
    }

    console.log("\n═══════════════════════════════════════════════════════");
    console.log("  ✅ Population complete!");
    console.log("═══════════════════════════════════════════════════════\n");
  } catch (error) {
    console.error("\n❌ Error:", error);
    throw error;
  } finally {
    await prisma.$disconnect();
  }
}

main();

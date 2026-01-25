/**
 * TikTok Ads Initial Data Import Script
 *
 * Import existing Excel files from docs/nusseyba folder
 * Run: npx ts-node backend/scripts/importTiktokAdsData.ts
 */

import * as fs from "fs";
import * as path from "path";
import { TiktokAdsExcelParser } from "../src/services/analytics/tiktokAdsExcelParser";
import { TiktokAdsRepository } from "../src/services/analytics/tiktokAdsRepository";
import { CreativeDataWithPeriod } from "../src/services/analytics/tiktokAdsTypes";
import { getDbManager } from "../src/services/dbConnectionManager";

function getPrisma(tenantId: string) {
  return getDbManager().getConnection(tenantId);
}

const TENANT_ID = "yumna_bertigamart";
const DATA_DIR = path.join(__dirname, "../../docs/nusseyba");

// Files to import
const FILES_TO_IMPORT = [
  "creative data for product campaigns 2025-11-01 00 ~ 2025-11-30 23.xlsx",
  "creative data for product campaigns 2025-12-01 00 ~ 2025-12-31 23.xlsx",
  "creative data for product campaigns 2026-01-01 00 ~ 2026-01-12 23.xlsx",
];

async function main() {
  console.log("=== TikTok Ads Data Import ===\n");

  const prisma = getPrisma(TENANT_ID);
  const repository = new TiktokAdsRepository(prisma, TENANT_ID);

  // Check if data already exists
  const existingCount = await prisma.tiktokAdsCreativeData.count({
    where: { tenantId: TENANT_ID },
  });

  if (existingCount > 0) {
    console.log(`⚠️  Found ${existingCount} existing records`);
    console.log("   Using 'skip' mode to avoid duplicates\n");
  }

  let totalProcessed = 0;
  let totalInserted = 0;
  let totalSkipped = 0;

  for (const filename of FILES_TO_IMPORT) {
    const filePath = path.join(DATA_DIR, filename);

    // Check if file exists
    if (!fs.existsSync(filePath)) {
      console.log(`❌ File not found: ${filename}`);
      continue;
    }

    console.log(`📄 Processing: ${filename}`);

    // Read file buffer
    const buffer = fs.readFileSync(filePath);

    // Parse Excel
    const parser = new TiktokAdsExcelParser();
    let parseResult: {
      data: CreativeDataWithPeriod[];
      period: { start: Date; end: Date };
    };

    try {
      parseResult = parser.parseExcelBuffer(buffer, filename);
    } catch (err) {
      console.log(`   ❌ Parse error: ${(err as Error).message}`);
      continue;
    }

    if (!parseResult.data || parseResult.data.length === 0) {
      console.log(`   ⚠️ No data found in file`);
      continue;
    }

    console.log(`   📊 Found ${parseResult.data.length} rows`);

    // Create upload batch
    const batch = await repository.createUploadBatch({
      fileName: filename,
      periodStart: parseResult.period.start,
      periodEnd: parseResult.period.end,
      totalRows: parseResult.data.length,
    });

    // Insert data with deduplication
    let inserted = 0;
    let skipped = 0;
    let errors = 0;

    for (const row of parseResult.data) {
      try {
        // Check if exists (deduplication key)
        const existing = await prisma.tiktokAdsCreativeData.findFirst({
          where: {
            tenantId: TENANT_ID,
            campaignId: row.campaignId,
            productId: row.productId,
            videoId: row.videoId || "",
            periodStart: parseResult.period.start,
            periodEnd: parseResult.period.end,
          },
        });

        if (existing) {
          skipped++;
          continue;
        }

        // Insert new record using Prisma schema fields
        await prisma.tiktokAdsCreativeData.create({
          data: {
            tenantId: TENANT_ID,
            uploadBatchId: batch.id,
            campaignId: row.campaignId,
            campaignName: row.campaignName,
            productId: row.productId,
            videoId: row.videoId || "",
            videoTitle: row.videoTitle,
            tiktokAccount: row.tiktokAccount,
            postingTime: row.postingTime,
            status: row.status,
            authorizationType: row.authorizationType,
            creativeType: row.creativeType,
            impressions: row.impressions,
            clicks: row.clicks,
            ctr: row.ctr,
            conversionRate: row.conversionRate,
            cost: row.cost,
            grossRevenue: row.grossRevenue,
            ordersSku: row.ordersSku,
            costPerOrder: row.costPerOrder,
            roi: row.roi,
            watchRate2s: row.watchRate2s,
            watchRate6s: row.watchRate6s,
            watchRate25pct: row.watchRate25pct,
            watchRate50pct: row.watchRate50pct,
            watchRate75pct: row.watchRate75pct,
            watchRate100pct: row.watchRate100pct,
            currency: row.currency,
            periodStart: parseResult.period.start,
            periodEnd: parseResult.period.end,
          },
        });

        inserted++;
      } catch (err) {
        errors++;
        // Log first few errors only
        if (errors <= 3) {
          console.log(`   ⚠️ Error inserting row: ${(err as Error).message}`);
        }
      }
    }

    // Update batch status
    await repository.updateUploadBatch(batch.id, {
      insertedRows: inserted,
      skippedRows: skipped,
      status: "completed",
    });

    console.log(
      `   ✅ Inserted: ${inserted}, Skipped: ${skipped}, Errors: ${errors}\n`
    );

    totalProcessed += parseResult.data.length;
    totalInserted += inserted;
    totalSkipped += skipped;
  }

  // Generate product summaries
  console.log("📈 Generating product summaries...");
  const summaryResult = await generateProductSummaries(prisma);
  console.log(`   ✅ Generated ${summaryResult} product summaries\n`);

  // Print summary
  console.log("=== Import Complete ===");
  console.log(`Total processed: ${totalProcessed}`);
  console.log(`Total inserted:  ${totalInserted}`);
  console.log(`Total skipped:   ${totalSkipped}`);

  // Final count
  const finalCount = await prisma.tiktokAdsCreativeData.count({
    where: { tenantId: TENANT_ID },
  });
  console.log(`\nTotal records in database: ${finalCount}`);
}

async function generateProductSummaries(
  prisma: ReturnType<typeof getPrisma>
): Promise<number> {
  // Get unique products with monthly aggregation
  const products = await prisma.tiktokAdsCreativeData.groupBy({
    by: ["productId"],
    where: { tenantId: TENANT_ID },
    _sum: {
      cost: true,
      grossRevenue: true,
      ordersSku: true,
      impressions: true,
      clicks: true,
    },
    _count: true,
  });

  let created = 0;
  const now = new Date();
  // Use first day of current month as periodDate
  const periodDate = new Date(now.getFullYear(), now.getMonth(), 1);

  for (const product of products) {
    const cost = product._sum.cost || 0;
    const revenue = product._sum.grossRevenue || 0;
    const orders = product._sum.ordersSku || 0;
    const impressions = product._sum.impressions || 0;
    const clicks = product._sum.clicks || 0;

    // Upsert product summary with proper unique key
    await prisma.tiktokAdsProductSummary.upsert({
      where: {
        tenantId_productId_periodType_periodDate: {
          tenantId: TENANT_ID,
          productId: product.productId,
          periodType: "monthly",
          periodDate: periodDate,
        },
      },
      update: {
        totalCost: cost,
        totalRevenue: revenue,
        totalOrders: orders,
        totalImpressions: impressions,
        totalClicks: clicks,
        avgRoi: cost > 0 ? revenue / cost : 0,
        avgCtr: impressions > 0 ? clicks / impressions : 0,
      },
      create: {
        tenantId: TENANT_ID,
        productId: product.productId,
        periodType: "monthly",
        periodDate: periodDate,
        totalCost: cost,
        totalRevenue: revenue,
        totalOrders: orders,
        totalImpressions: impressions,
        totalClicks: clicks,
        avgRoi: cost > 0 ? revenue / cost : 0,
        avgCtr: impressions > 0 ? clicks / impressions : 0,
      },
    });

    created++;
  }

  return created;
}

main().catch((err) => {
  console.error("Import failed:", err);
  process.exit(1);
});

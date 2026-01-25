/**
 * TikTok API vs Database Comparison
 * Find missing orders between API and Database
 */
const fetch = require("node-fetch");
const { PrismaClient } = require("@prisma/client");

const DATABASE_PATH = "/app/config/databases/yumna_bertigamart.db";
const API_URL = "http://localhost:3000";
const TENANT_ID = "yumna_bertigamart";

async function main() {
  console.log("🔍 Comparing TikTok API vs Database...\n");

  // Create Prisma client for tenant database
  const prisma = new PrismaClient({
    datasources: {
      db: { url: `file:${DATABASE_PATH}` },
    },
  });

  try {
    // 1. Fetch from TikTok API via sync endpoint
    console.log("📡 Fetching from TikTok API...");
    const response = await fetch(`${API_URL}/api/orders/sync/processed`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": TENANT_ID,
      },
      body: JSON.stringify({
        platform: "tiktok",
        forceRefresh: true,
      }),
    });

    const result = await response.json();
    console.log("API Response status:", result.success ? "SUCCESS" : "FAILED");

    // Get TikTok orders from API response
    const apiOrders =
      result.data?.orders?.filter((o) => o._platform === "tiktok") || [];
    const apiOrderSns = apiOrders.map((o) => o.order_sn);
    console.log(`API returned: ${apiOrders.length} TikTok PROCESSED orders`);

    // 2. Get from Database
    console.log("\n📊 Checking Database...");
    const dbOrders = await prisma.tiktokOrder.findMany({
      where: {
        tenantId: TENANT_ID,
        orderStatus: "AWAITING_COLLECTION",
      },
    });
    const dbOrderSns = dbOrders.map((o) => o.orderSn);
    console.log(`Database has: ${dbOrders.length} AWAITING_COLLECTION orders`);

    // 3. Compare
    console.log("\n🔄 Comparing...");

    // Orders in API but not in DB (missing from DB)
    const missingInDb = apiOrderSns.filter((sn) => !dbOrderSns.includes(sn));
    console.log(
      `\n❌ Missing in DB (in API but not DB): ${missingInDb.length}`
    );
    if (missingInDb.length > 0) {
      console.log("Missing order SNs:", missingInDb);
    }

    // Orders in DB but not in API (stale/extra in DB)
    const extraInDb = dbOrderSns.filter((sn) => !apiOrderSns.includes(sn));
    console.log(`\n⚠️  Extra in DB (in DB but not API): ${extraInDb.length}`);
    if (extraInDb.length > 0) {
      console.log("Extra order SNs:", extraInDb);
    }

    // Orders in both (synced correctly)
    const synced = apiOrderSns.filter((sn) => dbOrderSns.includes(sn));
    console.log(`\n✅ Synced correctly: ${synced.length}`);

    // Summary
    console.log("\n" + "=".repeat(60));
    console.log("📊 SUMMARY:");
    console.log(`   TikTok API PROCESSED orders: ${apiOrders.length}`);
    console.log(`   Database AWAITING_COLLECTION orders: ${dbOrders.length}`);
    console.log(`   Missing in DB: ${missingInDb.length}`);
    console.log(`   Extra in DB: ${extraInDb.length}`);
    console.log("=".repeat(60));

    // Also check raw API response
    if (result.data?.rawResponse?.data?.orders) {
      const rawOrders = result.data.rawResponse.data.orders;
      console.log(`\n📡 Raw TikTok API orders: ${rawOrders.length}`);
      const rawSns = rawOrders.map((o) => o.order_sn || o.id);
      console.log("Raw order SNs:", rawSns);
    }
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    await prisma.$disconnect();
  }
}

main();

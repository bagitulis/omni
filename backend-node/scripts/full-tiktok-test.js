/**
 * Full TikTok Sync Test with Database Check
 */
const fetch = require("node-fetch");
const { PrismaClient } = require("@prisma/client");

const API_URL = "http://localhost:3000";
const TENANT_ID = "yumna_bertigamart";
const DATABASE_PATH = "/app/config/databases/yumna_bertigamart.db";

async function main() {
  console.log("🔍 Full TikTok Sync Test...\n");
  console.log(`Using tenant: ${TENANT_ID}`);

  const prisma = new PrismaClient({
    datasources: {
      db: { url: `file:${DATABASE_PATH}` },
    },
  });

  try {
    // 1. Check DB before sync
    const beforeCount = await prisma.tiktokOrder.count({
      where: { tenantId: TENANT_ID },
    });
    console.log(`\n📊 BEFORE SYNC: ${beforeCount} TikTok orders in DB`);

    // 2. Call sync API
    console.log("\n📡 Calling POST /api/orders/sync/processed...");
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
    console.log("Response status:", response.status);
    console.log("API Success:", result.success);

    // Show sync details
    if (result.data?.results) {
      console.log("\n📊 Sync Results:");
      for (const [platform, data] of Object.entries(result.data.results)) {
        console.log(
          `   ${platform}: ${data.count || 0} orders (success: ${data.success})`
        );
        if (data.error) console.log(`   Error: ${data.error}`);
      }
    }

    // 3. Check DB after sync
    const afterCount = await prisma.tiktokOrder.count({
      where: { tenantId: TENANT_ID },
    });
    console.log(`\n📊 AFTER SYNC: ${afterCount} TikTok orders in DB`);

    // 4. List orders
    const orders = await prisma.tiktokOrder.findMany({
      where: { tenantId: TENANT_ID },
      take: 20,
    });
    console.log(`\n📦 TikTok Orders in DB (${orders.length}):`);
    for (const order of orders) {
      console.log(`   - ${order.orderSn} [${order.orderStatus}]`);
    }
  } catch (error) {
    console.error("Error:", error.message);
    console.error(error.stack);
  } finally {
    await prisma.$disconnect();
  }
}

main();

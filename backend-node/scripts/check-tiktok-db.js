/**
 * Check TikTok Database Status
 */
const { PrismaClient } = require("@prisma/client");

const DATABASE_PATH = "/app/config/databases/yumna_bertigamart.db";
const TENANT_ID = "yumna_bertigamart";

async function main() {
  const prisma = new PrismaClient({
    datasources: {
      db: { url: `file:${DATABASE_PATH}` },
    },
  });

  try {
    console.log("🔍 Checking TikTok Orders in Database...\n");

    // Get all TikTok orders grouped by status
    const allOrders = await prisma.tiktokOrder.findMany({
      where: { tenantId: TENANT_ID },
    });

    console.log(`Total TikTok orders: ${allOrders.length}`);

    // Group by status
    const statusGroups = {};
    for (const order of allOrders) {
      const status = order.orderStatus || "UNKNOWN";
      if (!statusGroups[status]) {
        statusGroups[status] = [];
      }
      statusGroups[status].push(order.orderSn);
    }

    console.log("\n📊 Orders by Status:");
    for (const [status, orders] of Object.entries(statusGroups)) {
      console.log(`   ${status}: ${orders.length}`);
    }

    // Show order SNs for AWAITING_COLLECTION
    if (statusGroups["AWAITING_COLLECTION"]) {
      console.log("\n📦 AWAITING_COLLECTION order SNs:");
      statusGroups["AWAITING_COLLECTION"].forEach((sn) =>
        console.log(`   - ${sn}`)
      );
    }

    console.log("\n" + "=".repeat(60));
  } catch (error) {
    console.error("Error:", error.message);
  } finally {
    await prisma.$disconnect();
  }
}

main();

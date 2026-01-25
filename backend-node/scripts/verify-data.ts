import { PrismaClient } from "@prisma/client";
import path from "path";

/**
 * Verify current Google Sheets data in databases
 * Test data should be: WALLET_TEST_NEW, SHIPPING_TEST_NEW, ORDER_TEST_NEW
 */

async function checkCurrentData() {
  console.log("\n╔════════════════════════════════════════════╗");
  console.log("║  VERIFY SAVED DATA IN DATABASES            ║");
  console.log("╚════════════════════════════════════════════╝\n");

  const tenants = {
    yumna: "config/databases/yumna_bertigamart.db",
  };

  for (const [tenantId, dbPath] of Object.entries(tenants)) {
    try {
      const dbUrl = `file:${path.resolve(dbPath)}`;
      const prisma = new PrismaClient({
        datasources: {
          db: {
            url: dbUrl,
          },
        },
      });

      const settings = await prisma.googleSheetsSettings.findFirst();

      console.log(`📊 Tenant ${tenantId}:`);
      if (!settings) {
        console.log(`   ⚠️ No data found`);
      } else {
        console.log(`   Wallet: ${settings.walletSpreadsheetId || "empty"}`);
        console.log(`   Shipping: ${settings.shippingSpreadsheetId || "empty"}`);
        console.log(`   Order: ${settings.orderSpreadsheetId || "empty"}`);
        console.log(`   Last updated: ${settings.updatedAt}`);

        const hasTestData =
          settings.walletSpreadsheetId === "WALLET_TEST_NEW" &&
          settings.shippingSpreadsheetId === "SHIPPING_TEST_NEW" &&
          settings.orderSpreadsheetId === "ORDER_TEST_NEW";

        if (hasTestData) {
          console.log(`   ✅ TEST DATA VERIFIED!\n`);
        } else {
          console.log(`   ⚠️ Different data than test\n`);
        }
      }

      await prisma.$disconnect();
    } catch (error: any) {
      console.log(`   ❌ Error: ${error.message}\n`);
    }
  }
}

checkCurrentData().catch(console.error);

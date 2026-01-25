// Direct database query script
// File: backend/scripts/check-filter-db.ts

import { getPrisma } from "../src/services/prismaClient";

async function checkFilterDB() {
  const prisma = getPrisma();

  try {
    console.log("🔍 Querying FilterPreference table...\n");

    // Get all records
    const allRecords = await prisma.filterPreference.findMany();
    console.log(`📊 Total records in database: ${allRecords.length}\n`);

    // Get specific inventory record
    const inventoryRecord = await prisma.filterPreference.findUnique({
      where: {
        platform_tab: { platform: "inventory", tab: "inventory" },
      },
    });

    if (!inventoryRecord) {
      console.log("❌ No record found for inventory/inventory");
      return;
    }

    console.log("✅ Found inventory record:");
    console.log(`   ID: ${inventoryRecord.id}`);
    console.log(`   Platform: ${inventoryRecord.platform}`);
    console.log(`   Tab: ${inventoryRecord.tab}`);
    console.log(`   Created: ${inventoryRecord.createdAt}`);
    console.log(`   Updated: ${inventoryRecord.updatedAt}`);

    // Parse and display visibleColumns
    console.log("\n📋 visibleColumns (raw):");
    console.log(`   Type: ${typeof inventoryRecord.visibleColumns}`);
    console.log(`   Length: ${inventoryRecord.visibleColumns?.length || 0}`);
    console.log(`   Content: ${inventoryRecord.visibleColumns}`);

    try {
      const parsed = JSON.parse(inventoryRecord.visibleColumns || "[]");
      console.log(`   Parsed type: ${Array.isArray(parsed) ? "Array" : typeof parsed}`);
      console.log(`   Parsed length: ${Array.isArray(parsed) ? parsed.length : Object.keys(parsed).length}`);
      console.log(`   Parsed content: ${JSON.stringify(parsed, null, 2)}`);
    } catch (e) {
      console.log(`   ❌ Parse error: ${e}`);
    }

    // Parse and display lockedColumns
    console.log("\n🔒 lockedColumns (raw):");
    console.log(`   Type: ${typeof inventoryRecord.lockedColumns}`);
    console.log(`   Length: ${inventoryRecord.lockedColumns?.length || 0}`);
    console.log(`   Content: ${inventoryRecord.lockedColumns}`);

    try {
      const parsed = JSON.parse(inventoryRecord.lockedColumns || "[]");
      console.log(`   Parsed type: ${Array.isArray(parsed) ? "Array" : typeof parsed}`);
      console.log(`   Parsed length: ${Array.isArray(parsed) ? parsed.length : 0}`);
      console.log(`   Parsed content: ${JSON.stringify(parsed, null, 2)}`);
    } catch (e) {
      console.log(`   ❌ Parse error: ${e}`);
    }

    // Parse and display columnFilters
    console.log("\n🔤 columnFilters (raw):");
    console.log(`   Type: ${typeof inventoryRecord.columnFilters}`);
    console.log(`   Length: ${inventoryRecord.columnFilters?.length || 0}`);

    try {
      const parsed = JSON.parse(inventoryRecord.columnFilters || "{}");
      console.log(`   Parsed keys: ${Object.keys(parsed).length}`);
      console.log(`   Parsed content: ${JSON.stringify(parsed, null, 2)}`);
    } catch (e) {
      console.log(`   ❌ Parse error: ${e}`);
    }

    console.log("\n✅ Database check complete");
  } catch (error) {
    console.error("❌ Error:", error);
  } finally {
    await prisma.$disconnect();
  }
}

checkFilterDB();

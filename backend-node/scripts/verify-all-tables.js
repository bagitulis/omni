const Database = require("better-sqlite3");

// Expected tables from Prisma schema
const expectedTables = {
  main: [
    "User",
    "PlatformConfig",
    "ShopeeOrder",
    "ShopeeOrderItem",
    "LazadaOrder",
    "LazadaOrderItem",
    "TiktokOrder",
    "TiktokOrderItem",
    "GoogleSheetsSettings",
    "InventorySettings",
    "InventoryRecord",
    "InventorySyncHistory",
    "SheetSnapshot",
    "LazadaProduct",
    "LazadaSku",
    "ShopeeProduct",
    "ShopeeSku",
    "TiktokProduct",
    "TiktokSku",
    "FilterPreference",
    "InventorySkuPlatformStatus",
    "RouteConfig",
    "Spreadsheet",
    "LockedOrder",
    "WholesaleSettings",
    "ShopeeEscrowSync",
    "ShopeeEscrowOrder",
    "ShopeeEscrowItem",
    "AnalyticsSettings",
  ],
  jobs: [
    "AnalyticsSettings",
    "ShopeeEscrowSync",
    "ShopeeEscrowOrder",
    "ShopeeEscrowItem",
    // Add any other job-specific tables here
  ],
};

const databases = {
  "yumna_bertigamart.db": "main",
  "tester_developer.db": "main",
  "yumna_jobs.db": "jobs",
  "tester_jobs.db": "jobs",
};

console.log("\n🔍 COMPREHENSIVE DATABASE SCHEMA VERIFICATION");
console.log("=".repeat(70) + "\n");

let totalMissing = 0;
let totalExtra = 0;

Object.entries(databases).forEach(([dbFile, dbType]) => {
  console.log(`\n📦 Database: ${dbFile} (type: ${dbType})`);
  console.log("-".repeat(70));

  const db = new Database(`./config/databases/${dbFile}`, { readonly: true });

  try {
    // Get all tables (excluding SQLite system tables and Prisma tables)
    const actualTables = db
      .prepare(
        `SELECT name FROM sqlite_master 
         WHERE type='table' 
         AND name NOT LIKE 'sqlite_%' 
         AND name NOT LIKE '_prisma_%'
         ORDER BY name`
      )
      .all()
      .map((t) => t.name);

    const expected = expectedTables[dbType] || [];

    // Find missing tables
    const missing = expected.filter((t) => !actualTables.includes(t));

    // Find extra tables (not in schema)
    const extra = actualTables.filter((t) => !expected.includes(t));

    console.log(`\n📊 Statistics:`);
    console.log(`   Expected tables: ${expected.length}`);
    console.log(`   Actual tables: ${actualTables.length}`);
    console.log(`   Missing: ${missing.length}`);
    console.log(`   Extra/Unknown: ${extra.length}`);

    if (missing.length > 0) {
      console.log(`\n❌ MISSING TABLES:`);
      missing.forEach((t) => console.log(`   - ${t}`));
      totalMissing += missing.length;
    } else {
      console.log(`\n✅ All expected tables present!`);
    }

    if (extra.length > 0) {
      console.log(`\n⚠️  EXTRA TABLES (not in schema):`);
      extra.forEach((t) => console.log(`   - ${t}`));
      totalExtra += extra.length;
    }

    // Show all tables
    console.log(`\n📋 All tables in database:`);
    actualTables.forEach((t, i) => {
      const isExpected = expected.includes(t);
      const icon = isExpected ? "✓" : "?";
      console.log(`   ${icon} ${i + 1}. ${t}`);
    });
  } catch (e) {
    console.error(`❌ Error: ${e.message}`);
  } finally {
    db.close();
  }
});

console.log("\n" + "=".repeat(70));
console.log("📊 OVERALL SUMMARY");
console.log("=".repeat(70));
console.log(`Total missing tables across all databases: ${totalMissing}`);
console.log(`Total extra tables across all databases: ${totalExtra}`);

if (totalMissing === 0 && totalExtra === 0) {
  console.log("\n🎉 ALL DATABASES MATCH SCHEMA PERFECTLY!");
} else if (totalMissing > 0) {
  console.log("\n⚠️  Some tables are missing - may need migration");
} else {
  console.log("\n✅ No missing tables, but some extra tables found");
}

console.log("=".repeat(70) + "\n");

const Database = require("better-sqlite3");

const databases = [
  "yumna_bertigamart.db",
  "tester_developer.db",
  "yumna_jobs.db",
  "tester_jobs.db",
];

console.log("\n🔧 FIXING SHOPEE ESCROW ITEM SCHEMA - ADDING MISSING COLUMNS\n");
console.log(
  "⚠️  This will add legacy columns to ShopeeEscrowItem for backward compatibility"
);
console.log("=".repeat(70) + "\n");

databases.forEach((dbFile) => {
  console.log(`\n📦 Processing: ${dbFile}`);
  console.log("-".repeat(70));

  const db = new Database(`./config/databases/${dbFile}`);

  try {
    // Check if ShopeeEscrowItem table exists
    const tableExists = db
      .prepare(
        `SELECT name FROM sqlite_master WHERE type='table' AND name='ShopeeEscrowItem'`
      )
      .get();

    if (!tableExists) {
      console.log("   ⏭️  No ShopeeEscrowItem table - skipping");
      db.close();
      return;
    }

    // Get current columns
    const columns = db.prepare(`PRAGMA table_info(ShopeeEscrowItem)`).all();
    const columnNames = columns.map((col) => col.name);

    console.log(`   Current columns: ${columns.length}`);

    // Define missing columns to add
    const columnsToAdd = [
      { name: "orderId", type: "TEXT", after: "escrowOrderId" },
      { name: "orderSn", type: "TEXT", after: "orderId" },
      { name: "month", type: "INTEGER", after: "orderSn" },
      { name: "year", type: "INTEGER", after: "month" },
    ];

    let addedCount = 0;

    // Add missing columns
    for (const col of columnsToAdd) {
      if (!columnNames.includes(col.name)) {
        console.log(`   ➕ Adding column: ${col.name} (${col.type})`);
        try {
          db.prepare(
            `ALTER TABLE ShopeeEscrowItem ADD COLUMN ${col.name} ${col.type}`
          ).run();
          addedCount++;
          console.log(`      ✅ Successfully added ${col.name}`);
        } catch (err) {
          console.log(`      ❌ Failed to add ${col.name}: ${err.message}`);
        }
      } else {
        console.log(`   ✓ Column ${col.name} already exists`);
      }
    }

    // Verify final schema
    const finalColumns = db
      .prepare(`PRAGMA table_info(ShopeeEscrowItem)`)
      .all();
    console.log(`\n   Final column count: ${finalColumns.length}`);
    console.log(`   Added ${addedCount} new columns`);

    // Verify orderId exists now
    const hasOrderId = finalColumns.some((col) => col.name === "orderId");
    console.log(
      `   ${hasOrderId ? "✅" : "❌"} orderId column ${
        hasOrderId ? "verified" : "still missing!"
      }`
    );
  } catch (e) {
    console.error(`   ❌ Error: ${e.message}`);
  } finally {
    db.close();
  }
});

console.log("\n" + "=".repeat(70));
console.log("✅ Schema fix complete!");
console.log("=".repeat(70));
console.log("\n💡 Tip: Run 'node check-escrow-schema.js' to verify changes\n");

const Database = require("better-sqlite3");

const db = new Database("/app/config/databases/yumna_bertigamart.db");

console.log("🔍 Verifying WebhookOrderEvent table...\n");

try {
  // Check if table exists
  const tableInfo = db
    .prepare(
      "SELECT sql FROM sqlite_master WHERE type='table' AND name='WebhookOrderEvent'"
    )
    .get();

  if (tableInfo) {
    console.log("✅ WebhookOrderEvent table exists");
    console.log("\nTable schema:");
    console.log(tableInfo.sql);

    // Check indexes
    const indexes = db
      .prepare(
        "SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name='WebhookOrderEvent'"
      )
      .all();
    console.log("\nIndexes:");
    for (const idx of indexes) {
      console.log("  -", idx.name);
    }

    // Count records
    const count = db
      .prepare("SELECT COUNT(*) as count FROM WebhookOrderEvent")
      .get();
    console.log("\nCurrent records:", count.count);
  } else {
    console.log("❌ WebhookOrderEvent table NOT found");
  }
} catch (err) {
  console.error("Error:", err.message);
}

db.close();

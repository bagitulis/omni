const Database = require("better-sqlite3");

const databases = [
  "yumna_bertigamart.db",
  "tester_developer.db",
  "yumna_jobs.db",
  "tester_jobs.db",
];

console.log("\n🔍 CHECKING SHOPEE ESCROW TABLES SCHEMA\n");

databases.forEach((dbFile) => {
  console.log(`\n${"=".repeat(70)}`);
  console.log(`DATABASE: ${dbFile}`);
  console.log("=".repeat(70));

  const db = new Database(`./config/databases/${dbFile}`, {
    readonly: true,
  });

  try {
    // Check if ShopeeEscrowItem table exists
    const tables = db
      .prepare(
        `SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%Escrow%'`
      )
      .all();

    if (tables.length === 0) {
      console.log("⚠️  No Escrow tables found in this database");
      db.close();
      return;
    }

    console.log(`\n📊 Escrow Tables Found: ${tables.length}`);
    tables.forEach((t) => console.log(`   - ${t.name}`));

    // Check ShopeeEscrowItem columns
    if (tables.some((t) => t.name === "ShopeeEscrowItem")) {
      console.log(`\n📋 ShopeeEscrowItem Columns:`);
      const columns = db.prepare(`PRAGMA table_info(ShopeeEscrowItem)`).all();
      columns.forEach((col) => {
        console.log(
          `   ${col.cid}. ${col.name.padEnd(30)} ${col.type.padEnd(10)} ${
            col.notnull ? "NOT NULL" : ""
          } ${col.dflt_value ? "DEFAULT " + col.dflt_value : ""}`
        );
      });

      // Check for orderId column specifically
      const hasOrderId = columns.some((col) => col.name === "orderId");
      console.log(
        `\n   ${hasOrderId ? "✅" : "❌"} orderId column ${
          hasOrderId ? "EXISTS" : "MISSING"
        }`
      );

      // Check row count
      const count = db
        .prepare(`SELECT COUNT(*) as count FROM ShopeeEscrowItem`)
        .get();
      console.log(`\n   Row count: ${count.count}`);
    }

    // Check ShopeeEscrowOrder columns
    if (tables.some((t) => t.name === "ShopeeEscrowOrder")) {
      console.log(`\n📋 ShopeeEscrowOrder Columns:`);
      const columns = db.prepare(`PRAGMA table_info(ShopeeEscrowOrder)`).all();
      columns.forEach((col) => {
        console.log(
          `   ${col.cid}. ${col.name.padEnd(30)} ${col.type.padEnd(10)} ${
            col.notnull ? "NOT NULL" : ""
          } ${col.dflt_value ? "DEFAULT " + col.dflt_value : ""}`
        );
      });

      // Check row count
      const count = db
        .prepare(`SELECT COUNT(*) as count FROM ShopeeEscrowOrder`)
        .get();
      console.log(`\n   Row count: ${count.count}`);
    }

    // Check ShopeeEscrowSync columns
    if (tables.some((t) => t.name === "ShopeeEscrowSync")) {
      console.log(`\n📋 ShopeeEscrowSync Columns:`);
      const columns = db.prepare(`PRAGMA table_info(ShopeeEscrowSync)`).all();
      columns.forEach((col) => {
        console.log(
          `   ${col.cid}. ${col.name.padEnd(30)} ${col.type.padEnd(10)} ${
            col.notnull ? "NOT NULL" : ""
          } ${col.dflt_value ? "DEFAULT " + col.dflt_value : ""}`
        );
      });

      // Check row count
      const count = db
        .prepare(`SELECT COUNT(*) as count FROM ShopeeEscrowSync`)
        .get();
      console.log(`\n   Row count: ${count.count}`);
    }
  } catch (e) {
    console.error("❌ Error:", e.message);
  } finally {
    db.close();
  }
});

console.log("\n" + "=".repeat(70));
console.log("✅ Schema inspection complete!");
console.log("=".repeat(70) + "\n");

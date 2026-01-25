const Database = require("better-sqlite3");

const databases = [
  "yumna_bertigamart.db",
  "tester_developer.db",
  "yumna_jobs.db",
  "tester_jobs.db",
];

databases.forEach((dbFile) => {
  console.log(`\n${"=".repeat(60)}`);
  console.log(`DATABASE: ${dbFile}`);
  console.log("=".repeat(60));

  const db = new Database(`./config/databases/${dbFile}`, {
    readonly: true,
  });

  try {
    // Get all tables
    const tables = db
      .prepare(
        `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE '_prisma_%' ORDER BY name`
      )
      .all();

    console.log(`\n📊 Total Tables: ${tables.length}`);
    console.log("\nTables:");
    tables.forEach((table, i) => {
      console.log(`  ${i + 1}. ${table.name}`);
    });

    // Get row counts for each table
    console.log("\n📈 Row Counts:");
    tables.forEach((table) => {
      try {
        const result = db
          .prepare(`SELECT COUNT(*) as count FROM ${table.name}`)
          .get();
        console.log(`  ${table.name}: ${result.count} rows`);
      } catch (e) {
        console.log(`  ${table.name}: Error - ${e.message}`);
      }
    });
  } catch (e) {
    console.error("Error:", e.message);
  } finally {
    db.close();
  }
});

console.log("\n" + "=".repeat(60));
console.log("✅ Schema check complete!");
console.log("=".repeat(60));

/**
 * Check Tables Script
 * List all tables in tenant databases and jobs databases
 */

const Database = require("better-sqlite3");
const path = require("path");
const fs = require("fs");

const dbDir = path.join(process.cwd(), "config/databases");

console.log("=== DATABASE TABLE ANALYSIS ===\n");

// List all database files
const files = fs.readdirSync(dbDir).filter((f) => f.endsWith(".db"));

for (const file of files) {
  const dbPath = path.join(dbDir, file);
  console.log(`📁 ${file}`);
  console.log("─".repeat(50));

  try {
    const db = new Database(dbPath, { readonly: true });
    const tables = db
      .prepare(
        "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
      )
      .all();

    for (const t of tables) {
      const countResult = db
        .prepare(`SELECT COUNT(*) as count FROM "${t.name}"`)
        .get();
      console.log(`  • ${t.name} (${countResult.count} rows)`);
    }

    db.close();
  } catch (err) {
    console.log(`  ❌ Error: ${err.message}`);
  }

  console.log("");
}

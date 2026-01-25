const Database = require("better-sqlite3");
const fs = require("fs");
const path = require("path");

const dbPath = path.join(__dirname, "dev.db");

// Delete and recreate database
if (fs.existsSync(dbPath)) {
  fs.unlinkSync(dbPath);
}

const db = new Database(dbPath);

// Read migration SQL
const migrationSql = fs.readFileSync(
  path.join(__dirname, "prisma/migrations/20251227064833_init/migration.sql"),
  "utf8"
);

// Execute entire migration at once
console.log("🔨 Executing migration SQL...");

try {
  db.exec(migrationSql);
  console.log("✅ Migration executed successfully");
} catch (error) {
  console.error("❌ Error executing migration:", error.message);
}

// Verify tables
const tables = db
  .prepare("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
  .all();

console.log(`\n📊 Database now has ${tables.length} tables:`);
tables.forEach((t) => {
  console.log(`  - ${t.name}`);
});

db.close();
console.log("\n✅ Database setup complete!");

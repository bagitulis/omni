const Database = require("better-sqlite3");
const db = new Database("./config/databases/yumna_bertigamart.db");
const tables = db
  .prepare("SELECT name FROM sqlite_master WHERE type='table'")
  .all();
console.log("Tables:", tables.map((t) => t.name).join(", "));
console.log("---");
const migrations = db
  .prepare("SELECT migration_name FROM _prisma_migrations")
  .all();
console.log("Migrations:", migrations.map((m) => m.migration_name).join(", "));
db.close();

const Database = require("better-sqlite3");
const path = require("path");

// Use relative path from backend-node
const dbPath = "./data/system.db";
console.log("Opening:", dbPath);

const db = new Database(dbPath);

// List all tables
const tables = db
  .prepare("SELECT name FROM sqlite_master WHERE type='table'")
  .all();
console.log("\nTables in system.db:");
tables.forEach((t) => console.log(" -", t.name));

// Check if GlobalConfig exists
if (tables.some((t) => t.name === "GlobalConfig")) {
  const configs = db
    .prepare("SELECT platform, configKey, isEncrypted FROM GlobalConfig")
    .all();
  console.log("\nGlobalConfig entries:");
  console.table(configs);
} else {
  console.log("\nNo GlobalConfig table found!");
}

db.close();

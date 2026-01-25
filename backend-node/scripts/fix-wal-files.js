/**
 * Fix WAL Locking Protocol Issues
 *
 * This script:
 * 1. Checkpoints all SQLite databases (merges WAL into main DB)
 * 2. Removes stale -shm and -wal files
 * 3. Prepares databases for clean restart
 *
 * Run AFTER stopping backend and n8n containers!
 */

const Database = require("better-sqlite3");
const fs = require("fs");
const path = require("path");

const dbDir = path.join(__dirname, "../config/databases");

console.log("🔧 Fixing SQLite WAL locking issues...\n");
console.log(`📂 Database directory: ${dbDir}\n`);

// Get all .db files (excluding backups and templates)
const dbFiles = fs
  .readdirSync(dbDir)
  .filter((f) => f.endsWith(".db"))
  .filter(
    (f) =>
      !f.includes("backup") &&
      !f.includes("corrupt") &&
      !f.includes("template"),
  );

console.log(`Found ${dbFiles.length} databases to fix:\n`);

let fixed = 0;
let errors = 0;

for (const dbFile of dbFiles) {
  const dbPath = path.join(dbDir, dbFile);
  const shmPath = `${dbPath}-shm`;
  const walPath = `${dbPath}-wal`;

  console.log(`\n📝 Processing: ${dbFile}`);

  try {
    // Open database and force checkpoint
    const db = new Database(dbPath);

    // Check current journal mode
    const journalMode = db.pragma("journal_mode", { simple: true });
    console.log(`   Journal mode: ${journalMode}`);

    if (journalMode === "wal") {
      // Force full checkpoint to merge WAL into main database
      const checkpoint = db.pragma("wal_checkpoint(TRUNCATE)");
      console.log(
        `   ✅ Checkpoint result: busy=${checkpoint[0].busy}, log=${checkpoint[0].log}, checkpointed=${checkpoint[0].checkpointed}`,
      );
    }

    // Ensure clean shutdown
    db.close();

    // Remove stale -shm and -wal files if they exist with 0 bytes or are now unnecessary
    if (fs.existsSync(shmPath)) {
      const shmSize = fs.statSync(shmPath).size;
      fs.unlinkSync(shmPath);
      console.log(`   🗑️  Removed ${dbFile}-shm (was ${shmSize} bytes)`);
    }

    if (fs.existsSync(walPath)) {
      const walSize = fs.statSync(walPath).size;
      fs.unlinkSync(walPath);
      console.log(`   🗑️  Removed ${dbFile}-wal (was ${walSize} bytes)`);
    }

    console.log(`   ✅ Fixed successfully`);
    fixed++;
  } catch (error) {
    console.log(`   ❌ Error: ${error.message}`);
    errors++;

    // If database is corrupt or locked, try to at least remove stale files
    try {
      if (fs.existsSync(shmPath)) {
        fs.unlinkSync(shmPath);
        console.log(`   🗑️  Force-removed ${dbFile}-shm`);
      }
      if (fs.existsSync(walPath)) {
        fs.unlinkSync(walPath);
        console.log(`   🗑️  Force-removed ${dbFile}-wal`);
      }
    } catch (cleanupError) {
      console.log(
        `   ⚠️  Could not clean up stale files: ${cleanupError.message}`,
      );
    }
  }
}

console.log("\n" + "=".repeat(50));
console.log(`\n✅ Fixed: ${fixed} databases`);
console.log(`❌ Errors: ${errors} databases`);
console.log("\n🚀 You can now restart the backend and n8n containers.");
console.log(
  "   docker compose -f docker-compose.n8n.production-full.yml up -d backend n8n",
);

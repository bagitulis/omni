import { getDbManager } from "../services/dbConnectionManager";
import Database from "better-sqlite3";

async function verifyLockedOrderTable() {
  try {
    console.log("🔍 Verifying LockedOrder table in all tenant databases...\n");

    const dbManager = getDbManager();
    const tenants = dbManager.getAllTenants();
    const processedDbs = new Set<string>();

    for (const tenantId of tenants) {
      const config = dbManager.getTenantConfig(tenantId);
      const dbPath = config.dbPath;

      if (processedDbs.has(dbPath.toLowerCase())) {
        console.log(`📌 ${tenantId}: Skipped (same database as another tenant)`);
        continue;
      }

      processedDbs.add(dbPath.toLowerCase());

      try {
        const db = new Database(dbPath);
        const tables = db
          .prepare(
            "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
          )
          .all() as Array<{ name: string }>;

        const tableNames = tables.map((t) => t.name);
        const hasLockedOrder = tableNames.includes("LockedOrder");

        console.log(`📋 ${tenantId} (${dbPath})`);
        console.log(`   Tables: ${tableNames.join(", ")}`);
        if (hasLockedOrder) {
          console.log(`   ✅ LockedOrder table EXISTS`);

          // Check table structure
          const columns = db
            .prepare("PRAGMA table_info(LockedOrder);")
            .all() as Array<{ name: string; type: string }>;
          console.log(`   Columns: ${columns.map((c) => c.name).join(", ")}`);
        } else {
          console.log(`   ❌ LockedOrder table NOT found`);
        }
        console.log("");

        db.close();
      } catch (error: any) {
        console.log(`❌ ${tenantId}: Error - ${error.message}\n`);
      }
    }

    console.log("✅ Verification complete!");
  } catch (error: any) {
    console.error("❌ Error:", error.message);
    process.exit(1);
  }
}

verifyLockedOrderTable();

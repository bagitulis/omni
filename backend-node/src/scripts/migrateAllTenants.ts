import { exec } from "child_process";
import { promisify } from "util";
import path from "path";
// import fs from "fs";
import { getDbManager } from "../services/dbConnectionManager";

const execAsync = promisify(exec);

async function migrateAllTenants() {
  try {
    console.log("🔄 Starting migration for all tenants...\n");

    const dbManager = getDbManager();
    const tenants = dbManager.getAllTenants();

    if (tenants.length === 0) {
      console.log("⚠️  No tenants found in configuration");
      return;
    }

    console.log(
      `📊 Found ${tenants.length} tenant(s): ${tenants.join(", ")}\n`
    );

    // Track which databases have been processed to skip duplicates (e.g., tester uses yumna's db)
    const processedDbs = new Set<string>();

    // Apply migration to each tenant database
    for (const tenantId of tenants) {
      try {
        console.log(`⏳ Migrating tenant: ${tenantId}`);
        const tenantConfig = dbManager.getTenantConfig(tenantId);

        // Get the database URL for this tenant (Windows compatible)
        const dbPath = path.resolve(tenantConfig.dbPath);
        const dbUrl = `file:${dbPath}`;

        // Check if this database has already been processed (handles tester using yumna's db)
        if (processedDbs.has(dbPath.toLowerCase())) {
          console.log(
            `  ℹ️  ${tenantId}: Skipped (already processed by another tenant)`
          );
          continue;
        }

        processedDbs.add(dbPath.toLowerCase());

        // Create env object with DATABASE_URL
        const env = {
          ...process.env,
          DATABASE_URL: dbUrl,
        };

        // SAFETY: We use 'migrate deploy' strictly.
        // We DO NOT use 'db push' because it bypasses migration history and risks data loss.
        // 'migrate deploy' ensures strict schema evolution based on committed migration files.

        // Optional: Resolve blocked migrations if known issues exist
        // (Only enable this if you are sure about specific stuck migrations)
        /*
        const failedMigrations = [
          "20251227064833_init",
          "20251228073142_add_worksheet_fields"
        ];
        // ... resolve logic ...
        */

        console.log(`  🚀 ${tenantId}: Deploying pending migrations...`);

        try {
          // Run Prisma migrate deploy
          // This command applies all pending migrations from prisma/migrations folder
          const { stdout, stderr } = await execAsync(
            `npx prisma migrate deploy`,
            {
              cwd: path.resolve(__dirname, "../.."),
              env,
            }
          );

          if (stdout) {
            // Check for "No pending migrations" message to reduce noise
            if (stdout.includes("No pending migrations")) {
              console.log(`  ✅ ${tenantId}: Already up to date`);
            } else {
              console.log(`  ✅ ${tenantId}: Migrations applied successfully`);
              // console.log(stdout); // Uncomment for verbose details
            }
          }

          if (stderr && !stderr.includes("warn")) {
            console.log(`  ⚠️  ${tenantId} stderr:`, stderr.trim());
          }
        } catch (error: any) {
          // Handle specific deployment errors
          const errorMsg = error.stderr || error.message || "Unknown error";

          if (errorMsg.includes("P3005")) {
            console.error(
              `  ❌ ${tenantId}: Database not clean. Non-empty database often requires manual baseline.`
            );
          } else {
            console.error(`  ❌ ${tenantId}: Migration failed!`);
            console.error(`     Error: ${errorMsg.substring(0, 200)}...`);
          }
        }
      } catch (error: any) {
        console.error(`❌ ${tenantId}: Error - ${error.message}`);
      }
    }

    console.log("\n✅ Migration completed for all tenants!");
  } catch (error: any) {
    console.error("❌ Error during migration:", error.message);
    process.exit(1);
  }
}

// Run the migration
migrateAllTenants();

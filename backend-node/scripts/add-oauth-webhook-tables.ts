/**
 * Migration Script: Add OAuth and Webhook Tables
 * Adds OAuthState and WebhookLog tables to all tenant databases
 *
 * Run: npx ts-node scripts/add-oauth-webhook-tables.ts
 */

import Database from "better-sqlite3";
import * as fs from "fs";
import * as path from "path";

const DB_DIR = path.join(__dirname, "..", "config", "databases");

// SQL to create OAuthState table
const CREATE_OAUTH_STATE = `
CREATE TABLE IF NOT EXISTS OAuthState (
  id TEXT PRIMARY KEY,
  tenantId TEXT NOT NULL,
  platform TEXT NOT NULL,
  state TEXT NOT NULL UNIQUE,
  redirectUrl TEXT,
  metadata TEXT,
  expiresAt DATETIME NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_oauth_tenant ON OAuthState(tenantId);
CREATE INDEX IF NOT EXISTS idx_oauth_platform ON OAuthState(platform);
CREATE INDEX IF NOT EXISTS idx_oauth_state ON OAuthState(state);
CREATE INDEX IF NOT EXISTS idx_oauth_expires ON OAuthState(expiresAt);
`;

// SQL to create WebhookLog table
const CREATE_WEBHOOK_LOG = `
CREATE TABLE IF NOT EXISTS WebhookLog (
  id TEXT PRIMARY KEY,
  tenantId TEXT,
  platform TEXT NOT NULL,
  eventType TEXT NOT NULL,
  payload TEXT NOT NULL,
  headers TEXT,
  status TEXT DEFAULT 'received',
  errorMsg TEXT,
  processedAt DATETIME,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_webhook_tenant ON WebhookLog(tenantId);
CREATE INDEX IF NOT EXISTS idx_webhook_platform ON WebhookLog(platform);
CREATE INDEX IF NOT EXISTS idx_webhook_event ON WebhookLog(eventType);
CREATE INDEX IF NOT EXISTS idx_webhook_status ON WebhookLog(status);
CREATE INDEX IF NOT EXISTS idx_webhook_created ON WebhookLog(createdAt);
`;

function migrateTables(): void {
  console.log("🔄 Starting OAuth & Webhook tables migration...\n");

  // Get all .db files (excluding jobs databases)
  const dbFiles = fs
    .readdirSync(DB_DIR)
    .filter(
      (f) =>
        f.endsWith(".db") &&
        !f.includes("_jobs") &&
        !f.includes("migration_template")
    );

  console.log(`Found ${dbFiles.length} tenant databases to migrate:\n`);

  let successCount = 0;
  let errorCount = 0;

  for (const dbFile of dbFiles) {
    const dbPath = path.join(DB_DIR, dbFile);
    console.log(`📦 Migrating: ${dbFile}`);

    try {
      const db = new Database(dbPath);

      // Create OAuthState table
      db.exec(CREATE_OAUTH_STATE);
      console.log(`   ✅ OAuthState table created/verified`);

      // Create WebhookLog table
      db.exec(CREATE_WEBHOOK_LOG);
      console.log(`   ✅ WebhookLog table created/verified`);

      db.close();
      successCount++;
    } catch (error: any) {
      console.log(`   ❌ Error: ${error.message}`);
      errorCount++;
    }
  }

  console.log("\n" + "=".repeat(50));
  console.log(`✅ Successfully migrated: ${successCount} databases`);
  if (errorCount > 0) {
    console.log(`❌ Failed: ${errorCount} databases`);
  }
  console.log("=".repeat(50));
}

// Run migration
migrateTables();

/**
 * Migration Script: Update tenantId from old format to new format
 *
 * Old format: 'yumna', 'tika'
 * New format: 'yumna_bertigamart', 'tika_nusseyba'
 *
 * This is a NON-DESTRUCTIVE migration that updates tenantId field in all tables
 */

const Database = require("better-sqlite3");

const migrations = [
  {
    dbFile: "config/databases/yumna_bertigamart.db",
    oldTenantId: "yumna",
    newTenantId: "yumna_bertigamart",
  },
  {
    dbFile: "config/databases/tika_nusseyba.db",
    oldTenantId: "tika",
    newTenantId: "tika_nusseyba",
  },
];

// Tables that have tenantId column
const tablesWithTenantId = [
  "InventoryRecord",
  "InventorySyncHistory",
  "InventorySettings",
  "ShopeeOrder",
  "ShopeeOrderItem",
  "LazadaOrder",
  "LazadaOrderItem",
  "TiktokOrder",
  "TiktokOrderItem",
  "ShopeeProduct",
  "ShopeeSku",
  "LazadaProduct",
  "LazadaSku",
  "TiktokProduct",
  "TiktokSku",
  "GoogleSheetsSettings",
  "FilterPreference",
  "RouteConfig",
  "LockedOrder",
  "SheetSnapshot",
  "Spreadsheet",
  "WholesaleSettings",
  "AnalyticsSettings",
  "ShopeeEscrowSync",
  "ShopeeEscrowOrder",
  "ShopeeEscrowItem",
  "TiktokEscrowSync",
  "TiktokEscrowOrder",
  "TiktokEscrowItem",
  "OAuthState",
  "OAuthLog",
  "WebhookLog",
  "WebhookOrderEvent",
  "WebhookProductEvent",
  "WebhookMarketingEvent",
  "WebhookReturnEvent",
  "WebhookShopeeEvent",
  "WebhookWebchatEvent",
  "WebhookFBSEvent",
  "OrderTodayItem",
  "InventorySkuPlatformStatus",
];

function migrateTenantId(migration) {
  const { dbFile, oldTenantId, newTenantId } = migration;

  console.log(`\n${"=".repeat(60)}`);
  console.log(`📦 Migrating: ${dbFile}`);
  console.log(
    `   Old tenantId: '${oldTenantId}' → New tenantId: '${newTenantId}'`
  );
  console.log("=".repeat(60));

  const db = new Database(dbFile);

  let totalUpdated = 0;

  for (const table of tablesWithTenantId) {
    try {
      // Check if table exists
      const tableExists = db
        .prepare(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`)
        .get(table);
      if (!tableExists) {
        continue;
      }

      // Check if table has tenantId column
      const columns = db.prepare(`PRAGMA table_info(${table})`).all();
      const hasTenantId = columns.some((col) => col.name === "tenantId");
      if (!hasTenantId) {
        continue;
      }

      // Count records to update
      const countBefore = db
        .prepare(`SELECT COUNT(*) as count FROM ${table} WHERE tenantId = ?`)
        .get(oldTenantId);

      if (countBefore.count > 0) {
        // Update tenantId
        const result = db
          .prepare(`UPDATE ${table} SET tenantId = ? WHERE tenantId = ?`)
          .run(newTenantId, oldTenantId);
        console.log(`   ✅ ${table}: ${result.changes} records updated`);
        totalUpdated += result.changes;
      }
    } catch (error) {
      console.log(`   ⚠️ ${table}: ${error.message}`);
    }
  }

  console.log(`\n   📊 Total records updated: ${totalUpdated}`);

  db.close();
  return totalUpdated;
}

console.log("🚀 Starting tenantId migration...\n");

let grandTotal = 0;
for (const migration of migrations) {
  grandTotal += migrateTenantId(migration);
}

console.log(`\n${"=".repeat(60)}`);
console.log(`✅ Migration complete! Total records updated: ${grandTotal}`);
console.log("=".repeat(60));

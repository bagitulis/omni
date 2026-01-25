import { PrismaClient } from "@prisma/client";
import Database from "better-sqlite3";
import path from "path";

/**
 * Migrate data from shared dev.db to tenant-specific databases
 * Currently only yumna has data in dev.db, so it gets migrated to yumna's database
 */

async function migrateDevDbData() {
  console.log(
    "🚀 Starting data migration from dev.db to tenant databases...\n"
  );

  // Read from shared dev.db
  const devDbPath = path.resolve("prisma/dev.db");
  const devDb = new Database(devDbPath);

  // Get yumna's Prisma client
  const yumnaPrisma = new PrismaClient({
    datasources: {
      db: {
        url: `file:${path.resolve("config/databases/yumna_bertigamart.db")}`,
      },
    },
  });

  try {
    // Get yumna user ID from yumna database (already seeded)
    const yumnaUser = await yumnaPrisma.user.findUnique({
      where: { username: "yumna" },
    });

    if (!yumnaUser) {
      throw new Error("yumna user not found in yumna database");
    }

    console.log(`✅ Found yumna user: ${yumnaUser.id}\n`);

    // 1. Migrate PlatformConfig
    console.log("📦 Migrating PlatformConfig...");
    const platformConfigs = devDb
      .prepare("SELECT * FROM PlatformConfig")
      .all() as any[];

    for (const config of platformConfigs) {
      const exists = await yumnaPrisma.platformConfig.findUnique({
        where: { id: config.id },
      });

      if (!exists) {
        await yumnaPrisma.platformConfig.create({
          data: {
            id: config.id,
            platform: config.platform,
            configKey: config.configKey,
            configValue: config.configValue,
            dataType: config.dataType,
            isEncrypted: Boolean(config.isEncrypted),
            metadata: config.metadata,
            createdAt: new Date(config.createdAt),
            updatedAt: new Date(config.updatedAt),
          },
        });
      }
    }
    console.log(
      `   ✅ Migrated ${platformConfigs.length} PlatformConfig records\n`
    );

    // 2. Migrate InventorySettings
    console.log("⚙️  Migrating InventorySettings...");
    const invSettings = devDb
      .prepare("SELECT * FROM InventorySettings")
      .all() as any[];

    for (const setting of invSettings) {
      const exists = await yumnaPrisma.inventorySettings.findUnique({
        where: { id: setting.id },
      });

      if (!exists) {
        await yumnaPrisma.inventorySettings.create({
          data: {
            id: setting.id,
            spreadsheetId: setting.spreadsheetId,
            sheetName: setting.sheetName,
            selectedColumns: setting.selectedColumns,
            headerRow: setting.headerRow,
            dataStartRow: setting.dataStartRow,
            keyColumn: setting.keyColumn,
            autoSync: Boolean(setting.autoSync),
            syncIntervalSeconds: setting.syncIntervalSeconds,
            lastSyncTimestamp: new Date(setting.lastSyncTimestamp),
            allColumns: setting.allColumns,
            lastHeadersHash: setting.lastHeadersHash,
            lastSyncStatus: setting.lastSyncStatus,
            createdAt: new Date(setting.createdAt),
            updatedAt: new Date(setting.updatedAt),
          },
        });
      }
    }
    console.log(
      `   ✅ Migrated ${invSettings.length} InventorySettings records\n`
    );

    // 3. Migrate FilterPreference
    console.log("🔍 Migrating FilterPreference...");
    const filters = devDb
      .prepare("SELECT * FROM FilterPreference")
      .all() as any[];

    for (const filter of filters) {
      const exists = await yumnaPrisma.filterPreference.findUnique({
        where: { id: filter.id },
      });

      if (!exists) {
        await yumnaPrisma.filterPreference.create({
          data: {
            id: filter.id,
            platform: filter.platform,
            tab: filter.tab,
            visibleColumns: filter.visibleColumns,
            columnFilters: filter.columnFilters,
            searchQuery: filter.searchQuery,
            lockedColumns: filter.lockedColumns,
            createdAt: new Date(filter.createdAt),
            updatedAt: new Date(filter.updatedAt),
          },
        });
      }
    }
    console.log(`   ✅ Migrated ${filters.length} FilterPreference records\n`);

    console.log("✅ Data migration complete!");
    console.log("\n📊 Migration Summary:");
    console.log(`   PlatformConfig: ${platformConfigs.length}`);
    console.log(`   InventorySettings: ${invSettings.length}`);
    console.log(`   FilterPreference: ${filters.length}`);
    console.log(
      `   Total: ${
        platformConfigs.length + invSettings.length + filters.length
      }`
    );
  } catch (error) {
    console.error("❌ Migration failed:", error);
    process.exit(1);
  } finally {
    devDb.close();
    await yumnaPrisma.$disconnect();
  }
}

migrateDevDbData();

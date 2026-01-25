const Database = require("better-sqlite3");
const fs = require("fs");

// Connect to yumna database
const db = new Database("./config/databases/yumna_bertigamart.db");

// Get all PlatformConfig with values
const rows = db
  .prepare(
    `
  SELECT id, platform, configKey, configValue, dataType, isEncrypted 
  FROM PlatformConfig 
  WHERE configValue != '' AND configValue IS NOT NULL
`,
  )
  .all();

// Generate SQL
let sql = "-- Migration: yumna_bertigamart PlatformConfig\n";
sql += "-- Run this in PostgreSQL\n\n";

rows.forEach((row) => {
  const escapedValue = (row.configValue || "").replace(/'/g, "''");
  sql += `INSERT INTO tenant_yumna_bertigamart.platform_configs (id, platform, config_key, config_value, data_type, is_encrypted, created_at, updated_at)\n`;
  sql += `VALUES ('${row.id}', '${row.platform}', '${
    row.configKey
  }', '${escapedValue}', '${row.dataType || "string"}', ${
    row.isEncrypted ? "true" : "false"
  }, NOW(), NOW())\n`;
  sql += `ON CONFLICT (id) DO NOTHING;\n\n`;
});

sql +=
  "\n-- Verify\nSELECT platform, config_key, SUBSTRING(config_value, 1, 30) as value_preview FROM tenant_yumna_bertigamart.platform_configs ORDER BY platform, config_key;\n";

fs.writeFileSync("../scripts/migrate_yumna_platform.sql", sql);
console.log(`Generated ${rows.length} INSERT statements`);

db.close();

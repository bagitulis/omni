require('dotenv').config({ path: '.env' });
const sqlite3 = require('sqlite3').verbose();
const path = require('path');

function checkDatabase(dbPath) {
  return new Promise((resolve) => {
    console.log(`\n📁 Checking database: ${dbPath}`);
    console.log('='.repeat(80));
    
    const db = new sqlite3.Database(dbPath, (err) => {
      if (err) {
        console.log(`❌ Cannot open database: ${err.message}`);
        resolve();
        return;
      }

      // Get all tables
      db.all(
        "SELECT name FROM sqlite_master WHERE type='table'",
        (err, tables) => {
          if (err) {
            console.log(`❌ Error querying tables: ${err.message}`);
            db.close();
            resolve();
            return;
          }

          console.log(`Found ${tables.length} tables:`);
          tables.forEach(t => console.log(`  - ${t.name}`));

          // Check if PlatformConfig exists
          const hasConfig = tables.some(t => t.name.toLowerCase() === 'platformconfig');
          
          if (hasConfig) {
            db.all('SELECT * FROM PlatformConfig WHERE platform = "tiktok"', (err, rows) => {
              if (err) {
                console.log(`❌ Error querying PlatformConfig: ${err.message}`);
              } else {
                if (rows && rows.length > 0) {
                  console.log(`\n✅ Found ${rows.length} TikTok configs`);
                  rows.forEach(row => {
                    console.log(`  Key: ${row.configKey}`);
                    console.log(`  Encrypted: ${row.isEncrypted}`);
                    console.log(`  Value: ${row.configValue.substring(0, 50)}...`);
                  });
                } else {
                  console.log('⚠️  No TikTok configs found in PlatformConfig');
                }
              }
              db.close();
              resolve();
            });
          } else {
            db.close();
            resolve();
          }
        }
      );
    });
  });
}

async function main() {
  const devDb = path.join(__dirname, 'dev.db');
  const omniDb = path.join(__dirname, 'omni.db');

  await checkDatabase(devDb);
  await checkDatabase(omniDb);
}

main();

/**
 * Detailed database check - see all orders by platform and status
 */

const sqlite3 = require("sqlite3").verbose();
const path = require("path");

const dbPath = path.join(
  __dirname,
  "config/databases/yumna_bertigamart.db"
);

async function detailedCheck() {
  return new Promise((resolve, reject) => {
    const db = new sqlite3.Database(dbPath, (err) => {
      if (err) reject(err);

      console.log("\n📊 DETAILED ORDER COUNT BY PLATFORM & STATUS:");
      console.log("=".repeat(70));

      // Shopee
      db.all(
        `SELECT orderStatus, COUNT(*) as count FROM ShopeeOrder WHERE tenantId = 'yumna' GROUP BY orderStatus`,
        (err, rows) => {
          if (err) {
            console.log(`❌ ShopeeOrder error: ${err.message}`);
          } else {
            console.log("\n🛍️  SHOPEE:");
            let total = 0;
            rows.forEach((r) => {
              console.log(`  - ${r.orderStatus}: ${r.count}`);
              total += r.count;
            });
            console.log(`  TOTAL: ${total}`);
          }

          // Lazada
          db.all(
            `SELECT orderStatus, COUNT(*) as count FROM LazadaOrder WHERE tenantId = 'yumna' GROUP BY orderStatus`,
            (err, rows) => {
              if (err) {
                console.log(`❌ LazadaOrder error: ${err.message}`);
              } else {
                console.log("\n📦 LAZADA:");
                let total = 0;
                rows.forEach((r) => {
                  console.log(`  - ${r.orderStatus}: ${r.count}`);
                  total += r.count;
                });
                console.log(`  TOTAL: ${total}`);
              }

              // TikTok
              db.all(
                `SELECT orderStatus, COUNT(*) as count FROM TiktokOrder WHERE tenantId = 'yumna' GROUP BY orderStatus`,
                (err, rows) => {
                  if (err) {
                    console.log(`❌ TiktokOrder error: ${err.message}`);
                  } else {
                    console.log("\n🎵 TIKTOK:");
                    let total = 0;
                    rows.forEach((r) => {
                      console.log(`  - ${r.orderStatus}: ${r.count}`);
                      total += r.count;
                    });
                    console.log(`  TOTAL: ${total}`);
                  }

                  // Summary - count PROCESSED across all platforms
                  db.all(
                    `
                    SELECT 'SHOPEE' as platform, COUNT(*) as count FROM ShopeeOrder WHERE tenantId = 'yumna' AND orderStatus = 'PROCESSED'
                    UNION ALL
                    SELECT 'LAZADA' as platform, COUNT(*) as count FROM LazadaOrder WHERE tenantId = 'yumna' AND orderStatus = 'toship'
                    UNION ALL
                    SELECT 'TIKTOK' as platform, COUNT(*) as count FROM TiktokOrder WHERE tenantId = 'yumna' AND orderStatus = 'AWAITING_COLLECTION'
                    `,
                    (err, rows) => {
                      if (err) {
                        console.log(`❌ PROCESSED count error: ${err.message}`);
                      } else {
                        console.log("\n📋 PROCESSED STATUS ACROSS ALL PLATFORMS:");
                        console.log("=".repeat(70));
                        let totalProcessed = 0;
                        rows.forEach((r) => {
                          console.log(`  ${r.platform}: ${r.count}`);
                          totalProcessed += r.count;
                        });
                        console.log(`\n  TOTAL PROCESSED: ${totalProcessed}`);
                      }

                      db.close(() => {
                        resolve();
                      });
                    }
                  );
                }
              );
            }
          );
        }
      );
    });
  });
}

detailedCheck().catch((err) => {
  console.error("❌ Error:", err);
  process.exit(1);
});

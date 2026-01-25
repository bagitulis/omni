/**
 * Test Order Count - Compare API vs Database
 * Check if API count matches actual database state
 */

const sqlite3 = require("sqlite3").verbose();
const path = require("path");

const dbPath = path.join(
  __dirname,
  "config/databases/yumna_bertigamart.db"
);

// Direct database query
async function testDatabase() {
  return new Promise((resolve, reject) => {
    const db = new sqlite3.Database(dbPath, (err) => {
      if (err) reject(err);

      console.log("\n📊 DIRECT DATABASE QUERY:");
      console.log("=".repeat(60));

      // Count total ShopeeOrder
      db.get(
        `SELECT COUNT(*) as total FROM ShopeeOrder WHERE tenantId = 'yumna'`,
        (err, row) => {
          if (err) reject(err);
          console.log(`✓ Total ShopeeOrder records: ${row.total}`);

          // Count by status
          db.all(
            `SELECT orderStatus, COUNT(*) as count FROM ShopeeOrder WHERE tenantId = 'yumna' GROUP BY orderStatus`,
            (err, rows) => {
              if (err) reject(err);
              console.log("\nBy status:");
              rows.forEach((r) => {
                console.log(`  - ${r.orderStatus}: ${r.count}`);
              });

              // Count items
              db.get(
                `SELECT COUNT(*) as total FROM ShopeeOrderItem WHERE tenantId = 'yumna'`,
                (err, row2) => {
                  if (err) {
                    console.log(
                      `\n⚠️  ShopeeOrderItem table error (might not exist): ${err.message}`
                    );
                    row2 = { total: 0 };
                  }
                  console.log(
                    `\n✓ Total ShopeeOrderItem records: ${row2?.total || 0}`
                  );

                  // Get distinct order_sn count
                  db.get(
                    `SELECT COUNT(DISTINCT orderSn) as unique_orders FROM ShopeeOrder WHERE tenantId = 'yumna'`,
                    (err, row3) => {
                      if (err) reject(err);
                      console.log(
                        `✓ Unique order_sn count: ${row3.unique_orders}`
                      );

                      // Get PROCESSED orders
                      db.get(
                        `SELECT COUNT(*) as total FROM ShopeeOrder WHERE tenantId = 'yumna' AND orderStatus = 'PROCESSED'`,
                        (err, row4) => {
                          if (err) reject(err);
                          console.log(
                            `\n✓ PROCESSED orders only: ${row4.total}`
                          );

                          db.close(() => {
                            resolve({
                              total: row.total,
                              processed: row4.total,
                              items: row2.total,
                              unique: row3.unique_orders,
                            });
                          });
                        }
                      );
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

// Test API endpoint
async function testAPI() {
  console.log("\n📡 API ENDPOINT TEST:");
  console.log("=".repeat(60));

  try {
    const response = await fetch("http://localhost:3001/api/orders/processed");
    const data = await response.json();

    console.log(`✓ Endpoint: GET /api/orders/processed`);
    console.log(`✓ Success: ${data.success}`);
    console.log(`✓ Returned count: ${data.count}`);
    console.log(`✓ Actual items returned: ${data.items?.length || 0}`);

    // Count unique order_no in items
    if (data.items && data.items.length > 0) {
      const uniqueOrderNos = new Set(
        data.items.map((item) => item.order_no)
      ).size;
      console.log(`✓ Unique order_no in items: ${uniqueOrderNos}`);

      // Show first few items
      console.log("\nFirst 3 items:");
      data.items.slice(0, 3).forEach((item, i) => {
        console.log(
          `  ${i + 1}. ${item.order_no} (${item.product_name}) - Qty: ${item.qty}`
        );
      });
    }

    return {
      count: data.count,
      itemsReturned: data.items?.length || 0,
    };
  } catch (error) {
    console.error("❌ API Error:", error.message);
    throw error;
  }
}

// Compare results
async function compare() {
  console.log("\n\n🔍 COMPARISON:");
  console.log("=".repeat(60));

  const dbResult = await testDatabase();
  const apiResult = await testAPI();

  console.log("\n📋 SUMMARY:");
  console.log(`Database - PROCESSED orders: ${dbResult.processed}`);
  console.log(`API - Returned count: ${apiResult.count}`);
  console.log(`API - Items returned: ${apiResult.itemsReturned}`);

  if (apiResult.count !== dbResult.processed) {
    console.log(
      `\n⚠️  MISMATCH! API count (${apiResult.count}) != DB PROCESSED (${dbResult.processed})`
    );
  } else {
    console.log(`\n✅ MATCH! API count = DB PROCESSED count`);
  }

  process.exit(0);
}

compare().catch((err) => {
  console.error("❌ Test failed:", err);
  process.exit(1);
});

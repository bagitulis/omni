const Database = require("better-sqlite3");

const db = new Database("/app/config/databases/yumna_bertigamart.db");

console.log("🔍 Checking Shopee orders in database...\n");

// Get tables
const tables = db
  .prepare("SELECT name FROM sqlite_master WHERE type='table'")
  .all();
console.log(
  "Tables:",
  tables.map((t) => t.name)
);

// Check ShopeeOrder table
try {
  const ordersCount = db
    .prepare(
      "SELECT COUNT(*) as count FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'"
    )
    .get();
  console.log("\n📦 PROCESSED Shopee orders:", ordersCount.count);

  // Count unique orderSn
  const uniqueCount = db
    .prepare(
      "SELECT COUNT(DISTINCT orderSn) as count FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'"
    )
    .get();
  console.log("🔢 Unique orderSn:", uniqueCount.count);

  // Count total items
  const itemsCount = db
    .prepare(
      "SELECT COUNT(*) as count FROM ShopeeOrderItem WHERE orderSn IN (SELECT orderSn FROM ShopeeOrder WHERE orderStatus = 'PROCESSED')"
    )
    .get();
  console.log("📋 Total items:", itemsCount.count);

  // Find orders with multiple items
  const multiItemOrders = db
    .prepare(
      "SELECT soi.orderSn, COUNT(*) as itemCount FROM ShopeeOrderItem soi JOIN ShopeeOrder so ON so.orderSn = soi.orderSn WHERE so.orderStatus = 'PROCESSED' GROUP BY soi.orderSn HAVING COUNT(*) > 1"
    )
    .all();

  console.log("\n📊 Orders with multiple items:", multiItemOrders.length);

  if (multiItemOrders.length > 0) {
    console.log("\n🔍 Multi-item orders:");
    for (const order of multiItemOrders) {
      console.log("  -", order.orderSn + ":", order.itemCount, "items");
    }
  }

  // Summary
  console.log("\n============================================================");
  console.log("📊 SUMMARY:");
  console.log("  - Orders in DB:", ordersCount.count);
  console.log("  - Total items (flattened rows):", itemsCount.count);
  console.log("  - Unique order_sn (UI count):", uniqueCount.count);
  console.log("  - Multi-item orders:", multiItemOrders.length);
  console.log(
    "  - Extra rows from multi-items:",
    itemsCount.count - ordersCount.count
  );
} catch (err) {
  console.error("Error:", err.message);
}

db.close();

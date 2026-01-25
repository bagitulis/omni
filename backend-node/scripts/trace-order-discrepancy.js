const Database = require("better-sqlite3");

const db = new Database("/app/config/databases/yumna_bertigamart.db");

console.log("🔍 Tracing order count discrepancy...\n");

try {
  // Get all unique orderSn
  const allOrders = db
    .prepare(
      "SELECT orderSn, orderStatus, createdAt, updatedAt FROM ShopeeOrder ORDER BY createdAt DESC"
    )
    .all();
  console.log("📦 Total Shopee orders in DB:", allOrders.length);

  // Check for duplicate orderSn
  const orderSnSet = new Set();
  const duplicates = [];
  for (const order of allOrders) {
    if (orderSnSet.has(order.orderSn)) {
      duplicates.push(order.orderSn);
    }
    orderSnSet.add(order.orderSn);
  }

  if (duplicates.length > 0) {
    console.log("\n⚠️  DUPLICATE orderSn found:", duplicates);
  } else {
    console.log("✅ No duplicate orderSn");
  }

  // Count by status
  const statusCounts = db
    .prepare(
      "SELECT orderStatus, COUNT(*) as count FROM ShopeeOrder GROUP BY orderStatus"
    )
    .all();
  console.log("\n📊 Orders by status:");
  let total = 0;
  for (const status of statusCounts) {
    console.log("  -", status.orderStatus + ":", status.count);
    total += status.count;
  }
  console.log("  TOTAL:", total);

  // Check orders without items
  const ordersWithoutItems = db
    .prepare(
      `
    SELECT so.orderSn, so.orderStatus 
    FROM ShopeeOrder so 
    LEFT JOIN ShopeeOrderItem soi ON so.orderSn = soi.orderSn 
    WHERE soi.orderSn IS NULL
  `
    )
    .all();

  console.log("\n📋 Orders WITHOUT items:", ordersWithoutItems.length);
  if (ordersWithoutItems.length > 0 && ordersWithoutItems.length <= 10) {
    for (const o of ordersWithoutItems) {
      console.log("  -", o.orderSn, "(", o.orderStatus, ")");
    }
  }

  // Check last 5 PROCESSED orders
  console.log("\n📦 Last 5 PROCESSED orders:");
  const lastProcessed = db
    .prepare(
      "SELECT orderSn, orderStatus, createdAt FROM ShopeeOrder WHERE orderStatus = 'PROCESSED' ORDER BY createdAt DESC LIMIT 5"
    )
    .all();
  for (const o of lastProcessed) {
    console.log("  -", o.orderSn, "| created:", o.createdAt);
  }

  // Check if any order has changed status recently (via webhook)
  console.log(
    "\n🔔 Orders with non-PROCESSED status (may have changed via webhook):"
  );
  const nonProcessed = db
    .prepare(
      "SELECT orderSn, orderStatus, updatedAt FROM ShopeeOrder WHERE orderStatus != 'PROCESSED' ORDER BY updatedAt DESC"
    )
    .all();
  for (const o of nonProcessed) {
    console.log(
      "  -",
      o.orderSn,
      "|",
      o.orderStatus,
      "| updated:",
      o.updatedAt
    );
  }
} catch (err) {
  console.error("Error:", err.message);
}

db.close();

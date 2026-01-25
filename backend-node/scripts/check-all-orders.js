const Database = require("better-sqlite3");

const db = new Database("/app/config/databases/yumna_bertigamart.db");

console.log("🔍 Checking ALL orders in database...\n");

// Get order-related tables
const tables = db
  .prepare("SELECT name FROM sqlite_master WHERE type='table'")
  .all();
console.log(
  "Order Tables:",
  tables
    .map((t) => t.name)
    .filter((t) => t.includes("Order") || t.includes("Item"))
);

try {
  // Check all Shopee order statuses
  const allStatuses = db
    .prepare(
      "SELECT orderStatus, COUNT(*) as count FROM ShopeeOrder GROUP BY orderStatus"
    )
    .all();
  console.log("\n📊 All Shopee order statuses:");
  let totalOrders = 0;
  for (const status of allStatuses) {
    console.log("  -", status.orderStatus + ":", status.count);
    totalOrders += status.count;
  }
  console.log("  TOTAL Shopee:", totalOrders);

  // Check PROCESSED specifically
  const processedCount = db
    .prepare(
      "SELECT COUNT(*) as count FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'"
    )
    .get();
  console.log("\n📦 PROCESSED Shopee orders:", processedCount.count);

  // Count items for PROCESSED
  const itemsCount = db
    .prepare(
      "SELECT COUNT(*) as count FROM ShopeeOrderItem WHERE orderSn IN (SELECT orderSn FROM ShopeeOrder WHERE orderStatus = 'PROCESSED')"
    )
    .get();
  console.log("📋 Total items (PROCESSED):", itemsCount.count);

  // Multi-item orders
  const multiItemOrders = db
    .prepare(
      "SELECT soi.orderSn, COUNT(*) as itemCount FROM ShopeeOrderItem soi JOIN ShopeeOrder so ON so.orderSn = soi.orderSn WHERE so.orderStatus = 'PROCESSED' GROUP BY soi.orderSn HAVING COUNT(*) > 1"
    )
    .all();
  console.log("📊 Multi-item orders:", multiItemOrders.length);

  // Check TikTok orders
  const tiktokStatuses = db
    .prepare(
      "SELECT orderStatus, COUNT(*) as count FROM TiktokOrder GROUP BY orderStatus"
    )
    .all();
  console.log("\n📊 TikTok order statuses:");
  for (const status of tiktokStatuses) {
    console.log("  -", status.orderStatus + ":", status.count);
  }

  // Check Lazada orders
  const lazadaStatuses = db
    .prepare(
      "SELECT orderStatus, COUNT(*) as count FROM LazadaOrder GROUP BY orderStatus"
    )
    .all();
  console.log("\n📊 Lazada order statuses:");
  if (lazadaStatuses.length === 0) {
    console.log("  (no orders)");
  }
  for (const status of lazadaStatuses) {
    console.log("  -", status.orderStatus + ":", status.count);
  }

  // Summary
  console.log("\n============================================================");
  console.log("📊 FINAL SUMMARY:");
  console.log("");
  console.log("Shopee:");
  console.log("  - Total orders in DB:", totalOrders);
  console.log("  - PROCESSED orders:", processedCount.count);
  console.log("  - PROCESSED items (flattened):", itemsCount.count);
  console.log("  - Multi-item orders:", multiItemOrders.length);
  console.log("");
  console.log("Expected UI display:");
  console.log("  - Shopee count (unique orders):", processedCount.count);
  console.log(
    "  - TikTok count:",
    tiktokStatuses.find((s) => s.orderStatus === "AWAITING_COLLECTION")
      ?.count || 0
  );
  console.log("  - Lazada count:", 0);
  console.log("");
  console.log('If backend log shows "68 orders fetched" but DB has 67,');
  console.log(
    "the missing order may have been filtered during save or status changed."
  );
} catch (err) {
  console.error("Error:", err.message);
}

db.close();

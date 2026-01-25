/**
 * Compare Shopee API orders vs Database using backend services
 * Must be run inside the backend context with proper config
 */

const path = require("path");
const Database = require("better-sqlite3");

// Load config from tenant database
function loadShopeeConfig(db) {
  try {
    const config = db
      .prepare(
        `
      SELECT * FROM PlatformConfig 
      WHERE platform = 'shopee' 
      LIMIT 1
    `
      )
      .get();

    if (!config) {
      console.log("❌ No Shopee config found in database");
      return null;
    }

    return {
      partnerId: config.partnerId,
      partnerKey: config.partnerKey,
      accessToken: config.accessToken,
      shopId: config.shopId,
    };
  } catch (err) {
    console.error("Error loading config:", err.message);
    return null;
  }
}

async function main() {
  console.log("🔍 Comparing Shopee API orders vs Database...\n");

  const dbPath = "/app/config/databases/yumna_bertigamart.db";
  const db = new Database(dbPath);

  // Get all orders grouped by status
  const allOrders = db
    .prepare(
      `
    SELECT orderSn, orderStatus, createdAt, updatedAt 
    FROM ShopeeOrder 
    ORDER BY createdAt DESC
  `
    )
    .all();

  console.log("📦 Total orders in DB:", allOrders.length);

  // Group by status
  const byStatus = {};
  for (const o of allOrders) {
    if (!byStatus[o.orderStatus]) {
      byStatus[o.orderStatus] = [];
    }
    byStatus[o.orderStatus].push(o.orderSn);
  }

  console.log("\n📊 Orders by status:");
  for (const [status, orders] of Object.entries(byStatus)) {
    console.log(`  - ${status}: ${orders.length}`);
  }

  // List PROCESSED orders
  const processedOrders = byStatus["PROCESSED"] || [];
  console.log("\n📋 PROCESSED orders (67 in DB):");
  console.log("   First 5:", processedOrders.slice(0, 5).join(", "));
  console.log("   Last 5:", processedOrders.slice(-5).join(", "));

  // Check for orders that might have status changed
  console.log(
    "\n🔍 Checking non-PROCESSED orders (potential moved from PROCESSED):"
  );

  const nonProcessed = allOrders.filter((o) => o.orderStatus !== "PROCESSED");
  for (const order of nonProcessed) {
    console.log(
      `  - ${order.orderSn} | ${order.orderStatus} | updated: ${
        order.updatedAt || "N/A"
      }`
    );
  }

  // Check if there are 68 unique order_sn total (API fetched 68)
  const uniqueOrderSns = new Set(allOrders.map((o) => o.orderSn));
  console.log("\n📊 Unique order_sn in DB:", uniqueOrderSns.size);

  // Check WebhookOrderEvent for any order updates
  const webhookEvents = db
    .prepare(
      `
    SELECT orderSn, eventType, fulfillmentStatus, newStatus, createdAt 
    FROM WebhookOrderEvent 
    ORDER BY createdAt DESC
  `
    )
    .all();

  console.log("\n🔔 Webhook events received:", webhookEvents.length);
  for (const evt of webhookEvents) {
    console.log(
      `  - ${evt.orderSn} | ${evt.eventType} | ${
        evt.fulfillmentStatus || evt.newStatus || ""
      }`
    );
  }

  // Cross-reference: check if any PROCESSED order has webhook update
  console.log("\n🔗 Cross-reference: PROCESSED orders with webhook events:");
  const processedSet = new Set(processedOrders);
  const webhookOrderSns = new Set(webhookEvents.map((e) => e.orderSn));

  const processedWithWebhook = [...processedSet].filter((sn) =>
    webhookOrderSns.has(sn)
  );
  console.log(
    `  - PROCESSED orders with webhook events: ${processedWithWebhook.length}`
  );
  for (const sn of processedWithWebhook) {
    console.log(`    • ${sn}`);
  }

  // The missing order analysis
  console.log("\n============================================================");
  console.log("📊 ANALYSIS:");
  console.log("");
  console.log('Backend log showed: "Fetched 68 items" from Shopee API');
  console.log("Database has: 67 PROCESSED orders");
  console.log("");
  console.log("Possible reasons for 1 missing order:");
  console.log("1. Order status changed between fetch and save");
  console.log("2. Order was deduplicated during save");
  console.log("3. Order failed to save silently");
  console.log("");

  // Check for any order_sn that appears multiple times
  const orderSnCounts = {};
  for (const o of allOrders) {
    orderSnCounts[o.orderSn] = (orderSnCounts[o.orderSn] || 0) + 1;
  }
  const duplicates = Object.entries(orderSnCounts).filter(
    ([sn, count]) => count > 1
  );
  console.log(
    "Duplicate order_sn in DB:",
    duplicates.length > 0 ? duplicates : "None"
  );

  db.close();
}

main().catch(console.error);

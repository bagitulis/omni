/**
 * Trigger sync for PROCESSED orders and verify database update
 */

const http = require("http");
const Database = require("better-sqlite3");

const DB_PATH = "/app/config/databases/yumna_bertigamart.db";

// Check database before sync
function checkDbBefore() {
  const db = new Database(DB_PATH);
  const count = db
    .prepare(
      "SELECT COUNT(*) as count FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'"
    )
    .get();
  db.close();
  return count.count;
}

// Check database after sync
function checkDbAfter() {
  const db = new Database(DB_PATH);
  const orders = db
    .prepare("SELECT orderSn FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'")
    .all();
  db.close();
  return orders.map((o) => o.orderSn);
}

// Trigger sync via internal API
function triggerSync() {
  return new Promise((resolve, reject) => {
    // Call internal sync endpoint - correct path
    const postData = JSON.stringify({ days: 7 });

    const options = {
      hostname: "127.0.0.1",
      port: 3000,
      path: "/api/orders/sync/processed", // POST /api/orders/sync/:category
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": "yumna",
        "Content-Length": Buffer.byteLength(postData),
      },
    };

    const req = http.request(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          resolve({ raw: data });
        }
      });
    });

    req.on("error", reject);
    req.write(postData);
    req.end();
  });
}

async function main() {
  console.log("🔄 Testing Sync Flow\n");
  console.log("=".repeat(60));

  // Before sync
  const beforeCount = checkDbBefore();
  console.log("📊 BEFORE SYNC:");
  console.log("   PROCESSED orders in DB:", beforeCount);

  // Trigger sync
  console.log("\n🔄 Triggering sync for PROCESSED orders...");
  try {
    const result = await triggerSync();
    console.log("   Sync response:", JSON.stringify(result, null, 2));
  } catch (err) {
    console.log("   Sync error:", err.message);
  }

  // Wait a bit for sync to complete
  await new Promise((r) => setTimeout(r, 2000));

  // After sync
  const afterOrders = checkDbAfter();
  console.log("\n📊 AFTER SYNC:");
  console.log("   PROCESSED orders in DB:", afterOrders.length);
  console.log(
    "   Order SNs:",
    afterOrders.slice(0, 5).join(", "),
    afterOrders.length > 5 ? "..." : ""
  );

  console.log("\n=".repeat(60));
  console.log("📊 RESULT:");
  console.log("   Before:", beforeCount);
  console.log("   After:", afterOrders.length);
  console.log("   Change:", afterOrders.length - beforeCount);

  if (afterOrders.length < beforeCount) {
    console.log("\n✅ Sync correctly removed stale orders!");
  } else if (afterOrders.length === beforeCount) {
    console.log(
      "\n⚠️  No change - sync may not have run or data was already current"
    );
  } else {
    console.log("\n📈 More orders after sync");
  }
}

main();

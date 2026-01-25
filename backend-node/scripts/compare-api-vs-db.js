/**
 * Script to compare orders from Shopee API vs Database
 * Finds missing orders that were fetched but not saved
 */

const Database = require("better-sqlite3");
const https = require("https");
const crypto = require("crypto");

// Shopee API credentials (from env/config)
const PARTNER_ID = process.env.SHOPEE_PARTNER_ID;
const PARTNER_KEY = process.env.SHOPEE_PARTNER_KEY;
const ACCESS_TOKEN = process.env.SHOPEE_ACCESS_TOKEN;
const SHOP_ID = process.env.SHOPEE_SHOP_ID;

const API_HOST = "partner.shopeemobile.com";

function generateSignature(path, timestamp) {
  const baseString = `${PARTNER_ID}${path}${timestamp}${ACCESS_TOKEN}${SHOP_ID}`;
  return crypto
    .createHmac("sha256", PARTNER_KEY)
    .update(baseString)
    .digest("hex");
}

async function fetchShopeeOrders(status) {
  return new Promise((resolve, reject) => {
    const timestamp = Math.floor(Date.now() / 1000);
    const path = "/api/v2/order/get_order_list";
    const sign = generateSignature(path, timestamp);

    // Calculate time range (last 7 days)
    const timeTo = timestamp;
    const timeFrom = timestamp - 7 * 24 * 60 * 60;

    const params = new URLSearchParams({
      time_range_field: "create_time",
      time_from: timeFrom.toString(),
      time_to: timeTo.toString(),
      page_size: "100",
      order_status: status,
      response_optional_fields: "order_status",
      partner_id: PARTNER_ID,
      shop_id: SHOP_ID,
      timestamp: timestamp.toString(),
      sign: sign,
      access_token: ACCESS_TOKEN,
    });

    const options = {
      hostname: API_HOST,
      path: `${path}?${params.toString()}`,
      method: "GET",
    };

    const req = https.request(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          const json = JSON.parse(data);
          resolve(json);
        } catch (e) {
          reject(e);
        }
      });
    });

    req.on("error", reject);
    req.end();
  });
}

async function main() {
  console.log("🔍 Comparing Shopee API orders vs Database...\n");

  // Check if we have credentials
  if (!PARTNER_ID || !PARTNER_KEY || !ACCESS_TOKEN || !SHOP_ID) {
    console.log("❌ Missing Shopee API credentials in environment");
    console.log("   Falling back to database-only analysis...\n");

    // Just analyze database
    const db = new Database("/app/config/databases/yumna_bertigamart.db");

    const processed = db
      .prepare(
        "SELECT orderSn, orderStatus FROM ShopeeOrder WHERE orderStatus = 'PROCESSED' ORDER BY createdAt DESC"
      )
      .all();
    console.log("📦 PROCESSED orders in DB:", processed.length);

    const allOrders = db
      .prepare(
        "SELECT orderSn, orderStatus FROM ShopeeOrder ORDER BY createdAt DESC"
      )
      .all();
    console.log("📦 Total orders in DB:", allOrders.length);

    // Group by status
    const byStatus = {};
    for (const o of allOrders) {
      byStatus[o.orderStatus] = (byStatus[o.orderStatus] || 0) + 1;
    }
    console.log("\n📊 By status:");
    for (const [status, count] of Object.entries(byStatus)) {
      console.log("  -", status + ":", count);
    }

    db.close();
    return;
  }

  try {
    // Fetch from API
    console.log("📡 Fetching PROCESSED orders from Shopee API...");
    const apiResponse = await fetchShopeeOrders("PROCESSED");

    if (apiResponse.error) {
      console.log("❌ API Error:", apiResponse.error, apiResponse.message);
      return;
    }

    const apiOrders = apiResponse.response?.order_list || [];
    console.log("📦 API returned:", apiOrders.length, "orders\n");

    // Get from database
    const db = new Database("/app/config/databases/yumna_bertigamart.db");
    const dbOrders = db
      .prepare(
        "SELECT orderSn FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'"
      )
      .all();
    console.log("📦 Database has:", dbOrders.length, "PROCESSED orders\n");

    // Compare
    const apiOrderSns = new Set(apiOrders.map((o) => o.order_sn));
    const dbOrderSns = new Set(dbOrders.map((o) => o.orderSn));

    // Find missing (in API but not in DB)
    const missingInDb = [...apiOrderSns].filter((sn) => !dbOrderSns.has(sn));
    console.log("❌ Missing in DB (in API but not saved):", missingInDb.length);
    for (const sn of missingInDb) {
      console.log("  -", sn);
    }

    // Find extra (in DB but not in API)
    const extraInDb = [...dbOrderSns].filter((sn) => !apiOrderSns.has(sn));
    console.log("\n⚠️  Extra in DB (not in API response):", extraInDb.length);
    for (const sn of extraInDb) {
      console.log("  -", sn);
    }

    // Check if missing orders exist with different status
    if (missingInDb.length > 0) {
      console.log(
        "\n🔍 Checking if missing orders exist with different status..."
      );
      for (const sn of missingInDb) {
        const found = db
          .prepare(
            "SELECT orderSn, orderStatus FROM ShopeeOrder WHERE orderSn = ?"
          )
          .get(sn);
        if (found) {
          console.log("  -", sn, "→ exists with status:", found.orderStatus);
        } else {
          console.log("  -", sn, "→ NOT FOUND in database at all!");
        }
      }
    }

    db.close();
  } catch (error) {
    console.error("Error:", error.message);
  }
}

main();

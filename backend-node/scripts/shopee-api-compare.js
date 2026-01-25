/**
 * Direct Shopee API comparison script
 * Fetches orders from Shopee API and compares with database
 */

const Database = require("better-sqlite3");
const https = require("https");
const crypto = require("crypto");

const DB_PATH = "/app/config/databases/yumna_bertigamart.db";

function loadConfig(db) {
  // Config is stored as key-value pairs
  const rows = db
    .prepare(
      `
    SELECT configKey, configValue 
    FROM PlatformConfig 
    WHERE platform = 'shopee'
  `
    )
    .all();

  const config = {};
  for (const row of rows) {
    config[row.configKey] = row.configValue;
  }

  return {
    partnerId: config.partner_id || config.partnerId,
    partnerKey: config.partner_key || config.partnerKey,
    accessToken: config.access_token || config.accessToken,
    shopId: config.shop_id || config.shopId,
  };
}

function generateSign(
  partnerId,
  partnerKey,
  path,
  timestamp,
  accessToken,
  shopId
) {
  const baseString = `${partnerId}${path}${timestamp}${accessToken}${shopId}`;
  return crypto
    .createHmac("sha256", partnerKey)
    .update(baseString)
    .digest("hex");
}

function fetchShopeeOrders(config, status) {
  return new Promise((resolve, reject) => {
    const timestamp = Math.floor(Date.now() / 1000);
    const path = "/api/v2/order/get_order_list";
    const sign = generateSign(
      config.partnerId,
      config.partnerKey,
      path,
      timestamp,
      config.accessToken,
      config.shopId
    );

    const timeFrom = timestamp - 7 * 24 * 60 * 60;
    const timeTo = timestamp;

    const params = new URLSearchParams({
      time_range_field: "create_time",
      time_from: timeFrom.toString(),
      time_to: timeTo.toString(),
      page_size: "100",
      order_status: status,
      response_optional_fields: "order_status",
      partner_id: config.partnerId,
      shop_id: config.shopId,
      timestamp: timestamp.toString(),
      sign: sign,
      access_token: config.accessToken,
    });

    const url = `${path}?${params.toString()}`;

    const options = {
      hostname: "partner.shopeemobile.com",
      path: url,
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    };

    const req = https.request(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          resolve(JSON.parse(data));
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
  console.log("🔍 Direct Shopee API vs Database Comparison\n");
  console.log("=".repeat(60));

  const db = new Database(DB_PATH);

  // Load config
  const config = loadConfig(db);
  if (!config) {
    console.log("❌ No Shopee config found");
    db.close();
    return;
  }

  console.log("✅ Config loaded: shopId =", config.shopId);

  // Fetch from API
  console.log("\n📡 Fetching PROCESSED orders from Shopee API...");

  try {
    const response = await fetchShopeeOrders(config, "PROCESSED");

    if (response.error) {
      console.log("❌ API Error:", response.error, response.message);
      db.close();
      return;
    }

    const apiOrders = response.response?.order_list || [];
    const apiOrderSns = apiOrders.map((o) => o.order_sn);

    console.log("📦 API returned:", apiOrders.length, "PROCESSED orders");

    // Get from database
    const dbOrders = db
      .prepare(
        `
      SELECT orderSn, orderStatus FROM ShopeeOrder WHERE orderStatus = 'PROCESSED'
    `
      )
      .all();
    const dbOrderSns = dbOrders.map((o) => o.orderSn);

    console.log("📦 Database has:", dbOrders.length, "PROCESSED orders");

    // Compare
    const apiSet = new Set(apiOrderSns);
    const dbSet = new Set(dbOrderSns);

    // Missing in DB (in API but not in DB)
    const missingInDb = apiOrderSns.filter((sn) => !dbSet.has(sn));

    // Extra in DB (in DB but not in API)
    const extraInDb = dbOrderSns.filter((sn) => !apiSet.has(sn));

    console.log("\n" + "=".repeat(60));
    console.log("📊 COMPARISON RESULTS:");
    console.log("=".repeat(60));

    console.log(
      "\n❌ Missing in DB (in API but NOT saved):",
      missingInDb.length
    );
    if (missingInDb.length > 0) {
      for (const sn of missingInDb) {
        // Check if it exists with different status
        const found = db
          .prepare(
            `SELECT orderSn, orderStatus FROM ShopeeOrder WHERE orderSn = ?`
          )
          .get(sn);
        if (found) {
          console.log(`  - ${sn} → EXISTS with status: ${found.orderStatus}`);
        } else {
          console.log(`  - ${sn} → NOT IN DATABASE AT ALL!`);
        }
      }
    }

    console.log("\n⚠️  Extra in DB (in DB but NOT in API):", extraInDb.length);
    if (extraInDb.length > 0) {
      for (const sn of extraInDb.slice(0, 10)) {
        console.log(`  - ${sn}`);
      }
      if (extraInDb.length > 10) {
        console.log(`  ... and ${extraInDb.length - 10} more`);
      }
    }

    // Summary
    console.log("\n" + "=".repeat(60));
    console.log("📊 SUMMARY:");
    console.log("  - API PROCESSED orders:", apiOrders.length);
    console.log("  - DB PROCESSED orders:", dbOrders.length);
    console.log("  - Difference:", apiOrders.length - dbOrders.length);
    console.log("  - Missing in DB:", missingInDb.length);
    console.log("  - Extra in DB:", extraInDb.length);
    console.log("=".repeat(60));
  } catch (error) {
    console.error("Error:", error.message);
  }

  db.close();
}

main();

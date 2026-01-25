/**
 * Direct TikTok Sync Test - Check what API returns
 */
const fetch = require("node-fetch");

const API_URL = "http://localhost:3000";
const TENANT_ID = "yumna_bertigamart";

async function main() {
  console.log("🔍 Testing TikTok Sync API...\n");

  try {
    // 1. Test /api/orders/sync/processed with platform=tiktok
    console.log("📡 POST /api/orders/sync/processed with platform=tiktok...");
    const response = await fetch(`${API_URL}/api/orders/sync/processed`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": TENANT_ID,
      },
      body: JSON.stringify({
        platform: "tiktok",
        forceRefresh: true,
      }),
    });

    const result = await response.json();
    console.log("Response status:", response.status);
    console.log("Success:", result.success);
    console.log("Category:", result.data?.category);
    console.log("Count:", result.data?.count);

    // Check orders
    const orders = result.data?.orders || [];
    console.log(`\nTotal orders in response: ${orders.length}`);

    // Filter by platform
    const tiktokOrders = orders.filter((o) => o._platform === "tiktok");
    const shopeeOrders = orders.filter((o) => o._platform === "shopee");
    console.log(`TikTok orders: ${tiktokOrders.length}`);
    console.log(`Shopee orders: ${shopeeOrders.length}`);

    // List TikTok order SNs
    if (tiktokOrders.length > 0) {
      console.log("\n📦 TikTok Order SNs:");
      tiktokOrders.forEach((o) => console.log(`   - ${o.order_sn}`));
    }

    // 2. Also test getOrdersByCategory endpoint
    console.log("\n\n📡 GET /api/orders/category/processed?platform=tiktok...");
    const response2 = await fetch(
      `${API_URL}/api/orders/category/processed?platform=tiktok`,
      {
        headers: {
          "x-tenant-id": TENANT_ID,
        },
      }
    );

    const result2 = await response2.json();
    console.log("Response status:", response2.status);
    console.log("Success:", result2.success);
    const orders2 = result2.data?.orders || [];
    console.log(`Orders count: ${orders2.length}`);

    // Show orders
    if (orders2.length > 0) {
      console.log("\n📦 TikTok Orders from GET:");
      orders2
        .slice(0, 10)
        .forEach((o) =>
          console.log(`   - ${o.order_sn} [${o._platform || o.platform}]`)
        );
      if (orders2.length > 10) {
        console.log(`   ... and ${orders2.length - 10} more`);
      }
    }
  } catch (error) {
    console.error("Error:", error.message);
  }
}

main();

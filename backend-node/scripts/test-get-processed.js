/**
 * Test GET /api/orders/processed - what frontend uses
 */
const fetch = require("node-fetch");

const API_URL = "http://localhost:3000";
const TENANT_ID = "yumna_bertigamart";

async function main() {
  console.log("🔍 Testing GET /api/orders/processed (what frontend uses)...\n");

  try {
    const response = await fetch(`${API_URL}/api/orders/processed`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": TENANT_ID,
      },
    });

    const result = await response.json();
    console.log("Response status:", response.status);
    console.log("Success:", result.success);
    console.log("Count:", result.count);
    console.log("Items length:", (result.items || result.data || []).length);

    const items = result.items || result.data || [];

    // Count by platform
    const tiktokItems = items.filter((o) => o.platform === "TIKTOK");
    const shopeeItems = items.filter((o) => o.platform === "SHOPEE");

    console.log(`\n📊 Items by Platform:`);
    console.log(`   TIKTOK: ${tiktokItems.length}`);
    console.log(`   SHOPEE: ${shopeeItems.length}`);

    // Count unique order numbers
    const tiktokOrderNos = new Set(tiktokItems.map((o) => o.order_no));
    const shopeeOrderNos = new Set(shopeeItems.map((o) => o.order_no));

    console.log(`\n📦 Unique Orders by Platform:`);
    console.log(`   TIKTOK unique orders: ${tiktokOrderNos.size}`);
    console.log(`   SHOPEE unique orders: ${shopeeOrderNos.size}`);

    // Show first few TikTok orders
    console.log(`\n📋 TikTok Order Numbers:`);
    Array.from(tiktokOrderNos)
      .slice(0, 20)
      .forEach((no) => console.log(`   - ${no}`));
  } catch (error) {
    console.error("Error:", error.message);
  }
}

main();

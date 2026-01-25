/**
 * Test Orders Endpoint - Check what /api/orders/processed returns
 */
const fetch = require("node-fetch");

const API_URL = "http://localhost:3000";
const TENANT_ID = "yumna_bertigamart";

async function main() {
  console.log("🔍 Testing /api/orders/processed endpoint...\n");

  try {
    // First do sync
    console.log("📡 1. Syncing all processed orders...");
    const syncResponse = await fetch(`${API_URL}/api/orders/sync/processed`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": TENANT_ID,
      },
      body: JSON.stringify({ days: 7 }),
    });
    const syncResult = await syncResponse.json();
    console.log("Sync response:", syncResult.success ? "SUCCESS" : "FAILED");

    // Then fetch
    console.log("\n📡 2. Fetching GET /api/orders/processed...");
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

    const items = result.items || result.data || [];
    console.log(`Total items: ${items.length}`);

    // Group by platform
    const byPlatform = {};
    for (const item of items) {
      const platform = item.platform || "UNKNOWN";
      if (!byPlatform[platform]) {
        byPlatform[platform] = new Set();
      }
      byPlatform[platform].add(item.order_no);
    }

    console.log("\n📊 Orders by Platform (unique order_no):");
    for (const [platform, orderNos] of Object.entries(byPlatform)) {
      console.log(`   ${platform}: ${orderNos.size} unique orders`);
    }

    // Check TikTok orders specifically
    const tiktokItems = items.filter((i) => i.platform === "TIKTOK");
    console.log(`\n📦 TikTok items: ${tiktokItems.length}`);

    const tiktokOrderNos = new Set(tiktokItems.map((i) => i.order_no));
    console.log(`📦 TikTok unique orders: ${tiktokOrderNos.size}`);

    // List TikTok order_no
    console.log("\n📋 TikTok Order Numbers:");
    Array.from(tiktokOrderNos).forEach((no) => console.log(`   - ${no}`));
  } catch (error) {
    console.error("Error:", error.message);
  }
}

main();

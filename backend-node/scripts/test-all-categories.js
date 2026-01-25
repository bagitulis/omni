/**
 * Test ALL categories
 */

const fetch = require("node-fetch");

async function testAllCategories() {
  const categories = ["unpaid", "unprocess", "processed"];

  console.log("\n📡 TESTING ALL CATEGORY ENDPOINTS:");
  console.log("=".repeat(70));

  for (const category of categories) {
    try {
      const response = await fetch(
        `http://localhost:3001/api/orders/${category}`
      );
      const data = await response.json();

      if (data.success) {
        console.log(`\n✓ GET /api/orders/${category}`);
        console.log(`  Count: ${data.count}`);
        console.log(`  Items returned: ${data.items?.length || 0}`);

        // Break down by platform
        if (data.items && data.items.length > 0) {
          const platformCount = {};
          data.items.forEach((item) => {
            platformCount[item.platform] =
              (platformCount[item.platform] || 0) + 1;
          });

          console.log(`  By platform:`);
          for (const [platform, count] of Object.entries(platformCount)) {
            // Count unique orders for this platform
            const platformItems = data.items.filter(
              (i) => i.platform === platform
            );
            const uniqueOrders = new Set(
              platformItems.map((i) => i.order_no)
            ).size;
            console.log(
              `    - ${platform}: ${count} items, ${uniqueOrders} unique orders`
            );
          }
        }
      } else {
        console.log(`❌ Error getting ${category}: ${data.error}`);
      }
    } catch (error) {
      console.log(`❌ Failed to fetch ${category}: ${error.message}`);
    }
  }

  console.log("\n" + "=".repeat(70));
  console.log(
    "NOTE: Frontend shows count ONLY for Shopee in 'SHOPEE: XX Order' card"
  );
}

testAllCategories().catch((err) => {
  console.error("Error:", err);
  process.exit(1);
});

/**
 * Test Locked Orders Endpoint & Database
 * Debug why productName and variationName are empty
 */

const axios = require("axios");

const API_BASE = "http://localhost:3001/api/orders";

// Get token from arguments or use default
const token = process.argv[2];

if (!token) {
  console.error("❌ Usage: node test-locked-orders.js <YOUR_TOKEN>");
  console.log("\n💡 Get token from browser localStorage.getItem('authToken')");
  process.exit(1);
}

const headers = {
  Authorization: `Bearer ${token}`,
  "Content-Type": "application/json",
};

async function testLockedOrders() {
  console.log("🧪 Testing Locked Orders\n");

  try {
    // 1. Get locked orders
    console.log("1️⃣ Fetching locked orders from database...");
    const response = await axios.get(`${API_BASE}/locked-today`, { headers });

    console.log("\n📦 Response:");
    console.log(JSON.stringify(response.data, null, 2));

    // 2. Check data structure
    if (response.data.success && response.data.data?.items) {
      const items = response.data.data.items;
      console.log(`\n✅ Found ${items.length} locked items`);

      // 3. Inspect first 3 items
      console.log("\n🔍 Sample items:");
      items.slice(0, 3).forEach((item, idx) => {
        console.log(`\n  Item ${idx + 1}:`);
        console.log(`    SKU: "${item.sku}"`);
        console.log(
          `    Product Name: "${item.productName || item.product_name}"`
        );
        console.log(
          `    Variation: "${
            item.variationName || item.variation_name || "N/A"
          }"`
        );
        console.log(`    QTY: ${item.qty}`);
      });

      // 4. Check for empty names
      const emptyNames = items.filter(
        (item) => !item.productName && !item.product_name
      );
      const emptyVariations = items.filter(
        (item) => !item.variationName && !item.variation_name
      );

      console.log(`\n⚠️  Items with empty product names: ${emptyNames.length}`);
      console.log(`⚠️  Items with empty variations: ${emptyVariations.length}`);

      if (emptyNames.length > 0) {
        console.log("\n❌ PROBLEM: Product names are missing!");
        console.log("   Check database lockedOrder table");
      }
    } else {
      console.log("\n❌ No locked orders data found");
    }
  } catch (error) {
    console.error("\n❌ Error:", error.response?.data || error.message);
  }
}

testLockedOrders();

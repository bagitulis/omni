/**
 * Check Shopee unprocess orders - find multi-item order
 */

const fetch = require("node-fetch");

async function checkShopeeUnprocess() {
  try {
    const response = await fetch(`http://localhost:3001/api/orders/unprocess`);
    const data = await response.json();

    const shopeeItems = data.items.filter((i) => i.platform === "SHOPEE");

    console.log("\n📋 SHOPEE UNPROCESS BREAKDOWN:");
    console.log("=".repeat(70));
    console.log(`Total items: ${shopeeItems.length}`);

    // Group by order_no
    const orderMap = {};
    shopeeItems.forEach((item) => {
      if (!orderMap[item.order_no]) {
        orderMap[item.order_no] = [];
      }
      orderMap[item.order_no].push(item);
    });

    console.log(`Unique orders: ${Object.keys(orderMap).length}`);

    console.log("\nOrder breakdown:");
    for (const [orderNo, items] of Object.entries(orderMap)) {
      console.log(`  ${orderNo}: ${items.length} item(s)`);
      if (items.length > 1) {
        console.log(`    ↳ ${items.map((i) => i.product_name).join(", ")}`);
      }
    }

    console.log("\n" + "=".repeat(70));
    console.log("Summary:");
    console.log(`  If you count ITEMS: 11`);
    console.log(`  If you count ORDERS: 10`);
  } catch (error) {
    console.error("Error:", error);
  }
}

checkShopeeUnprocess();

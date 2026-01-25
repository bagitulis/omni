/**
 * Quick Debug: Dump Raw API Responses
 * To understand actual field structure
 */

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";
import { LazadaOrderManager } from "./src/services/orders/lazadaOrderManager";

async function dumpRawResponse() {
  console.log("📦 Fetching raw order data to inspect response structure...\n");

  const configManager = new LazadaConfigManager();
  await configManager.loadConfig();

  const apiClient = new LazadaAPIClient(configManager);
  const orderManager = new LazadaOrderManager(apiClient);

  // Get SHIPPED orders
  const orders = await orderManager.getOrderList("SHIPPED", 7);

  if (orders.length === 0) {
    console.log("❌ No orders found");
    return;
  }

  console.log(`✅ Found ${orders.length} SHIPPED orders\n`);

  // Get items for first order
  const firstOrderId = String(orders[0].order_id);
  console.log(`📋 Fetching items for order ID: ${firstOrderId}\n`);

  const items = await orderManager.getOrderItems([firstOrderId]);

  if (items.length === 0) {
    console.log("❌ No items found");
    return;
  }

  console.log(`✅ Found ${items.length} items\n`);
  console.log("📊 FIRST ITEM STRUCTURE:");
  console.log(JSON.stringify(items[0], null, 2));

  console.log("\n📊 ALL ITEMS:");
  items.forEach((item, index) => {
    console.log(`\nItem ${index + 1}:`);
    console.log(`  - item_id: ${item.item_id}`);
    console.log(`  - seller_sku: ${item.seller_sku}`);
    console.log(`  - product_name: ${item.product_name}`);
    console.log(`  - price: ${item.price}`);
  });
}

dumpRawResponse().catch((error) => {
  console.error("Error:", error);
  process.exit(1);
});

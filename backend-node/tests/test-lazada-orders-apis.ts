/**
 * Test Lazada Order APIs
 * 1. GET /order/get - Get single order details
 * 2. GET /orders/items/get - Get items from multiple orders
 */

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";

class LazadaOrderAPITester {
  private configManager: LazadaConfigManager;
  private apiClient: LazadaAPIClient;

  constructor() {
    this.configManager = new LazadaConfigManager();
    this.apiClient = new LazadaAPIClient(this.configManager);
  }

  async initialize(): Promise<void> {
    await this.configManager.loadConfig();
    if (!this.configManager.accessToken) {
      throw new Error("No access token found");
    }
  }

  /**
   * Test 1: /order/get - Get single order details
   */
  async testOrderGet(orderId: string): Promise<void> {
    console.log("\n╔════════════════════════════════════════════════════╗");
    console.log("║  TEST 1: GET /order/get                           ║");
    console.log("║  Get single order details                          ║");
    console.log("╚════════════════════════════════════════════════════╝\n");

    try {
      console.log(`📦 Fetching order: ${orderId}`);

      const response = await this.apiClient.request("/order/get", "GET", {
        order_id: orderId,
      });

      console.log("\n📡 RAW RESPONSE:");
      console.log(JSON.stringify(response, null, 2));

      if (!response || response.code !== "0") {
        console.log(`❌ API Error: ${response?.message || "Unknown"}`);
        return;
      }

      const order = response.data;
      console.log("\n✅ Order retrieved\n");

      console.log(`📋 Order Details:`);
      console.log(`   Order ID: ${order.order_id}`);
      console.log(`   Order Number: ${order.order_number}`);
      console.log(`   Status: ${order.status}`);
      console.log(`   Created: ${order.created_at}`);
      console.log(`   Items Count: ${order.items_count}`);
      console.log(`   Total Price: ${order.price} ${order.currency}`);
      console.log(`   Shipping Fee: ${order.shipping_fee}`);
      console.log(`   Payment Method: ${order.payment_method}`);

      if (order.address_shipping) {
        console.log(`\n📍 Shipping Address:`);
        const addr = order.address_shipping;
        console.log(`   ${addr.first_name} ${addr.last_name}`);
        console.log(`   ${addr.address1}`);
        console.log(`   ${addr.city}, ${addr.post_code}`);
        console.log(`   ${addr.country}`);
        console.log(`   Phone: ${addr.phone}`);
      }

      if (order.buyer_note) {
        console.log(`\n💬 Buyer Note: ${order.buyer_note}`);
      }
    } catch (error) {
      console.error("❌ Error:", error);
    }
  }

  /**
   * Test 2: /orders/items/get - Get items from multiple orders
   */
  async testOrdersItemsGet(orderIds: string[]): Promise<void> {
    console.log("\n╔════════════════════════════════════════════════════╗");
    console.log("║  TEST 2: GET /orders/items/get                    ║");
    console.log("║  Get items from multiple orders (max 50)           ║");
    console.log("╚════════════════════════════════════════════════════╝\n");

    try {
      const idsStr = JSON.stringify(orderIds);
      console.log(`📦 Fetching items for orders: ${idsStr}`);

      const response = await this.apiClient.request(
        "/orders/items/get",
        "GET",
        {
          order_ids: idsStr,
        }
      );

      console.log("\n📡 RAW RESPONSE:");
      console.log(JSON.stringify(response, null, 2));

      if (!response || response.code !== "0") {
        console.log(`❌ API Error: ${response?.message || "Unknown"}`);
        return;
      }

      const ordersData = response.data;
      console.log(`\n✅ Data retrieved for ${ordersData.length} order(s)\n`);

      ordersData.forEach((orderData: any, orderIdx: number) => {
        console.log(`📋 Order ${orderIdx + 1}:`);
        console.log(`   Order ID: ${orderData.order_id}`);
        console.log(`   Order Number: ${orderData.order_number}`);
        console.log(`   Items: ${orderData.order_items.length}\n`);

        orderData.order_items.forEach((item: any, itemIdx: number) => {
          console.log(`   Item ${itemIdx + 1}:`);
          console.log(`      Product: ${item.name}`);
          console.log(`      SKU: ${item.sku}`);
          console.log(`      Status: ${item.status}`);
          console.log(`      Price: ${item.item_price} ${item.currency}`);
          console.log(`      Quantity: ${item.quantity}`);
          console.log(`      Tracking: ${item.tracking_code || "N/A"}`);
          console.log();
        });
      });
    } catch (error) {
      console.error("❌ Error:", error);
    }
  }

  /**
   * Run all tests
   */
  async runTests(orderId: string): Promise<void> {
    try {
      await this.initialize();
      console.log("✅ Configuration loaded\n");

      // Test 1
      await this.testOrderGet(orderId);

      // Test 2
      await this.testOrdersItemsGet([orderId]);

      console.log("\n╔════════════════════════════════════════════════════╗");
      console.log("║  ✅ ALL TESTS COMPLETED                            ║");
      console.log("╚════════════════════════════════════════════════════╝\n");
    } catch (error) {
      console.error("❌ Fatal error:", error);
      process.exit(1);
    }
  }
}

const orderId = process.argv[2] || "2672780399902269";
const tester = new LazadaOrderAPITester();
tester.runTests(orderId);

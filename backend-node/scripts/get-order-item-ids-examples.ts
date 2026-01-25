/**
 * GETTING REAL ORDER ITEM IDs - PRACTICAL EXAMPLES
 *
 * This file shows you exactly how to get real order item IDs
 * from your Lazada shop to use in the document API
 */

import { LazadaOrderManager } from "./src/services/orders/lazadaOrderManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";
import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";

// ============================================================================
// METHOD 1: Get Order Item IDs from Recently Processed Orders
// ============================================================================

async function getOrderItemsFromProcessedOrders(): Promise<number[]> {
  console.log("📦 Fetching order item IDs from processed orders...\n");

  try {
    const configManager = new LazadaConfigManager();
    await configManager.loadConfig();

    const apiClient = new LazadaAPIClient(configManager);
    const orderManager = new LazadaOrderManager(apiClient);

    // Step 1: Get processed orders from last 7 days
    console.log("Step 1: Fetching processed orders...");
    const processedOrders = await orderManager.getOrderList("processed", 7);

    if (processedOrders.length === 0) {
      console.log("❌ No processed orders found in last 7 days");
      return [];
    }

    console.log(`✅ Found ${processedOrders.length} processed orders\n`);

    // Step 2: Extract order IDs
    const orderIds = processedOrders.map((order) => String(order.order_id));
    console.log(`Order IDs: ${orderIds.join(", ")}\n`);

    // Step 3: Get order items (this includes order_item_id)
    console.log("Step 2: Fetching order items...");
    const allItems = await orderManager.getOrderItems(orderIds);

    if (allItems.length === 0) {
      console.log("❌ No order items found");
      return [];
    }

    console.log(`✅ Found ${allItems.length} items\n`);

    // Step 4: Extract order_item_ids
    const orderItemIds = allItems.map(
      (item) => item.item_id || item.order_item_id
    );

    console.log("📋 Sample Order Items:");
    allItems.slice(0, 3).forEach((item) => {
      console.log(`   - Item ID: ${item.order_item_id}`);
      console.log(`     SKU: ${item.seller_sku}`);
      console.log(`     Product: ${item.product_name}`);
    });

    if (allItems.length > 3) {
      console.log(`   ... and ${allItems.length - 3} more items`);
    }

    console.log(`\n✅ Extracted order_item_ids:`, orderItemIds);
    return orderItemIds;
  } catch (error) {
    console.error("❌ Error fetching order items:", error);
    return [];
  }
}

// ============================================================================
// METHOD 2: Get Order Item IDs from a Specific Order
// ============================================================================

async function getOrderItemsFromSpecificOrder(
  orderId: string
): Promise<number[]> {
  console.log(`📦 Fetching items from order ${orderId}...\n`);

  try {
    const configManager = new LazadaConfigManager();
    await configManager.loadConfig();

    const apiClient = new LazadaAPIClient(configManager);
    const orderManager = new LazadaOrderManager(apiClient);

    // Fetch items for this specific order
    const items = await orderManager.getOrderItems([orderId]);

    if (items.length === 0) {
      console.log(`❌ No items found for order ${orderId}`);
      return [];
    }

    console.log(`✅ Found ${items.length} items:\n`);

    const orderItemIds = items.map((item) => item.order_item_id);

    items.forEach((item, index) => {
      console.log(`${index + 1}. Item ID: ${item.order_item_id}`);
      console.log(`   SKU: ${item.seller_sku}`);
      console.log(`   Product: ${item.product_name}`);
      console.log(`   Price: ${item.price}`);
      console.log("");
    });

    return orderItemIds;
  } catch (error) {
    console.error("❌ Error fetching order items:", error);
    return [];
  }
}

// ============================================================================
// METHOD 3: Get Order Item IDs from All Order Statuses
// ============================================================================

async function getAllOrderItems(): Promise<
  Record<string, { orders: any[]; items: any[] }>
> {
  console.log("📦 Fetching order items from ALL statuses...\n");

  const statuses = [
    "PENDING_PAYMENT",
    "PAYMENT_CONFIRMED",
    "PROCESSING",
    "READY_TO_SHIP",
    "SHIPPED",
    "DELIVERED",
    "CANCELLED",
    "RETURNED",
  ];

  const configManager = new LazadaConfigManager();
  await configManager.loadConfig();

  const apiClient = new LazadaAPIClient(configManager);
  const orderManager = new LazadaOrderManager(apiClient);

  const results: Record<string, { orders: any[]; items: any[] }> = {};

  for (const status of statuses) {
    try {
      console.log(`⏳ Fetching ${status} orders...`);

      // Get orders with this status
      const orders = await orderManager.getOrderList(status, 7);

      if (orders.length === 0) {
        console.log(`   No ${status} orders found`);
        continue;
      }

      console.log(`   ✅ Found ${orders.length} orders`);

      // Get items for these orders
      const orderIds = orders.map((o) => String(o.order_id));
      const items = await orderManager.getOrderItems(orderIds);

      console.log(`   ✅ Found ${items.length} items\n`);

      results[status] = { orders, items };
    } catch (error) {
      console.error(`   ❌ Error: ${error}`);
    }
  }

  return results;
}

// ============================================================================
// METHOD 4: Get Items Ready for Document Retrieval
// ============================================================================

async function getItemsReadyForDocuments(): Promise<
  Array<{
    status: string;
    orderItemId: number;
    productName: string;
    orderStatus: string;
  }>
> {
  console.log("📦 Fetching items that are ready for document retrieval...\n");

  // Documents are typically available for:
  // - READY_TO_SHIP: Can get shipping label
  // - SHIPPED: Can get shipping label
  // - DELIVERED: Can get all documents (label, invoice)
  const readyStatuses = ["READY_TO_SHIP", "SHIPPED", "DELIVERED"];

  const configManager = new LazadaConfigManager();
  await configManager.loadConfig();

  const apiClient = new LazadaAPIClient(configManager);
  const orderManager = new LazadaOrderManager(apiClient);

  const readyItems: Array<{
    status: string;
    orderItemId: number;
    productName: string;
    orderStatus: string;
  }> = [];

  for (const status of readyStatuses) {
    try {
      console.log(`⏳ Checking ${status} orders...`);

      const orders = await orderManager.getOrderList(status, 7);

      if (orders.length === 0) {
        console.log(`   No ${status} orders\n`);
        continue;
      }

      const orderIds = orders.map((o) => String(o.order_id));
      const items = await orderManager.getOrderItems(orderIds);

      items.forEach((item) => {
        readyItems.push({
          status,
          orderItemId: item.item_id || item.order_item_id,
          productName: item.product_name || item.name,
          orderStatus: status,
        });
      });

      console.log(`   ✅ Added ${items.length} items\n`);
    } catch (error) {
      console.error(`   ❌ Error: ${error}`);
    }
  }

  console.log(`✅ Total items ready for documents: ${readyItems.length}\n`);

  if (readyItems.length > 0) {
    console.log("Sample items:");
    readyItems.slice(0, 5).forEach((item) => {
      console.log(`  - ID: ${item.orderItemId} (${item.status})`);
      console.log(`    Product: ${item.productName}`);
    });

    if (readyItems.length > 5) {
      console.log(`  ... and ${readyItems.length - 5} more`);
    }
  }

  return readyItems;
}

// ============================================================================
// EXAMPLE USAGE & TESTING
// ============================================================================

export async function runExamples() {
  console.log(
    "\n╔════════════════════════════════════════════════════════════╗"
  );
  console.log("║  GETTING ORDER ITEM IDs - PRACTICAL EXAMPLES             ║");
  console.log(
    "╚════════════════════════════════════════════════════════════╝\n"
  );

  // Example 1: Get items from processed orders
  console.log("═══════════════════════════════════════════════════════════\n");
  console.log("EXAMPLE 1: Items from Processed Orders\n");
  const processedItems = await getOrderItemsFromProcessedOrders();

  if (processedItems.length > 0) {
    console.log("\n💡 Now you can use these IDs for testing document API:");
    console.log(
      `   const itemIds = [${processedItems.slice(0, 3).join(", ")}];`
    );
  }

  // Example 2: Get items from specific order
  console.log(
    "\n═══════════════════════════════════════════════════════════\n"
  );
  console.log("EXAMPLE 2: Items from Specific Order\n");

  if (processedItems.length > 0) {
    // Get all orders first to find an order ID
    const configManager = new LazadaConfigManager();
    await configManager.loadConfig();

    const apiClient = new LazadaAPIClient(configManager);
    const orderManager = new LazadaOrderManager(apiClient);

    const orders = await orderManager.getOrderList("processed", 7);
    if (orders.length > 0) {
      const specificOrderItems = await getOrderItemsFromSpecificOrder(
        String(orders[0].order_id)
      );

      if (specificOrderItems.length > 0) {
        console.log("\n💡 Use these for document API:");
        console.log(`   apiClient.request("/order/document/get", "GET", {`);
        console.log(`     doc_type: "shippingLabel",`);
        console.log(`     order_item_ids: "[${specificOrderItems[0]}]"`);
        console.log(`   });`);
      }
    }
  }

  // Example 3: Get items ready for documents
  console.log(
    "\n═══════════════════════════════════════════════════════════\n"
  );
  console.log("EXAMPLE 3: Items Ready for Document Retrieval\n");
  const readyItems = await getItemsReadyForDocuments();

  if (readyItems.length > 0) {
    console.log("\n💡 Best items for testing:");
    const deliveredItems = readyItems.filter((i) => i.status === "DELIVERED");

    if (deliveredItems.length > 0) {
      const firstItem = deliveredItems[0];
      console.log(`   - ID: ${firstItem.orderItemId}`);
      console.log(
        `     Status: ${firstItem.status} (supports all document types)`
      );
      console.log(`     Product: ${firstItem.productName}`);
    }
  }

  console.log(
    "\n╔════════════════════════════════════════════════════════════╗"
  );
  console.log("║  ✅ EXAMPLES COMPLETED                                    ║");
  console.log(
    "╚════════════════════════════════════════════════════════════╝\n"
  );
}

// ============================================================================
// INTEGRATION WITH DOCUMENT API TEST
// ============================================================================

/**
 * Here's how to use the order item IDs with the document API:
 *
 * 1. Get order item IDs:
 *    const itemIds = await getOrderItemsFromProcessedOrders();
 *
 * 2. Use with API client:
 *    const response = await apiClient.request("/order/document/get", "GET", {
 *      doc_type: "shippingLabel",
 *      order_item_ids: JSON.stringify(itemIds) // [123, 456, 789]
 *    });
 *
 * 3. Check response:
 *    if (response.code === "0") {
 *      const buffer = Buffer.from(response.data.document.file, "base64");
 *      fs.writeFileSync("document.html", buffer);
 *      console.log("✅ Document saved!");
 *    }
 *
 * ORDER STATUS → DOCUMENT AVAILABILITY:
 *
 * PENDING_PAYMENT           ❌ No documents
 * PAYMENT_CONFIRMED         ❌ No documents
 * PROCESSING                ⚠️  May have partial docs
 * READY_TO_SHIP             ✅ Shipping label available
 * SHIPPED                   ✅ Shipping label + invoice
 * DELIVERED                 ✅ All documents available
 * CANCELLED                 ❌ May not have documents
 * RETURNED                  ❌ Limited documents
 */

// Run if executed directly
if (require.main === module) {
  runExamples().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
  });
}

export {
  getOrderItemsFromProcessedOrders,
  getOrderItemsFromSpecificOrder,
  getAllOrderItems,
  getItemsReadyForDocuments,
};

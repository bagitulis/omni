/**
 * Test script to verify order items count vs unique orders
 * This checks if 68 orders -> 67 unique order_no is due to multi-item orders
 */

import { PrismaClient } from "@prisma/client";

async function testOrderItemsCount() {
  // Connect to yumna tenant database - absolute path
  const dbPath = "C:/Users/PC/Desktop/omni/backend/config/databases/yumna.db";
  const prisma = new PrismaClient({
    datasources: {
      db: { url: `file:${dbPath}` },
    },
  });

  try {
    console.log("🔍 Checking Shopee orders with PROCESSED status...\n");

    // Get all PROCESSED Shopee orders
    const orders = await prisma.shopeeOrder.findMany({
      where: { orderStatus: "PROCESSED" },
      include: { items: true },
      orderBy: { createdAt: "desc" },
    });

    console.log(`📦 Total PROCESSED orders in database: ${orders.length}`);

    // Count unique order_sn
    const uniqueOrderSn = new Set(orders.map((o) => o.orderSn));
    console.log(`🔢 Unique order_sn: ${uniqueOrderSn.size}`);

    // Count total items
    let totalItems = 0;
    const ordersWithMultipleItems: any[] = [];

    for (const order of orders) {
      const itemCount = order.items?.length || 0;
      totalItems += itemCount;

      if (itemCount > 1) {
        ordersWithMultipleItems.push({
          orderSn: order.orderSn,
          itemCount,
          items: order.items?.map((i) => ({
            itemName: i.itemName,
            modelName: i.modelName,
            quantity: i.quantity,
          })),
        });
      }
    }

    console.log(`📋 Total items across all orders: ${totalItems}`);
    console.log(
      `📊 Orders with multiple items: ${ordersWithMultipleItems.length}`
    );

    if (ordersWithMultipleItems.length > 0) {
      console.log("\n🔍 Orders with multiple items:");
      for (const order of ordersWithMultipleItems) {
        console.log(`\n  Order: ${order.orderSn} (${order.itemCount} items)`);
        for (const item of order.items) {
          console.log(
            `    - ${item.itemName} | ${item.modelName} | qty: ${item.quantity}`
          );
        }
      }
    }

    // Now check how many rows would be returned by formatter
    // Each item becomes a separate row
    console.log("\n📊 Summary:");
    console.log(`  - Orders in DB: ${orders.length}`);
    console.log(`  - Items flattened: ${totalItems}`);
    console.log(`  - Unique order_sn: ${uniqueOrderSn.size}`);
    console.log(`  - Expected UI count (unique orders): ${uniqueOrderSn.size}`);
  } catch (error) {
    console.error("❌ Error:", error);
  } finally {
    await prisma.$disconnect();
  }
}

testOrderItemsCount();

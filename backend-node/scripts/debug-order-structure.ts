/**
 * Debug Script: Check ShopeeOrderItem data structure
 * Purpose: Investigate why order prices are showing as 0
 */

import { PrismaClient } from '@prisma/client';
import path from 'path';

async function main() {
  const tenantId = 'yumna';
  const databasesDir = path.resolve(__dirname, '../config/databases');
  const dbPath = path.join(databasesDir, `${tenantId}_bertigamart.db`);
  
  process.env.DATABASE_URL = `file:${dbPath}`;
  const prisma = new PrismaClient();

  console.log('🔍 Checking Shopee Order data structure...\n');

  // Get one recent order with items
  const order = await prisma.shopeeOrder.findFirst({
    include: {
      items: true,
    },
    orderBy: {
      createdAt: 'desc',
    },
  });

  if (!order) {
    console.log('❌ No orders found');
    return;
  }

  console.log('📦 Sample Order:');
  console.log('Order SN:', order.orderSn);
  console.log('Order Status:', order.orderStatus);
  console.log('Items count:', order.items.length);

  console.log('\n📊 Sample Order Item Fields:');
  if (order.items.length > 0) {
    const item = order.items[0];
    console.log('- itemId:', item.itemId);
    console.log('- modelId:', item.modelId);
    console.log('- itemName:', item.itemName);
    console.log('- modelName:', item.modelName);
    console.log('- itemSku:', item.itemSku);
    console.log('- modelSku:', item.modelSku);
    console.log('- quantity:', item.quantity);
    console.log('- price:', item.price);
  }

  console.log('\n🔍 Checking Inventory data structure...\n');

  // Get one inventory record
  const inventory = await prisma.inventoryRecord.findFirst({
    where: {
      tenantId,
    },
  });

  if (!inventory) {
    console.log('❌ No inventory found');
    return;
  }

  console.log('📦 Sample Inventory Record:');
  console.log('- keyValue (SKU):', inventory.keyValue);
  console.log('- keyColumnName:', inventory.keyColumnName);
  console.log('- data (parsed):');
  console.log(JSON.stringify(JSON.parse(inventory.data), null, 2));

  await prisma.$disconnect();
}

main().catch(console.error);

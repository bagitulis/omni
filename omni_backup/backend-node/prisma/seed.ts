// prisma/seed.ts
import { PrismaClient } from "@prisma/client";
import * as bcrypt from "bcryptjs";

const prisma = new PrismaClient();

async function main() {
  console.log("🌱 Seeding database...");

  // Clear existing data
  await prisma.orderItem.deleteMany();
  await prisma.order.deleteMany();
  await prisma.inventory.deleteMany();
  await prisma.product.deleteMany();
  await prisma.shop.deleteMany();
  await prisma.user.deleteMany();

  // Create test user
  const hashedPassword = await bcrypt.hash("password123", 10);
  const user = await prisma.user.create({
    data: {
      email: "test@example.com",
      password: hashedPassword,
      name: "Test User",
      role: "admin",
      isActive: true,
    },
  });
  console.log("✅ User created:", user.id);

  // Create Shopee shop
  const shopeeShop = await prisma.shop.create({
    data: {
      userId: user.id,
      platform: "shopee",
      shopName: "Test Shopee Store",
      shopId: "shopee_123456",
      shopToken: "test_shopee_token_xxx",
      isActive: true,
    },
  });
  console.log("✅ Shopee shop created:", shopeeShop.id);

  // Create Lazada shop
  const lazadaShop = await prisma.shop.create({
    data: {
      userId: user.id,
      platform: "lazada",
      shopName: "Test Lazada Store",
      shopId: "lazada_789012",
      shopToken: "test_lazada_token_xxx",
      isActive: true,
    },
  });
  console.log("✅ Lazada shop created:", lazadaShop.id);

  // Create TikTok shop
  const tiktokShop = await prisma.shop.create({
    data: {
      userId: user.id,
      platform: "tiktok",
      shopName: "Test TikTok Store",
      shopId: "tiktok_345678",
      shopToken: "test_tiktok_token_xxx",
      isActive: true,
    },
  });
  console.log("✅ TikTok shop created:", tiktokShop.id);

  // Create sample Shopee products
  const shopeeProducts = [];
  for (let i = 1; i <= 5; i++) {
    const product = await prisma.product.create({
      data: {
        shopId: shopeeShop.id,
        itemId: `shopee_item_${i}`,
        itemName: `Shopee Product ${i}`,
        itemSku: `SHOP-SKU-${i}`,
        platform: "shopee",
        price: 50000 + i * 10000,
        stock: 100 - i * 10,
        status: "active",
        category: "Electronics",
        brand: "TestBrand",
        description: `This is a test Shopee product number ${i}`,
        imageUrls: JSON.stringify([
          `https://via.placeholder.com/300?text=Shopee+Product+${i}`,
        ]),
        variations: JSON.stringify([
          { id: 1, name: "Size", values: ["S", "M", "L"] },
          { id: 2, name: "Color", values: ["Black", "White", "Red"] },
        ]),
        isActive: true,
      },
    });
    shopeeProducts.push(product);
  }
  console.log(`✅ ${shopeeProducts.length} Shopee products created`);

  // Create sample Lazada products
  const lazadaProducts = [];
  for (let i = 1; i <= 5; i++) {
    const product = await prisma.product.create({
      data: {
        shopId: lazadaShop.id,
        itemId: `lazada_item_${i}`,
        itemName: `Lazada Product ${i}`,
        itemSku: `LAZADA-SKU-${i}`,
        platform: "lazada",
        price: 40000 + i * 8000,
        stock: 80 - i * 5,
        status: "active",
        category: "Fashion",
        brand: "TestBrand",
        description: `This is a test Lazada product number ${i}`,
        imageUrls: JSON.stringify([
          `https://via.placeholder.com/300?text=Lazada+Product+${i}`,
        ]),
        variations: JSON.stringify([
          { id: 1, name: "Size", values: ["XS", "S", "M", "L", "XL"] },
        ]),
        isActive: true,
      },
    });
    lazadaProducts.push(product);
  }
  console.log(`✅ ${lazadaProducts.length} Lazada products created`);

  // Create sample TikTok products
  const tiktokProducts = [];
  for (let i = 1; i <= 5; i++) {
    const product = await prisma.product.create({
      data: {
        shopId: tiktokShop.id,
        itemId: `tiktok_item_${i}`,
        itemName: `TikTok Product ${i}`,
        itemSku: `TIKTOK-SKU-${i}`,
        platform: "tiktok",
        price: 30000 + i * 5000,
        stock: 50 - i * 3,
        status: "active",
        category: "Accessories",
        brand: "TestBrand",
        description: `This is a test TikTok product number ${i}`,
        imageUrls: JSON.stringify([
          `https://via.placeholder.com/300?text=TikTok+Product+${i}`,
        ]),
        variations: JSON.stringify([
          { id: 1, name: "Color", values: ["Red", "Blue", "Green", "Yellow"] },
        ]),
        isActive: true,
      },
    });
    tiktokProducts.push(product);
  }
  console.log(`✅ ${tiktokProducts.length} TikTok products created`);

  // Create sample inventory
  for (const product of [
    ...shopeeProducts,
    ...lazadaProducts,
    ...tiktokProducts,
  ]) {
    await prisma.inventory.create({
      data: {
        userId: user.id,
        productId: product.id,
        quantity: product.stock,
        location: "Warehouse A",
      },
    });
  }
  console.log("✅ Inventory records created");

  // Create sample orders
  const order1 = await prisma.order.create({
    data: {
      userId: user.id,
      shopId: shopeeShop.id,
      orderId: "shopee_order_001",
      platform: "shopee",
      customerName: "John Doe",
      customerPhone: "08123456789",
      customerEmail: "john@example.com",
      shippingAddress: "123 Main St, City, Country",
      totalAmount: 150000,
      status: "unpaid",
      paymentMethod: "COD",
      notes: "Please pack carefully",
      isActive: true,
    },
  });
  console.log("✅ Order 1 created");

  // Add items to order
  await prisma.orderItem.create({
    data: {
      orderId: order1.id,
      productId: shopeeProducts[0].id,
      quantity: 2,
      price: shopeeProducts[0].price,
    },
  });

  // Create more sample orders for different statuses
  const order2 = await prisma.order.create({
    data: {
      userId: user.id,
      shopId: shopeeShop.id,
      orderId: "shopee_order_002",
      platform: "shopee",
      customerName: "Jane Smith",
      customerPhone: "08987654321",
      customerEmail: "jane@example.com",
      shippingAddress: "456 Oak Ave, Town, Country",
      totalAmount: 250000,
      status: "cancelled",
      paymentMethod: "Bank Transfer",
      notes: "",
      isActive: true,
    },
  });
  console.log("✅ Order 2 created (cancelled)");

  const order3 = await prisma.order.create({
    data: {
      userId: user.id,
      shopId: lazadaShop.id,
      orderId: "lazada_order_001",
      platform: "lazada",
      customerName: "Bob Johnson",
      customerPhone: "08111222333",
      customerEmail: "bob@example.com",
      shippingAddress: "789 Pine Rd, Village, Country",
      totalAmount: 180000,
      status: "unpaid",
      paymentMethod: "Credit Card",
      notes: "Urgent delivery",
      isActive: true,
    },
  });
  console.log("✅ Order 3 created (Lazada)");

  const order4 = await prisma.order.create({
    data: {
      userId: user.id,
      shopId: tiktokShop.id,
      orderId: "tiktok_order_001",
      platform: "tiktok",
      customerName: "Alice Brown",
      customerPhone: "08444555666",
      customerEmail: "alice@example.com",
      shippingAddress: "321 Elm St, District, Country",
      totalAmount: 120000,
      status: "unpaid",
      paymentMethod: "E-wallet",
      notes: "",
      isActive: true,
    },
  });
  console.log("✅ Order 4 created (TikTok)");

  console.log("\n🎉 Database seeding completed successfully!");
  console.log("\nSummary:");
  console.log(`- Users: 1`);
  console.log(`- Shops: 3 (Shopee, Lazada, TikTok)`);
  console.log(`- Products: 15 (5 per platform)`);
  console.log(`- Orders: 4`);
  console.log(`- Inventory records: 15`);
}

main()
  .then(async () => {
    await prisma.$disconnect();
  })
  .catch(async (e) => {
    console.error("❌ Seeding error:", e);
    await prisma.$disconnect();
    process.exit(1);
  });

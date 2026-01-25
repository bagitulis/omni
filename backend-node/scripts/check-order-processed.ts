import { getPrisma } from "./src/services/prismaClient";

async function checkOrderProcessed() {
  const prisma = getPrisma("yumna");

  const orderSn = "26010212PXASTK";

  console.log("\n=== Checking if Webhook was Processed ===\n");
  console.log(`Order SN: ${orderSn}\n`);

  // Check webhook logs
  console.log("1️⃣ Checking webhook_logs table...");
  const webhookLog = await prisma.webhookLog.findMany({
    where: {
      payload: { contains: orderSn },
    },
    orderBy: { receivedAt: "desc" },
    take: 5,
  });

  if (webhookLog.length > 0) {
    console.log(`   ✅ Found ${webhookLog.length} webhook log(s):`);
    webhookLog.forEach((log) => {
      console.log(`      - ID: ${log.id}`);
      console.log(`      - Platform: ${log.platform}`);
      console.log(`      - Status: ${log.status}`);
      console.log(`      - Received: ${log.receivedAt}`);
      console.log(`      - Processed: ${log.processedAt || "NOT PROCESSED"}`);
      console.log(`      - Error: ${log.errorMessage || "None"}`);
      console.log();
    });
  } else {
    console.log("   ❌ No webhook logs found - WEBHOOK NOT LOGGED!\n");
  }

  // Check orders table
  console.log("2️⃣ Checking orders table...");
  try {
    const order = await prisma.order.findFirst({
      where: { orderSn: orderSn },
    });

    if (order) {
      console.log(`   ✅ Order FOUND in database:`);
      console.log(`      - Order SN: ${order.orderSn}`);
      console.log(`      - Status: ${order.status}`);
      console.log(`      - Created: ${order.createdAt}`);
    } else {
      console.log(`   ❌ Order NOT found in database - NOT PROCESSED!\n`);
    }
  } catch (error) {
    console.log(`   ⚠️ Error checking orders: ${error}`);
  }

  console.log("\n=== Conclusion ===");
  console.log("If webhook log shows 'NOT PROCESSED' or order not in database,");
  console.log(
    "it means signature verification FAILED and webhook was REJECTED."
  );
  console.log(
    "Backend returned 200 OK to Shopee to avoid retries, but did NOT process the webhook.\n"
  );

  await prisma.$disconnect();
}

checkOrderProcessed();

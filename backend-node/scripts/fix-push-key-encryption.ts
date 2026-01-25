import { getPrisma } from "./src/services/prismaClient";

async function fixPushKeyEncryption() {
  const prisma = getPrisma("yumna");

  console.log("\n=== Fixing Push Partner Key Encryption Flag ===\n");

  const result = await prisma.platformConfig.updateMany({
    where: {
      platform: "shopee",
      configKey: "pushPartnerKey",
    },
    data: {
      isEncrypted: false, // Set to false since value is plain text
    },
  });

  console.log(`✅ Updated ${result.count} record(s)`);
  console.log("\n=== Verification ===\n");

  const updated = await prisma.platformConfig.findFirst({
    where: {
      platform: "shopee",
      configKey: "pushPartnerKey",
    },
  });

  if (updated) {
    console.log(`Config Key: ${updated.configKey}`);
    console.log(`Value: ${updated.configValue}`);
    console.log(`Is Encrypted: ${updated.isEncrypted}`);
    console.log(`\n✅ Fix applied successfully!`);
    console.log(`\n🔄 Restart backend to apply changes:`);
    console.log(`   docker-compose restart backend`);
  }

  await prisma.$disconnect();
}

fixPushKeyEncryption();

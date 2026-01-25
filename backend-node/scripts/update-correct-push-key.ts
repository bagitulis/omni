import { getPrisma } from "./src/services/prismaClient";

async function updateCorrectPushKey() {
  const prisma = getPrisma("yumna");

  // Key yang BENAR dari Shopee Console screenshot
  const CORRECT_KEY =
    "544c48794b62515666614e6246666d616b76447676674a54794677574d786d58";

  // Key yang SALAH di database saat ini
  const OLD_KEY =
    "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";

  console.log("\n=== Updating Push Partner Key ===\n");
  console.log(`Old Key: ${OLD_KEY}`);
  console.log(`New Key: ${CORRECT_KEY}`);
  console.log(`\nDifference at position 14-15:`);
  console.log(`  Old: ...4b6251566...`);
  console.log(`  New: ...4b62515666... ✅ CORRECT\n`);

  const result = await prisma.platformConfig.updateMany({
    where: {
      platform: "shopee",
      configKey: "pushPartnerKey",
    },
    data: {
      configValue: CORRECT_KEY,
    },
  });

  console.log(`✅ Updated ${result.count} record(s)\n`);

  // Verify
  const updated = await prisma.platformConfig.findFirst({
    where: {
      platform: "shopee",
      configKey: "pushPartnerKey",
    },
  });

  if (updated) {
    console.log("=== Verification ===");
    console.log(`Config Key: ${updated.configKey}`);
    console.log(`Value: ${updated.configValue}`);
    console.log(`Is Encrypted: ${updated.isEncrypted}`);
    console.log(`\n✅ Push Partner Key updated successfully!`);
    console.log(`\n🔄 Restart backend:`);
    console.log(`   docker-compose restart backend\n`);
  }

  await prisma.$disconnect();
}

updateCorrectPushKey();

import { getPrisma } from "./src/services/prismaClient";

async function revertToOriginalKey() {
  const prisma = getPrisma("yumna");

  // Key yang BENAR (original di database)
  const CORRECT_KEY =
    "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";

  console.log("\n=== Reverting to Original Push Partner Key ===\n");
  console.log(`Correct Key: ${CORRECT_KEY}\n`);

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
    console.log(`\n✅ Key reverted to original!`);
  }

  await prisma.$disconnect();
}

revertToOriginalKey();

import { getPrisma } from "./src/services/prismaClient";

async function checkPushKeyEncryption() {
  const prisma = getPrisma("yumna");

  const pushKey = await prisma.platformConfig.findFirst({
    where: {
      platform: "shopee",
      configKey: "pushPartnerKey",
    },
  });

  console.log("\n=== Push Partner Key Encryption Status ===\n");

  if (pushKey) {
    console.log(`Config Key: ${pushKey.configKey}`);
    console.log(`Value: ${pushKey.configValue}`);
    console.log(`Is Encrypted: ${pushKey.isEncrypted}`);
    console.log(`Created: ${pushKey.createdAt}`);
    console.log(`Updated: ${pushKey.updatedAt}`);

    console.log("\n=== Problem ===");
    if (pushKey.isEncrypted && !pushKey.configValue.startsWith("gAAAAA")) {
      console.log("❌ isEncrypted = true, but value is PLAIN TEXT!");
      console.log("   Fernet encrypted values start with 'gAAAAA'");
      console.log("\n=== Solution ===");
      console.log("Set isEncrypted = false for this key:");
      console.log(`
UPDATE platformConfig 
SET isEncrypted = false 
WHERE platform = 'shopee' 
AND configKey = 'pushPartnerKey'
`);
    } else if (pushKey.isEncrypted) {
      console.log("✅ Value appears to be encrypted (starts with gAAAAA)");
    } else {
      console.log("✅ isEncrypted = false, value is plain text (correct)");
    }
  } else {
    console.log("❌ pushPartnerKey NOT FOUND in database");
  }

  await prisma.$disconnect();
}

checkPushKeyEncryption();

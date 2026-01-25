import { getPrisma } from "./src/services/prismaClient";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("CheckYumnaConfig");

async function checkShopeeConfig() {
  try {
    const prisma = getPrisma("yumna");

    console.log("\n=== Checking Shopee Config for Tenant: yumna ===\n");

    const configs = await prisma.platformConfig.findMany({
      where: {
        platform: "shopee",
      },
    });

    console.log(`Found ${configs.length} Shopee configurations:\n`);

    configs.forEach((config) => {
      console.log(`Key: ${config.configKey}`);
      console.log(`Value: ${config.configValue}`);
      console.log(`Created: ${config.createdAt}`);
      console.log(`Updated: ${config.updatedAt}`);
      console.log("---");
    });

    // Check specifically for pushPartnerKey
    const pushKey = await prisma.platformConfig.findFirst({
      where: {
        platform: "shopee",
        configKey: "pushPartnerKey",
      },
    });

    console.log("\n=== Push Partner Key Status ===");
    if (pushKey) {
      console.log("✅ pushPartnerKey found:");
      console.log(`   Value: ${pushKey.configValue}`);
    } else {
      console.log("❌ pushPartnerKey NOT found");
    }

    // Check for legacy partnerKey
    const legacyKey = await prisma.platformConfig.findFirst({
      where: {
        platform: "shopee",
        configKey: "partnerKey",
      },
    });

    console.log("\n=== Legacy Partner Key Status ===");
    if (legacyKey) {
      console.log("✅ partnerKey found:");
      console.log(`   Value: ${legacyKey.configValue}`);
    } else {
      console.log("❌ partnerKey NOT found");
    }

    await prisma.$disconnect();
  } catch (error) {
    console.error("Error checking config:", error);
    process.exit(1);
  }
}

checkShopeeConfig();

#!/usr/bin/env node
import { getPrisma } from "../src/services/prismaClient";

async function main() {
  try {
    const prisma = getPrisma();

    // Update Lazada config dengan correct refresh_token dari plaintext database
    const updated = await prisma.platformConfig.update({
      where: {
        platform_configKey: {
          platform: "lazada",
          configKey: "refreshToken",
        },
      },
      data: {
        configValue:
          "500011000052EFdwU12f29e65OiDTvdy7hFpxtucC2GXElsArqCdieviNGlZsD9q", // Correct token dari plaintext DB
        isEncrypted: false,
      },
    });

    console.log(`✅ Updated refreshToken for lazada`);
    console.log(
      `   New value: 500011000052EFdwU12f29e65OiDTvdy7hFpxtucC2GXElsArqCdieviNGlZsD9q`
    );

    // Also verify appSecret is correct
    const appSecret = await prisma.platformConfig.findUnique({
      where: {
        platform_configKey: {
          platform: "lazada",
          configKey: "appSecret",
        },
      },
    });

    if (appSecret) {
      console.log(`\n📌 Current appSecret: ${appSecret.configValue}`);
    }

    // Get all Lazada config
    const all = await prisma.platformConfig.findMany({
      where: { platform: "lazada" },
    });

    console.log(`\n📋 Current Lazada Config:`);
    all.forEach((item) => {
      if (
        item.configKey === "appSecret" ||
        item.configKey === "refreshToken" ||
        item.configKey === "accessToken" ||
        item.configKey === "appKey"
      ) {
        const preview =
          item.configValue.length > 20
            ? item.configValue.substring(0, 20) + "..."
            : item.configValue;
        console.log(
          `  ${item.configKey}: ${preview} (encrypted: ${item.isEncrypted})`
        );
      }
    });
  } catch (error) {
    console.error("Error:", error);
    process.exit(1);
  }
}

main();

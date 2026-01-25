#!/usr/bin/env node
import { getPrisma } from "../src/services/prismaClient";
import sqlite3 from "sqlite3";

async function main() {
  try {
    const prisma = getPrisma();

    // Read TikTok credentials dari plaintext database
    const plaintextDb = new sqlite3.Database(
      `C:/Users/yumna/Desktop/omni/backend/config/ecommerce_config-before enskrip.db`
    );

    const tiktokData: Record<string, string> = {};

    await new Promise((resolve, reject) => {
      plaintextDb.all(
        "SELECT config_key, config_value FROM platform_config WHERE platform='tiktok'",
        (err, rows: any[]) => {
          if (err) reject(err);
          else {
            rows.forEach((row) => {
              tiktokData[row.config_key] = row.config_value;
            });
            resolve(null);
          }
        }
      );
    });

    plaintextDb.close();

    console.log("✅ Extracted TikTok credentials from plaintext database:");
    console.log(`   appSecret: ${tiktokData.app_secret}`);
    console.log(
      `   refreshToken: ${tiktokData.refresh_token.substring(0, 30)}...`
    );

    // Update production database with plaintext values
    // appSecret
    await prisma.platformConfig.update({
      where: {
        platform_configKey: {
          platform: "tiktok",
          configKey: "appSecret",
        },
      },
      data: {
        configValue: tiktokData.app_secret,
        isEncrypted: false,
      },
    });
    console.log(`\n✅ Updated appSecret (plaintext)`);

    // refreshToken
    await prisma.platformConfig.update({
      where: {
        platform_configKey: {
          platform: "tiktok",
          configKey: "refreshToken",
        },
      },
      data: {
        configValue: tiktokData.refresh_token,
        isEncrypted: false,
      },
    });
    console.log(`✅ Updated refreshToken (plaintext)`);

    // accessToken
    await prisma.platformConfig.update({
      where: {
        platform_configKey: {
          platform: "tiktok",
          configKey: "accessToken",
        },
      },
      data: {
        configValue: tiktokData.access_token,
        isEncrypted: false,
      },
    });
    console.log(`✅ Updated accessToken (plaintext)`);

    console.log("\n📋 Current TikTok Config (updated):");
    const all = await prisma.platformConfig.findMany({
      where: { platform: "tiktok" },
    });

    all.forEach((item) => {
      if (
        item.configKey === "appSecret" ||
        item.configKey === "refreshToken" ||
        item.configKey === "accessToken" ||
        item.configKey === "appKey"
      ) {
        const preview =
          item.configValue.length > 30
            ? item.configValue.substring(0, 30) + "..."
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

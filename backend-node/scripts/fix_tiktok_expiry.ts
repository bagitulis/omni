#!/usr/bin/env node

import { getPrisma } from "../src/services/prismaClient";

async function fixTiktokExpiry() {
  const prisma = getPrisma();

  // TikTok token should have been accessed_token_expire_in: 604800 (7 days)
  const expiresIn = 604800; // 7 days in seconds

  const expiryMs = Date.now() + expiresIn * 1000;

  console.log(`Fixing TikTok tokenExpiry...`);
  console.log(
    `  expiresIn: ${expiresIn}s (${(expiresIn / 86400).toFixed(1)}d)`
  );
  console.log(`  Current time: ${new Date().toISOString()}`);
  console.log(`  New expiry (ms): ${expiryMs}`);
  console.log(`  New expiry (date): ${new Date(expiryMs).toISOString()}`);

  try {
    const result = await prisma.platformConfig.update({
      where: {
        platform_configKey: {
          platform: "tiktok",
          configKey: "tokenExpiry",
        },
      },
      data: {
        configValue: String(expiryMs),
      },
    });

    console.log(`\n✅ TikTok tokenExpiry updated:`);
    console.log(
      `  ${result.configValue} (${new Date(parseInt(result.configValue)).toISOString()})`
    );
  } catch (error) {
    console.error(`❌ Error:`, error);
  }
}

fixTiktokExpiry().catch(console.error);

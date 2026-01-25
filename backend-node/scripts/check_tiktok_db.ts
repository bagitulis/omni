#!/usr/bin/env node

import { getPrisma } from "../src/services/prismaClient";

async function checkTiktokDb() {
  const prisma = getPrisma();

  const result = await prisma.platformConfig.findMany({
    where: { platform: "tiktok" },
    select: { configKey: true, configValue: true },
  });

  console.log("TikTok config in database:");
  result.forEach((r) => {
    if (r.configKey === "tokenExpiry") {
      const val = parseInt(r.configValue, 10);
      console.log(`  ${r.configKey}: ${r.configValue}`);
      console.log(`    Parsed: ${val}`);
      if (!isNaN(val)) {
        console.log(`    Date: ${new Date(val).toISOString()}`);
      }
    } else {
      console.log(`  ${r.configKey}: ${r.configValue.substring(0, 50)}`);
    }
  });
}

checkTiktokDb().catch(console.error);

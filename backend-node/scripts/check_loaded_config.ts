#!/usr/bin/env node

import { TiktokConfigManager } from "../src/config/managers/tiktokConfigManager";

async function checkLoaded() {
  const config = new TiktokConfigManager();
  await config.loadConfig();

  console.log("TikTok config loaded in memory:");
  console.log(
    `  accessToken: ${config.accessToken ? "✅ LOADED" : "❌ MISSING"}`
  );
  console.log(
    `  refreshToken: ${config.refreshToken ? "✅ LOADED" : "❌ MISSING"}`
  );
  console.log(`  tokenExpiry (ms): ${config.tokenExpiry}`);

  if (config.tokenExpiry) {
    console.log(`  Expiry date: ${new Date(config.tokenExpiry).toISOString()}`);
    const now = Date.now();
    const diff = config.tokenExpiry - now;
    const days = Math.floor(diff / (1000 * 86400));
    const hours = Math.floor((diff % (1000 * 86400)) / (1000 * 3600));
    console.log(`  Time remaining: ${days}d ${hours}h`);
  }
}

checkLoaded().catch(console.error);

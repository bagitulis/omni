/**
 * Check Users and Tenant Configuration
 */
const { PrismaClient } = require("@prisma/client");
const fs = require("fs");

async function main() {
  console.log("🔍 Checking Users and Tenant Configuration...\n");

  // Read tenants.json
  const tenantsConfig = JSON.parse(
    fs.readFileSync("/app/config/static/tenants.json", "utf-8")
  );

  console.log("📋 Tenants Configuration:");
  for (const [tenantId, config] of Object.entries(tenantsConfig)) {
    console.log(`   ${tenantId}: ${config.dbPath}`);
  }

  console.log("\n📊 Users in each tenant database:");

  for (const [tenantId, config] of Object.entries(tenantsConfig)) {
    try {
      const prisma = new PrismaClient({
        datasources: {
          db: { url: `file:/app/${config.dbPath}` },
        },
      });

      const users = await prisma.user.findMany({
        select: {
          id: true,
          username: true,
          role: true,
        },
      });

      console.log(`\n   [${tenantId}] - ${users.length} users:`);
      for (const user of users) {
        console.log(`      - ${user.username} (role: ${user.role})`);
      }

      await prisma.$disconnect();
    } catch (err) {
      console.log(`   [${tenantId}] - Error: ${err.message}`);
    }
  }
}

main();

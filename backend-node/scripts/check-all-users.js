const { PrismaClient } = require('@prisma/client');
const fs = require('fs');
const path = require('path');

async function main() {
  // Read tenants config
  const tenantsPath = path.join(__dirname, '../config/static/tenants.json');
  const tenants = JSON.parse(fs.readFileSync(tenantsPath, 'utf8'));
  
  console.log('=== Tenant Configuration ===');
  console.table(Object.entries(tenants).map(([id, config]) => ({
    tenantId: id,
    dbPath: config.dbPath,
    shopName: config.shopName
  })));
  
  // Track unique databases
  const dbPaths = new Set();
  const uniqueDbs = [];
  
  for (const [tenantId, config] of Object.entries(tenants)) {
    if (!dbPaths.has(config.dbPath)) {
      dbPaths.add(config.dbPath);
      uniqueDbs.push({ tenantId, dbPath: config.dbPath });
    }
  }
  
  console.log('\n=== Unique Databases ===');
  for (const db of uniqueDbs) {
    const fullPath = path.join(__dirname, '..', db.dbPath);
    console.log(`\nDatabase: ${db.dbPath}`);
    
    if (fs.existsSync(fullPath)) {
      const prisma = new PrismaClient({
        datasources: { db: { url: `file:${fullPath}` } }
      });
      
      try {
        const users = await prisma.user.findMany();
        console.log(`Users found: ${users.length}`);
        users.forEach(u => {
          console.log(`  - ${u.username} (${u.email}) [${u.role}]`);
        });
      } catch (e) {
        console.log(`  Error reading: ${e.message}`);
      } finally {
        await prisma.$disconnect();
      }
    } else {
      console.log('  ❌ Database file not found');
    }
  }
}

main().catch(console.error);

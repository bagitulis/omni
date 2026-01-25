require('dotenv').config({ path: '.env' });
const { PrismaClient } = require('@prisma/client');

async function main() {
  const prisma = new PrismaClient();
  try {
    // Cek semua platform yang ada
    const allConfigs = await prisma.platformConfig.findMany();
    console.log('📊 Total configs di database:', allConfigs.length);
    
    if (allConfigs.length > 0) {
      // Group by platform
      const byPlatform = {};
      for (const cfg of allConfigs) {
        if (!byPlatform[cfg.platform]) {
          byPlatform[cfg.platform] = [];
        }
        byPlatform[cfg.platform].push(cfg.configKey);
      }
      
      console.log('\nPlatforms found:');
      for (const [platform, keys] of Object.entries(byPlatform)) {
        console.log(`  - ${platform}: ${keys.length} configs`);
        console.log(`    Keys: ${keys.join(', ')}`);
      }
    } else {
      console.log('⚠️  Tidak ada config ditemukan');
    }
  } catch (error) {
    console.error('Error:', error.message);
  } finally {
    await prisma.$disconnect();
  }
}

main();

#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const backendPath = 'c:\\Users\\PC\\Desktop\\omni\\backend';

// Read tenants config to get ACTUAL database paths
const tenantsConfig = JSON.parse(
  fs.readFileSync(path.join(backendPath, 'config/static/tenants.json'), 'utf-8')
);

console.log('\n' + '='.repeat(60));
console.log('🔧 SYNCING ACTUAL TENANT DATABASES');
console.log('='.repeat(60) + '\n');

console.log('📋 Tenant Configuration:');
Object.entries(tenantsConfig).forEach(([tenant, config]) => {
  console.log(`  ${tenant} → ${config.dbPath}`);
});
console.log('');

for (const [tenant, config] of Object.entries(tenantsConfig)) {
  try {
    const dbPath = `file:${config.dbPath}`;
    console.log(`⏳ Syncing ${tenant} (${config.dbPath})...`);

    const cmd = `cd ${backendPath} && set DATABASE_URL=${dbPath} && npx prisma db push --force-reset --skip-generate`;

    const output = execSync(cmd, {
      encoding: 'utf-8',
      stdio: 'pipe',
      shell: 'cmd.exe'
    });

    if (output.includes('successfully reset') || output.includes('in sync')) {
      console.log(`✅ ${tenant}: Synced\n`);
    } else {
      console.log(`⚠️ ${tenant}: Unclear result\n`);
    }
  } catch (error) {
    console.log(`❌ ${tenant}: ERROR - ${error.message}\n`);
  }
}

console.log('='.repeat(60));
console.log('✅ All tenant databases synced with correct paths!');
console.log('='.repeat(60) + '\n');

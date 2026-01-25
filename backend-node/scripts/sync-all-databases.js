#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');

const tenants = ['yumna', 'tika', 'tester'];
const backendPath = 'c:\\Users\\PC\\Desktop\\omni\\backend';

console.log('\n' + '='.repeat(60));
console.log('🔧 SYNCING ALL TENANT DATABASES');
console.log('='.repeat(60) + '\n');

for (const tenant of tenants) {
  try {
    const dbPath = `file:config/databases/${tenant}_bertigamart.db`;
    console.log(`\n⏳ Syncing ${tenant}...`);

    const cmd = `cd ${backendPath} && set DATABASE_URL=${dbPath} && npx prisma db push --force-reset --skip-generate`;
    
    const output = execSync(cmd, { 
      encoding: 'utf-8',
      stdio: 'pipe',
      shell: 'cmd.exe'
    });

    if (output.includes('successfully reset') || output.includes('in sync')) {
      console.log(`✅ ${tenant}: Database synced`);
    } else {
      console.log(`⚠️ ${tenant}: Sync output unclear`);
      console.log(output);
    }
  } catch (error) {
    console.log(`❌ ${tenant}: ${error.message}`);
  }
}

console.log('\n' + '='.repeat(60));
console.log('✅ All databases synced!');
console.log('='.repeat(60) + '\n');

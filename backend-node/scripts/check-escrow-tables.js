const Database = require('better-sqlite3');

const tenants = [
  { name: 'yumna', db: './config/databases/yumna_bertigamart.db' },
  { name: 'tester', db: './config/databases/tester_developer.db' },
];

for (const tenant of tenants) {
  console.log(`\n=== Tenant: ${tenant.name} ===`);
  try {
    const db = new Database(tenant.db);
    
    // Check all tables
    const allTables = db.prepare("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").all();
    console.log('All tables:', allTables.map(t => t.name));
    
    // Check Escrow tables specifically
    const escrowTables = allTables.filter(t => t.name.includes('Escrow'));
    console.log('Escrow tables:', escrowTables.map(t => t.name));
    
    // Check if data exists
    for (const table of escrowTables) {
      try {
        const count = db.prepare(`SELECT COUNT(*) as cnt FROM ${table.name}`).get();
        console.log(`  ${table.name}: ${count.cnt} rows`);
      } catch (e) {
        console.log(`  ${table.name}: ERROR - ${e.message}`);
      }
    }
    
    db.close();
  } catch (e) {
    console.log(`Error: ${e.message}`);
  }
}

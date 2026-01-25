const Database = require('better-sqlite3');

const db = new Database('./config/databases/yumna_bertigamart.db', { readonly: true });

// List tables
const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all();
console.log('Tables in database:');
tables.forEach(t => console.log('  -', t.name));

// Check ShopeeEscrowSync
console.log('\n--- ShopeeEscrowSync ---');
try {
  const sync = db.prepare('SELECT * FROM ShopeeEscrowSync').all();
  console.log('Records:', sync.length);
  if (sync.length > 0) console.table(sync);
} catch(e) {
  console.log('Table not found or error:', e.message);
}

// Check ShopeeEscrowOrder  
console.log('\n--- ShopeeEscrowOrder ---');
try {
  const orders = db.prepare('SELECT COUNT(*) as cnt FROM ShopeeEscrowOrder').get();
  console.log('Total orders:', orders.cnt);
} catch(e) {
  console.log('Table not found or error:', e.message);
}

// Check ShopeeEscrowItem
console.log('\n--- ShopeeEscrowItem ---');
try {
  const items = db.prepare('SELECT COUNT(*) as cnt FROM ShopeeEscrowItem').get();
  console.log('Total items:', items.cnt);
} catch(e) {
  console.log('Table not found or error:', e.message);
}

// Check AnalyticsSettings
console.log('\n--- AnalyticsSettings ---');
try {
  const settings = db.prepare('SELECT * FROM AnalyticsSettings').all();
  console.log('Records:', settings.length);
  if (settings.length > 0) console.table(settings);
} catch(e) {
  console.log('Table not found or error:', e.message);
}

db.close();

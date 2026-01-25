const Database = require('better-sqlite3');

const db = new Database('./config/databases/yumna_bertigamart.db', { readonly: true });

// Check inventory for BAKIW6135
console.log('='.repeat(60));
console.log('INVENTORY LOOKUP DEBUG');
console.log('='.repeat(60));

const inv = db.prepare(`
  SELECT keyValue, keyColumnName, tenantId, data 
  FROM InventoryRecord 
  WHERE keyValue LIKE '%BAKIW6135%'
`).all();

console.log('\n1. Records with BAKIW6135 in keyValue:');
inv.forEach(i => {
  console.log('   keyValue:', i.keyValue);
  console.log('   keyColumnName:', i.keyColumnName);
  console.log('   tenantId:', i.tenantId);
  try {
    const d = JSON.parse(i.data);
    console.log('   HARGA:', d.HARGA);
    console.log('   SKU in data:', d.SKU);
  } catch(e) {}
  console.log();
});

// Check escrow items for this SKU
const escrow = db.prepare(`
  SELECT modelSku, sku, itemName, originalPrice, COUNT(*) as cnt
  FROM ShopeeEscrowItem 
  WHERE modelSku LIKE '%BAKIW6135%' OR sku LIKE '%BAKIW6135%'
  GROUP BY modelSku
  LIMIT 5
`).all();

console.log('2. Escrow items with BAKIW6135:');
escrow.forEach(e => {
  console.log('   modelSku:', e.modelSku);
  console.log('   sku:', e.sku);
  console.log('   originalPrice:', e.originalPrice);
  console.log('   count:', e.cnt);
  console.log();
});

// Check exact match lookup
const exactLookup = db.prepare(`
  SELECT keyValue, tenantId
  FROM InventoryRecord 
  WHERE keyValue = 'BAKIW6135' AND keyColumnName = 'SKU'
`).all();

console.log('3. Exact lookup (keyValue=BAKIW6135, keyColumnName=SKU):');
console.log('   Found:', exactLookup.length, 'records');
exactLookup.forEach(e => {
  console.log('   tenantId:', e.tenantId);
});

// Check all tenantIds in inventory
const tenants = db.prepare(`
  SELECT DISTINCT tenantId, COUNT(*) as cnt
  FROM InventoryRecord 
  GROUP BY tenantId
`).all();

console.log('\n4. TenantIds in InventoryRecord:');
tenants.forEach(t => {
  console.log('   ', t.tenantId, ':', t.cnt, 'records');
});

// Check if the escrow service uses correct tenantId
console.log('\n5. TenantIds in ShopeeEscrowItem:');
const escrowTenants = db.prepare(`
  SELECT DISTINCT tenantId, COUNT(*) as cnt
  FROM ShopeeEscrowItem 
  GROUP BY tenantId
`).all();
escrowTenants.forEach(t => {
  console.log('   ', t.tenantId, ':', t.cnt, 'records');
});

db.close();

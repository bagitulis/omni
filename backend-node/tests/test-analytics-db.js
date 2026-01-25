/**
 * Test Analytics Database Direct
 * Bypass API, test services directly
 */

const Database = require('better-sqlite3');
const path = require('path');

const DB_PATH = './config/databases/yumna_bertigamart.db';
const TENANT_ID = 'yumna';

async function main() {
  console.log('\n' + '='.repeat(60));
  console.log('ANALYTICS DATABASE DIRECT TEST');
  console.log('='.repeat(60));
  
  const db = new Database(DB_PATH, { readonly: true });
  
  // 1. Check Sync Status
  console.log('\n📊 1. SYNC STATUS');
  console.log('-'.repeat(40));
  try {
    const sync = db.prepare(`
      SELECT * FROM ShopeeEscrowSync 
      WHERE tenantId = ?
      ORDER BY year DESC, month DESC
      LIMIT 5
    `).all(TENANT_ID);
    
    if (sync.length === 0) {
      console.log('   ❌ No sync records found');
      console.log('   → Need to call POST /api/analytics/shopee/sync first');
    } else {
      console.log('   ✅ Sync records found:');
      sync.forEach(s => {
        console.log(`      ${s.year}-${String(s.month).padStart(2, '0')}: ${s.totalOrders} orders (synced: ${s.syncedAt})`);
      });
    }
  } catch(e) {
    console.log('   Error:', e.message);
  }
  
  // 2. Check Escrow Orders
  console.log('\n📦 2. ESCROW ORDERS');
  console.log('-'.repeat(40));
  try {
    const orders = db.prepare(`
      SELECT month, year, COUNT(*) as cnt, SUM(escrowAmount) as totalEscrow
      FROM ShopeeEscrowOrder 
      WHERE tenantId = ?
      GROUP BY month, year
      ORDER BY year DESC, month DESC
    `).all(TENANT_ID);
    
    if (orders.length === 0) {
      console.log('   ❌ No escrow orders found');
    } else {
      console.log('   ✅ Escrow orders by month:');
      orders.forEach(o => {
        console.log(`      ${o.year}-${String(o.month).padStart(2, '0')}: ${o.cnt} orders, Rp ${o.totalEscrow?.toLocaleString()}`);
      });
    }
  } catch(e) {
    console.log('   Error:', e.message);
  }
  
  // 3. Check Escrow Items
  console.log('\n📋 3. ESCROW ITEMS');
  console.log('-'.repeat(40));
  try {
    const items = db.prepare(`
      SELECT COUNT(*) as cnt, COUNT(DISTINCT modelSku) as uniqueSku
      FROM ShopeeEscrowItem 
      WHERE tenantId = ?
    `).get(TENANT_ID);
    
    if (items.cnt === 0) {
      console.log('   ❌ No escrow items found');
    } else {
      console.log(`   ✅ Total items: ${items.cnt}`);
      console.log(`   ✅ Unique SKUs: ${items.uniqueSku}`);
      
      // Sample items
      const sample = db.prepare(`
        SELECT modelSku, itemName, originalPrice, COUNT(*) as cnt
        FROM ShopeeEscrowItem 
        WHERE tenantId = ?
        GROUP BY modelSku
        ORDER BY cnt DESC
        LIMIT 5
      `).all(TENANT_ID);
      
      console.log('\n   Top 5 SKUs by transaction count:');
      sample.forEach(s => {
        console.log(`      ${s.modelSku}: ${s.cnt}x @ Rp ${s.originalPrice?.toLocaleString()} - ${s.itemName?.substring(0, 30)}...`);
      });
    }
  } catch(e) {
    console.log('   Error:', e.message);
  }
  
  // 4. Check Inventory (for reconciliation)
  console.log('\n📊 4. INVENTORY RECORDS');
  console.log('-'.repeat(40));
  try {
    const inv = db.prepare(`
      SELECT COUNT(*) as cnt
      FROM InventoryRecord 
      WHERE tenantId = ? AND keyColumnName = 'SKU'
    `).get(TENANT_ID);
    
    console.log(`   Total inventory records: ${inv.cnt}`);
    
    if (inv.cnt > 0) {
      const sample = db.prepare(`
        SELECT keyValue, data
        FROM InventoryRecord 
        WHERE tenantId = ? AND keyColumnName = 'SKU'
        LIMIT 3
      `).all(TENANT_ID);
      
      console.log('\n   Sample inventory:');
      sample.forEach(s => {
        try {
          const data = JSON.parse(s.data);
          console.log(`      ${s.keyValue}: HARGA = Rp ${data.HARGA?.toLocaleString()} - ${data['Nama Barang']?.substring(0, 30)}...`);
        } catch(e) {
          console.log(`      ${s.keyValue}: [parse error]`);
        }
      });
    }
  } catch(e) {
    console.log('   Error:', e.message);
  }
  
  // 5. Check Analytics Settings
  console.log('\n⚙️ 5. ANALYTICS SETTINGS');
  console.log('-'.repeat(40));
  try {
    const settings = db.prepare(`
      SELECT * FROM AnalyticsSettings 
      WHERE tenantId = ? AND platform = 'shopee'
    `).get(TENANT_ID);
    
    if (!settings) {
      console.log('   ❌ No settings found (will use defaults)');
      console.log('   Default: priceColumn=HARGA, deduction=1500, multiplier=0.84');
    } else {
      console.log(`   priceColumn: ${settings.priceColumn}`);
      console.log(`   formulaDeduction: ${settings.formulaDeduction}`);
      console.log(`   formulaMultiplier: ${settings.formulaMultiplier}`);
    }
  } catch(e) {
    console.log('   Error:', e.message);
  }
  
  db.close();
  
  // Summary
  console.log('\n' + '='.repeat(60));
  console.log('SUMMARY');
  console.log('='.repeat(60));
  console.log(`
Flow untuk Analytics/Shopee:

1. User buka http://localhost:4173/analytics/shopee
2. Pilih bulan (e.g., December 2025)
3. Klik "🔄 Sync Escrow Data"
   → POST /api/analytics/shopee/sync
   → Fetch wallet transactions
   → Fetch escrow details batch
   → Save to ShopeeEscrowOrder & ShopeeEscrowItem
4. Setelah sync selesai, klik "📈 Analyze Prices"
   → GET /api/analytics/shopee/reconciliation
   → Query escrow items + lookup inventory
   → Return summary + skuGroups
5. Frontend render hasil di table

Status saat ini: Database KOSONG
→ Perlu sync data dulu dari frontend/API
`);
}

main().catch(console.error);

const sqlite3 = require('sqlite3').verbose();
const path = require('path');

const dbPath = path.join(__dirname, 'config/databases/yumna_bertigamart.db');

const db = new sqlite3.Database(dbPath, (err) => {
  if (err) {
    console.error('Error opening database:', err);
    process.exit(1);
  }
  
  console.log('📊 Checking Shopee Order Discrepancy...\n');
  
  // Check ShopeeOrder count
  db.get(`SELECT COUNT(*) as count FROM ShopeeOrder`, (err, row) => {
    if (err) {
      console.error('Error:', err);
      db.close();
      process.exit(1);
    }
    
    console.log(`📦 Total ShopeeOrder records: ${row.count}`);
    
    // Check ShopeeOrderItem count
    db.get(`SELECT COUNT(*) as count FROM ShopeeOrderItem`, (err, itemRow) => {
      if (err) {
        console.error('Error:', err);
        db.close();
        process.exit(1);
      }
      
      console.log(`📋 Total ShopeeOrderItem records: ${itemRow.count}`);
      
      // Find orders WITHOUT items
      db.all(`
        SELECT so.orderSn, so.orderStatus
        FROM ShopeeOrder so
        LEFT JOIN ShopeeOrderItem soi ON so.orderSn = soi.orderSn
        WHERE soi.id IS NULL
        GROUP BY so.orderSn
      `, (err, rows) => {
        if (err) {
          console.error('Error:', err);
          db.close();
          process.exit(1);
        }
        
        console.log(`\n❌ Orders WITHOUT items (${rows.length}):`);
        rows.forEach(r => {
          console.log(`   - ${r.orderSn} (${r.orderStatus})`);
        });
        
        // Show all order SNs with item count
        console.log(`\n📊 Order breakdown:`);
        db.all(`
          SELECT 
            so.orderSn,
            so.orderStatus,
            COUNT(soi.id) as itemCount
          FROM ShopeeOrder so
          LEFT JOIN ShopeeOrderItem soi ON so.orderSn = soi.orderSn
          GROUP BY so.orderSn
          ORDER BY itemCount ASC, so.orderSn
        `, (err, breakdown) => {
          if (err) {
            console.error('Error:', err);
            db.close();
            process.exit(1);
          }
          
          breakdown.forEach(b => {
            const itemStr = b.itemCount === 0 ? '❌ NO ITEMS' : `${b.itemCount} item(s)`;
            console.log(`   ${b.orderSn}: ${itemStr} (${b.orderStatus})`);
          });
          
          console.log(`\n✅ Analysis complete`);
          db.close();
        });
      });
    });
  });
});

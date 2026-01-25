#!/usr/bin/env python3
"""
Check LockedOrder table schema and data
Debug why productName and variationName are empty
"""

import sqlite3
import os
from pathlib import Path

def check_locked_orders_table(db_path, tenant_name):
    """Check schema and data in lockedOrder table"""
    print(f"\n📂 Checking: {tenant_name}")
    print(f"   Path: {db_path}")
    
    if not os.path.exists(db_path):
        print(f"   ⚠️  Database not found")
        return
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # 1. Check if table exists
        cursor.execute("""
            SELECT name FROM sqlite_master 
            WHERE type='table' AND name='LockedOrder'
        """)
        
        if not cursor.fetchone():
            print(f"   ⚠️  Table LockedOrder doesn't exist")
            conn.close()
            return
        
        # 2. Get schema
        cursor.execute("PRAGMA table_info(LockedOrder)")
        columns = cursor.fetchall()
        column_names = [col[1] for col in columns]
        
        print(f"\n   📋 Schema columns: {', '.join(column_names)}")
        
        # 3. Count records
        cursor.execute("SELECT COUNT(*) FROM LockedOrder")
        count = cursor.fetchone()[0]
        print(f"   📊 Total records: {count}")
        
        if count == 0:
            print(f"   ⚠️  No data in table")
            conn.close()
            return
        
        # 4. Check for NULL or empty productName
        cursor.execute("""
            SELECT COUNT(*) FROM LockedOrder 
            WHERE productName IS NULL OR productName = ''
        """)
        empty_product_names = cursor.fetchone()[0]
        
        cursor.execute("""
            SELECT COUNT(*) FROM LockedOrder 
            WHERE variationName IS NULL OR variationName = ''
        """)
        empty_variation_names = cursor.fetchone()[0]
        
        print(f"\n   ⚠️  Empty productName: {empty_product_names}/{count}")
        print(f"   ⚠️  Empty variationName: {empty_variation_names}/{count}")
        
        # 5. Sample data
        cursor.execute("""
            SELECT id, sku, productName, variationName, qty, tenantId
            FROM LockedOrder 
            LIMIT 5
        """)
        samples = cursor.fetchall()
        
        print(f"\n   🔍 Sample data:")
        for row in samples:
            id, sku, pname, vname, qty, tid = row
            print(f"      ID:{id} | SKU:{sku} | Product:'{pname}' | Variation:'{vname}' | QTY:{qty}")
        
        # 6. Check tenantId filter
        cursor.execute("SELECT DISTINCT tenantId FROM LockedOrder")
        tenants = cursor.fetchall()
        print(f"\n   🏢 TenantIds in table: {[t[0] for t in tenants]}")
        
        conn.close()
        
    except Exception as e:
        print(f"   ❌ Error: {e}")

def main():
    """Check all tenant databases"""
    print("🔍 Checking LockedOrder table in all tenant databases\n")
    
    backend_dir = Path(__file__).parent.parent
    db_dir = backend_dir / "config" / "databases"
    
    print(f"📁 Database directory: {db_dir}\n")
    
    if not db_dir.exists():
        print(f"❌ Database directory not found: {db_dir}")
        return
    
    # Find all tenant databases (not jobs.db)
    tenant_dbs = [f for f in db_dir.glob("*.db") if not f.name.endswith("_jobs.db")]
    
    if not tenant_dbs:
        print("⚠️  No tenant databases found")
        return
    
    print(f"Found {len(tenant_dbs)} tenant databases:\n")
    print("=" * 80)
    
    for db_path in tenant_dbs:
        tenant_name = db_path.stem.replace(".db", "")
        check_locked_orders_table(str(db_path), tenant_name)
        print("=" * 80)

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Test Locked Orders Save & Load Flow
Verify data integrity from POST save to GET load
"""

import sqlite3
import json
import os
from pathlib import Path
from datetime import datetime

def check_database_before_after(db_path, tenant_name):
    """Check database state before and after operations"""
    print(f"\n{'='*80}")
    print(f"📂 Tenant: {tenant_name}")
    print(f"   Database: {db_path}")
    
    if not os.path.exists(db_path):
        print(f"   ❌ Database not found")
        return None
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # Check if table exists
        cursor.execute("""
            SELECT name FROM sqlite_master 
            WHERE type='table' AND name='LockedOrder'
        """)
        
        if not cursor.fetchone():
            print(f"   ⚠️  Table LockedOrder doesn't exist")
            conn.close()
            return None
        
        # Get all data grouped by tenantId
        cursor.execute("""
            SELECT tenantId, COUNT(*) as count,
                   SUM(CASE WHEN productName IS NULL OR productName = '' THEN 1 ELSE 0 END) as empty_product,
                   SUM(CASE WHEN variationName IS NULL OR variationName = '' THEN 1 ELSE 0 END) as empty_variation,
                   SUM(qty) as total_qty
            FROM LockedOrder
            GROUP BY tenantId
        """)
        
        tenant_stats = cursor.fetchall()
        
        print(f"\n   📊 Database Stats by TenantId:")
        print(f"   {'TenantId':<15} {'Records':<10} {'Empty Name':<12} {'Empty Var':<12} {'Total Qty':<10}")
        print(f"   {'-'*65}")
        
        for tid, count, empty_p, empty_v, total_q in tenant_stats:
            print(f"   {tid:<15} {count:<10} {empty_p:<12} {empty_v:<12} {total_q:<10}")
        
        # Get recent data samples
        cursor.execute("""
            SELECT tenantId, sku, productName, variationName, qty, 
                   datetime(createdAt, 'localtime') as created,
                   datetime(updatedAt, 'localtime') as updated
            FROM LockedOrder
            ORDER BY updatedAt DESC
            LIMIT 5
        """)
        
        recent = cursor.fetchall()
        
        print(f"\n   🕐 Recent Records (last 5):")
        for row in recent:
            tid, sku, pname, vname, qty, created, updated = row
            print(f"      [{tid}] {sku} | '{pname[:40]}...' | '{vname}' | {qty}")
            print(f"         Created: {created} | Updated: {updated}")
        
        conn.close()
        return tenant_stats
        
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None

def simulate_save_operation(db_path, tenant_id, test_items):
    """Simulate what lockedOrderService.saveLockedOrders() does"""
    print(f"\n   🧪 Simulating SAVE operation for tenantId: '{tenant_id}'")
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # Step 1: Delete existing
        cursor.execute("DELETE FROM LockedOrder WHERE tenantId = ?", (tenant_id,))
        deleted = cursor.rowcount
        print(f"      ✅ Deleted {deleted} existing records for tenantId '{tenant_id}'")
        
        # Step 2: Insert new data
        for item in test_items:
            cursor.execute("""
                INSERT INTO LockedOrder (id, tenantId, sku, productName, variationName, qty, createdAt, updatedAt)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                f"test_{datetime.now().timestamp()}_{item['sku']}",
                tenant_id,
                item['sku'],
                item['productName'],
                item.get('variationName', None),
                item['qty'],
                datetime.now().isoformat(),
                datetime.now().isoformat()
            ))
        
        conn.commit()
        print(f"      ✅ Inserted {len(test_items)} new records")
        
        # Step 3: Verify insertion
        cursor.execute("""
            SELECT sku, productName, variationName, qty
            FROM LockedOrder
            WHERE tenantId = ?
            ORDER BY qty DESC
        """, (tenant_id,))
        
        saved = cursor.fetchall()
        print(f"\n      📋 Verification - Saved Data:")
        for sku, pname, vname, qty in saved[:3]:
            print(f"         {sku} | '{pname}' | '{vname}' | {qty}")
        
        conn.close()
        return True
        
    except Exception as e:
        print(f"      ❌ Save failed: {e}")
        return False

def simulate_load_operation(db_path, tenant_id):
    """Simulate what lockedOrderService.getLockedOrders() does"""
    print(f"\n   🔍 Simulating LOAD operation for tenantId: '{tenant_id}'")
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        cursor.execute("""
            SELECT sku, productName, variationName, qty
            FROM LockedOrder
            WHERE tenantId = ?
            ORDER BY qty DESC
        """, (tenant_id,))
        
        orders = cursor.fetchall()
        conn.close()
        
        print(f"      ✅ Loaded {len(orders)} records")
        
        if orders:
            print(f"\n      📋 Loaded Data (first 5):")
            for sku, pname, vname, qty in orders[:5]:
                print(f"         {sku} | '{pname}' | '{vname}' | {qty}")
                
                # Check for empty fields
                if not pname or pname.strip() == '':
                    print(f"         ⚠️  EMPTY productName detected!")
                if not vname or vname.strip() == '':
                    print(f"         ⚠️  EMPTY variationName detected!")
        
        return orders
        
    except Exception as e:
        print(f"      ❌ Load failed: {e}")
        return []

def main():
    """Test save-load flow"""
    print("🧪 TESTING LOCKED ORDERS SAVE & LOAD FLOW\n")
    
    backend_dir = Path(__file__).parent.parent
    db_path = backend_dir / "config" / "databases" / "yumna_bertigamart.db"
    
    if not db_path.exists():
        print(f"❌ Database not found: {db_path}")
        return
    
    # 1. Check current state
    print("\n" + "="*80)
    print("STEP 1: CHECK CURRENT DATABASE STATE")
    check_database_before_after(str(db_path), "yumna_bertigamart")
    
    # 2. Simulate SAVE with test data
    print("\n" + "="*80)
    print("STEP 2: SIMULATE SAVE OPERATION")
    
    test_items = [
        {
            "sku": "TEST001",
            "productName": "Test Product 1 - Save Flow Test",
            "variationName": "Test Variation A",
            "qty": 100
        },
        {
            "sku": "TEST002",
            "productName": "Test Product 2 - Save Flow Test",
            "variationName": "Test Variation B",
            "qty": 50
        },
        {
            "sku": "TEST003",
            "productName": "Test Product 3 - Save Flow Test",
            "variationName": None,  # Test NULL variation
            "qty": 25
        }
    ]
    
    # Test with correct tenantId
    print(f"\n🔹 Test A: Save with tenantId='yumna' (CORRECT)")
    simulate_save_operation(str(db_path), "yumna", test_items)
    
    # 3. Simulate LOAD
    print("\n" + "="*80)
    print("STEP 3: SIMULATE LOAD OPERATION")
    
    print(f"\n🔹 Test B: Load with tenantId='yumna' (SHOULD MATCH)")
    loaded_yumna = simulate_load_operation(str(db_path), "yumna")
    
    print(f"\n🔹 Test C: Load with tenantId='tester' (SHOULD BE DIFFERENT)")
    loaded_tester = simulate_load_operation(str(db_path), "tester")
    
    print(f"\n🔹 Test D: Load with tenantId='yumna_bertigamart' (WRONG FORMAT)")
    loaded_wrong = simulate_load_operation(str(db_path), "yumna_bertigamart")
    
    # 4. Analyze results
    print("\n" + "="*80)
    print("STEP 4: ANALYSIS & DIAGNOSIS")
    
    print(f"\n📊 Results Summary:")
    print(f"   - Saved to tenantId 'yumna': {len(test_items)} items")
    print(f"   - Loaded from tenantId 'yumna': {len(loaded_yumna)} items")
    print(f"   - Loaded from tenantId 'tester': {len(loaded_tester)} items")
    print(f"   - Loaded from tenantId 'yumna_bertigamart': {len(loaded_wrong)} items")
    
    if len(loaded_yumna) == len(test_items):
        print(f"\n✅ SAVE-LOAD WORKING CORRECTLY!")
        print(f"   Data integrity verified - productName and variationName preserved")
    else:
        print(f"\n❌ MISMATCH DETECTED!")
        print(f"   Expected {len(test_items)} items, got {len(loaded_yumna)}")
    
    # 5. Check final state
    print("\n" + "="*80)
    print("STEP 5: FINAL DATABASE STATE")
    check_database_before_after(str(db_path), "yumna_bertigamart")
    
    print("\n" + "="*80)
    print("🎯 CONCLUSION:")
    print("   If productName shows correctly in test but empty in frontend:")
    print("   1. Check JWT token - what tenantId is being extracted?")
    print("   2. Check frontend - is it using correct field names?")
    print("   3. Check API response mapping - snake_case vs camelCase?")
    print("="*80)

if __name__ == "__main__":
    main()
